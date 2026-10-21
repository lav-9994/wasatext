package database

import (
	"database/sql"
	"errors"
)

func (db *appdbimpl) checkMember(conversationID int64, userID int64) error {
	member, err := db.IsMember(conversationID, userID)
	if err != nil {
		return err
	}
	if !member {
		return ErrConversationNotFound
	}
	return nil
}

type messageInfo struct {
	msgType  string
	senderID int64
	deleted  bool
}

func (db *appdbimpl) messageInConversation(conversationID int64, messageID int64) (messageInfo, error) {
	var info messageInfo
	err := db.c.QueryRow(`
		SELECT type, sender_id, deleted FROM messages
		WHERE id = ? AND conversation_id = ?`, messageID, conversationID).Scan(&info.msgType, &info.senderID, &info.deleted)
	if errors.Is(err, sql.ErrNoRows) {
		return messageInfo{}, ErrMessageNotFound
	}
	return info, err
}

func insertMessage(tx *sql.Tx, conversationID int64, senderID int64, m NewMessage, forwarded bool) (int64, error) {
	var fileName, fileMime, fileData sql.NullString
	if m.File != nil {
		fileName = sql.NullString{String: m.File.Name, Valid: true}
		fileMime = sql.NullString{String: m.File.MimeType, Valid: true}
		fileData = sql.NullString{String: m.File.Data, Valid: true}
	}

	res, err := tx.Exec(`
		INSERT INTO messages (conversation_id, sender_id, timestamp, type, text, photo,
		                      file_name, file_mime, file_data, in_reply_to, forwarded)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		conversationID, senderID, now(), m.Type, m.Text, m.Photo,
		fileName, fileMime, fileData, m.InReplyTo, forwarded)
	if err != nil {
		return 0, err
	}
	messageID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(`
		INSERT INTO receipts (message_id, user_id)
		SELECT ?, user_id FROM members WHERE conversation_id = ? AND user_id <> ?`,
		messageID, conversationID, senderID)
	return messageID, err
}

func (db *appdbimpl) SendMessage(conversationID int64, userID int64, m NewMessage) (Message, error) {
	if err := db.checkMember(conversationID, userID); err != nil {
		return Message{}, err
	}
	if m.InReplyTo != nil {
		replied, err := db.messageInConversation(conversationID, *m.InReplyTo)
		if errors.Is(err, ErrMessageNotFound) {
			return Message{}, ErrInvalidReply
		}
		if err != nil {
			return Message{}, err
		}
		if replied.deleted {
			return Message{}, ErrInvalidReply
		}
	}

	tx, err := db.c.Begin()
	if err != nil {
		return Message{}, err
	}
	defer func() { _ = tx.Rollback() }()

	messageID, err := insertMessage(tx, conversationID, userID, m, false)
	if err != nil {
		return Message{}, err
	}
	if err := tx.Commit(); err != nil {
		return Message{}, err
	}
	return db.getMessage(messageID, userID)
}

func (db *appdbimpl) ForwardMessage(conversationID int64, messageID int64, targetID int64, userID int64) (Message, error) {
	if err := db.checkMember(conversationID, userID); err != nil {
		return Message{}, err
	}
	if _, err := db.messageInConversation(conversationID, messageID); err != nil {
		return Message{}, err
	}
	original, err := db.getMessage(messageID, userID)
	if err != nil {
		return Message{}, err
	}
	if original.Type == MessageSystem || original.Deleted {
		return Message{}, ErrMessageUnavailable
	}
	if err := db.checkMember(targetID, userID); err != nil {
		return Message{}, err
	}

	tx, err := db.c.Begin()
	if err != nil {
		return Message{}, err
	}
	defer func() { _ = tx.Rollback() }()

	copied := NewMessage{Type: original.Type, Text: original.Text, Photo: original.Photo, File: original.File}
	newID, err := insertMessage(tx, targetID, userID, copied, true)
	if err != nil {
		return Message{}, err
	}
	if err := tx.Commit(); err != nil {
		return Message{}, err
	}
	return db.getMessage(newID, userID)
}

func (db *appdbimpl) DeleteMessage(conversationID int64, messageID int64, userID int64) error {
	if err := db.checkMember(conversationID, userID); err != nil {
		return err
	}
	info, err := db.messageInConversation(conversationID, messageID)
	if err != nil {
		return err
	}
	if info.senderID != userID || info.msgType == MessageSystem {
		return ErrNotMessageOwner
	}

	tx, err := db.c.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`
		UPDATE messages
		SET deleted = 1, text = NULL, photo = NULL, file_name = NULL, file_mime = NULL, file_data = NULL
		WHERE id = ?`, messageID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM reactions WHERE message_id = ?`, messageID); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *appdbimpl) CommentMessage(conversationID int64, messageID int64, userID int64, emoji string) (Comment, error) {
	if err := db.checkMember(conversationID, userID); err != nil {
		return Comment{}, err
	}
	info, err := db.messageInConversation(conversationID, messageID)
	if err != nil {
		return Comment{}, err
	}
	if info.msgType == MessageSystem || info.deleted {
		return Comment{}, ErrMessageUnavailable
	}

	if _, err := db.c.Exec(`
		INSERT INTO reactions (message_id, user_id, emoji) VALUES (?, ?, ?)
		ON CONFLICT (message_id, user_id) DO UPDATE SET emoji = excluded.emoji`,
		messageID, userID, emoji); err != nil {
		return Comment{}, err
	}

	user, err := db.GetUser(userID)
	if err != nil {
		return Comment{}, err
	}
	return Comment{User: UserSummary{ID: user.ID, Username: user.Username}, Emoji: emoji}, nil
}

