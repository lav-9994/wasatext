package database

import (
	"database/sql"
	"time"

	"github.com/lav-9994/wasatext/service/globaltime"
)

func nullToPtr(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	return &s.String
}

func emptyToNull(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func now() string {
	return globaltime.Now().UTC().Format(time.RFC3339)
}

func insertSystemMessage(tx *sql.Tx, conversationID int64, senderID int64, text string) error {
	_, err := tx.Exec(`
		INSERT INTO messages (conversation_id, sender_id, timestamp, type, text)
		VALUES (?, ?, ?, ?, ?)`, conversationID, senderID, now(), MessageSystem, text)
	return err
}
