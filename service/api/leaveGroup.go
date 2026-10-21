package api

import (
	"errors"
	"net/http"

	"github.com/lav-9994/wasatext/service/api/reqcontext"
	"github.com/lav-9994/wasatext/service/database"
)

func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request) {
	ctx := reqcontext.FromContext(r.Context())

	groupID, err := pathID(r, paramGroupID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = rt.db.LeaveGroup(groupID, ctx.UserID)
	switch {
	case errors.Is(err, database.ErrConversationNotFound):
		w.WriteHeader(http.StatusNotFound)
	case err != nil:
		serverError(w, r, err, "cannot leave the group")
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}
