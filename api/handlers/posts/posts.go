package posts

import (
	"context"
	"encoding/json"
	"fmt"
	"ncp/backend/api"
	"ncp/backend/api/mongo"
	"ncp/backend/imap"
	"ncp/backend/imap/client"
	"net/http"

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

		imapClient, err := client.DialWithTLS("tcp", "modsoussi.com:993")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer imapClient.Logout()

		err = imapClient.Login(user.Username+"@modsoussi.com", user.InternalPassword)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = imapClient.Select("INBOX")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		seqset := []imap.SeqSet{}
		if imapClient.Mailbox().Exists > 0 {
			if imapClient.Mailbox().Exists == 1 {
				seqset = append(seqset, &imap.SeqNumber{1})
			} else {
				var from, to uint64
				to = imapClient.Mailbox().Exists
				if to < 6 {
					from = 0
				}

				seqset = append(seqset, &imap.SeqRange{From: from, To: to})
			}
		}

		msgs, err := imapClient.Fetch(
			seqset,
			[]*imap.DataItem{
				{
					Name: imap.DataItemNameEnvelope,
				},
				{
					Name:    imap.DataItemNameBody,
					Section: imap.BodySectionText,
				},
			},
			"",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b, _ := json.Marshal(msgs)
		fmt.Fprintf(w, `{"status":"success", "posts": %s}`, string(b))
	}
}
