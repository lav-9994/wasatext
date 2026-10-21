package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/lav-9994/wasatext/service/api/reqcontext"
	"github.com/lav-9994/wasatext/service/database"
)

type sendMessageRequest struct {
	Type      string               `json:"type"`
	Text      *string              `json:"text"`
	Photo     *string              `json:"photo"`
	File      *database.Attachment `json:"file"`
	InReplyTo *int64               `json:"inReplyTo"`
}

func validAttachment(f database.Attachment) bool {
	return validText(f.Name, 1, 255, false) && !strings.ContainsAny(f.Name, `/\`) &&
		len(f.MimeType) >= 3 && len(f.MimeType) <= 255 && mimeTypePattern.MatchString(f.MimeType) &&
		len(f.Data) <= maxImageLength && base64Pattern.MatchString(f.Data)
}

func (req sendMessageRequest) valid() bool {
	if req.Text != nil && !validText(*req.Text, 1, 2000, true) {
		return false
	}
	if req.InReplyTo != nil && *req.InReplyTo <= 0 {
		return false
	}
	switch req.Type {
	case database.MessageText:
		return req.Text != nil && req.Photo == nil && req.File == nil
	case database.MessagePhoto:
		return req.Photo != nil && validImage(*req.Photo) && req.File == nil
	case database.MessageFile:
		return req.File != nil && validAttachment(*req.File) && req.Photo == nil
	default:
		return false
	}
}

func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request) {
	ctx := reqcontext.FromContext(r.Context())

	conversationID, err := pathID(r, paramConversationID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var req sendMessageRequest
	if err := readJSON(w, r, &req); err != nil || !req.valid() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	message, err := rt.db.SendMessage(conversationID, ctx.UserID, database.NewMessage{
		Type:      req.Type,
		Text:      req.Text,
		Photo:     req.Photo,
		File:      req.File,
		InReplyTo: req.InReplyTo,
	})
	switch {
	case errors.Is(err, database.ErrConversationNotFound):
		w.WriteHeader(http.StatusNotFound)
	case errors.Is(err, database.ErrInvalidReply):
		w.WriteHeader(http.StatusBadRequest)
	case err != nil:
		serverError(w, r, err, "cannot send the message")
	default:
		writeJSON(w, r, http.StatusCreated, message)
	}
}
