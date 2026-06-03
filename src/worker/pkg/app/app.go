package app

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"codeberg.org/federico-paolillo/archivist/internal/artifacts"
	"codeberg.org/federico-paolillo/archivist/internal/fetcher"
	"codeberg.org/federico-paolillo/archivist/internal/markdown"
	"codeberg.org/federico-paolillo/archivist/internal/observability"
	"codeberg.org/federico-paolillo/archivist/internal/pipeline"
	"codeberg.org/federico-paolillo/archivist/internal/ssrf"
	"codeberg.org/federico-paolillo/archivist/internal/summary"
	"codeberg.org/federico-paolillo/archivist/pkg/app/config"
	"codeberg.org/federico-paolillo/archivist/pkg/db"
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
	"github.com/imroc/req/v3"
)

var ErrNilConfig = errors.New("app: config is nil")

// App is the composition root. Add your application's dependencies as fields.
type App struct {
	Logger   *slog.Logger
	LogLevel *slog.LevelVar
	Config   *config.Root

	HTTPClient              *req.Client
	SSRFGuard               *ssrf.Guard
	DB                      *sql.DB
	NotificationIDGenerator jobs.IDGenerator
	Jobs                    jobs.Repository
	Enqueuer                jobs.Enqueuer
	Fetcher                 *fetcher.Fetcher
	ArtifactStore           *artifacts.Store
	LocalMarkdown           markdown.MarkdownExtractor
	JinaMarkdown            markdown.MarkdownExtractor
	Summarizer              summary.SummarizerService
	SummaryHandoff          pipeline.SummaryHandoff
	MarkdownHandoff         pipeline.MarkdownHandoff
	SnapshotPipeline        *pipeline.SnapshotPipeline
}

func NewApp(logger *slog.Logger, logLevel *slog.LevelVar, cfg *config.Root) (*App, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	validateErr := cfg.Validate()
	if validateErr != nil {
		return nil, fmt.Errorf("app: invalid config: %w", validateErr)
	}

	database, err := createDB(cfg.SQLite.Path)
	if err != nil {
		return nil, err
	}

	store, err := artifacts.NewStore(cfg.Data.Dir)
	if err != nil {
		_ = database.Close()

		return nil, fmt.Errorf("app: failed to create artifact store: %w", err)
	}

	ssrfGuard, httpClient := createHTTPClient(logger)

	application := &App{
		Logger:   logger,
		LogLevel: logLevel,
		Config:   cfg,

		HTTPClient: httpClient,
		SSRFGuard:  ssrfGuard,
		DB:         database,
	}

	notificationIDs := jobs.NewULIDGenerator()
	fetcherService := createFetcher(application)
	localMarkdown := markdown.NewGoReadabilityExtractor()
	jinaMarkdown := markdown.NewJinaExtractor(application.HTTPClient, cfg.Jina.API.Key)
	summarizer := summary.NewAnthropicAdapter(application.HTTPClient, cfg.LLM.API.Key, cfg.LLM.Model)
	jobsRepository := jobs.NewSQLiteRepository(database, notificationIDs)
	summaryHandoff := createSummaryHandoff(logger, jobsRepository, store, summarizer)
	markdownHandoff := createMarkdownHandoff(logger, jobsRepository, store, localMarkdown, jinaMarkdown, summaryHandoff)

	applyProcessingServices(
		application,
		notificationIDs,
		jobsRepository,
		fetcherService,
		store,
		localMarkdown,
		jinaMarkdown,
		summarizer,
		summaryHandoff,
		markdownHandoff,
	)

	application.SnapshotPipeline = pipeline.NewSnapshotPipeline(
		logger,
		application.Jobs,
		application.ArtifactStore,
		application.Fetcher,
		application.MarkdownHandoff,
	)

	return application, nil
}

func createSummaryHandoff(
	logger *slog.Logger,
	jobsRepository jobs.Repository,
	store *artifacts.Store,
	summarizer summary.SummarizerService,
) *pipeline.SummaryGenerationHandoff {
	return pipeline.NewSummaryGenerationHandoff(
		logger,
		jobsRepository,
		store,
		summarizer,
	)
}

func createMarkdownHandoff(
	logger *slog.Logger,
	jobsRepository jobs.Repository,
	store *artifacts.Store,
	localMarkdown markdown.MarkdownExtractor,
	jinaMarkdown markdown.MarkdownExtractor,
	summaryHandoff pipeline.SummaryHandoff,
) *pipeline.MarkdownExtractionHandoff {
	return pipeline.NewMarkdownExtractionHandoff(
		logger,
		jobsRepository,
		store,
		localMarkdown,
		jinaMarkdown,
		summaryHandoff,
	)
}

func createFetcher(application *App) *fetcher.Fetcher {
	return fetcher.New(application.HTTPClient, func(rawURL string) error {
		_, err := application.SSRFGuard.ValidateURL(rawURL, ssrf.PhaseInitialURL)
		if err != nil {
			return fmt.Errorf("app: validate article URL: %w", err)
		}

		return nil
	})
}

func applyProcessingServices(
	application *App,
	notificationIDs jobs.IDGenerator,
	jobsRepository *jobs.SQLiteRepository,
	fetcherService *fetcher.Fetcher,
	store *artifacts.Store,
	localMarkdown markdown.MarkdownExtractor,
	jinaMarkdown markdown.MarkdownExtractor,
	summarizer summary.SummarizerService,
	summaryHandoff pipeline.SummaryHandoff,
	markdownHandoff pipeline.MarkdownHandoff,
) {
	application.NotificationIDGenerator = notificationIDs
	application.Jobs = jobsRepository
	application.Enqueuer = jobsRepository
	application.Fetcher = fetcherService
	application.ArtifactStore = store
	application.LocalMarkdown = localMarkdown
	application.JinaMarkdown = jinaMarkdown
	application.Summarizer = summarizer
	application.SummaryHandoff = summaryHandoff
	application.MarkdownHandoff = markdownHandoff
}

func createHTTPClient(logger *slog.Logger) (*ssrf.Guard, *req.Client) {
	ssrfGuard := ssrf.New(logger)
	httpClient := req.NewClient().
		OnBeforeRequest(ssrfGuard.RequestMiddleware()).
		WrapRoundTripFunc(observability.ReqRoundTripWrapper()).
		SetRedirectPolicy(ssrfGuard.RedirectPolicy()).
		SetDial(ssrfGuard.DialContext).
		SetTimeout(20 * time.Second).
		DisableForceHttpVersion().
		DisableHTTP3()

	return ssrfGuard, httpClient
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
	return errors.Join(
		a.ArtifactStore.Close(),
		a.DB.Close(),
	)
}
