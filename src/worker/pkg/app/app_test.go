package app_test

import (
	"log/slog"
	"testing"

	"codeberg.org/federico-paolillo/archivist/internal/testutils/slogt"
	"codeberg.org/federico-paolillo/archivist/pkg/app"
	"codeberg.org/federico-paolillo/archivist/pkg/app/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAppReturnsApp(t *testing.T) {
	logger := slogt.New(t)

	logLevel := new(slog.LevelVar)

	cfg := config.Default()

	application, err := app.NewApp(logger, logLevel, cfg)

	require.NoError(t, err)

	require.NotNil(t, application)

	require.Equal(t, logger, application.Logger)
	require.Equal(t, logLevel, application.LogLevel)
	require.Equal(t, cfg, application.Config)

	// Without SqlitePath, DB and Jobs are nil (no database configured).
	assert.Nil(t, application.DB)
	assert.Nil(t, application.Jobs)

	require.NotNil(t, application.Fetcher)

	// Without DataDir, ArtifactStore is nil.
	assert.Nil(t, application.ArtifactStore)

	// Without both SqlitePath and DataDir, SnapshotPipeline is nil.
	assert.Nil(t, application.SnapshotPipeline)
}

func TestNewAppWithSQLitePathOpensDatabase(t *testing.T) {
	logger := slogt.New(t)

	logLevel := new(slog.LevelVar)

	cfg := config.Default()
	cfg.SqlitePath = ":memory:"

	application, err := app.NewApp(logger, logLevel, cfg)

	require.NoError(t, err)
	require.NotNil(t, application)

	t.Cleanup(func() {
		application.Close()
	})

	assert.NotNil(t, application.DB)
	assert.NotNil(t, application.Jobs)

	// DataDir not configured, so ArtifactStore and SnapshotPipeline are nil.
	assert.Nil(t, application.ArtifactStore)
	assert.Nil(t, application.SnapshotPipeline)
}

func TestNewAppWithSQLiteAndDataDirWiresSnapshotPipeline(t *testing.T) {
	logger := slogt.New(t)

	logLevel := new(slog.LevelVar)

	cfg := config.Default()
	cfg.SqlitePath = ":memory:"
	cfg.DataDir = t.TempDir()

	application, err := app.NewApp(logger, logLevel, cfg)

	require.NoError(t, err)
	require.NotNil(t, application)

	t.Cleanup(func() {
		application.Close()
	})

	assert.NotNil(t, application.DB)
	assert.NotNil(t, application.Jobs)
	assert.NotNil(t, application.ArtifactStore)
	assert.NotNil(t, application.SnapshotPipeline)
}
