package database

type User struct {
	ID             int64   `json:"id"`
	Username       string  `json:"username"`
	ProfilePicture *string `json:"profilePicture,omitempty"`
	Bio            *string `json:"bio,omitempty"`
}

type UserSummary struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

type Attachment struct {
	Name     string `json:"name"`
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type Comment struct {
	User  UserSummary `json:"user"`
	Emoji string      `json:"emoji"`
}

type Message struct {
	ID        int64       `json:"id"`
	Timestamp string      `json:"timestamp"`
	Sender    UserSummary `json:"sender"`
	Type      string      `json:"type"`
	Text      *string     `json:"text,omitempty"`
	Photo     *string     `json:"photo,omitempty"`
	File      *Attachment `json:"file,omitempty"`
	InReplyTo *int64      `json:"inReplyTo,omitempty"`
	Forwarded bool        `json:"forwarded"`
	Deleted   bool        `json:"deleted"`
	Status    *string     `json:"status,omitempty"`
	Comments  []Comment   `json:"comments"`
}

type MessagePreview struct {
	Timestamp string      `json:"timestamp"`
	Type      string      `json:"type"`
	Sender    UserSummary `json:"sender"`
	Text      *string     `json:"text,omitempty"`
	Deleted   bool        `json:"deleted"`
}

type ConversationPreview struct {
	ID             int64          `json:"id"`
	Type           string         `json:"type"`
	DisplayName    string         `json:"displayName"`
	DisplayPicture *string        `json:"displayPicture,omitempty"`
	LastMessage    MessagePreview `json:"lastMessage"`
}

type Conversation struct {
	ID             int64        `json:"id"`
	Type           string       `json:"type"`
	DisplayName    string       `json:"displayName"`
	DisplayPicture *string      `json:"displayPicture,omitempty"`
	Description    *string      `json:"description,omitempty"`
	Creator        *UserSummary `json:"creator,omitempty"`
	Members        []User       `json:"members"`
	Messages       []Message    `json:"messages"`
}

type NewMessage struct {
	Type      string
	Text      *string
	Photo     *string
	File      *Attachment
	InReplyTo *int64
}

const (
	ConversationDirect = "direct"
	ConversationGroup  = "group"

	MessageText   = "text"
	MessagePhoto  = "photo"
	MessageFile   = "file"
	MessageSystem = "system"

	StatusSent     = "sent"
	StatusReceived = "received"
	StatusRead     = "read"

	MaxGroupMembers = 500
)
