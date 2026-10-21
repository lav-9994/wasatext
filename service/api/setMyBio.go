package api

import (
	"net/http"

	"github.com/lav-9994/wasatext/service/api/reqcontext"
)

type bioRequest struct {
	Bio *string `json:"bio"`
}

func (rt *_router) setMyBio(w http.ResponseWriter, r *http.Request) {
	ctx := reqcontext.FromContext(r.Context())

	var req bioRequest
	if err := readJSON(w, r, &req); err != nil || req.Bio == nil || !validText(*req.Bio, 0, 150, true) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := rt.db.SetBio(ctx.UserID, *req.Bio)
	if err != nil {
		serverError(w, r, err, "cannot set the bio")
		return
	}
	writeJSON(w, r, http.StatusOK, user)
}
