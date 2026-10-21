package api

import (
	"errors"
	"net/http"

	"github.com/lav-9994/wasatext/service/api/reqcontext"
	"github.com/lav-9994/wasatext/service/database"
)

type addToGroupRequest struct {
	UserIDs []int64 `json:"userIds"`
}

func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request) {
	ctx := reqcontext.FromContext(r.Context())

	groupID, err := pathID(r, paramGroupID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var req addToGroupRequest
	if err := readJSON(w, r, &req); err != nil || !validIDList(req.UserIDs) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = rt.db.AddToGroup(groupID, ctx.UserID, req.UserIDs)
	switch {
	case errors.Is(err, database.ErrConversationNotFound):
		w.WriteHeader(http.StatusNotFound)
	case errors.Is(err, database.ErrInvalidMembers), errors.Is(err, database.ErrGroupFull):
		w.WriteHeader(http.StatusBadRequest)
	case err != nil:
		serverError(w, r, err, "cannot add users to the group")
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}
