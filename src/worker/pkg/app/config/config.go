package config

import (
	"errors"
	"fmt"
	"strings"
)

const (
	ProviderAnthropic = "anthropic"

	DefaultLLMModel = "claude-3-5-haiku-20241022"
)

// Root is the top-level configuration structure.
type Root struct {
	App    App    `config:"APP"`
	Debug  bool   `config:"DEBUG"`
	SQLite SQLite `config:"SQLITE"`
	Data   Data   `config:"DATA"`
	Jina   Jina   `config:"JINA"`
	LLM    LLM    `config:"LLM"`
}

type App struct {
	Name    string
	Version string
}

type SQLite struct {
	Path string `config:"PATH"`
}

type Data struct {
	Dir string `config:"DIR"`
}

type API struct {
	Key string `config:"KEY"`
}

type Jina struct {
	Enabled bool `config:"ENABLED"`
	API     API  `config:"API"`
}

type LLM struct {
	Provider string `config:"PROVIDER"`
	API      API    `config:"API"`
	Model    string `config:"MODEL"`
}

func Default() *Root {
	return &Root{
		App: App{
			Name: "archivist-worker",
		},
		Debug: true,
		Jina: Jina{
			Enabled: false,
		},
		LLM: LLM{
			Provider: ProviderAnthropic,
			Model:    DefaultLLMModel,
		},
	}
}

func (r Root) Validate() error {
	var problems []string

	if strings.TrimSpace(r.SQLite.Path) == "" {
		problems = append(problems, "SQLITE_PATH is required")
	}

	if strings.TrimSpace(r.Data.Dir) == "" {
		problems = append(problems, "DATA_DIR is required")
	}

	switch r.LLM.Provider {
	case ProviderAnthropic:
		if strings.TrimSpace(r.LLM.API.Key) == "" {
			problems = append(problems, "LLM_API_KEY is required when LLM_PROVIDER=anthropic")
		}
	default:
		problems = append(problems, fmt.Sprintf("LLM_PROVIDER %q is not supported", r.LLM.Provider))
	}

	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}

	return nil
}
