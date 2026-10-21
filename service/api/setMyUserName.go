package api

import (
	"errors"
	"net/http"

	"github.com/lav-9994/wasatext/service/api/reqcontext"
	"github.com/lav-9994/wasatext/service/database"
)

type usernameRequest struct {
	Username string `json:"username"`
}

func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request) {
	ctx := reqcontext.FromContext(r.Context())

	var req usernameRequest
	if err := readJSON(w, r, &req); err != nil || !validText(req.Username, 3, 16, false) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := rt.db.SetUsername(ctx.UserID, req.Username)
	if errors.Is(err, database.ErrUsernameTaken) {
		w.WriteHeader(http.StatusConflict)
		return
	}
	if err != nil {
		serverError(w, r, err, "cannot set the username")
		return
	}
	writeJSON(w, r, http.StatusOK, user)
}
