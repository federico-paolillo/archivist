package pipeline_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
	"codeberg.org/federico-paolillo/archivist/internal/artifacts"
	"codeberg.org/federico-paolillo/archivist/internal/fetcher"
	"codeberg.org/federico-paolillo/archivist/internal/pipeline"
	"codeberg.org/federico-paolillo/archivist/internal/ssrf"
	"codeberg.org/federico-paolillo/archivist/internal/testutils/slogt"
	"codeberg.org/federico-paolillo/archivist/pkg/db"
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

const (
	personalUserID = "01ASB2XFCZJY7WHZ2FNRTMQJCT"
	otherUserID    = "01ASB2XFCZJY7WHZ2FNRTMQJCX"
	articleID      = "01ASB2XFCZJY7WHZ2FNRTMQJA1"
	jobID          = "01ASB2XFCZJY7WHZ2FNRTMQJB1"
)

// openTestDB opens a temporary in-memory SQLite database with the Archivist schema.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	database, err := db.Open(":memory:")
	require.NoError(t, err)

	t.Cleanup(func() {
		database.Close()
	})

	err = db.ApplySchema(database)
	require.NoError(t, err)

	return database
}

// seedUser inserts the personal Archivist user required by foreign keys.
func seedUser(t *testing.T, database *sql.DB) {
	t.Helper()

	_, err := database.Exec(`INSERT INTO users (id) VALUES (?)`, personalUserID)
	require.NoError(t, err)
}

func seedOtherUser(t *testing.T, database *sql.DB) {
	t.Helper()

	_, err := database.Exec(`INSERT INTO users (id) VALUES (?)`, otherUserID)
	require.NoError(t, err)
}

// seedArticle inserts a queued article row with the given original URL.
func seedArticle(t *testing.T, database *sql.DB, id string, originalURL string) {
	t.Helper()

	_, err := database.Exec(
		`INSERT INTO articles (id, user_id, original_url, status, created_at)
		 VALUES (?, ?, ?, 'queued', ?)`,
		id, personalUserID, originalURL, time.Now().UTC().Format(time.RFC3339Nano),
	)
	require.NoError(t, err)
}

