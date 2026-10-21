package api

import (
	"errors"
	"net/http"

	"github.com/lav-9994/wasatext/service/api/reqcontext"
	"github.com/lav-9994/wasatext/service/database"
)

func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request) {
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

	err = rt.db.UncommentMessage(conversationID, messageID, ctx.UserID)
	switch {
	case errors.Is(err, database.ErrConversationNotFound), errors.Is(err, database.ErrMessageNotFound):
		w.WriteHeader(http.StatusNotFound)
	case err != nil:
		serverError(w, r, err, "cannot uncomment the message")
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}
