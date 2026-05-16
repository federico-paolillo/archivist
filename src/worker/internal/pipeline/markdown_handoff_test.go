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
		},
	}
	jina := &fakeMarkdownExtractor{
		provider: markdown.ProviderJina,
		output: markdown.ExtractOutput{
			Markdown: "# Fallback",
		},
	}

	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, nil), store, local, jina)
	p := newTestPipeline(t, database, store, newTestFetcher(srv.URL), handoff)

	err = p.ProcessOne(t.Context())
	require.NoError(t, err)

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
	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, &logs), store, local, jina)

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
	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, &logs), store, local, jina)

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

	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, nil), store, local, jina)

	err = handoff.Handoff(t.Context(), testJob(), "https://example.com/article")
	require.ErrorIs(t, err, arc.ErrJinaReaderFailure)
	pipelineErr := requirePipelineError(t, err, "markdown", "provider jina")
	require.Equal(t, articleID, pipelineErr.ArticleID)
	require.Equal(t, jobID, pipelineErr.JobID)
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

	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, nil), store, local, jina)

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

	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, nil), store, local, nil)

	err = handoff.Handoff(t.Context(), testJob(), "https://example.com/article")
	require.ErrorIs(t, err, arc.ErrMarkdownWrite)
	pipelineErr := requirePipelineError(t, err, "markdown", "write markdown")

	var storeErr *artifacts.StoreError
	require.True(t, errors.As(pipelineErr, &storeErr))
	require.Equal(t, "promote artifact", storeErr.Op)
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

	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, nil), store, local, jina)
	p := newTestPipeline(t, database, store, newTestFetcher(srv.URL), handoff)

	err = p.ProcessOne(t.Context())
	require.NoError(t, err)

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

	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, nil), store, local, jina)
	p := newTestPipeline(t, database, store, newTestFetcher(srv.URL), handoff)

	processErr := p.ProcessOne(t.Context())
	require.Error(t, processErr)

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
