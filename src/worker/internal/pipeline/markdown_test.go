package pipeline_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
	"codeberg.org/federico-paolillo/archivist/internal/artifacts"
	"codeberg.org/federico-paolillo/archivist/internal/markdown"
	"codeberg.org/federico-paolillo/archivist/internal/pipeline"
	"codeberg.org/federico-paolillo/archivist/internal/summary"
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeMarkdownExtractor struct {
	provider markdown.Provider
	calls    int
	inputs   []markdown.ExtractInput
	output   markdown.ExtractOutput
	err      error
}

func (e *fakeMarkdownExtractor) Provider() markdown.Provider {
	return e.provider
}

func (e *fakeMarkdownExtractor) ExtractMarkdown(
	_ context.Context,
	input markdown.ExtractInput,
) (markdown.ExtractOutput, error) {
	e.calls++
	e.inputs = append(e.inputs, input)

	return e.output, e.err
}

type fakeSummarizer struct {
	provider summary.Provider
	calls    int
	inputs   []summary.SummarizerRequest
	output   summary.SummarizerOutput
	err      error
}

func (s *fakeSummarizer) Provider() summary.Provider {
	return s.provider
}

func (s *fakeSummarizer) Model() string {
	return "claude-test"
}

func (s *fakeSummarizer) Summarize(
	_ context.Context,
	req summary.SummarizerRequest,
) (summary.SummarizerOutput, error) {
	s.calls++
	s.inputs = append(s.inputs, req)

	return s.output, s.err
}

type fakeJobsRepository struct {
	titleArticleIDs []string
	titles          []string
	updateTitleErr  error
}

func (r *fakeJobsRepository) ClaimQueued(_ context.Context) (*jobs.Job, error) {
	return nil, errors.New("fake jobs repository does not claim jobs")
}

func (r *fakeJobsRepository) CompleteTerminal(_ context.Context, _ *jobs.Job, _ jobs.TerminalOutcome) error {
	return errors.New("fake jobs repository does not complete jobs")
}

func (r *fakeJobsRepository) ArticleURL(_ context.Context, _ string, _ string) (string, error) {
	return "", errors.New("fake jobs repository does not load article URLs")
}

func (r *fakeJobsRepository) UpdateCanonicalURL(_ context.Context, _ string, _ string, _ string) error {
	return errors.New("fake jobs repository does not update canonical URLs")
}

func (r *fakeJobsRepository) UpdateArticleTitle(_ context.Context, articleID string, _ string, title string) error {
	if r.updateTitleErr != nil {
		return r.updateTitleErr
	}

	r.titleArticleIDs = append(r.titleArticleIDs, articleID)
	r.titles = append(r.titles, title)

	return nil
}

func TestMarkdownExtractionHandoffLocalSuccessWritesMarkdownAndStaysNonTerminal(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	srv := newHTMLServer("<html><body><article><h1>Article</h1><p>Readable text.</p></article></body></html>")
	defer srv.Close()

	seedArticle(t, database, articleID, srv.URL+"/article")
	seedTelegramJob(t, database, jobID, articleID)

	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	local := &fakeMarkdownExtractor{
		provider: markdown.ProviderGoReadability,
		output: markdown.ExtractOutput{
			Markdown: "# Article\n\nReadable text.",
			Title:    "  Extracted Article Title  ",
		},
	}
	jina := &fakeMarkdownExtractor{
		provider: markdown.ProviderJina,
		output: markdown.ExtractOutput{
			Markdown: "# Fallback",
		},
	}

	repo := jobs.NewSQLiteRepository(database, jobs.NewULIDGenerator())
	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, nil), repo, store, local, jina, pipeline.NoOpSummaryHandoff)
	p := newTestPipeline(t, database, store, newTestFetcher(srv.URL), handoff)

	processed, err := p.ProcessOne(t.Context())
	require.NoError(t, err)
	require.True(t, processed)

	rc, openErr := store.OpenMarkdown(articleID)
	require.NoError(t, openErr)
	defer rc.Close()

	content, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, "# Article\n\nReadable text.", string(content))

	assert.Equal(t, 1, local.calls)
	assert.Equal(t, 0, jina.calls)

	articleStatus := scalarString(t, database, `SELECT status FROM articles WHERE id = ?`, articleID)
	assert.Equal(t, "queued", articleStatus)

	jobStatus := scalarString(t, database, `SELECT status FROM jobs WHERE id = ?`, jobID)
	assert.Equal(t, "running", jobStatus)

	notifCount := scalarInt(t, database, `SELECT COUNT(*) FROM notifications WHERE job_id = ?`, jobID)
	assert.Equal(t, 0, notifCount)

	title := scalarNullableString(t, database, `SELECT title FROM articles WHERE id = ?`, articleID)
	assert.Equal(t, "Extracted Article Title", title)
}

