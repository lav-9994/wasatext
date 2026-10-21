package api

import (
	"net/http"

	"github.com/lav-9994/wasatext/service/api/reqcontext"
)

type profilePictureRequest struct {
	ProfilePicture string `json:"profilePicture"`
}

func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request) {
	ctx := reqcontext.FromContext(r.Context())

	var req profilePictureRequest
	if err := readJSON(w, r, &req); err != nil || !validImage(req.ProfilePicture) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := rt.db.SetProfilePicture(ctx.UserID, req.ProfilePicture)
	if err != nil {
		serverError(w, r, err, "cannot set the profile picture")
		return
	}
	writeJSON(w, r, http.StatusOK, user)
}
