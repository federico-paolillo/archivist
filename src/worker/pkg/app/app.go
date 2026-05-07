package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"codeberg.org/federico-paolillo/archivist/internal/markdown"
	"codeberg.org/federico-paolillo/archivist/pkg/app/artifacts"
	"codeberg.org/federico-paolillo/archivist/pkg/app/config"
	"codeberg.org/federico-paolillo/archivist/pkg/app/persistence"
)

// App is the composition root. Add your application's dependencies as fields.
type App struct {
	Logger        *slog.Logger
	LogLevel      *slog.LevelVar
	Config        *config.Root
	DB            *sql.DB
	Jobs          *persistence.Repository
	ArtifactPaths *artifacts.ArticlePaths
	JinaExtractor *markdown.JinaExtractor
}

func NewApp(ctx context.Context, logger *slog.Logger, logLevel *slog.LevelVar, cfg *config.Root) (*App, error) {
	var db *sql.DB

	var jobs *persistence.Repository

	if cfg.Persistence.SQLitePath != "" {
		opened, err := persistence.OpenSQLite(ctx, cfg.Persistence.SQLitePath)
		if err != nil {
			return nil, fmt.Errorf("app: open persistence: %w", err)
		}

		err = persistence.EnsureSchema(ctx, opened)
		if err != nil {
			closeErr := opened.Close()
			if closeErr != nil {
				logger.Error("app: failed to close sqlite after schema initialization failure", "error", closeErr)
			}

			return nil, fmt.Errorf("app: initialize persistence schema: %w", err)
		}

		db = opened
		jobs = persistence.NewRepository(db, persistence.SystemIDGenerator{})
	}

	app := &App{
		Logger:        logger,
		LogLevel:      logLevel,
		Config:        cfg,
		DB:            db,
		Jobs:          jobs,
		ArtifactPaths: artifacts.NewArticlePaths(cfg.Artifacts.DataDir),
		JinaExtractor: markdown.NewJinaExtractor(cfg.Jina.Enabled, cfg.Jina.APIKey),
	}

	return app, nil
}

// Close releases any resources held by the App (os.Root handles, etc.).
func (a *App) Close() error {
	if a.DB == nil {
		return nil
	}

	err := a.DB.Close()
	if err != nil {
		return fmt.Errorf("app: close sqlite: %w", err)
	}

	return nil
}
