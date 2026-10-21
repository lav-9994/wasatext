package api

import (
	"errors"
	"net/http"

	"github.com/lav-9994/wasatext/service/api/reqcontext"
	"github.com/lav-9994/wasatext/service/database"
)

type forwardMessageRequest struct {
	ConversationID int64 `json:"conversationId"`
}

func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request) {
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
	var req forwardMessageRequest
	if err := readJSON(w, r, &req); err != nil || req.ConversationID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	message, err := rt.db.ForwardMessage(conversationID, messageID, req.ConversationID, ctx.UserID)
	switch {
	case errors.Is(err, database.ErrConversationNotFound), errors.Is(err, database.ErrMessageNotFound):
		w.WriteHeader(http.StatusNotFound)
	case errors.Is(err, database.ErrMessageUnavailable):
		w.WriteHeader(http.StatusBadRequest)
	case err != nil:
		serverError(w, r, err, "cannot forward the message")
	default:
		writeJSON(w, r, http.StatusCreated, message)
	}
}
