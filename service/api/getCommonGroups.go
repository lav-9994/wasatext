package api

import (
	"errors"
	"net/http"

	"github.com/lav-9994/wasatext/service/api/reqcontext"
	"github.com/lav-9994/wasatext/service/database"
)

func (rt *_router) getCommonGroups(w http.ResponseWriter, r *http.Request) {
	ctx := reqcontext.FromContext(r.Context())

	otherID, err := pathID(r, paramUserID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	groups, err := rt.db.GetCommonGroups(ctx.UserID, otherID)
	if errors.Is(err, database.ErrUserNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		serverError(w, r, err, "cannot get the groups in common")
		return
	}
	writeJSON(w, r, http.StatusOK, groups)
}
