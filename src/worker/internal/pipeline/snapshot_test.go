package pipeline_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"codeberg.org/federico-paolillo/archivist/internal/artifacts"
	"codeberg.org/federico-paolillo/archivist/internal/fetcher"
	"codeberg.org/federico-paolillo/archivist/internal/pipeline"
	"codeberg.org/federico-paolillo/archivist/internal/testutils/slogt"
	"codeberg.org/federico-paolillo/archivist/pkg/db"
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	personalUserID = "01ASB2XFCZJY7WHZ2FNRTMQJCT"
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
	repo := jobs.NewSQLiteRepository(database)

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

	err = p.ProcessOne(t.Context())
	require.NoError(t, err)

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

	err = p.ProcessOne(t.Context())
	require.NoError(t, err)

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

	err = p.ProcessOne(t.Context())
	require.NoError(t, err)

	articleError := scalarNullableString(t, database, `SELECT error_message FROM articles WHERE id = ?`, articleID)
	assert.Contains(t, articleError, "[ARC-002]")
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

	err = p.ProcessOne(t.Context())
	require.NoError(t, err)

	articleError := scalarNullableString(t, database, `SELECT error_message FROM articles WHERE id = ?`, articleID)
	assert.Contains(t, articleError, "[ARC-005]")

	// snapshot.html must NOT exist.
	_, openErr := store.OpenSnapshot(articleID)
	assert.Error(t, openErr)
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

	p := newTestPipeline(t, database, store, fetch, nil)

	// The pipeline will attempt to insert a second notification, violating UNIQUE.
	// The transaction should roll back, so articles.status remains in its pre-terminal state.
	processErr := p.ProcessOne(t.Context())

	// ProcessOne must return an error because CompleteTerminal failed.
	require.Error(t, processErr)

	// Since the transaction rolled back, articles.status must not be 'failed'.
	// The job was claimed (status=running) but the terminal transition failed.
	articleStatus := scalarString(t, database, `SELECT status FROM articles WHERE id = ?`, articleID)
	assert.NotEqual(t, "failed", articleStatus)
}

// TestSnapshotNoQueuedJobReturnsNil verifies that ProcessOne returns nil when there are no jobs.
func TestSnapshotNoQueuedJobReturnsNil(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	// Point fetcher at anything — it won't be called.
	fetch := fetcher.New(req.NewClient())

	p := newTestPipeline(t, database, store, fetch, nil)

	err = p.ProcessOne(t.Context())
	require.NoError(t, err)
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

	err = p.ProcessOne(t.Context())
	require.NoError(t, err)

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
		return pipeline.ErrSnapshotWrite // reuse as a convenient ARC-coded error for testing
	})

	p := newTestPipeline(t, database, store, fetch, handoff)

	err = p.ProcessOne(t.Context())
	require.NoError(t, err)

	articleStatus := scalarString(t, database, `SELECT status FROM articles WHERE id = ?`, articleID)
	assert.Equal(t, "failed", articleStatus)

	jobStatus := scalarString(t, database, `SELECT status FROM jobs WHERE id = ?`, jobID)
	assert.Equal(t, "failed", jobStatus)

	notifCount := scalarInt(t, database, `SELECT COUNT(*) FROM notifications WHERE job_id = ? AND status = 'pending'`, jobID)
	assert.Equal(t, 1, notifCount)
}
