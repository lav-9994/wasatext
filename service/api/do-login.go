package api

import (
	"net/http"
)

type loginRequest struct {
	Name string `json:"name"`
}

type loginResponse struct {
	Identifier int64 `json:"identifier"`
}

func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := readJSON(w, r, &req); err != nil || !validText(req.Name, 3, 16, false) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := rt.db.LoginUser(req.Name)
	if err != nil {
		serverError(w, r, err, "cannot log in the user")
		return
	}

	writeJSON(w, r, http.StatusCreated, loginResponse{Identifier: id})
}
