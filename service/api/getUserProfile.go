package api

import (
	"errors"
	"net/http"

	"github.com/lav-9994/wasatext/service/database"
)

func (rt *_router) getUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := pathID(r, paramUserID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := rt.db.GetUser(userID)
	if errors.Is(err, database.ErrUserNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		serverError(w, r, err, "cannot get the user")
		return
	}
	writeJSON(w, r, http.StatusOK, user)
}
