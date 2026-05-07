package app_test

import (
	"log/slog"
	"path/filepath"
	"testing"

	"codeberg.org/federico-paolillo/archivist/internal/testutils/slogt"
	"codeberg.org/federico-paolillo/archivist/pkg/app"
	"codeberg.org/federico-paolillo/archivist/pkg/app/config"
	"github.com/stretchr/testify/require"
)

func TestNewAppReturnsApp(t *testing.T) {
	logger := slogt.New(t)

	logLevel := new(slog.LevelVar)

	cfg := config.Default()

	application, err := app.NewApp(t.Context(), logger, logLevel, cfg)

	require.NoError(t, err)

	require.NotNil(t, application)

	require.Equal(t, logger, application.Logger)
	require.Equal(t, logLevel, application.LogLevel)
	require.Equal(t, cfg, application.Config)
	require.NotNil(t, application.ArtifactPaths)
	require.Nil(t, application.DB)
	require.Nil(t, application.Jobs)
}

func TestNewAppCreatesPersistenceWhenSQLitePathIsConfigured(t *testing.T) {
	logger := slogt.New(t)
	logLevel := new(slog.LevelVar)
	cfg := config.Default()
	cfg.Persistence.SQLitePath = filepath.Join(t.TempDir(), "archive.db")

	application, err := app.NewApp(t.Context(), logger, logLevel, cfg)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, application.Close())
	})

	require.NotNil(t, application.DB)
	require.NotNil(t, application.Jobs)
}