func TestMarkdownExtractionHandoffPersistsFirstMarkdownH1WhenExtractorTitleIsEmpty(t *testing.T) {
	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	require.NoError(t, store.WriteSnapshot(articleID, strings.NewReader("<html>saved</html>")))

	local := &fakeMarkdownExtractor{
		provider: markdown.ProviderGoReadability,
		output: markdown.ExtractOutput{
			Markdown: "intro\n\n# Markdown Article Title\n\nReadable text.",
		},
	}
	jina := &fakeMarkdownExtractor{provider: markdown.ProviderJina}
	repo := &fakeJobsRepository{}

	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, nil), repo, store, local, jina, pipeline.NoOpSummaryHandoff)

	err = handoff.Handoff(t.Context(), testJob(), "https://example.com/article")
	require.NoError(t, err)

	require.Equal(t, []string{articleID}, repo.titleArticleIDs)
	require.Equal(t, []string{"Markdown Article Title"}, repo.titles)
}

func TestMarkdownExtractionHandoffLeavesTitleNullWhenNoTitleIsDiscovered(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)
	seedArticle(t, database, articleID, "https://example.com/article")

	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	require.NoError(t, store.WriteSnapshot(articleID, strings.NewReader("<html>saved</html>")))

	local := &fakeMarkdownExtractor{
		provider: markdown.ProviderGoReadability,
		output: markdown.ExtractOutput{
			Markdown: "Readable text without a heading.",
		},
	}
	jina := &fakeMarkdownExtractor{provider: markdown.ProviderJina}
	repo := jobs.NewSQLiteRepository(database, jobs.NewULIDGenerator())

	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, nil), repo, store, local, jina, pipeline.NoOpSummaryHandoff)

	err = handoff.Handoff(t.Context(), testJob(), "https://example.com/article")
	require.NoError(t, err)

	title := scalarNullableString(t, database, `SELECT title FROM articles WHERE id = ?`, articleID)
	assert.Empty(t, title)
}

func TestMarkdownExtractionHandoffSnapshotReadFailureLogsMappedARCCode(t *testing.T) {
	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	local := &fakeMarkdownExtractor{provider: markdown.ProviderGoReadability}
	jina := &fakeMarkdownExtractor{provider: markdown.ProviderJina}
	repo := &fakeJobsRepository{}

	var logs bytes.Buffer
	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, &logs), repo, store, local, jina, pipeline.NoOpSummaryHandoff)

	err = handoff.Handoff(t.Context(), testJob(), "https://example.com/article")
	require.ErrorIs(t, err, arc.ErrLocalExtractionFailed)

	logText := logs.String()
	assert.Contains(t, logText, "pipeline: markdown snapshot read failed")
	assert.Contains(t, logText, `"stage":"markdown"`)
	assert.Contains(t, logText, `"status":"failure"`)
	assert.Contains(t, logText, `"arc_code":"ARC-009"`)
	assert.NotContains(t, logText, `"arc_code":"ARC-999"`)
}

