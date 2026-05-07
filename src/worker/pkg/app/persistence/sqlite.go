package persistence

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func OpenSQLite(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("persistence: open sqlite: %w", err)
	}

	_, err = db.ExecContext(ctx, "PRAGMA foreign_keys = ON;")
	if err != nil {
		_ = db.Close()

		return nil, fmt.Errorf("persistence: enable foreign keys: %w", err)
	}

	return db, nil
}

func EnsureSchema(ctx context.Context, db *sql.DB) error {
	statements := []string{
		`PRAGMA foreign_keys = ON;`,
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			telegram_user_id INTEGER UNIQUE NULL,
			password_hash TEXT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS articles (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			original_url TEXT NOT NULL,
			canonical_url TEXT NULL,
			title TEXT NULL,
			status TEXT NOT NULL CHECK (status IN ('queued', 'ready', 'failed')),
			error_message TEXT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
		);`,
		`CREATE TABLE IF NOT EXISTS jobs (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			article_id TEXT NOT NULL,
			type TEXT NOT NULL,
			status TEXT NOT NULL CHECK (status IN ('queued', 'running', 'succeeded', 'failed')),
			telegram_update_id INTEGER UNIQUE NULL,
			telegram_chat_id INTEGER NULL,
			telegram_message_id INTEGER NULL,
			telegram_user_id INTEGER NULL,
			error_message TEXT NULL,
			created_at TEXT NOT NULL,
			started_at TEXT NULL,
			completed_at TEXT NULL,
			expires_at TEXT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
			FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS notifications (
			id TEXT PRIMARY KEY,
			job_id TEXT NOT NULL UNIQUE,
			status TEXT NOT NULL CHECK (status IN ('pending', 'sent', 'failed')),
			error_message TEXT NULL,
			created_at TEXT NOT NULL,
			sent_at TEXT NULL,
			expires_at TEXT NULL,
			FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
		);`,
	}

	for _, statement := range statements {
		_, err := db.ExecContext(ctx, statement)
		if err != nil {
			return fmt.Errorf("persistence: ensure schema: %w", err)
		}
	}

	return nil
}
