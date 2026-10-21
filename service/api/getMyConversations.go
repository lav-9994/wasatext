package api

import (
	"net/http"

	"github.com/lav-9994/wasatext/service/api/reqcontext"
)

func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request) {
	ctx := reqcontext.FromContext(r.Context())

	conversations, err := rt.db.GetMyConversations(ctx.UserID)
	if err != nil {
		serverError(w, r, err, "cannot get the conversations")
		return
	}
	writeJSON(w, r, http.StatusOK, conversations)
}
