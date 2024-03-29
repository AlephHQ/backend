package posts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"aleph/backend/api"
	"aleph/backend/api/mongo"
	"aleph/backend/imap"
	"aleph/backend/imap/sessions"

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
		userID := r.Context().Value(api.ContextKeyNameUserID)
		uoid, err := primitive.ObjectIDFromHex(userID.(string))
		if err != nil {
			api.Error(w, err.Error(), http.StatusBadRequest)
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

		seqset := []imap.SeqSet{}
		if imapClient.Mailbox().Exists > 0 {
			if imapClient.Mailbox().Exists == 1 {
				seqset = append(seqset, &imap.SeqNumber{1})
			} else {
				to := imapClient.Mailbox().Exists
				from := to - 9
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
					Name: imap.DataItemNameInternalDate,
				},
				{
					Name: imap.DataItemNameEnvelope,
				},
				{
					Name: imap.DataItemNameBodyStructure,
				},
				{
					Name:    imap.DataItemNameBodyPeek, // BODY.PEEK doesn't set the \Seen flag
					Section: imap.BodySection("1"),
					Partial: "0.512", // for preview purposes
				},
				{
					Name: imap.DataItemNameFlags,
				},
			},
			"",
		)
		if err != nil {
			api.Error(w, err.Error(), http.StatusInternalServerError)
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
