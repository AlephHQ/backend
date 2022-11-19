package auth

import (
	"context"
	"fmt"
	"log"
	"ncp/backend/api/mongo"
	"ncp/backend/utils"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct{}

func NewHandlerAuth() *AuthHandler {
	return &AuthHandler{}
}

func signup(email, password string) (*User, error) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)

	user := &User{
		EmailAddresses: []EmailAddress{
			{Addr: email, Primary: true},
		},
		Password:         string(hash),
		InternalPassword: utils.RandStr(12),
		Username:         utils.RandStr(10),
	}

	result, err := mongo.AuthCollection().InsertOne(
		context.Background(),
		user,
	)

	if err != nil {
		return nil, err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)

	return user, nil
}

func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodPost:
		r.ParseMultipartForm(0)

		email := r.PostFormValue("email")
		password := r.PostFormValue("password")
		action := r.PostFormValue("action")

		switch action {
		case "signin":

			user := &User{}
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
		}
	}
}
