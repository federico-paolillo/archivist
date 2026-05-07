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

func TestConfigurationJinaDefaultsToDisabled(t *testing.T) {
	cfg, err := config.Load()

	require.NoError(t, err)
	require.False(t, cfg.Jina.Enabled)
	require.Empty(t, cfg.Jina.APIKey)
}

func TestConfigurationLoadsJinaEnabledFromEnvVar(t *testing.T) {
	t.Setenv("APP_JINA_JINA__ENABLED", "true")

	cfg, err := config.Load()

	require.NoError(t, err)
	require.True(t, cfg.Jina.Enabled)
}

func TestConfigurationLoadsJinaAPIKeyFromEnvVar(t *testing.T) {
	t.Setenv("APP_JINA_JINA__API__KEY", "test-api-key")

	cfg, err := config.Load()

	require.NoError(t, err)
	require.Equal(t, "test-api-key", cfg.Jina.APIKey)
}

func TestLLMProviderDefault(t *testing.T) {
	cfg, err := config.Load()

	require.NoError(t, err)
	require.Equal(t, "anthropic", cfg.LLM.Provider)
}

func TestLLMModelDefault(t *testing.T) {
	cfg, err := config.Load()

	require.NoError(t, err)
	require.Equal(t, "claude-3-5-haiku-20241022", cfg.LLM.Model)
}

func TestLLMAPIKeyDefaultIsEmpty(t *testing.T) {
	cfg, err := config.Load()

	require.NoError(t, err)
	require.Empty(t, cfg.LLM.APIKey)
}

func TestLLMConfigurationLoadsFromEnvVars(t *testing.T) {
	t.Setenv("APP_LLM_PROVIDER", "anthropic")
	t.Setenv("APP_LLM_MODEL", "claude-3-opus-20240229")
	t.Setenv("APP_LLM_APIKEY", "sk-test-key")

	cfg, err := config.Load()

	require.NoError(t, err)
	require.Equal(t, "anthropic", cfg.LLM.Provider)
	require.Equal(t, "claude-3-opus-20240229", cfg.LLM.Model)
	require.Equal(t, "sk-test-key", cfg.LLM.APIKey)
}