func (db *appdbimpl) UncommentMessage(conversationID int64, messageID int64, userID int64) error {
	if err := db.checkMember(conversationID, userID); err != nil {
		return err
	}
	if _, err := db.messageInConversation(conversationID, messageID); err != nil {
		return err
	}
	_, err := db.c.Exec(`DELETE FROM reactions WHERE message_id = ? AND user_id = ?`, messageID, userID)
	return err
}

func (db *appdbimpl) getMessage(messageID int64, userID int64) (Message, error) {
	m, err := scanMessage(db.c.QueryRow(`
		SELECT m.id, m.timestamp, u.id, u.username, m.type, m.text, m.photo,
		       m.file_name, m.file_mime, m.file_data, m.in_reply_to, m.forwarded, m.deleted
		FROM messages m JOIN users u ON u.id = m.sender_id
		WHERE m.id = ?`, messageID))
	if err != nil {
		return Message{}, err
	}

	rows, err := db.c.Query(`
		SELECT u.id, u.username, r.emoji
		FROM reactions r JOIN users u ON u.id = r.user_id
		WHERE r.message_id = ?
		ORDER BY u.username COLLATE NOCASE`, messageID)
	if err != nil {
		return Message{}, err
	}
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.User.ID, &c.User.Username, &c.Emoji); err != nil {
			_ = rows.Close()
			return Message{}, err
		}
		m.Comments = append(m.Comments, c)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return Message{}, err
	}
	_ = rows.Close()

	if m.Sender.ID == userID && m.Type != MessageSystem {
		status := StatusSent
		var count int
		var received, read sql.NullBool
		err := db.c.QueryRow(`
			SELECT COUNT(*), MIN(is_received), MIN(is_read)
			FROM receipts WHERE message_id = ?`, messageID).Scan(&count, &received, &read)
		if err != nil {
			return Message{}, err
		}
		if count > 0 {
			switch {
			case read.Bool:
				status = StatusRead
			case received.Bool:
				status = StatusReceived
			}
		}
		m.Status = &status
	}
	return m, nil
}
