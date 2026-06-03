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
    traceparent TEXT,
    tracestate TEXT,
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

	err = ensureColumn(database, "jobs", "traceparent", "TEXT")
	if err != nil {
		return err
	}

	err = ensureColumn(database, "jobs", "tracestate", "TEXT")
	if err != nil {
		return err
	}

	return nil
}

func ensureColumn(database *sql.DB, table string, column string, definition string) error {
	ctx := context.Background()

	exists, err := columnExists(ctx, database, table, column)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	_, err = database.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition))
	if err != nil {
		return fmt.Errorf("db: failed to add %s.%s column: %w", table, column, err)
	}

	return nil
}

func columnExists(ctx context.Context, database *sql.DB, table string, column string) (bool, error) {
	rows, err := database.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, fmt.Errorf("db: failed to inspect %s columns: %w", table, err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var (
			cid        int
			name       string
			columnType string
			notNull    int
			defaultVal sql.NullString
			pk         int
		)

		err = rows.Scan(&cid, &name, &columnType, &notNull, &defaultVal, &pk)
		if err != nil {
			return false, fmt.Errorf("db: failed to scan %s column metadata: %w", table, err)
		}

		if name == column {
			return true, nil
		}
	}

	err = rows.Err()
	if err != nil {
		return false, fmt.Errorf("db: failed to iterate %s column metadata: %w", table, err)
	}

	return false, nil
}
