package post

import (
	"context"
	"encoding/json"
	"fmt"
	"ncp/backend/api"
	"ncp/backend/api/mongo"
	"ncp/backend/imap"
	"ncp/backend/imap/sessions"
	"net/http"
	"strconv"

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
		r.ParseForm()
		userID := r.FormValue("user_id")
		if userID == "" {
			api.Error(w, "missing user_id param", http.StatusBadRequest)
			return
		}

		uoid, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			api.Error(w, "invalid user_id param", http.StatusBadRequest)
			return
		}

		user := &api.User{}
		err = mongo.AuthCollection().FindOne(
			context.Background(),
			bson.D{{"_id", uoid}},
		).Decode(&user)
		if err == mongo.ErrNoDocuments {
			api.Error(w, "user not found", http.StatusNotFound)
			return
		}

		params := r.Context().Value(api.ContextKeyNameParams)
		if m, ok := params.(map[string]string); ok {
			err = mongo.AuthCollection().FindOne(
				context.Background(),
				bson.D{{"_id", uoid}},
			).Decode(&user)
			if err != nil {
				api.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			imapClient, err := sessions.Session(&sessions.Params{
				Username: user.Username,
				Password: user.InternalPassword,
			})
			if err != nil {
				api.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			seqnum, err := strconv.ParseUint(m["seqnum"], 10, 64)
			if err != nil {
				api.Error(w, "invalid seqnum", http.StatusBadRequest)
				return
			}

			if seqnum > imapClient.Mailbox().Exists {
				api.Error(w, "invalid seqnum", http.StatusBadRequest)
				return
			}

			messages, err := imapClient.Fetch(
				[]imap.SeqSet{
					&imap.SeqNumber{seqnum},
				},
				[]*imap.DataItem{
					{
						Name:    imap.DataItemNameBody,
						Section: imap.BodySectionText,
					},
					{
						Name: imap.DataItemNameEnvelope,
					},
					{
						Name: imap.DataItemNameInternalDate,
					},
					{
						Name: imap.DataItemNameFlags,
					},
					{
						Name: imap.DataItemNameRFC822Size,
					},
				},
				"",
			)

			b, _ := json.Marshal(api.MessageToPost(messages[0]))
			fmt.Fprintf(w, `{"status":"success", "post": %s}`, string(b))
			return
		}

		api.Error(w, "bad context values", http.StatusBadRequest)
	}
}
