package app_test

import (
	"log/slog"
	"testing"

	"codeberg.org/federico-paolillo/archivist/internal/summary"
	"codeberg.org/federico-paolillo/archivist/internal/testutils/slogt"
	"codeberg.org/federico-paolillo/archivist/pkg/app"
	"codeberg.org/federico-paolillo/archivist/pkg/app/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAppReturnsApp(t *testing.T) {
	logger := slogt.New(t)

	logLevel := new(slog.LevelVar)

	cfg := newValidConfig(t)

	application, err := app.NewApp(logger, logLevel, cfg)
	require.NoError(t, err)
	require.NotNil(t, application)

	t.Cleanup(func() {
		require.NoError(t, application.Close())
	})

	require.Equal(t, logger, application.Logger)
	require.Equal(t, logLevel, application.LogLevel)
	require.Equal(t, cfg, application.Config)

	require.NotNil(t, application.Fetcher)
	require.NotNil(t, application.LocalMarkdown)
	require.NotNil(t, application.JinaMarkdown)
	assert.NotNil(t, application.DB)
	assert.NotNil(t, application.Jobs)
	assert.NotNil(t, application.ArtifactStore)
	assert.NotNil(t, application.SnapshotPipeline)
	require.NotNil(t, application.Summarizer)
	assert.Equal(t, summary.ProviderAnthropic, application.Summarizer.Provider())
}

func TestNewAppRejectsNilConfig(t *testing.T) {
	logger := slogt.New(t)
	logLevel := new(slog.LevelVar)

	application, err := app.NewApp(logger, logLevel, nil)

	require.ErrorIs(t, err, app.ErrNilConfig)
	assert.Nil(t, application)
}

func TestNewAppRejectsMissingSQLitePath(t *testing.T) {
	cfg := newValidConfig(t)
	cfg.SQLite.Path = ""

	application, err := app.NewApp(slogt.New(t), new(slog.LevelVar), cfg)

	require.Error(t, err)
	require.ErrorContains(t, err, "SQLITE_PATH is required")
	assert.Nil(t, application)
}

func TestNewAppRejectsMissingDataDir(t *testing.T) {
	cfg := newValidConfig(t)
	cfg.Data.Dir = ""

	application, err := app.NewApp(slogt.New(t), new(slog.LevelVar), cfg)

	require.Error(t, err)
	require.ErrorContains(t, err, "DATA_DIR is required")
	assert.Nil(t, application)
}

func TestNewAppRejectsMissingAnthropicAPIKey(t *testing.T) {
	cfg := newValidConfig(t)
	cfg.LLM.API.Key = ""

	application, err := app.NewApp(slogt.New(t), new(slog.LevelVar), cfg)

	require.Error(t, err)
	require.ErrorContains(t, err, "LLM_API_KEY is required when LLM_PROVIDER=anthropic")
	assert.Nil(t, application)
}

func TestNewAppRejectsMissingJinaAPIKey(t *testing.T) {
	cfg := newValidConfig(t)
	cfg.Jina.API.Key = ""

	application, err := app.NewApp(slogt.New(t), new(slog.LevelVar), cfg)

	require.Error(t, err)
	require.ErrorContains(t, err, "JINA_API_KEY is required")
	assert.Nil(t, application)
}

func TestNewAppRejectsUnsupportedLLMProvider(t *testing.T) {
	cfg := newValidConfig(t)
	cfg.LLM.Provider = "unsupported"

	application, err := app.NewApp(slogt.New(t), new(slog.LevelVar), cfg)

	require.Error(t, err)
	require.ErrorContains(t, err, `LLM_PROVIDER "unsupported" is not supported`)
	assert.Nil(t, application)
}

func newValidConfig(t *testing.T) *config.Root {
	t.Helper()

	cfg := config.Default()
	cfg.SQLite.Path = ":memory:"
	cfg.Data.Dir = t.TempDir()
	cfg.Jina.API.Key = "jina-secret"
	cfg.LLM.API.Key = "llm-secret"

	return cfg
}