// seedTelegramJob inserts a queued Telegram-originated job.
func seedTelegramJob(t *testing.T, database *sql.DB, jID, aID string) {
	t.Helper()

	_, err := database.Exec(
		`INSERT INTO jobs (id, user_id, article_id, type, status,
		                   telegram_update_id, telegram_chat_id,
		                   telegram_message_id, telegram_user_id, created_at)
		 VALUES (?, ?, ?, ?, 'queued', 1001, 2001, 3001, 4001, ?)`,
		jID, personalUserID, aID, jobs.TypeArticleProcessing,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	require.NoError(t, err)
}

// newTestFetcher creates a fetcher pointing at the given test server URL.
func newTestFetcher(serverURL string) *fetcher.Fetcher {
	client := req.NewClient().
		SetBaseURL(serverURL).
		SetRedirectPolicy(req.MaxRedirectPolicy(10))

	return fetcher.New(client)
}

// newTestPipeline wires up a SnapshotPipeline for tests.
func newTestPipeline(
	t *testing.T,
	database *sql.DB,
	store *artifacts.Store,
	fetch *fetcher.Fetcher,
	mdHandoff pipeline.MarkdownHandoff,
) *pipeline.SnapshotPipeline {
	t.Helper()

	logger := slogt.New(t)
	repo := jobs.NewSQLiteRepository(database, jobs.NewULIDGenerator())

	if mdHandoff == nil {
		mdHandoff = pipeline.NoOpMarkdownHandoff
	}

	return pipeline.NewSnapshotPipeline(logger, repo, store, fetch, mdHandoff)
}

// scalarString is a test helper to read a single string column.
func scalarString(t *testing.T, database *sql.DB, query string, args ...any) string {
	t.Helper()

	var value string
	require.NoError(t, database.QueryRow(query, args...).Scan(&value))

	return value
}

// scalarNullableString reads a nullable string column; returns "" when NULL.
func scalarNullableString(t *testing.T, database *sql.DB, query string, args ...any) string {
	t.Helper()

	var value sql.NullString
	require.NoError(t, database.QueryRow(query, args...).Scan(&value))

	return value.String
}

// scalarInt reads a single int column.
func scalarInt(t *testing.T, database *sql.DB, query string, args ...any) int {
	t.Helper()

	var value int
	require.NoError(t, database.QueryRow(query, args...).Scan(&value))

	return value
}

// TestSnapshotSuccessWritesSnapshotAndUpdatesCanonicalURL verifies the happy path:
// fetch succeeds → snapshot.html written → canonical_url updated → no terminal success.
func TestSnapshotSuccessWritesSnapshotAndUpdatesCanonicalURL(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html><body><p>Hello article</p></body></html>`))
	}))
	defer srv.Close()

	seedArticle(t, database, articleID, srv.URL+"/article")
	seedTelegramJob(t, database, jobID, articleID)

	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	fetch := newTestFetcher(srv.URL)

	p := newTestPipeline(t, database, store, fetch, nil)

	processed, err := p.ProcessOne(t.Context())
	require.NoError(t, err)
	require.True(t, processed)

	// snapshot.html must exist and be readable.
	rc, openErr := store.OpenSnapshot(articleID)
	require.NoError(t, openErr)
	defer rc.Close()

	// canonical_url must be set to the final URL.
	canonicalURL := scalarNullableString(t, database, `SELECT canonical_url FROM articles WHERE id = ?`, articleID)
	assert.NotEmpty(t, canonicalURL)

	// articles.status must NOT be ready at the snapshot boundary.
	articleStatus := scalarString(t, database, `SELECT status FROM articles WHERE id = ?`, articleID)
	assert.NotEqual(t, "ready", articleStatus)

	// jobs.status must NOT be succeeded at the snapshot boundary.
	jobStatus := scalarString(t, database, `SELECT status FROM jobs WHERE id = ?`, jobID)
	assert.NotEqual(t, "succeeded", jobStatus)

	// No success notification must be inserted at the snapshot boundary.
	notifCount := scalarInt(t, database, `SELECT COUNT(*) FROM notifications WHERE job_id = ?`, jobID)
	assert.Equal(t, 0, notifCount)
}

// TestSnapshotFetchFailureCommitsARCCodedFailureTransactionally verifies ARC-coded
// failure: article failed, job failed, notification inserted, all in one transaction.
func TestSnapshotFetchFailureCommitsARCCodedFailureTransactionally(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	// Server returns 404 → ARC-003.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	seedArticle(t, database, articleID, srv.URL+"/notfound")
	seedTelegramJob(t, database, jobID, articleID)

	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	fetch := newTestFetcher(srv.URL)

	p := newTestPipeline(t, database, store, fetch, nil)

	processed, err := p.ProcessOne(t.Context())
	require.NoError(t, err)
	require.True(t, processed)

	articleStatus := scalarString(t, database, `SELECT status FROM articles WHERE id = ?`, articleID)
	assert.Equal(t, "failed", articleStatus)

	articleError := scalarNullableString(t, database, `SELECT error_message FROM articles WHERE id = ?`, articleID)
	assert.Contains(t, articleError, "[ARC-003]")

	jobStatus := scalarString(t, database, `SELECT status FROM jobs WHERE id = ?`, jobID)
	assert.Equal(t, "failed", jobStatus)

	jobError := scalarNullableString(t, database, `SELECT error_message FROM jobs WHERE id = ?`, jobID)
	assert.Contains(t, jobError, "[ARC-003]")

	notifCount := scalarInt(t, database, `SELECT COUNT(*) FROM notifications WHERE job_id = ? AND status = 'pending'`, jobID)
	assert.Equal(t, 1, notifCount)
}

// TestSnapshotForbiddenFailureMapsToARC002 verifies that HTTP 403 maps to ARC-002.
func TestSnapshotForbiddenFailureMapsToARC002(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	seedArticle(t, database, articleID, srv.URL+"/forbidden")
	seedTelegramJob(t, database, jobID, articleID)

	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	fetch := newTestFetcher(srv.URL)

	p := newTestPipeline(t, database, store, fetch, nil)

	processed, err := p.ProcessOne(t.Context())
	require.NoError(t, err)
	require.True(t, processed)

	articleError := scalarNullableString(t, database, `SELECT error_message FROM articles WHERE id = ?`, articleID)
	assert.Contains(t, articleError, "[ARC-002]")
}

func TestSnapshotSSRFPolicyBlockPersistsARC017(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)
	seedArticle(t, database, articleID, "http://example.com/article")
	seedTelegramJob(t, database, jobID, articleID)

	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	guard := ssrf.New(slogt.New(t))
	fetch := fetcher.New(req.NewClient(), func(rawURL string) error {
		_, validateErr := guard.ValidateURL(rawURL, ssrf.PhaseInitialURL)
		return validateErr
	})
	p := newTestPipeline(t, database, store, fetch, nil)

	processed, err := p.ProcessOne(t.Context())
	require.NoError(t, err)
	require.True(t, processed)

	articleStatus := scalarString(t, database, `SELECT status FROM articles WHERE id = ?`, articleID)
	assert.Equal(t, "failed", articleStatus)

	const publicMessage = "[ARC-017] Archivist refused to process suspicious URL."
	assert.Equal(t, publicMessage, scalarNullableString(t, database, `SELECT error_message FROM articles WHERE id = ?`, articleID))
	assert.Equal(t, publicMessage, scalarNullableString(t, database, `SELECT error_message FROM jobs WHERE id = ?`, jobID))
	assert.Equal(t, 1, scalarInt(t, database, `SELECT COUNT(*) FROM notifications WHERE job_id = ? AND status = 'pending'`, jobID))
}

func TestSnapshotReportsProcessedAndFailsJobWhenArticleOwnershipDoesNotMatch(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)
	seedOtherUser(t, database)

	_, err := database.Exec(
		`INSERT INTO articles (id, user_id, original_url, status, created_at)
		 VALUES (?, ?, ?, 'queued', ?)`,
		articleID,
		otherUserID,
		"https://example.com/article",
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	require.NoError(t, err)

	seedTelegramJob(t, database, jobID, articleID)

	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	p := newTestPipeline(t, database, store, fetcher.New(req.NewClient()), nil)

	processed, err := p.ProcessOne(t.Context())
	require.NoError(t, err)
	require.True(t, processed)

	assert.Equal(t, "queued", scalarString(t, database, `SELECT status FROM articles WHERE id = ?`, articleID))
	assert.Empty(t, scalarNullableString(t, database, `SELECT error_message FROM articles WHERE id = ?`, articleID))
	assert.Equal(t, jobs.StatusFailed, scalarString(t, database, `SELECT status FROM jobs WHERE id = ?`, jobID))
	assert.Equal(t, "[ARC-999] Archivist could not process the URL.", scalarNullableString(t, database, `SELECT error_message FROM jobs WHERE id = ?`, jobID))
	assert.Empty(t, scalarNullableString(t, database, `SELECT canonical_url FROM articles WHERE id = ?`, articleID))
	assert.Equal(t, 1, scalarInt(t, database, `SELECT COUNT(*) FROM notifications WHERE job_id = ? AND status = 'pending'`, jobID))
}

// TestSnapshotNonHTMLFailureMapsToARC005 verifies non-HTML content type maps to ARC-005.
func TestSnapshotNonHTMLFailureMapsToARC005(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`%PDF-1.4 fake`))
	}))
	defer srv.Close()

	seedArticle(t, database, articleID, srv.URL+"/doc.pdf")
	seedTelegramJob(t, database, jobID, articleID)

	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	fetch := newTestFetcher(srv.URL)

	p := newTestPipeline(t, database, store, fetch, nil)

	processed, err := p.ProcessOne(t.Context())
	require.NoError(t, err)
	require.True(t, processed)

	articleError := scalarNullableString(t, database, `SELECT error_message FROM articles WHERE id = ?`, articleID)
	assert.Contains(t, articleError, "[ARC-005]")

	// snapshot.html must NOT exist.
	_, openErr := store.OpenSnapshot(articleID)
	assert.Error(t, openErr)
}

func TestSnapshotWriteFailureLogsCanonicalErrorFieldAndARCCode(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html><body><p>Hello article</p></body></html>`))
	}))
	defer srv.Close()

	seedArticle(t, database, articleID, srv.URL+"/article")
	seedTelegramJob(t, database, jobID, articleID)

	dataDir := t.TempDir()
	store, err := artifacts.NewStore(dataDir)
	require.NoError(t, err)
	defer store.Close()

	require.NoError(t, os.MkdirAll(filepath.Join(dataDir, "articles", articleID, "snapshot.html"), 0o700))

	var logs bytes.Buffer
	logger := newBufferLogger(t, &logs)
	repo := jobs.NewSQLiteRepository(database, jobs.NewULIDGenerator())
	p := pipeline.NewSnapshotPipeline(logger, repo, store, newTestFetcher(srv.URL), nil)

	processed, err := p.ProcessOne(t.Context())
	require.NoError(t, err)
	require.True(t, processed)

	logText := logs.String()
	assert.Contains(t, logText, `"artifact_result":"failure"`)
	assert.Contains(t, logText, `"arc_code":"ARC-007"`)
	assert.Contains(t, logText, `"error":`)
	assert.NotContains(t, logText, `"err":`)
}

