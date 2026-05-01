package app

import (
	"log/slog"

	"codeberg.org/federico-paolillo/archivist/pkg/app/config"
)

// App is the composition root. Add your application's dependencies as fields.
type App struct {
	Logger   *slog.Logger
	LogLevel *slog.LevelVar
	Config   *config.Root
}

func NewApp(logger *slog.Logger, logLevel *slog.LevelVar, cfg *config.Root) (*App, error) {
	app := &App{
		Logger:   logger,
		LogLevel: logLevel,
		Config:   cfg,
	}

	return app, nil
}

// Close releases any resources held by the App (os.Root handles, etc.).
func (a *App) Close() error {
	return nil
}
