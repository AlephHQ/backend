package posts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"ncp/backend/api"
	"ncp/backend/api/mongo"
	"ncp/backend/imap"
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

		uoid, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user := &api.User{}
		err = mongo.AuthCollection().FindOne(
			context.Background(),
			bson.D{{"_id", uoid}},
		).Decode(&user)
		if err == mongo.ErrNoDocuments {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

		seqset := []imap.SeqSet{}
		if imapClient.Mailbox().Exists > 0 {
			if imapClient.Mailbox().Exists == 1 {
				seqset = append(seqset, &imap.SeqNumber{1})
			} else {
				to := imapClient.Mailbox().Exists
				from := to - 5
				if to < 10 {
					from = 1
				}

				seqset = append(seqset, &imap.SeqRange{From: from, To: to})
			}
		}

		msgs, err := imapClient.Fetch(
			seqset,
			[]*imap.DataItem{
				{
					Name: imap.DataItemNameRFC822,
				},
				{
					Name: imap.DataItemNameInternalDate,
				},
				{
					Name: imap.DataItemNameEnvelope,
				},
				{
					Name: imap.DataItemNamePreview,
				},
				{
					Name: imap.DataItemNameFlags,
				},
			},
			"",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		posts := make([]*api.Post, 0)
		for _, msg := range msgs {
			posts = append(posts, api.MessageToPost(msg))
		}

		b, _ := json.Marshal(posts)
		fmt.Fprintf(w, `{"status":"success", "posts": %s}`, string(b))
	}
}
