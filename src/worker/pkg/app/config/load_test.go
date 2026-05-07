package config_test

import (
	"testing"

	"codeberg.org/federico-paolillo/archivist/pkg/app/config"
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
	require.Equal(t, "archivist-worker", cfg.App.Name)
	require.Equal(t, "/data", cfg.Artifacts.DataDir)
	require.True(t, cfg.Debug)
}

func TestConfigurationLoadsNewFieldsFromEnvVars(t *testing.T) {
	t.Setenv("APP_DEBUG", "false")

	cfg, err := config.Load()

	require.NoError(t, err)
	require.False(t, cfg.Debug)
}