func TestSnapshotLogsClaimBeforeArticleURLLoad(t *testing.T) {
	var logs bytes.Buffer
	logger := newBufferLogger(t, &logs)
	repo := &claimBeforeURLLoadRepo{
		job: &jobs.Job{
			ID:        jobID,
			ArticleID: articleID,
			UserID:    personalUserID,
			Type:      jobs.TypeArticleProcessing,
			Status:    jobs.StatusRunning,
			CreatedAt: time.Now().UTC(),
			StartedAt: new(time.Time),
		},
	}

	p := pipeline.NewSnapshotPipeline(logger, repo, nil, nil, pipeline.NoOpMarkdownHandoff)

	processed, err := p.ProcessOne(t.Context())
	require.Error(t, err)
	require.False(t, processed)
	require.True(t, repo.articleURLCalled)

	logText := logs.String()
	assert.Contains(t, logText, `"article_id":"`+articleID+`"`)
	assert.Contains(t, logText, `"job_id":"`+jobID+`"`)
	assert.Contains(t, logText, `"user_id":"`+personalUserID+`"`)
	assert.Contains(t, logText, `"stage":"claim"`)
	assert.Contains(t, logText, `"status":"claimed"`)
}

// TestSnapshotTransactionRollbackOnNotificationFailure demonstrates that a failed
// notification insert causes the entire terminal transaction to roll back.
// We simulate this by seeding a duplicate notification row that will cause the
// INSERT to violate the UNIQUE constraint on notifications.job_id.
func TestSnapshotTransactionRollbackOnNotificationFailure(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	seedArticle(t, database, articleID, srv.URL+"/notfound")
	seedTelegramJob(t, database, jobID, articleID)

	// Pre-seed a notification row for the job to cause UNIQUE constraint violation.
	_, err := database.Exec(
		`INSERT INTO notifications (id, job_id, status, created_at, expires_at) VALUES (?, ?, 'pending', ?, ?)`,
		"01ASB2XFCZJY7WHZ2FNRTMQJZ9",
		jobID,
		time.Now().UTC().Format(time.RFC3339Nano),
		time.Now().UTC().Add(7*24*time.Hour).Format(time.RFC3339Nano),
	)
	require.NoError(t, err)

	store, storeErr := artifacts.NewStore(t.TempDir())
	require.NoError(t, storeErr)
	defer store.Close()

	fetch := newTestFetcher(srv.URL)

	var logs bytes.Buffer
	logger := newBufferLogger(t, &logs)
	repo := jobs.NewSQLiteRepository(database, jobs.NewULIDGenerator())
	p := pipeline.NewSnapshotPipeline(logger, repo, store, fetch, nil)

	// The pipeline will attempt to insert a second notification, violating UNIQUE.
	// The transaction should roll back, so articles.status remains in its pre-terminal state.
	processed, processErr := p.ProcessOne(t.Context())

	// ProcessOne must return an error because CompleteTerminal failed.
	require.Error(t, processErr)
	require.False(t, processed)

	// Since the transaction rolled back, articles.status must not be 'failed'.
	// The job was claimed (status=running) but the terminal transition failed.
	articleStatus := scalarString(t, database, `SELECT status FROM articles WHERE id = ?`, articleID)
	assert.NotEqual(t, "failed", articleStatus)

	logText := logs.String()
	assert.Contains(t, logText, `"article_id":"`+articleID+`"`)
	assert.Contains(t, logText, `"job_id":"`+jobID+`"`)
	assert.Contains(t, logText, `"stage":"terminal_failure"`)
	assert.Contains(t, logText, `"status":"terminal_persist_failed"`)
	assert.Contains(t, logText, `"error":`)
}

