package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/sherifabdlnaby/configuro"
)

func Load() (*Root, error) {
	cfguro, err := configuro.NewConfig(
		configuro.WithLoadFromEnvVars("APP"),
		configuro.WithLoadFromConfigFile("config.yml", false),
		configuro.WithoutEnvConfigPathOverload(),
		configuro.WithoutLoadDotEnv(),
	)
	if err != nil {
		//nolint:errorlint // We do not want to wrap and leak errors that are not under our control
		return nil, fmt.Errorf(
			"config: failed to setup configuro. %v",
			err,
		)
	}

	cfg := Default()

	err = cfguro.Load(cfg)
	if err != nil {
		//nolint:errorlint // We do not want to wrap and leak errors that are not under our control
		return nil, fmt.Errorf(
			"config: failed to bind configuration. %v",
			err,
		)
	}

	err = applyJinaEnvOverrides(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func applyJinaEnvOverrides(cfg *Root) error {
	if enabled, ok := os.LookupEnv("APP_JINA_ENABLED"); ok {
		parsed, err := strconv.ParseBool(enabled)
		if err != nil {
			return fmt.Errorf("config: failed to parse APP_JINA_ENABLED: %w", err)
		}

		cfg.JinaEnabled = parsed
	}

	if apiKey, ok := os.LookupEnv("APP_JINA_API_KEY"); ok {
		cfg.JinaAPIKey = apiKey
	}

	return nil
}
