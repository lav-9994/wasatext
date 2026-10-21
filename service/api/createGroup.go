package api

import (
	"errors"
	"net/http"

	"github.com/lav-9994/wasatext/service/api/reqcontext"
	"github.com/lav-9994/wasatext/service/database"
)

type createGroupRequest struct {
	Name    string  `json:"name"`
	Members []int64 `json:"members"`
}

func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request) {
	ctx := reqcontext.FromContext(r.Context())

	var req createGroupRequest
	if err := readJSON(w, r, &req); err != nil || !validText(req.Name, 1, 32, false) || !validIDList(req.Members) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	group, err := rt.db.CreateGroup(ctx.UserID, req.Name, req.Members)
	switch {
	case errors.Is(err, database.ErrInvalidMembers), errors.Is(err, database.ErrGroupFull):
		w.WriteHeader(http.StatusBadRequest)
	case err != nil:
		serverError(w, r, err, "cannot create the group")
	default:
		writeJSON(w, r, http.StatusCreated, group)
	}
}
