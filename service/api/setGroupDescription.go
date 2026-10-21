package api

import (
	"errors"
	"net/http"

	"github.com/lav-9994/wasatext/service/api/reqcontext"
	"github.com/lav-9994/wasatext/service/database"
)

type groupDescriptionRequest struct {
	Description *string `json:"description"`
}

func (rt *_router) setGroupDescription(w http.ResponseWriter, r *http.Request) {
	ctx := reqcontext.FromContext(r.Context())

	groupID, err := pathID(r, paramGroupID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var req groupDescriptionRequest
	if err := readJSON(w, r, &req); err != nil || req.Description == nil || !validText(*req.Description, 0, 150, true) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = rt.db.SetGroupDescription(groupID, ctx.UserID, *req.Description)
	switch {
	case errors.Is(err, database.ErrConversationNotFound):
		w.WriteHeader(http.StatusNotFound)
	case err != nil:
		serverError(w, r, err, "cannot set the group description")
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}