// TestSnapshotNoQueuedJobReturnsFalse verifies that ProcessOne returns false when there are no jobs.
func TestSnapshotNoQueuedJobReturnsFalse(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)
	spanRecorder := installSpanRecorder(t)

	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	// Point fetcher at anything — it won't be called.
	fetch := fetcher.New(req.NewClient())

	p := newTestPipeline(t, database, store, fetch, nil)

	processed, err := p.ProcessOne(t.Context())
	require.NoError(t, err)
	require.False(t, processed)
	requireSpanStatusNotError(t, spanRecorder.Ended(), "worker.pipeline.claim")
	requireSpanStatusNotError(t, spanRecorder.Ended(), "worker.jobs.claim")
}

// TestMarkdownHandoffIsCalledOnSnapshotSuccess verifies that the MarkdownHandoff
// extension point is invoked after a successful snapshot.
func TestMarkdownHandoffIsCalledOnSnapshotSuccess(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html><body><p>Article text</p></body></html>`))
	}))
	defer srv.Close()

	seedArticle(t, database, articleID, srv.URL+"/article")
	seedTelegramJob(t, database, jobID, articleID)

	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	fetch := newTestFetcher(srv.URL)

	var handoffCalled bool

	handoff := pipeline.MarkdownHandoffFunc(func(_ context.Context, j *jobs.Job, _ string) error {
		handoffCalled = true
		assert.Equal(t, articleID, j.ArticleID)

		return nil
	})

	p := newTestPipeline(t, database, store, fetch, handoff)

	processed, err := p.ProcessOne(t.Context())
	require.NoError(t, err)
	require.True(t, processed)

	assert.True(t, handoffCalled, "expected markdown handoff to be called after snapshot success")
}

// TestMarkdownHandoffFailureCommitsTerminalFailure verifies that when the markdown handoff
// returns an ARC-coded error, the pipeline commits a terminal failure.
func TestMarkdownHandoffFailureCommitsTerminalFailure(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html><body><p>Article text</p></body></html>`))
	}))
	defer srv.Close()

	seedArticle(t, database, articleID, srv.URL+"/article")
	seedTelegramJob(t, database, jobID, articleID)

	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	fetch := newTestFetcher(srv.URL)

	// Simulate markdown extraction failure.
	handoff := pipeline.MarkdownHandoffFunc(func(_ context.Context, _ *jobs.Job, _ string) error {
		return arc.ErrSnapshotWrite // reuse as a convenient ARC-coded error for testing
	})

	p := newTestPipeline(t, database, store, fetch, handoff)

	processed, err := p.ProcessOne(t.Context())
	require.NoError(t, err)
	require.True(t, processed)

	articleStatus := scalarString(t, database, `SELECT status FROM articles WHERE id = ?`, articleID)
	assert.Equal(t, "failed", articleStatus)

	jobStatus := scalarString(t, database, `SELECT status FROM jobs WHERE id = ?`, jobID)
	assert.Equal(t, "failed", jobStatus)

	notifCount := scalarInt(t, database, `SELECT COUNT(*) FROM notifications WHERE job_id = ? AND status = 'pending'`, jobID)
	assert.Equal(t, 1, notifCount)
}

