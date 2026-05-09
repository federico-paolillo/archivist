package db

import (
	"context"
	"database/sql"
	"fmt"
)

// schema is the Archivist SQLite DDL.
// Schema matches the contract defined in TELING-001 persistence contracts.
const schema = `
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY NOT NULL,
    telegram_user_id TEXT UNIQUE,
    password_hash TEXT
) STRICT;

CREATE TABLE IF NOT EXISTS articles (
    id TEXT PRIMARY KEY NOT NULL,
    user_id TEXT NOT NULL REFERENCES users(id),
    original_url TEXT NOT NULL,
    canonical_url TEXT,
    title TEXT,
    status TEXT NOT NULL CHECK(status IN ('queued', 'ready', 'failed')),
    error_message TEXT,
    created_at TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY NOT NULL,
    user_id TEXT NOT NULL REFERENCES users(id),
    article_id TEXT NOT NULL REFERENCES articles(id),
    type TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('queued', 'running', 'succeeded', 'failed')),
    telegram_update_id INTEGER UNIQUE,
    telegram_chat_id INTEGER,
    telegram_message_id INTEGER,
    telegram_user_id INTEGER,
    error_message TEXT,
    created_at TEXT NOT NULL,
    started_at TEXT,
    completed_at TEXT,
    expires_at TEXT
) STRICT;

CREATE TABLE IF NOT EXISTS notifications (
    id TEXT PRIMARY KEY NOT NULL,
    job_id TEXT NOT NULL UNIQUE REFERENCES jobs(id),
    status TEXT NOT NULL CHECK(status IN ('pending', 'sent', 'failed')),
    error_message TEXT,
    created_at TEXT NOT NULL,
    sent_at TEXT,
    expires_at TEXT NOT NULL
) STRICT;
`

// ApplySchema applies the Archivist schema DDL to the given database.
func ApplySchema(database *sql.DB) error {
	_, err := database.ExecContext(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("db: failed to apply schema: %w", err)
	}

	return nil
}
