/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

The database schema is defined by the SQL migrations embedded in this package
(see the migrations/ directory and the Migrate function). Apply them before
using an AppDatabase instance:

	// Apply pending schema migrations first.
	if err := database.Migrate(ctx, db); err != nil {
		return fmt.Errorf("migrating database: %w", err)
	}

	appdb, err := database.New(db)

The concrete implementation returned by New must satisfy the AppDatabase
interface. The rest of the program depends only on that interface, so a
different data layer (for example, sqlc or an ORM) can be plugged in by
providing another implementation of AppDatabase. See the "Database migrations"
section of the README for details.
*/
package database

import (
	"database/sql"
	"errors"
)

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	// Users (users.go)
	LoginUser(username string) (int64, error)
	UserExists(id int64) (bool, error)
	GetUser(id int64) (User, error)
	SetUsername(id int64, username string) (User, error)
	SetProfilePicture(id int64, picture string) (User, error)
	SetBio(id int64, bio string) (User, error)
	SearchUsers(text string, excludeID int64) ([]User, error)

	IsMember(conversationID int64, userID int64) (bool, error)
	GetMyConversations(userID int64) ([]ConversationPreview, error)
	GetCommonGroups(userID int64, otherID int64) ([]ConversationPreview, error)
	StartConversation(userID int64, otherID int64) (Conversation, bool, error)
	GetConversation(conversationID int64, userID int64) (Conversation, error)

	SendMessage(conversationID int64, userID int64, m NewMessage) (Message, error)
	ForwardMessage(conversationID int64, messageID int64, targetID int64, userID int64) (Message, error)
	DeleteMessage(conversationID int64, messageID int64, userID int64) error
	CommentMessage(conversationID int64, messageID int64, userID int64, emoji string) (Comment, error)
	UncommentMessage(conversationID int64, messageID int64, userID int64) error

	CreateGroup(userID int64, name string, memberIDs []int64) (Conversation, error)
	AddToGroup(groupID int64, userID int64, memberIDs []int64) error
	LeaveGroup(groupID int64, userID int64) error
	SetGroupName(groupID int64, userID int64, name string) error
	SetGroupPicture(groupID int64, userID int64, picture string) error
	SetGroupDescription(groupID int64, userID int64, description string) error

	Ping() error
}

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
//
// New does not create or update the schema: call Migrate before New so that the
// tables the methods rely on actually exist.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// SQLite allows only one writer at a time: using a single connection
	// avoids "database is locked" errors when requests arrive together.
	db.SetMaxOpenConns(1)

	return &appdbimpl{
		c: db,
	}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
