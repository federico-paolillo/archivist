package config_test

import (
	"testing"

	"codeberg.org/federico-paolillo/archivist/pkg/app/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigurationLoadsFromEnvVars(t *testing.T) {
	setRequiredWorkerConfig(t)
	t.Setenv("ARCHIVIST_DEBUG", "true")

	cfg, err := config.Load()

	require.NoError(t, err)
	require.Equal(t, "archivist-worker", cfg.App.Name)
	require.True(t, cfg.Debug)
}

func TestConfigurationDefaults(t *testing.T) {
	cfg := config.Default()

	assert.Equal(t, "archivist-worker", cfg.App.Name)
	assert.True(t, cfg.Debug)
	assert.Equal(t, "", cfg.SQLite.Path)
	assert.Equal(t, "", cfg.Data.Dir)
	assert.False(t, cfg.Jina.Enabled)
	assert.Equal(t, "", cfg.Jina.API.Key)
	assert.Equal(t, config.ProviderAnthropic, cfg.LLM.Provider)
	assert.Equal(t, config.DefaultLLMModel, cfg.LLM.Model)
	assert.Equal(t, "", cfg.LLM.API.Key)
}

func TestConfigurationLoadsNewFieldsFromEnvVars(t *testing.T) {
	setRequiredWorkerConfig(t)
	t.Setenv("ARCHIVIST_DEBUG", "false")

	cfg, err := config.Load()

	require.NoError(t, err)
	require.False(t, cfg.Debug)
}

func TestConfigurationLoadsSQLitePath(t *testing.T) {
	setRequiredWorkerConfig(t)
	t.Setenv("ARCHIVIST_SQLITE_PATH", "/data/archive.db")

	cfg, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, "/data/archive.db", cfg.SQLite.Path)
}

func TestConfigurationLoadsDataDir(t *testing.T) {
	setRequiredWorkerConfig(t)
	t.Setenv("ARCHIVIST_DATA_DIR", "/data")

	cfg, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, "/data", cfg.Data.Dir)
}

func TestConfigurationLoadsJina(t *testing.T) {
	setRequiredWorkerConfig(t)
	t.Setenv("ARCHIVIST_JINA_ENABLED", "true")
	t.Setenv("ARCHIVIST_JINA_API_KEY", "jina-secret")

	cfg, err := config.Load()

	require.NoError(t, err)
	assert.True(t, cfg.Jina.Enabled)
	assert.Equal(t, "jina-secret", cfg.Jina.API.Key)
}

func TestConfigurationLoadsLLM(t *testing.T) {
	setRequiredWorkerConfig(t)
	t.Setenv("ARCHIVIST_LLM_PROVIDER", config.ProviderAnthropic)
	t.Setenv("ARCHIVIST_LLM_API_KEY", "llm-secret")
	t.Setenv("ARCHIVIST_LLM_MODEL", "claude-test-model")

	cfg, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, config.ProviderAnthropic, cfg.LLM.Provider)
	assert.Equal(t, "llm-secret", cfg.LLM.API.Key)
	assert.Equal(t, "claude-test-model", cfg.LLM.Model)
}

func TestConfigurationRequiresSQLitePath(t *testing.T) {
	clearWorkerConfigEnv(t)
	t.Setenv("ARCHIVIST_DATA_DIR", "/data")
	t.Setenv("ARCHIVIST_LLM_API_KEY", "llm-secret")

	_, err := config.Load()

	require.Error(t, err)
	require.ErrorContains(t, err, "SQLITE_PATH is required")
}

func TestConfigurationRequiresDataDir(t *testing.T) {
	clearWorkerConfigEnv(t)
	t.Setenv("ARCHIVIST_SQLITE_PATH", "/data/archive.db")
	t.Setenv("ARCHIVIST_LLM_API_KEY", "llm-secret")

	_, err := config.Load()

	require.Error(t, err)
	require.ErrorContains(t, err, "DATA_DIR is required")
}

func TestConfigurationRequiresLLMAPIKeyForAnthropic(t *testing.T) {
	clearWorkerConfigEnv(t)
	t.Setenv("ARCHIVIST_SQLITE_PATH", "/data/archive.db")
	t.Setenv("ARCHIVIST_DATA_DIR", "/data")

	_, err := config.Load()

	require.Error(t, err)
	require.ErrorContains(t, err, "LLM_API_KEY is required when LLM_PROVIDER=anthropic")
}

func TestConfigurationRejectsUnsupportedLLMProvider(t *testing.T) {
	setRequiredWorkerConfig(t)
	t.Setenv("ARCHIVIST_LLM_PROVIDER", "unsupported")

	_, err := config.Load()

	require.Error(t, err)
	require.ErrorContains(t, err, `LLM_PROVIDER "unsupported" is not supported`)
}

func setRequiredWorkerConfig(t *testing.T) {
	t.Helper()

	clearWorkerConfigEnv(t)

	t.Setenv("ARCHIVIST_SQLITE_PATH", "/data/archive.db")
	t.Setenv("ARCHIVIST_DATA_DIR", "/data")
	t.Setenv("ARCHIVIST_LLM_API_KEY", "llm-secret")
}

func clearWorkerConfigEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"ARCHIVIST_" + "APP" + "_NAME",
		"ARCHIVIST_DEBUG",
		"ARCHIVIST_SQLITE_PATH",
		"ARCHIVIST_DATA_DIR",
		"ARCHIVIST_JINA_ENABLED",
		"ARCHIVIST_JINA_API_KEY",
		"ARCHIVIST_LLM_PROVIDER",
		"ARCHIVIST_LLM_API_KEY",
		"ARCHIVIST_LLM_MODEL",
	} {
		t.Setenv(key, "")
	}
}
