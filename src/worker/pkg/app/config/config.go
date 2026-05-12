package config

// Root is the top-level configuration structure.
type Root struct {
	App         App
	Debug       bool
	SqlitePath  string
	DataDir     string
	JinaEnabled bool   `config:"JINA_ENABLED"`
	JinaAPIKey  string `config:"JINA_API_KEY"`
}

type App struct {
	Name    string
	Version string
}

func Default() *Root {
	return &Root{
		App: App{
			Name: "archivist-worker",
		},
		Debug:       true,
		SqlitePath:  "",
		DataDir:     "",
		JinaEnabled: false,
		JinaAPIKey:  "",
	}
}
