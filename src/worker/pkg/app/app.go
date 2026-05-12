package app

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"codeberg.org/federico-paolillo/archivist/internal/artifacts"
	"codeberg.org/federico-paolillo/archivist/internal/fetcher"
	"codeberg.org/federico-paolillo/archivist/internal/markdown"
	"codeberg.org/federico-paolillo/archivist/internal/pipeline"
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

	DB               *sql.DB
	Jobs             jobs.Repository
	Fetcher          *fetcher.Fetcher
	ArtifactStore    *artifacts.Store
	LocalMarkdown    markdown.MarkdownExtractor
	JinaMarkdown     markdown.MarkdownExtractor
	SnapshotPipeline *pipeline.SnapshotPipeline
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
	application.LocalMarkdown = markdown.NewGoReadabilityExtractor()
	application.JinaMarkdown = markdown.NewJinaExtractor(httpClient, cfg.JinaEnabled, cfg.JinaAPIKey)

	if cfg.SqlitePath != "" {
		database, err := createDB(cfg.SqlitePath)
		if err != nil {
			return nil, err
		}

		application.DB = database
		application.Jobs = jobs.NewSQLiteRepository(database)
	}

	if cfg.DataDir != "" {
		store, err := artifacts.NewStore(cfg.DataDir)
		if err != nil {
			return nil, fmt.Errorf("app: failed to create artifact store: %w", err)
		}

		application.ArtifactStore = store
	}

	if application.Jobs != nil && application.ArtifactStore != nil {
		markdownHandoff := pipeline.NewMarkdownExtractionHandoff(
			logger,
			application.ArtifactStore,
			application.LocalMarkdown,
			application.JinaMarkdown,
		)

		application.SnapshotPipeline = pipeline.NewSnapshotPipeline(
			logger,
			application.Jobs,
			application.ArtifactStore,
			application.Fetcher,
			markdownHandoff,
		)
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
	var closeErr error

	if a.ArtifactStore != nil {
		err := a.ArtifactStore.Close()
		if err != nil {
			closeErr = fmt.Errorf("app: failed to close artifact store: %w", err)
		}
	}

	if a.DB != nil {
		err := a.DB.Close()
		if err != nil {
			closeErr = fmt.Errorf("app: failed to close database: %w", err)
		}
	}

	return closeErr
}
