package inbox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"ncp/backend/api"
	"ncp/backend/api/mongo"
	"ncp/backend/imap/sessions"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		r.ParseForm()
		userID := r.FormValue("user_id")
		if userID == "" {
			http.Error(w, "missing user_id param", http.StatusBadRequest)
			return
		}

		user := &api.User{}
		uoid, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			http.Error(w, "invalid user_id param", http.StatusBadRequest)
			return
		}

		err = mongo.AuthCollection().FindOne(
			context.Background(),
			bson.D{{"_id", uoid}},
		).Decode(&user)
		if err == mongo.ErrNoDocuments {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		imapClient, err := sessions.Session(&sessions.Params{
			Username: user.Username,
			Password: user.InternalPassword,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b, _ := json.Marshal(imapClient.Mailbox())
		fmt.Fprintf(w, `{"status":"success", "inbox": %s}`, string(b))
	}
}
