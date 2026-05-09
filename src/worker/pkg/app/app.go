package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"codeberg.org/federico-paolillo/archivist/internal/artifacts"
	"codeberg.org/federico-paolillo/archivist/internal/markdown"
	"codeberg.org/federico-paolillo/archivist/internal/summary"
	"codeberg.org/federico-paolillo/archivist/pkg/app/config"
	"codeberg.org/federico-paolillo/archivist/pkg/app/persistence"
	"github.com/imroc/req/v3"
)

// App is the composition root. Add your application's dependencies as fields.
type App struct {
	Logger        *slog.Logger
	LogLevel      *slog.LevelVar
	Config        *config.Root
	DB            *sql.DB
	Jobs          *persistence.Repository
	Artifacts     *artifacts.Store
	HTTP          *req.Client
	JinaExtractor *markdown.JinaExtractor
	Summarizer    summary.SummarizerService
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

	httpClient := req.NewClient().
		SetUserAgent("archivist-worker/0.1").
		SetTimeout(30 * time.Second)

	summarizer, err := createSummarizer(httpClient, cfg)
	if err != nil {
		return nil, fmt.Errorf("app: create summarizer: %w", err)
	}

	store, err := artifacts.NewStore(cfg.Artifacts.DataDir)
	if err != nil {
		return nil, fmt.Errorf("app: open artifact store: %w", err)
	}

	app := &App{
		Logger:        logger,
		LogLevel:      logLevel,
		Config:        cfg,
		DB:            db,
		Jobs:          jobs,
		Artifacts:     store,
		HTTP:          httpClient,
		JinaExtractor: markdown.NewJinaExtractor(httpClient, cfg.Jina.Enabled, cfg.Jina.APIKey),
		Summarizer:    summarizer,
	}

	return app, nil
}

func createSummarizer(httpClient *req.Client, cfg *config.Root) (*summary.AnthropicAdapter, error) {
	if cfg.LLM.Provider != "anthropic" {
		return nil, fmt.Errorf("unsupported LLM provider: %q", cfg.LLM.Provider)
	}

	return summary.NewAnthropicAdapter(httpClient, cfg.LLM.APIKey, cfg.LLM.Model), nil
}

// Close releases any resources held by the App (os.Root handles, etc.).
func (a *App) Close() error {
	if a.DB != nil {
		err := a.DB.Close()
		if err != nil {
			return fmt.Errorf("app: close sqlite: %w", err)
		}
	}

	if a.Artifacts != nil {
		err := a.Artifacts.Close()
		if err != nil {
			return fmt.Errorf("app: close artifact store: %w", err)
		}
	}

	return nil
}
