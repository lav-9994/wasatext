package api

import (
	"errors"
	"net/http"

	"github.com/lav-9994/wasatext/service/api/reqcontext"
	"github.com/lav-9994/wasatext/service/database"
)

func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request) {
	ctx := reqcontext.FromContext(r.Context())

	conversationID, err := pathID(r, paramConversationID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	conversation, err := rt.db.GetConversation(conversationID, ctx.UserID)
	if errors.Is(err, database.ErrConversationNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		serverError(w, r, err, "cannot get the conversation")
		return
	}
	writeJSON(w, r, http.StatusOK, conversation)
}
