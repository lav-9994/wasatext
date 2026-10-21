package api

import (
	"errors"
	"net/http"

	"github.com/lav-9994/wasatext/service/api/reqcontext"
	"github.com/lav-9994/wasatext/service/database"
)

type commentRequest struct {
	Emoji string `json:"emoji"`
}

func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request) {
	ctx := reqcontext.FromContext(r.Context())

	conversationID, err := pathID(r, paramConversationID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	messageID, err := pathID(r, paramMessageID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var req commentRequest
	if err := readJSON(w, r, &req); err != nil || !allowedEmojis[req.Emoji] {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	comment, err := rt.db.CommentMessage(conversationID, messageID, ctx.UserID, req.Emoji)
	switch {
	case errors.Is(err, database.ErrConversationNotFound), errors.Is(err, database.ErrMessageNotFound):
		w.WriteHeader(http.StatusNotFound)
	case errors.Is(err, database.ErrMessageUnavailable):
		w.WriteHeader(http.StatusBadRequest)
	case err != nil:
		serverError(w, r, err, "cannot comment the message")
	default:
		writeJSON(w, r, http.StatusOK, comment)
	}
}
