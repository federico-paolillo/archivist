package app

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"codeberg.org/federico-paolillo/archivist/internal/fetcher"
	"codeberg.org/federico-paolillo/archivist/pkg/app/config"
	"codeberg.org/federico-paolillo/archivist/pkg/db"
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
	"github.com/imroc/req/v3"
)

// App is the composition root. Add your application's dependencies as fields.
type App struct {
	Logger   *slog.Logger
	LogLevel *slog.LevelVar
	Config   *config.Root

	DB      *sql.DB
	Jobs    jobs.Repository
	Fetcher *fetcher.Fetcher
}

func NewApp(logger *slog.Logger, logLevel *slog.LevelVar, cfg *config.Root) (*App, error) {
	application := &App{
		Logger:   logger,
		LogLevel: logLevel,
		Config:   cfg,
	}

	httpClient := req.NewClient().
		SetRedirectPolicy(req.MaxRedirectPolicy(10)).
		SetTimeout(20 * time.Second).
		DisableForceHttpVersion()

	application.Fetcher = fetcher.New(httpClient)

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