func TestWrappedARCFailurePersistsCleanPublicMessage(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html><body><p>Article text</p></body></html>`))
	}))
	defer srv.Close()

	seedArticle(t, database, articleID, srv.URL+"/article")
	seedTelegramJob(t, database, jobID, articleID)

	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	handoff := pipeline.MarkdownHandoffFunc(func(_ context.Context, _ *jobs.Job, _ string) error {
		return fmt.Errorf("jina: HTTP 500 from provider: %w", arc.ErrJinaReaderFailure)
	})

	p := newTestPipeline(t, database, store, newTestFetcher(srv.URL), handoff)

	processed, err := p.ProcessOne(t.Context())
	require.NoError(t, err)
	require.True(t, processed)

	const publicMessage = "[ARC-010] Archivist could not extract this page with the fallback reader."
	assert.Equal(t, publicMessage, scalarNullableString(t, database, `SELECT error_message FROM articles WHERE id = ?`, articleID))
	assert.Equal(t, publicMessage, scalarNullableString(t, database, `SELECT error_message FROM jobs WHERE id = ?`, jobID))
}

type claimBeforeURLLoadRepo struct {
	job              *jobs.Job
	articleURLCalled bool
}

func (r *claimBeforeURLLoadRepo) ClaimQueued(context.Context) (*jobs.Job, error) {
	return r.job, nil
}

func (r *claimBeforeURLLoadRepo) CompleteTerminal(context.Context, *jobs.Job, jobs.TerminalOutcome) error {
	return nil
}

func (r *claimBeforeURLLoadRepo) CompleteJobFailure(context.Context, *jobs.Job, string) error {
	return nil
}

func (r *claimBeforeURLLoadRepo) ArticleURL(context.Context, string, string) (string, error) {
	r.articleURLCalled = true

	return "", fmt.Errorf("article URL load failed")
}

func (r *claimBeforeURLLoadRepo) UpdateCanonicalURL(context.Context, string, string, string) error {
	return nil
}

func (r *claimBeforeURLLoadRepo) UpdateArticleTitle(context.Context, string, string, string) error {
	return nil
}

func installSpanRecorder(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()

	spanRecorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(tracerProvider)

	t.Cleanup(func() {
		otel.SetTracerProvider(previous)
		require.NoError(t, tracerProvider.Shutdown(context.Background()))
	})

	return spanRecorder
}

func requireSpanStatusNotError(t *testing.T, spans []sdktrace.ReadOnlySpan, spanName string) {
	t.Helper()

	for _, span := range spans {
		if span.Name() != spanName {
			continue
		}

		assert.NotEqual(t, codes.Error, span.Status().Code)

		return
	}

	t.Fatalf("span %q was not recorded", spanName)
}
