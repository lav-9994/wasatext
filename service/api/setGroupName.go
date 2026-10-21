package api

import (
	"errors"
	"net/http"

	"github.com/lav-9994/wasatext/service/api/reqcontext"
	"github.com/lav-9994/wasatext/service/database"
)

type groupNameRequest struct {
	Name string `json:"name"`
}

func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request) {
	ctx := reqcontext.FromContext(r.Context())

	groupID, err := pathID(r, paramGroupID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var req groupNameRequest
	if err := readJSON(w, r, &req); err != nil || !validText(req.Name, 1, 32, false) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = rt.db.SetGroupName(groupID, ctx.UserID, req.Name)
	switch {
	case errors.Is(err, database.ErrConversationNotFound):
		w.WriteHeader(http.StatusNotFound)
	case err != nil:
		serverError(w, r, err, "cannot set the group name")
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}
