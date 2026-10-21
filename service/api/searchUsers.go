package api

import (
	"net/http"

	"github.com/lav-9994/wasatext/service/api/reqcontext"
)

func (rt *_router) searchUsers(w http.ResponseWriter, r *http.Request) {
	ctx := reqcontext.FromContext(r.Context())

	text := r.URL.Query().Get("username")
	if !validText(text, 0, 16, false) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	users, err := rt.db.SearchUsers(text, ctx.UserID)
	if err != nil {
		serverError(w, r, err, "cannot search users")
		return
	}
	writeJSON(w, r, http.StatusOK, users)
}
