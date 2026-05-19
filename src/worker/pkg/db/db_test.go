package db

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAppliesSQLitePragmas(t *testing.T) {
	database, err := Open(filepath.Join(t.TempDir(), "archive.db"))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, database.Close())
	})

	var foreignKeys int
	require.NoError(t, database.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys))
	assert.Equal(t, 1, foreignKeys)

	var busyTimeout int
	require.NoError(t, database.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout))
	assert.Equal(t, 5000, busyTimeout)

	var synchronous int
	require.NoError(t, database.QueryRow("PRAGMA synchronous").Scan(&synchronous))
	assert.Equal(t, 1, synchronous)

	var cacheSize int
	require.NoError(t, database.QueryRow("PRAGMA cache_size").Scan(&cacheSize))
	assert.Equal(t, -40000, cacheSize)

	var tempStore int
	require.NoError(t, database.QueryRow("PRAGMA temp_store").Scan(&tempStore))
	assert.Equal(t, 2, tempStore)

	var mmapSize int
	require.NoError(t, database.QueryRow("PRAGMA mmap_size").Scan(&mmapSize))
	assert.Equal(t, 268435456, mmapSize)

	var journalMode string
	require.NoError(t, database.QueryRow("PRAGMA journal_mode").Scan(&journalMode))
	assert.Equal(t, "wal", journalMode)
}