func TestMarkdownExtractionHandoffTitleUpdateFailureDoesNotBlockSummaryHandoff(t *testing.T) {
	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	require.NoError(t, store.WriteSnapshot(articleID, strings.NewReader("<html>saved</html>")))

	local := &fakeMarkdownExtractor{
		provider: markdown.ProviderGoReadability,
		output: markdown.ExtractOutput{
			Markdown: "# Article\n\nReadable text.",
			Title:    "Article",
		},
	}
	jina := &fakeMarkdownExtractor{provider: markdown.ProviderJina}
	repo := &fakeJobsRepository{updateTitleErr: errors.New("title update failed")}
	var summaryCalls int
	summaryHandoff := pipeline.SummaryHandoffFunc(func(_ context.Context, j *jobs.Job, canonicalURL string) error {
		summaryCalls++
		assert.Equal(t, articleID, j.ArticleID)
		assert.Equal(t, "https://example.com/article", canonicalURL)

		return nil
	})

	var logs bytes.Buffer
	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, &logs), repo, store, local, jina, summaryHandoff)

	err = handoff.Handoff(t.Context(), testJob(), "https://example.com/article")
	require.NoError(t, err)

	assert.Equal(t, 1, summaryCalls)
	assert.Contains(t, logs.String(), "pipeline: article title update failed")

	rc, openErr := store.OpenMarkdown(articleID)
	require.NoError(t, openErr)
	defer rc.Close()

	content, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, "# Article\n\nReadable text.", string(content))
}

func TestMarkdownExtractionHandoffLocalUnreadableFallsBackToJina(t *testing.T) {
	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	require.NoError(t, store.WriteSnapshot(articleID, strings.NewReader("<html>saved</html>")))

	local := &fakeMarkdownExtractor{
		provider: markdown.ProviderGoReadability,
		err:      arc.ErrLocalUnreadable,
	}
	jina := &fakeMarkdownExtractor{
		provider: markdown.ProviderJina,
		output: markdown.ExtractOutput{
			Markdown: "# From Jina",
		},
	}

	var logs bytes.Buffer
	repo := &fakeJobsRepository{}
	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, &logs), repo, store, local, jina, pipeline.NoOpSummaryHandoff)

	err = handoff.Handoff(t.Context(), testJob(), "https://example.com/article")
	require.NoError(t, err)

	assert.Equal(t, 1, local.calls)
	assert.Equal(t, 1, jina.calls)
	assert.Equal(t, "https://example.com/article", jina.inputs[0].CanonicalURL)

	rc, openErr := store.OpenMarkdown(articleID)
	require.NoError(t, openErr)
	defer rc.Close()

	content, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, "# From Jina", string(content))

	logText := logs.String()
	assert.Contains(t, logText, `"provider":"go-readability"`)
	assert.Contains(t, logText, `"provider":"jina"`)
	assert.Contains(t, logText, `"fallback_reason":"local unreadable"`)
	assert.Contains(t, logText, `"selected_provider":"jina"`)
	assert.Contains(t, logText, `"artifact_result":"success"`)
}

func TestMarkdownExtractionHandoffLocalFailureFallsBackToJina(t *testing.T) {
	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	require.NoError(t, store.WriteSnapshot(articleID, strings.NewReader("<html>saved</html>")))

	local := &fakeMarkdownExtractor{
		provider: markdown.ProviderGoReadability,
		err: &markdown.ExtractionError{
			Provider: markdown.ProviderGoReadability,
			Reason:   "convert readable HTML to Markdown: failed",
			Err:      arc.ErrLocalExtractionFailed,
		},
	}
	jina := &fakeMarkdownExtractor{
		provider: markdown.ProviderJina,
		output: markdown.ExtractOutput{
			Markdown: "# From Jina",
		},
	}

	var logs bytes.Buffer
	repo := &fakeJobsRepository{}
	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, &logs), repo, store, local, jina, pipeline.NoOpSummaryHandoff)

	err = handoff.Handoff(t.Context(), testJob(), "https://example.com/article")
	require.NoError(t, err)

	assert.Equal(t, 1, local.calls)
	assert.Equal(t, 1, jina.calls)
	assert.Contains(t, logs.String(), `"fallback_reason":"convert readable HTML to Markdown: failed"`)
}

