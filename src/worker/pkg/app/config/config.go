package config

// Root is the top-level configuration structure.
type Root struct {
	App         App
	Persistence Persistence
	Artifacts   Artifacts
	LLM         LLM
	Debug       bool
}

type App struct {
	Name    string
	Version string
}

type Persistence struct {
	SQLitePath string `config:"SQLITE_PATH"`
}

type Artifacts struct {
	DataDir string `config:"DATA_DIR"`
}

// LLM holds configuration for the summarizer LLM provider.
// APIKey must not be logged.
type LLM struct {
	Provider string
	Model    string
	APIKey   string
}

func Default() *Root {
	return &Root{
		App: App{
			Name: "archivist-worker",
		},
		Artifacts: Artifacts{
			DataDir: "/data",
		},
		LLM: LLM{
			Provider: "anthropic",
			Model:    "claude-3-5-haiku-20241022",
		},
		Debug: true,
	}
}
