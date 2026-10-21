
-- Conceptual model: the entities are divided into User, Conversation with the subtypes direct
-- conversation and groups, Message. The relationships are "member of", "creates",
-- "reacts to" (with emoji), "replies to".
-- Foreign keys document the relationships while the integrity is checked in code.

CREATE TABLE users (
    id              INTEGER NOT NULL PRIMARY KEY,
    username        TEXT    NOT NULL UNIQUE,
    profile_picture TEXT,
    bio             TEXT
);

CREATE TABLE conversations (
    id          INTEGER NOT NULL PRIMARY KEY,
    type        TEXT    NOT NULL CHECK (type IN ('direct', 'group')),
    name        TEXT,
    picture     TEXT,
    description TEXT,
    creator_id  INTEGER REFERENCES users(id),
    created_at  TEXT    NOT NULL
);


CREATE TABLE members (
    conversation_id INTEGER NOT NULL REFERENCES conversations(id),
    user_id         INTEGER NOT NULL REFERENCES users(id),
    PRIMARY KEY (conversation_id, user_id)
);
CREATE INDEX idx_members_user ON members(user_id);

CREATE TABLE messages (
    id              INTEGER NOT NULL PRIMARY KEY,
    conversation_id INTEGER NOT NULL REFERENCES conversations(id),
    sender_id       INTEGER NOT NULL REFERENCES users(id),
    timestamp       TEXT    NOT NULL,
    type            TEXT    NOT NULL CHECK (type IN ('text', 'photo', 'file', 'system')),
    text            TEXT,
    photo           TEXT,
    file_name       TEXT,
    file_mime       TEXT,
    file_data       TEXT,
    in_reply_to     INTEGER REFERENCES messages(id),
    forwarded       INTEGER NOT NULL DEFAULT 0,
    deleted         INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX idx_messages_conversation ON messages(conversation_id, id);

CREATE TABLE reactions (
    message_id INTEGER NOT NULL REFERENCES messages(id),
    user_id    INTEGER NOT NULL REFERENCES users(id),
    emoji      TEXT    NOT NULL,
    PRIMARY KEY (message_id, user_id)
);

CREATE TABLE receipts (
    message_id  INTEGER NOT NULL REFERENCES messages(id),
    user_id     INTEGER NOT NULL REFERENCES users(id),
    is_received INTEGER NOT NULL DEFAULT 0,
    is_read     INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (message_id, user_id)
);
CREATE INDEX idx_receipts_user ON receipts(user_id);