func TestMarkdownExtractionHandoffJinaFailureReturnsARC010(t *testing.T) {
	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	require.NoError(t, store.WriteSnapshot(articleID, strings.NewReader("<html>saved</html>")))

	local := &fakeMarkdownExtractor{
		provider: markdown.ProviderGoReadability,
		err:      arc.ErrLocalUnreadable,
	}
	jina := &fakeMarkdownExtractor{
		provider: markdown.ProviderJina,
		err: &markdown.ExtractionError{
			Provider:   markdown.ProviderJina,
			Reason:     "jina reader returned unexpected HTTP status",
			StatusCode: http.StatusInternalServerError,
			Err:        arc.ErrJinaReaderFailure,
		},
	}

	var logs bytes.Buffer
	repo := &fakeJobsRepository{}
	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, &logs), repo, store, local, jina, pipeline.NoOpSummaryHandoff)

	err = handoff.Handoff(t.Context(), testJob(), "https://example.com/article")
	require.ErrorIs(t, err, arc.ErrJinaReaderFailure)
	pipelineErr := requirePipelineError(t, err, "markdown", "provider jina")
	require.Equal(t, articleID, pipelineErr.ArticleID)
	require.Equal(t, jobID, pipelineErr.JobID)

	logText := logs.String()
	assert.Contains(t, logText, `"provider":"jina"`)
	assert.Contains(t, logText, `"arc_code":"ARC-010"`)
	assert.Contains(t, logText, `"error":`)
	assert.NotContains(t, logText, `"err":`)
}

func TestMarkdownExtractionHandoffJinaInsufficientBalanceReturnsARC011(t *testing.T) {
	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	require.NoError(t, store.WriteSnapshot(articleID, strings.NewReader("<html>saved</html>")))

	local := &fakeMarkdownExtractor{
		provider: markdown.ProviderGoReadability,
		err:      arc.ErrLocalUnreadable,
	}
	jina := &fakeMarkdownExtractor{
		provider: markdown.ProviderJina,
		err: &markdown.ExtractionError{
			Provider:   markdown.ProviderJina,
			Reason:     "jina reader insufficient balance",
			StatusCode: http.StatusPaymentRequired,
			Err:        arc.ErrJinaInsufficientBalance,
		},
	}

	repo := &fakeJobsRepository{}
	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, nil), repo, store, local, jina, pipeline.NoOpSummaryHandoff)

	err = handoff.Handoff(t.Context(), testJob(), "https://example.com/article")
	require.ErrorIs(t, err, arc.ErrJinaInsufficientBalance)
	requirePipelineError(t, err, "markdown", "provider jina")
}

func TestMarkdownExtractionHandoffMarkdownWriteFailureReturnsARC012(t *testing.T) {
	dataDir := t.TempDir()

	store, err := artifacts.NewStore(dataDir)
	require.NoError(t, err)
	defer store.Close()

	require.NoError(t, store.WriteSnapshot(articleID, strings.NewReader("<html>saved</html>")))
	require.NoError(t, os.Mkdir(filepath.Join(dataDir, "articles", articleID, "content.md"), 0o700))

	local := &fakeMarkdownExtractor{
		provider: markdown.ProviderGoReadability,
		output: markdown.ExtractOutput{
			Markdown: "# Article",
		},
	}
	jina := &fakeMarkdownExtractor{
		provider: markdown.ProviderJina,
		output: markdown.ExtractOutput{
			Markdown: "# Fallback",
		},
	}

	var logs bytes.Buffer
	repo := &fakeJobsRepository{}
	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, &logs), repo, store, local, jina, pipeline.NoOpSummaryHandoff)

	err = handoff.Handoff(t.Context(), testJob(), "https://example.com/article")
	require.ErrorIs(t, err, arc.ErrMarkdownWrite)
	pipelineErr := requirePipelineError(t, err, "markdown", "write markdown")

	var storeErr *artifacts.StoreError
	require.True(t, errors.As(pipelineErr, &storeErr))
	require.Equal(t, "promote artifact", storeErr.Op)

	logText := logs.String()
	assert.Contains(t, logText, `"artifact_result":"failure"`)
	assert.Contains(t, logText, `"arc_code":"ARC-012"`)
	assert.Contains(t, logText, `"error":`)
	assert.NotContains(t, logText, `"err":`)
}

