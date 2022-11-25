package auth

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"ncp/backend/api"
	"ncp/backend/api/mongo"
	"ncp/backend/api/session"
	"ncp/backend/env"
	"ncp/backend/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
)

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
			user := &api.User{}
			err := mongo.AuthCollection().FindOne(
				context.Background(),
				bson.D{
					{"email_addresses.addr", email},
				},
			).Decode(&user)
			if err == mongo.ErrNoDocuments {
				fmt.Fprint(w, `{"status":"error", "message":"user not found"}`)
				return
			}

			if err != nil {
				log.Panic(err)
			}

			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err == nil {
				session.SetCookie(w)
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

			session.SetCookie(w)
			fmt.Fprintf(w, `{"status":"success", "user": %s}`, user.JSON())
		}
	}
}
