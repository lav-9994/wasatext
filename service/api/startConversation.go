package api

import (
	"errors"
	"net/http"

	"github.com/lav-9994/wasatext/service/api/reqcontext"
	"github.com/lav-9994/wasatext/service/database"
)

type startConversationRequest struct {
	UserID int64 `json:"userId"`
}

func (rt *_router) startConversation(w http.ResponseWriter, r *http.Request) {
	ctx := reqcontext.FromContext(r.Context())

	var req startConversationRequest
	if err := readJSON(w, r, &req); err != nil || req.UserID <= 0 || req.UserID == ctx.UserID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	conversation, created, err := rt.db.StartConversation(ctx.UserID, req.UserID)
	if errors.Is(err, database.ErrUserNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		serverError(w, r, err, "cannot start the conversation")
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, r, status, conversation)
}
