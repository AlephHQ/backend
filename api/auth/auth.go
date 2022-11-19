package auth

import (
	"context"
	"fmt"
	"log"
	"ncp/backend/api/mongo"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

type AuthHandler struct{}

func NewHandlerAuth() *AuthHandler {
	return &AuthHandler{}
}

func signup(email, password string) error {
	return nil
}

func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodPost:
		r.ParseMultipartForm(0)

		email := r.PostFormValue("email")
		// password := r.PostFormValue("password")
		action := r.PostFormValue("action")

		switch action {
		case "signin":
			var result bson.M

			err := mongo.AuthCollection().FindOne(
				context.Background(),
				bson.D{
					{"email", email},
				},
			).Decode(&result)
			if err == mongo.ErrNoDocuments {
				fmt.Fprint(w, `{"status":"error", "message":"user not found"}`)
				return
			}

			if err != nil {
				log.Panic(err)
			}

			fmt.Fprint(w, `{"status":"success"}`)
		case "signup":

		}
	}
}