func TestMarkdownExtractionFailureCommitsTerminalFailureTransactionally(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	srv := newHTMLServer("<html><body><article><p>Article text</p></article></body></html>")
	defer srv.Close()

	seedArticle(t, database, articleID, srv.URL+"/article")
	seedTelegramJob(t, database, jobID, articleID)

	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	local := &fakeMarkdownExtractor{
		provider: markdown.ProviderGoReadability,
		err:      arc.ErrLocalUnreadable,
	}
	jina := &fakeMarkdownExtractor{
		provider: markdown.ProviderJina,
		err: &markdown.ExtractionError{
			Provider:   markdown.ProviderJina,
			Reason:     "jina reader returned unexpected HTTP status",
			StatusCode: http.StatusInternalServerError,
			Err:        arc.ErrJinaReaderFailure,
		},
	}

	repo := jobs.NewSQLiteRepository(database, jobs.NewULIDGenerator())
	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, nil), repo, store, local, jina, pipeline.NoOpSummaryHandoff)
	p := newTestPipeline(t, database, store, newTestFetcher(srv.URL), handoff)

	processed, err := p.ProcessOne(t.Context())
	require.NoError(t, err)
	require.True(t, processed)

	assert.Equal(t, "failed", scalarString(t, database, `SELECT status FROM articles WHERE id = ?`, articleID))
	assert.Equal(t, "failed", scalarString(t, database, `SELECT status FROM jobs WHERE id = ?`, jobID))
	assert.Equal(
		t,
		"[ARC-010] Archivist could not extract this page with the fallback reader.",
		scalarNullableString(t, database, `SELECT error_message FROM articles WHERE id = ?`, articleID),
	)
	assert.Equal(t, 1, scalarInt(t, database, `SELECT COUNT(*) FROM notifications WHERE job_id = ? AND status = 'pending'`, jobID))
}

func TestMarkdownExtractionFailureRollsBackWhenNotificationInsertFails(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	srv := newHTMLServer("<html><body><article><p>Article text</p></article></body></html>")
	defer srv.Close()

	seedArticle(t, database, articleID, srv.URL+"/article")
	seedTelegramJob(t, database, jobID, articleID)

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

	local := &fakeMarkdownExtractor{
		provider: markdown.ProviderGoReadability,
		err:      arc.ErrLocalUnreadable,
	}
	jina := &fakeMarkdownExtractor{
		provider: markdown.ProviderJina,
		err: &markdown.ExtractionError{
			Provider:   markdown.ProviderJina,
			Reason:     "jina reader returned unexpected HTTP status",
			StatusCode: http.StatusInternalServerError,
			Err:        arc.ErrJinaReaderFailure,
		},
	}

	repo := jobs.NewSQLiteRepository(database, jobs.NewULIDGenerator())
	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, nil), repo, store, local, jina, pipeline.NoOpSummaryHandoff)
	p := newTestPipeline(t, database, store, newTestFetcher(srv.URL), handoff)

	processed, processErr := p.ProcessOne(t.Context())
	require.Error(t, processErr)
	require.False(t, processed)

	assert.NotEqual(t, "failed", scalarString(t, database, `SELECT status FROM articles WHERE id = ?`, articleID))
	assert.Equal(t, "running", scalarString(t, database, `SELECT status FROM jobs WHERE id = ?`, jobID))
}

func newHTMLServer(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
}

func testJob() *jobs.Job {
	return &jobs.Job{
		ID:        jobID,
		ArticleID: articleID,
	}
}

func newBufferLogger(t *testing.T, buf *bytes.Buffer) *slog.Logger {
	t.Helper()

	if buf == nil {
		return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	return slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func requirePipelineError(t *testing.T, err error, stage string, op string) *pipeline.PipelineError {
	t.Helper()

	pipelineErr, ok := errors.AsType[*pipeline.PipelineError](err)
	require.True(t, ok)
	require.Equal(t, stage, pipelineErr.Stage)
	require.Equal(t, op, pipelineErr.Op)

	return pipelineErr
}
