package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/lav-9994/wasatext/service/api/reqcontext"
	"github.com/lav-9994/wasatext/service/database"
)

const (
	headerAuthorization = "Authorization"
	headerContentType   = "Content-Type"
	contentTypeJSON     = "application/json"
	bearerScheme        = "Bearer"

	paramUserID         = "userId"
	paramConversationID = "conversationId"
	paramMessageID      = "messageId"
	paramGroupID        = "groupId"

	maxRequestSize = 10 << 20
	maxImageLength = 7000000
)

var (
	imagePattern    = regexp.MustCompile(`^data:image/[a-zA-Z0-9.+-]+;base64,[A-Za-z0-9+/]+=*$`)
	mimeTypePattern = regexp.MustCompile(`^[a-zA-Z0-9.+-]+/[a-zA-Z0-9.+-]+$`)
	base64Pattern   = regexp.MustCompile(`^[A-Za-z0-9+/]+=*$`)
)

var allowedEmojis = map[string]bool{"👍": true, "❤️": true, "😂": true, "😮": true, "😢": true, "🙏": true}

var errInvalidID = errors.New("invalid identifier")

func writeJSON(w http.ResponseWriter, r *http.Request, status int, value any) {
	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		reqcontext.FromContext(r.Context()).Logger.WithError(err).Error("cannot encode the response")
	}
}

func readJSON(w http.ResponseWriter, r *http.Request, dest any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)
	return json.NewDecoder(r.Body).Decode(dest)
}

func pathID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, name), 10, 64)
	if err != nil || id <= 0 {
		return 0, errInvalidID
	}
	return id, nil
}

func validText(s string, minLength int, maxLength int, multiline bool) bool {
	n := utf8.RuneCountInString(s)
	if n < minLength || n > maxLength || !utf8.ValidString(s) {
		return false
	}
	return multiline || !strings.ContainsAny(s, "\r\n")
}

func validImage(s string) bool {
	return len(s) <= maxImageLength && imagePattern.MatchString(s)
}

func serverError(w http.ResponseWriter, r *http.Request, err error, message string) {
	reqcontext.FromContext(r.Context()).Logger.WithError(err).Error(message)
	w.WriteHeader(http.StatusInternalServerError)
}

func validIDList(ids []int64) bool {
	if len(ids) < 1 || len(ids) > database.MaxGroupMembers-1 {
		return false
	}
	seen := map[int64]bool{}
	for _, id := range ids {
		if id <= 0 || seen[id] {
			return false
		}
		seen[id] = true
	}
	return true
}
