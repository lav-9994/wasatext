package api

import (
	"errors"
	"net/http"

	"github.com/lav-9994/wasatext/service/api/reqcontext"
	"github.com/lav-9994/wasatext/service/database"
)

type groupPictureRequest struct {
	Picture string `json:"picture"`
}

func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request) {
	ctx := reqcontext.FromContext(r.Context())

	groupID, err := pathID(r, paramGroupID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var req groupPictureRequest
	if err := readJSON(w, r, &req); err != nil || !validImage(req.Picture) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = rt.db.SetGroupPicture(groupID, ctx.UserID, req.Picture)
	switch {
	case errors.Is(err, database.ErrConversationNotFound):
		w.WriteHeader(http.StatusNotFound)
	case err != nil:
		serverError(w, r, err, "cannot set the group picture")
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}
