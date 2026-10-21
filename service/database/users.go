package database

import (
	"database/sql"
	"errors"
	"strings"
)

func (db *appdbimpl) LoginUser(username string) (int64, error) {
	_, err := db.c.Exec(`
		INSERT INTO users (username) VALUES (?)
		ON CONFLICT (username) DO NOTHING`, username)
	if err != nil {
		return 0, err
	}

	var id int64
	err = db.c.QueryRow(`SELECT id FROM users WHERE username = ?`, username).Scan(&id)
	return id, err
}

func (db *appdbimpl) UserExists(id int64) (bool, error) {
	var exists bool
	err := db.c.QueryRow(`SELECT EXISTS (SELECT 1 FROM users WHERE id = ?)`, id).Scan(&exists)
	return exists, err
}

func (db *appdbimpl) GetUser(id int64) (User, error) {
	var u User
	var picture, bio sql.NullString
	err := db.c.QueryRow(`
		SELECT id, username, profile_picture, bio
		FROM users WHERE id = ?`, id).Scan(&u.ID, &u.Username, &picture, &bio)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, err
	}
	u.ProfilePicture = nullToPtr(picture)
	u.Bio = nullToPtr(bio)
	return u, nil
}

func (db *appdbimpl) SetUsername(id int64, username string) (User, error) {
	var taken bool
	err := db.c.QueryRow(`
		SELECT EXISTS (SELECT 1 FROM users WHERE username = ? AND id <> ?)`,
		username, id).Scan(&taken)
	if err != nil {
		return User{}, err
	}
	if taken {
		return User{}, ErrUsernameTaken
	}

	if _, err := db.c.Exec(`UPDATE users SET username = ? WHERE id = ?`, username, id); err != nil {
		return User{}, err
	}
	return db.GetUser(id)
}

func (db *appdbimpl) SetProfilePicture(id int64, picture string) (User, error) {
	if _, err := db.c.Exec(`UPDATE users SET profile_picture = ? WHERE id = ?`, picture, id); err != nil {
		return User{}, err
	}
	return db.GetUser(id)
}

func (db *appdbimpl) SetBio(id int64, bio string) (User, error) {

	if _, err := db.c.Exec(`UPDATE users SET bio = ? WHERE id = ?`, emptyToNull(bio), id); err != nil {
		return User{}, err
	}
	return db.GetUser(id)

}

func (db *appdbimpl) SearchUsers(text string, excludeID int64) ([]User, error) {

	rows, err := db.c.Query(`
		SELECT id, username, profile_picture, bio
		FROM users
		WHERE username LIKE '%' || ? || '%' ESCAPE '\' AND id <> ?
		ORDER BY username COLLATE NOCASE`, escapeLike(text), excludeID)
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

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	return strings.ReplaceAll(s, `_`, `\_`)
}
