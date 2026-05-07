package config

// Root is the top-level configuration structure.
type Root struct {
	App         App
	Persistence Persistence
	Artifacts   Artifacts
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

func Default() *Root {
	return &Root{
		App: App{
			Name: "archivist-worker",
		},
		Artifacts: Artifacts{
			DataDir: "/data",
		},
		Debug: true,
	}
}
