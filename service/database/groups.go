package database

import (
	"database/sql"
	"errors"
	"fmt"
)

func (db *appdbimpl) checkGroup(groupID int64, userID int64) error {
	var ctype string
	err := db.c.QueryRow(`SELECT type FROM conversations WHERE id = ?`, groupID).Scan(&ctype)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrConversationNotFound
	}
	if err != nil {
		return err
	}
	if ctype != ConversationGroup {
		return ErrConversationNotFound
	}
	return db.checkMember(groupID, userID)
}

func (db *appdbimpl) otherUsers(userID int64, ids []int64) ([]int64, error) {
	seen := map[int64]bool{}
	var others []int64
	for _, id := range ids {
		if id == userID || seen[id] {
			continue
		}
		exists, err := db.UserExists(id)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrInvalidMembers
		}
		seen[id] = true
		others = append(others, id)
	}
	return others, nil
}

func usernameTx(tx *sql.Tx, userID int64) (string, error) {
	var username string
	err := tx.QueryRow(`SELECT username FROM users WHERE id = ?`, userID).Scan(&username)
	return username, err
}

func (db *appdbimpl) CreateGroup(userID int64, name string, memberIDs []int64) (Conversation, error) {
	others, err := db.otherUsers(userID, memberIDs)
	if err != nil {
		return Conversation{}, err
	}
	if len(others) == 0 {
		return Conversation{}, ErrInvalidMembers
	}
	if len(others)+1 > MaxGroupMembers {
		return Conversation{}, ErrGroupFull
	}

	tx, err := db.c.Begin()
	if err != nil {
		return Conversation{}, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.Exec(`
		INSERT INTO conversations (type, name, creator_id, created_at) VALUES (?, ?, ?, ?)`,
		ConversationGroup, name, userID, now())
	if err != nil {
		return Conversation{}, err
	}
	groupID, err := res.LastInsertId()
	if err != nil {
		return Conversation{}, err
	}
	for _, id := range append([]int64{userID}, others...) {
		if _, err := tx.Exec(`INSERT INTO members (conversation_id, user_id) VALUES (?, ?)`, groupID, id); err != nil {
			return Conversation{}, err
		}
	}
	creator, err := usernameTx(tx, userID)
	if err != nil {
		return Conversation{}, err
	}
	text := fmt.Sprintf("%s ha creato il gruppo \"%s\"", creator, name)
	if err := insertSystemMessage(tx, groupID, userID, text); err != nil {
		return Conversation{}, err
	}
	if err := tx.Commit(); err != nil {
		return Conversation{}, err
	}
	return db.GetConversation(groupID, userID)
}

func (db *appdbimpl) AddToGroup(groupID int64, userID int64, memberIDs []int64) error {
	if err := db.checkGroup(groupID, userID); err != nil {
		return err
	}
	others, err := db.otherUsers(userID, memberIDs)
	if err != nil {
		return err
	}

	var newMembers []int64
	for _, id := range others {
		member, err := db.IsMember(groupID, id)
		if err != nil {
			return err
		}
		if !member {
			newMembers = append(newMembers, id)
		}
	}
	if len(newMembers) == 0 {
		return nil
	}

	var count int
	if err := db.c.QueryRow(`SELECT COUNT(*) FROM members WHERE conversation_id = ?`, groupID).Scan(&count); err != nil {
		return err
	}
	if count+len(newMembers) > MaxGroupMembers {
		return ErrGroupFull
	}

	tx, err := db.c.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	adder, err := usernameTx(tx, userID)
	if err != nil {
		return err
	}
	for _, id := range newMembers {
		if _, err := tx.Exec(`INSERT INTO members (conversation_id, user_id) VALUES (?, ?)`, groupID, id); err != nil {
			return err
		}
		added, err := usernameTx(tx, id)
		if err != nil {
			return err
		}
		if err := insertSystemMessage(tx, groupID, userID, fmt.Sprintf("%s ha aggiunto %s", adder, added)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (db *appdbimpl) LeaveGroup(groupID int64, userID int64) error {
	if err := db.checkGroup(groupID, userID); err != nil {
		return err
	}

	tx, err := db.c.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM members WHERE conversation_id = ? AND user_id = ?`, groupID, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		DELETE FROM receipts
		WHERE user_id = ? AND message_id IN (SELECT id FROM messages WHERE conversation_id = ?)`,
		userID, groupID); err != nil {
		return err
	}

	var remaining int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM members WHERE conversation_id = ?`, groupID).Scan(&remaining); err != nil {
		return err
	}

	if remaining == 0 {
		statements := []string{
			`DELETE FROM receipts WHERE message_id IN (SELECT id FROM messages WHERE conversation_id = ?)`,
			`DELETE FROM reactions WHERE message_id IN (SELECT id FROM messages WHERE conversation_id = ?)`,
			`DELETE FROM messages WHERE conversation_id = ?`,
			`DELETE FROM conversations WHERE id = ?`,
		}
		for _, stmt := range statements {
			if _, err := tx.Exec(stmt, groupID); err != nil {
				return err
			}
		}
		return tx.Commit()
	}

	leaver, err := usernameTx(tx, userID)
	if err != nil {
		return err
	}
	if err := insertSystemMessage(tx, groupID, userID, fmt.Sprintf("%s ha lasciato il gruppo", leaver)); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *appdbimpl) updateGroup(groupID int64, userID int64, query string, value any, message string) error {
	if err := db.checkGroup(groupID, userID); err != nil {
		return err
	}

	tx, err := db.c.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(query, value, groupID); err != nil {
		return err
	}
	username, err := usernameTx(tx, userID)
	if err != nil {
		return err
	}
	if err := insertSystemMessage(tx, groupID, userID, username+" "+message); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *appdbimpl) SetGroupName(groupID int64, userID int64, name string) error {
	return db.updateGroup(groupID, userID, `UPDATE conversations SET name = ? WHERE id = ?`, name,
		fmt.Sprintf("ha cambiato il nome del gruppo in \"%s\"", name))
}

func (db *appdbimpl) SetGroupPicture(groupID int64, userID int64, picture string) error {
	return db.updateGroup(groupID, userID, `UPDATE conversations SET picture = ? WHERE id = ?`, picture,
		"ha cambiato la foto del gruppo")
}

func (db *appdbimpl) SetGroupDescription(groupID int64, userID int64, description string) error {
	message := "ha cambiato la descrizione del gruppo"
	if description == "" {
		message = "ha rimosso la descrizione del gruppo"
	}
	return db.updateGroup(groupID, userID, `UPDATE conversations SET description = ? WHERE id = ?`,
		emptyToNull(description), message)
}
