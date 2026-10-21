package database

import (
	"database/sql"
	"errors"
)

func (db *appdbimpl) IsMember(conversationID int64, userID int64) (bool, error) {
	var member bool
	err := db.c.QueryRow(`
		SELECT EXISTS (SELECT 1 FROM members WHERE conversation_id = ? AND user_id = ?)`,
		conversationID, userID).Scan(&member)
	return member, err
}

func (db *appdbimpl) GetMyConversations(userID int64) ([]ConversationPreview, error) {

	if _, err := db.c.Exec(`
		UPDATE receipts SET is_received = 1
		WHERE user_id = ? AND is_received = 0`, userID); err != nil {
		return nil, err

	}

	return db.conversationPreviews(`
		SELECT c.id, c.type, c.name, c.picture, MAX(m.id) AS last_id
		FROM conversations c
		JOIN members mb ON mb.conversation_id = c.id AND mb.user_id = ?
		JOIN messages m ON m.conversation_id = c.id
		GROUP BY c.id
		ORDER BY last_id DESC`, userID, userID)

}

func (db *appdbimpl) GetCommonGroups(userID int64, otherID int64) ([]ConversationPreview, error) {
	exists, err := db.UserExists(otherID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	return db.conversationPreviews(`
		SELECT c.id, c.type, c.name, c.picture, MAX(m.id) AS last_id
		FROM conversations c
		JOIN members m1 ON m1.conversation_id = c.id AND m1.user_id = ?
		JOIN members m2 ON m2.conversation_id = c.id AND m2.user_id = ?
		JOIN messages m ON m.conversation_id = c.id
		WHERE c.type = 'group'
		GROUP BY c.id
		ORDER BY last_id DESC`, userID, userID, otherID)
}

type previewRow struct {
	id      int64
	ctype   string
	name    sql.NullString
	picture sql.NullString
	lastID  int64
}

func (db *appdbimpl) conversationPreviews(query string, userID int64, args ...any) ([]ConversationPreview, error) {
	rows, err := db.c.Query(query, args...)
	if err != nil {
		return nil, err
	}
	var list []previewRow
	for rows.Next() {
		var r previewRow
		if err := rows.Scan(&r.id, &r.ctype, &r.name, &r.picture, &r.lastID); err != nil {
			_ = rows.Close()
			return nil, err
		}
		list = append(list, r)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	_ = rows.Close()

	previews := []ConversationPreview{}
	for _, r := range list {
		p := ConversationPreview{ID: r.id, Type: r.ctype}
		p.DisplayName, p.DisplayPicture, err = db.displayInfo(r.id, r.ctype, r.name, r.picture, userID)
		if err != nil {
			return nil, err
		}
		p.LastMessage, err = db.messagePreview(r.lastID)
		if err != nil {
			return nil, err
		}
		previews = append(previews, p)
	}
	return previews, nil
}

func (db *appdbimpl) displayInfo(conversationID int64, ctype string, name sql.NullString, picture sql.NullString, userID int64) (string, *string, error) {
	if ctype == ConversationGroup {
		return name.String, nullToPtr(picture), nil
	}

	var username string
	var otherPicture sql.NullString
	err := db.c.QueryRow(`
		SELECT u.username, u.profile_picture
		FROM members m JOIN users u ON u.id = m.user_id
		WHERE m.conversation_id = ? AND m.user_id <> ?`, conversationID, userID).Scan(&username, &otherPicture)
	if err != nil {
		return "", nil, err
	}
	return username, nullToPtr(otherPicture), nil
}

func (db *appdbimpl) messagePreview(messageID int64) (MessagePreview, error) {
	var p MessagePreview
	var text sql.NullString
	err := db.c.QueryRow(`
		SELECT m.timestamp, m.type, u.id, u.username, m.text, m.deleted
		FROM messages m JOIN users u ON u.id = m.sender_id
		WHERE m.id = ?`, messageID).Scan(&p.Timestamp, &p.Type, &p.Sender.ID, &p.Sender.Username, &text, &p.Deleted)
	if err != nil {
		return MessagePreview{}, err
	}
	if text.Valid {
		t := previewText(text.String)
		p.Text = &t
	}
	return p, nil
}

func (db *appdbimpl) StartConversation(userID int64, otherID int64) (Conversation, bool, error) {
	exists, err := db.UserExists(otherID)
	if err != nil {
		return Conversation{}, false, err
	}
	if !exists {
		return Conversation{}, false, ErrUserNotFound
	}

	var existingID int64
	err = db.c.QueryRow(`
		SELECT c.id
		FROM conversations c
		JOIN members m1 ON m1.conversation_id = c.id AND m1.user_id = ?
		JOIN members m2 ON m2.conversation_id = c.id AND m2.user_id = ?
		WHERE c.type = 'direct'`, userID, otherID).Scan(&existingID)
	if err == nil {
		conv, err := db.GetConversation(existingID, userID)
		return conv, false, err
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Conversation{}, false, err
	}

	tx, err := db.c.Begin()
	if err != nil {
		return Conversation{}, false, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.Exec(`INSERT INTO conversations (type, created_at) VALUES (?, ?)`, ConversationDirect, now())
	if err != nil {
		return Conversation{}, false, err
	}
	conversationID, err := res.LastInsertId()
	if err != nil {
		return Conversation{}, false, err
	}
	if _, err := tx.Exec(`
		INSERT INTO members (conversation_id, user_id) VALUES (?, ?), (?, ?)`,
		conversationID, userID, conversationID, otherID); err != nil {
		return Conversation{}, false, err
	}
	if err := insertSystemMessage(tx, conversationID, userID, "Conversazione iniziata"); err != nil {
		return Conversation{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return Conversation{}, false, err
	}

	conv, err := db.GetConversation(conversationID, userID)
	return conv, true, err
}

func (db *appdbimpl) GetConversation(conversationID int64, userID int64) (Conversation, error) {
	member, err := db.IsMember(conversationID, userID)
	if err != nil {
		return Conversation{}, err
	}
	if !member {
		return Conversation{}, ErrConversationNotFound
	}

	if _, err := db.c.Exec(`
		UPDATE receipts SET is_received = 1, is_read = 1
		WHERE user_id = ? AND is_read = 0
		AND message_id IN (SELECT id FROM messages WHERE conversation_id = ?)`,
		userID, conversationID); err != nil {
		return Conversation{}, err
	}

	var c Conversation
	var name, picture, description sql.NullString
	var creatorID sql.NullInt64
	err = db.c.QueryRow(`
		SELECT id, type, name, picture, description, creator_id
		FROM conversations WHERE id = ?`, conversationID).Scan(&c.ID, &c.Type, &name, &picture, &description, &creatorID)
	if err != nil {
		return Conversation{}, err
	}

	c.DisplayName, c.DisplayPicture, err = db.displayInfo(c.ID, c.Type, name, picture, userID)
	if err != nil {
		return Conversation{}, err
	}
	if c.Type == ConversationGroup {
		c.Description = nullToPtr(description)
		if creatorID.Valid {
			creator, err := db.GetUser(creatorID.Int64)
			if err != nil {
				return Conversation{}, err
			}
			c.Creator = &UserSummary{ID: creator.ID, Username: creator.Username}
		}
	}

	c.Members, err = db.members(c.ID)
	if err != nil {
		return Conversation{}, err
	}
	c.Messages, err = db.messages(c.ID, userID)
	if err != nil {
		return Conversation{}, err
	}
	return c, nil
}

func (db *appdbimpl) members(conversationID int64) ([]User, error) {
	rows, err := db.c.Query(`
		SELECT u.id, u.username, u.profile_picture, u.bio
		FROM members m JOIN users u ON u.id = m.user_id
		WHERE m.conversation_id = ?
		ORDER BY u.username COLLATE NOCASE`, conversationID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	users := []User{}
	for rows.Next() {
		var u User
		var picture, bio sql.NullString
		if err := rows.Scan(&u.ID, &u.Username, &picture, &bio); err != nil {
			return nil, err
		}
		u.ProfilePicture = nullToPtr(picture)
		u.Bio = nullToPtr(bio)
		users = append(users, u)
	}
	return users, rows.Err()
}

func (db *appdbimpl) messages(conversationID int64, userID int64) ([]Message, error) {
	rows, err := db.c.Query(`
		SELECT m.id, m.timestamp, u.id, u.username, m.type, m.text, m.photo,
		       m.file_name, m.file_mime, m.file_data, m.in_reply_to, m.forwarded, m.deleted
		FROM messages m JOIN users u ON u.id = m.sender_id
		WHERE m.conversation_id = ?
		ORDER BY m.id DESC`, conversationID)
	if err != nil {
		return nil, err
	}

	messages := []Message{}
	for rows.Next() {
		m, err := scanMessage(rows)
		if err != nil {
			_ = rows.Close()
			return nil, err
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	_ = rows.Close()

	comments, err := db.conversationComments(conversationID)
	if err != nil {
		return nil, err
	}
	statuses, err := db.conversationStatuses(conversationID, userID)
	if err != nil {
		return nil, err
	}

	for i := range messages {
		m := &messages[i]
		if c, ok := comments[m.ID]; ok {
			m.Comments = c
		}
		if m.Sender.ID == userID && m.Type != MessageSystem {
			s := statuses[m.ID]
			if s == "" {
				s = StatusSent
			}
			m.Status = &s
		}
	}
	return messages, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanMessage(row rowScanner) (Message, error) {
	var m Message
	var text, photo, fileName, fileMime, fileData sql.NullString
	var inReplyTo sql.NullInt64
	err := row.Scan(&m.ID, &m.Timestamp, &m.Sender.ID, &m.Sender.Username, &m.Type, &text, &photo,
		&fileName, &fileMime, &fileData, &inReplyTo, &m.Forwarded, &m.Deleted)
	if err != nil {
		return Message{}, err
	}
	m.Text = nullToPtr(text)
	m.Photo = nullToPtr(photo)
	if fileData.Valid {
		m.File = &Attachment{Name: fileName.String, MimeType: fileMime.String, Data: fileData.String}
	}
	if inReplyTo.Valid {
		id := inReplyTo.Int64
		m.InReplyTo = &id
	}
	m.Comments = []Comment{}
	return m, nil
}

func (db *appdbimpl) conversationComments(conversationID int64) (map[int64][]Comment, error) {
	rows, err := db.c.Query(`
		SELECT r.message_id, u.id, u.username, r.emoji
		FROM reactions r
		JOIN messages m ON m.id = r.message_id
		JOIN users u ON u.id = r.user_id
		WHERE m.conversation_id = ?
		ORDER BY u.username COLLATE NOCASE`, conversationID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	comments := map[int64][]Comment{}
	for rows.Next() {
		var messageID int64
		var c Comment
		if err := rows.Scan(&messageID, &c.User.ID, &c.User.Username, &c.Emoji); err != nil {
			return nil, err
		}
		comments[messageID] = append(comments[messageID], c)
	}
	return comments, rows.Err()
}

func (db *appdbimpl) conversationStatuses(conversationID int64, userID int64) (map[int64]string, error) {
	rows, err := db.c.Query(`
		SELECT r.message_id, MIN(r.is_received), MIN(r.is_read)
		FROM receipts r JOIN messages m ON m.id = r.message_id
		WHERE m.conversation_id = ? AND m.sender_id = ?
		GROUP BY r.message_id`, conversationID, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	statuses := map[int64]string{}
	for rows.Next() {
		var messageID int64
		var received, read bool
		if err := rows.Scan(&messageID, &received, &read); err != nil {
			return nil, err
		}
		switch {
		case read:
			statuses[messageID] = StatusRead
		case received:
			statuses[messageID] = StatusReceived
		default:
			statuses[messageID] = StatusSent
		}
	}
	return statuses, rows.Err()
}

func previewText(s string) string {
	r := []rune(s)
	if len(r) <= 50 {
		return s
	}
	return string(r[:50]) + "…"
}
