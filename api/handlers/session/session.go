package session

import (
	"aleph/backend/api"
	"aleph/backend/api/mongo"
	"context"
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct{}

func NewHandler() *Handler {
	return new(Handler)
}

func (Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		userID := r.Context().Value(api.ContextKeyNameUserID)

		uoid, err := primitive.ObjectIDFromHex(userID.(string))
		if err != nil {
			api.Error(w, "invalid user_id param", http.StatusBadRequest)
			return
		}

		user := &api.User{}
		err = mongo.AuthCollection().FindOne(
			context.Background(),
			bson.D{
				{"_id", uoid},
			},
		).Decode(&user)
		if err == mongo.ErrNoDocuments {
			api.Error(w, "user not found", http.StatusNotFound)
			return
		}

		if err != nil {
			api.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, `{"status":"success", "user": %s}`, user.JSON())
	default:
		api.Error(w, "method not alowed", http.StatusMethodNotAllowed)
		return
	}
}
