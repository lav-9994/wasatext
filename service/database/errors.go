package database

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")

	ErrUsernameTaken = errors.New("username already taken")

	ErrConversationNotFound = errors.New("conversation not found")

	ErrMessageNotFound = errors.New("message not found")

	ErrInvalidReply = errors.New("invalid reply")

	ErrMessageUnavailable = errors.New("message is a system message or has been deleted")

	ErrNotMessageOwner = errors.New("message not sent by the user")

	ErrInvalidMembers = errors.New("invalid list of members")

	ErrGroupFull = errors.New("too many members in the group")
)
