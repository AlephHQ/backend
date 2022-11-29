package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"ncp/backend/api"
	"ncp/backend/api/mongo"
	"ncp/backend/env"
	"ncp/backend/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
)

var errUserNotFound = errors.New("user not found")

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func insertNewUserInMailAuthTable(u *api.User) error {
	db, err := sql.Open("mysql", env.MySQLURI())
	if err != nil {
		return err
	}

	defer db.Close()

	hash, _ := bcrypt.GenerateFromPassword([]byte(u.InternalPassword), bcrypt.MinCost)
	sql := fmt.Sprintf("INSERT INTO users (username, password) VALUES ('%s', '%s')", u.Username, hash)
	_, err = db.Exec(sql)
	return err
}

func signup(email, password string) (*api.User, error) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)

	user := &api.User{
		EmailAddresses: []api.EmailAddress{
			{Addr: email, Primary: true},
		},
		Password:         string(hash),
		InternalPassword: utils.RandStr(12),
		Username:         utils.RandStr(10),
		CreatedAt:        time.Now(),
	}

	result, err := mongo.AuthCollection().InsertOne(
		context.Background(),
		user,
	)

	if err != nil {
		return nil, err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)

	err = insertNewUserInMailAuthTable(user)
	if err != nil {
		_, derr := mongo.AuthCollection().DeleteOne(context.Background(), bson.D{{"_id", user.ID}})
		if derr != nil {
			return nil, derr
		}

		return nil, err
	}

	return user, nil
}

func signin(email, password string) (u *api.User, err error) {
	err = mongo.AuthCollection().FindOne(
		context.Background(),
		bson.D{
			{"email_addresses.addr", email},
		},
	).Decode(&u)
	if err == mongo.ErrNoDocuments {
		return
	}

	if err != nil {
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodPost:
		r.ParseMultipartForm(0)

		email := r.PostFormValue("email")
		password := r.PostFormValue("password")
		action := r.PostFormValue("action")

		switch action {
		case "signin":
			if user, err := signin(email, password); err == nil {
				api.SetAuthCookies(w, user)
				fmt.Fprintf(w, `{"status":"success", "user": %s}`, user.JSON())
				return
			}

			fmt.Fprint(w, `{"status":"error", "message":"authentication failed"}`)
		case "signup":
			user, err := signup(email, password)
			if err != nil {
				fmt.Fprintf(w, `{"status":"error", "message": "%s"}`, err.Error())
				return
			}

			fmt.Fprintf(w, `{"status":"success", "user": %s}`, user.JSON())
		case "":
			api.Error(w, "missing action param", http.StatusBadRequest)
		default:
			api.Error(
				w,
				fmt.Sprintf("action %s not implemented", action),
				http.StatusNotImplemented,
			)
		}
	}
}
