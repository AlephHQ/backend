package post

import (
	"context"
	"fmt"
	"io"
	"mime/quotedprintable"
	"ncp/backend/api"
	"ncp/backend/api/mongo"
	"ncp/backend/imap"
	"ncp/backend/imap/sessions"
	"net/http"
	"strconv"
	"strings"

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

			// fetching messages is a two step process.
			// step 1: fetch (BODY). This will send the MIME multipart
			// 				 body structure.
			// step 2: use information in step 1 to fetch content in one
			// 				 of the data formats available. Usually this will be one
			//				 of text/plain, text/html.
			//
			// Worth noting that right now we prioritize html. Send text/html if
			// it exists, otherwise send text/plain.
			messages, err := imapClient.Fetch(
				[]imap.SeqSet{
					&imap.SeqNumber{seqnum},
				},
				[]*imap.DataItem{
					{
						Name: imap.DataItemNameBody,
					},
				},
				"",
			)
			if err != nil {
				api.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			msg := messages[0] // we know there's exactly one message in the result
			fetchPart := "1"
			if msg.Body != nil {
				for i, part := range msg.Body.Parts {
					if part.Type == "text" && part.Subtype == "html" {
						fetchPart = strconv.Itoa(i + 1)
					}
				}
			}

			messages, err = imapClient.Fetch(
				[]imap.SeqSet{
					&imap.SeqNumber{seqnum},
				},
				[]*imap.DataItem{
					{
						Name:    imap.DataItemNameBody,
						Section: imap.BodySection(fetchPart),
					},
					{
						Name: imap.DataItemNameBody,
					},
					{
						Name: imap.DataItemNameEnvelope,
					},
					{
						Name: imap.DataItemNameFlags,
					},
					{
						Name: imap.DataItemNameInternalDate,
					},
				},
				"",
			)
			if err != nil {
				api.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			msg = messages[0]
			content := msg.Body.Sections[fetchPart]
			switch imap.Encoding(msg.Body.Parts[0].Encoding) {
			case imap.EncodingQuotePrintable:
				b, err := io.ReadAll(quotedprintable.NewReader(strings.NewReader(content)))
				if err != nil {
					api.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				msg.Body.Sections[fetchPart] = string(b)
			}

			// b, _ := json.Marshal(api.MessageToPost(msg))
			// fmt.Fprintf(w, `{"status":"success", "post": %s}`, string(b))
			post := api.MessageToPost(msg)
			w.Header().Add("Content-Type", post.Body.Type+"/"+post.Body.Subtype+"; charset="+post.Body.Params["charset"])
			fmt.Fprint(w, post.Body.Content)
			return
		}

		api.Error(w, "bad context values", http.StatusBadRequest)
	}
}
