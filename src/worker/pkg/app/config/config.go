package config

// Root is the top-level configuration structure.
type Root struct {
	App        App
	Debug      bool
	SqlitePath string
	DataDir    string
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
		Debug:      true,
		SqlitePath: "",
		DataDir:    "",
	}
}
