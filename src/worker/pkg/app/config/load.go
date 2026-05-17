package config

import (
	"fmt"

	"github.com/sherifabdlnaby/configuro"
)

func Load() (*Root, error) {
	cfguro, err := configuro.NewConfig(
		configuro.WithLoadFromEnvVars("ARCHIVIST"),
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

	err = cfguro.Validate(cfg)
	if err != nil {
		//nolint:errorlint // We do not want to wrap and leak errors that are not under our control
		return nil, fmt.Errorf(
			"config: failed to validate configuration. %v",
			err,
		)
	}

	return cfg, nil
}
