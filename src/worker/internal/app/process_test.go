package app

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"testing"
	"time"

	"codeberg.org/federico-paolillo/archivist/internal/fetcher"
	"codeberg.org/federico-paolillo/archivist/internal/pipeline"
	"codeberg.org/federico-paolillo/archivist/internal/ssrf"
	pkgapp "codeberg.org/federico-paolillo/archivist/pkg/app"
	"codeberg.org/federico-paolillo/archivist/pkg/app/config"
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	processTestUserID    = "01ASB2XFCZJY7WHZ2FNRTMQJCT"
	processTestArticleID = "01ASB2XFCZJY7WHZ2FNRTMQJP1"
	processTestJobID     = "01ASB2XFCZJY7WHZ2FNRTMQJP2"
)

func TestProcessCommandOnceProcessesQueuedJob(t *testing.T) {
	application, cfg := newProcessTestApp(t)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<!doctype html>
<html>
<head><title>Readable Test Article</title></head>
<body>
<article>
<h1>Readable Test Article</h1>
<p>This article has enough readable text for the local extractor to accept it.</p>
<p>Archivist should fetch it through the process command and persist a snapshot.</p>
</article>
</body>
</html>`))
	}))
	defer srv.Close()
	installProcessTestFetcher(t, application, srv.Listener.Addr().String())

	seedProcessUser(t, application.DB)
	seedProcessArticle(t, application.DB, "https://article.example/article")
	seedProcessJob(t, application.DB)

	withArgs(t, "archivist-worker", "process", "--once")

	err := CliProgram(t.Context(), application, cfg)
	require.NoError(t, err)

	snapshot, openErr := application.ArtifactStore.OpenSnapshot(processTestArticleID)
	require.NoError(t, openErr)
	defer snapshot.Close()

	_, err = io.ReadAll(snapshot)
	require.NoError(t, err)

	assert.NotEmpty(
		t,
		scalarNullableString(t, application.DB, `SELECT canonical_url FROM articles WHERE id = ?`, processTestArticleID),
	)
	assert.NotEqual(
		t,
		"queued",
		scalarString(t, application.DB, `SELECT status FROM jobs WHERE id = ?`, processTestJobID),
	)
}

func installProcessTestFetcher(t *testing.T, application *pkgapp.App, serverAddress string) {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	resolver := processTestResolver{ips: map[string][]netip.Addr{
		"article.example": {netip.MustParseAddr("93.184.216.34")},
	}}
	dialer := processTestDialer{targetAddress: serverAddress}
	guard := ssrf.New(logger, ssrf.WithResolver(resolver), ssrf.WithDialer(dialer))
	client := req.NewClient().
		EnableInsecureSkipVerify().
		OnBeforeRequest(guard.RequestMiddleware()).
		SetRedirectPolicy(guard.RedirectPolicy()).
		SetDial(guard.DialContext).
		SetTimeout(20 * time.Second).
		DisableForceHttpVersion().
		DisableHTTP3()

	application.HTTPClient = client
	application.SSRFGuard = guard
	application.Fetcher = fetcher.New(client, func(rawURL string) error {
		_, validateErr := guard.ValidateURL(rawURL, ssrf.PhaseInitialURL)
		return validateErr
	})
	application.SnapshotPipeline = pipeline.NewSnapshotPipeline(
		logger,
		application.Jobs,
		application.ArtifactStore,
		application.Fetcher,
		pipeline.NoOpMarkdownHandoff,
	)
}

func TestProcessLoopExitsWhenContextIsCanceledWhileIdle(t *testing.T) {
	application, _ := newProcessTestApp(t)
	seedProcessUser(t, application.DB)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	errCh := make(chan error, 1)

	go func() {
		errCh <- process(ctx, application, false, time.Hour)
	}()

	time.AfterFunc(10*time.Millisecond, cancel)

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("process loop did not exit after context cancellation")
	}
}

func newProcessTestApp(t *testing.T) (*pkgapp.App, *config.Root) {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	logLevel := new(slog.LevelVar)

	cfg := config.Default()
	cfg.SQLite.Path = filepath.Join(t.TempDir(), "archive.db")
	cfg.Data.Dir = t.TempDir()
	cfg.Jina.API.Key = "jina-secret"
	cfg.LLM.API.Key = "llm-secret"

	application, err := pkgapp.NewApp(logger, logLevel, cfg)
	require.NoError(t, err)
	require.NotNil(t, application.SnapshotPipeline)

	t.Cleanup(func() {
		require.NoError(t, application.Close())
	})

	return application, cfg
}

func withArgs(t *testing.T, args ...string) {
	t.Helper()

	original := os.Args
	os.Args = args

	t.Cleanup(func() {
		os.Args = original
	})
}

func seedProcessUser(t *testing.T, database *sql.DB) {
	t.Helper()

	_, err := database.Exec(`INSERT INTO users (id) VALUES (?)`, processTestUserID)
	require.NoError(t, err)
}

func seedProcessArticle(t *testing.T, database *sql.DB, originalURL string) {
	t.Helper()

	_, err := database.Exec(
		`INSERT INTO articles (id, user_id, original_url, status, created_at)
		 VALUES (?, ?, ?, 'queued', ?)`,
		processTestArticleID,
		processTestUserID,
		originalURL,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	require.NoError(t, err)
}

func seedProcessJob(t *testing.T, database *sql.DB) {
	t.Helper()

	_, err := database.Exec(
		`INSERT INTO jobs (id, user_id, article_id, type, status,
		                   telegram_update_id, telegram_chat_id,
		                   telegram_message_id, telegram_user_id, created_at)
		 VALUES (?, ?, ?, ?, 'queued', 2001, 3001, 4001, 5001, ?)`,
		processTestJobID,
		processTestUserID,
		processTestArticleID,
		jobs.TypeArticleProcessing,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	require.NoError(t, err)
}

type processTestResolver struct {
	ips map[string][]netip.Addr
}

func (r processTestResolver) LookupNetIP(_ context.Context, _ string, host string) ([]netip.Addr, error) {
	return r.ips[host], nil
}

type processTestDialer struct {
	targetAddress string
}

func (d processTestDialer) DialContext(ctx context.Context, network string, _ string) (net.Conn, error) {
	var dialer net.Dialer

	return dialer.DialContext(ctx, network, d.targetAddress)
}

func scalarString(t *testing.T, database *sql.DB, query string, args ...any) string {
	t.Helper()

	var value string
	require.NoError(t, database.QueryRow(query, args...).Scan(&value))

	return value
}

func scalarNullableString(t *testing.T, database *sql.DB, query string, args ...any) string {
	t.Helper()

	var value sql.NullString
	require.NoError(t, database.QueryRow(query, args...).Scan(&value))

	return value.String
}
