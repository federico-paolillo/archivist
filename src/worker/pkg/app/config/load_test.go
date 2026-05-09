package config_test

import (
	"testing"

	"codeberg.org/federico-paolillo/archivist/pkg/app/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigurationLoadsFromEnvVars(t *testing.T) {
	t.Setenv("APP_APP_NAME", "testapp")
	t.Setenv("APP_DEBUG", "true")

	cfg, err := config.Load()

	require.NoError(t, err)
	require.Equal(t, "testapp", cfg.App.Name)
	require.True(t, cfg.Debug)
}

func TestConfigurationDefaults(t *testing.T) {
	cfg, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, "archivist-worker", cfg.App.Name)
	assert.True(t, cfg.Debug)
	assert.Equal(t, "", cfg.SqlitePath)
	assert.Equal(t, "", cfg.DataDir)
}

func TestConfigurationLoadsNewFieldsFromEnvVars(t *testing.T) {
	t.Setenv("APP_DEBUG", "false")

	cfg, err := config.Load()

	require.NoError(t, err)
	require.False(t, cfg.Debug)
}

func TestConfigurationLoadsSQLitePath(t *testing.T) {
	t.Setenv("APP_SQLITEPATH", "/data/archive.db")

	cfg, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, "/data/archive.db", cfg.SqlitePath)
}

func TestConfigurationLoadsDataDir(t *testing.T) {
	t.Setenv("APP_DATADIR", "/data")

	cfg, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, "/data", cfg.DataDir)
}
