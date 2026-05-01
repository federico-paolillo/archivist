package app_test

import (
	"log/slog"
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

	application, err := app.NewApp(logger, logLevel, cfg)

	require.NoError(t, err)

	require.NotNil(t, application)

	require.Equal(t, logger, application.Logger)
	require.Equal(t, logLevel, application.LogLevel)
	require.Equal(t, cfg, application.Config)
}
