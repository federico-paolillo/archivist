package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // SQLite driver (CGO-free)
)

// Open opens (or creates) a SQLite database at the given path with
// settings suitable for the Archivist worker: WAL mode, foreign keys,
// and a short busy timeout so atomic claims fail fast on contention.
func Open(path string) (*sql.DB, error) {
	dsn := path + "?_journal_mode=WAL&_foreign_keys=on&_busy_timeout=5000"

	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("db: failed to open SQLite at %q: %w", path, err)
	}

	err = database.PingContext(context.Background())
	if err != nil {
		_ = database.Close()

		return nil, fmt.Errorf("db: failed to ping SQLite at %q: %w", path, err)
	}

	// Single writer in v0: one connection avoids lock contention.
	database.SetMaxOpenConns(1)

	return database, nil
}
