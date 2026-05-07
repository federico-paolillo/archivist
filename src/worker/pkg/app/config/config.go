package config

// Root is the top-level configuration structure.
type Root struct {
	App         App
	Persistence Persistence
	Artifacts   Artifacts
	Jina        Jina
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

// Jina holds configuration for the Jina Reader fallback extractor.
// APIKey is optional and must never be logged.
type Jina struct {
	Enabled bool   `config:"JINA_ENABLED"`
	APIKey  string `config:"JINA_API_KEY"`
}

func Default() *Root {
	return &Root{
		App: App{
			Name: "archivist-worker",
		},
		Artifacts: Artifacts{
			DataDir: "/data",
		},
		Jina: Jina{
			Enabled: false,
		},
		Debug: true,
	}
}
