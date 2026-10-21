package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/lav-9994/wasatext/service/api/reqcontext"
)

func (rt *_router) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scheme, token, found := strings.Cut(r.Header.Get(headerAuthorization), " ")
		if !found || !strings.EqualFold(scheme, bearerScheme) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userID, err := strconv.ParseInt(strings.TrimSpace(token), 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		exists, err := rt.db.UserExists(userID)
		if err != nil {
			serverError(w, r, err, "cannot check the user")
			return
		}
		if !exists {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := reqcontext.FromContext(r.Context())
		ctx.UserID = userID
		ctx.Logger = ctx.Logger.WithField("user", userID)
		next.ServeHTTP(w, r.WithContext(reqcontext.WithContext(r.Context(), ctx)))
	})
}
