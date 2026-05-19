package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "modernc.org/sqlite" // SQLite driver (CGO-free)
)

// Open opens (or creates) a SQLite database at the given path with
// settings suitable for the Archivist worker: WAL mode, foreign keys,
// and a short busy timeout so atomic claims fail fast on contention.
func Open(path string) (*sql.DB, error) {
	dsn := buildDSN(path)

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

func buildDSN(path string) string {
	values := url.Values{}
	values.Add("_pragma", "busy_timeout=5000")
	values.Add("_pragma", "foreign_keys=ON")
	values.Add("_pragma", "journal_mode=WAL")
	values.Add("_pragma", "synchronous=NORMAL")
	values.Add("_pragma", "cache_size=-40000")
	values.Add("_pragma", "temp_store=MEMORY")
	values.Add("_pragma", "mmap_size=268435456")

	separator := "?"
	if strings.Contains(path, "?") {
		separator = "&"
	}

	return path + separator + values.Encode()
}
