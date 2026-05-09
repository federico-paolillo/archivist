package app

import (
	"database/sql"
	"fmt"
	"log/slog"

	"codeberg.org/federico-paolillo/archivist/pkg/app/config"
	"codeberg.org/federico-paolillo/archivist/pkg/db"
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
)

// App is the composition root. Add your application's dependencies as fields.
type App struct {
	Logger   *slog.Logger
	LogLevel *slog.LevelVar
	Config   *config.Root

	DB   *sql.DB
	Jobs jobs.Repository
}

func NewApp(logger *slog.Logger, logLevel *slog.LevelVar, cfg *config.Root) (*App, error) {
	application := &App{
		Logger:   logger,
		LogLevel: logLevel,
		Config:   cfg,
	}

	if cfg.SqlitePath != "" {
		database, err := createDB(cfg.SqlitePath)
		if err != nil {
			return nil, err
		}

		application.DB = database
		application.Jobs = jobs.NewSQLiteRepository(database)
	}

	return application, nil
}

func createDB(path string) (*sql.DB, error) {
	database, err := db.Open(path)
	if err != nil {
		return nil, fmt.Errorf("app: failed to open database: %w", err)
	}

	schemaErr := db.ApplySchema(database)
	if schemaErr != nil {
		_ = database.Close()

		return nil, fmt.Errorf("app: failed to apply database schema: %w", schemaErr)
	}

	return database, nil
}

// Close releases any resources held by the App (database connections, os.Root handles, etc.).
func (a *App) Close() error {
	if a.DB == nil {
		return nil
	}

	closeErr := a.DB.Close()
	if closeErr != nil {
		return fmt.Errorf("app: failed to close database: %w", closeErr)
	}

	return nil
}
