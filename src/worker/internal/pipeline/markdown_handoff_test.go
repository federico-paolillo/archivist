package pipeline_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"codeberg.org/federico-paolillo/archivist/internal/artifacts"
	"codeberg.org/federico-paolillo/archivist/internal/markdown"
	"codeberg.org/federico-paolillo/archivist/internal/pipeline"
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeMarkdownExtractor struct {
	calls  int
	inputs []markdown.ExtractInput
	result markdown.ExtractResult
}

func (e *fakeMarkdownExtractor) ExtractMarkdown(_ context.Context, input markdown.ExtractInput) markdown.ExtractResult {
	e.calls++
	e.inputs = append(e.inputs, input)

	return e.result
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
		result: markdown.ExtractResult{
			Status:   markdown.ResultStatusSuccess,
			Provider: markdown.ProviderGoReadability,
			Markdown: "# Article\n\nReadable text.",
		},
	}
	jina := &fakeMarkdownExtractor{
		result: markdown.ExtractResult{
			Status:   markdown.ResultStatusSuccess,
			Provider: markdown.ProviderJina,
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
		result: markdown.ExtractResult{
			Status:   markdown.ResultStatusLocalUnreadable,
			Provider: markdown.ProviderGoReadability,
		},
	}
	jina := &fakeMarkdownExtractor{
		result: markdown.ExtractResult{
			Status:   markdown.ResultStatusSuccess,
			Provider: markdown.ProviderJina,
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
		result: markdown.ExtractResult{
			Status:        markdown.ResultStatusFailure,
			Provider:      markdown.ProviderGoReadability,
			ErrorCode:     markdown.ErrorCodeLocalExtractionFailed,
			FailureReason: "convert readable HTML to Markdown: failed",
		},
	}
	jina := &fakeMarkdownExtractor{
		result: markdown.ExtractResult{
			Status:   markdown.ResultStatusSuccess,
			Provider: markdown.ProviderJina,
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
		result: markdown.ExtractResult{
			Status:   markdown.ResultStatusLocalUnreadable,
			Provider: markdown.ProviderGoReadability,
		},
	}
	jina := &fakeMarkdownExtractor{
		result: markdown.ExtractResult{
			Status:        markdown.ResultStatusFailure,
			Provider:      markdown.ProviderJina,
			ErrorCode:     markdown.ErrorCodeJinaFailed,
			FailureReason: "jina reader returned HTTP 500",
		},
	}

	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, nil), store, local, jina)

	err = handoff.Handoff(t.Context(), testJob(), "https://example.com/article")
	require.ErrorIs(t, err, pipeline.ErrJinaReaderFailure)
}

func TestMarkdownExtractionHandoffJinaInsufficientBalanceReturnsARC011(t *testing.T) {
	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	require.NoError(t, store.WriteSnapshot(articleID, strings.NewReader("<html>saved</html>")))

	local := &fakeMarkdownExtractor{
		result: markdown.ExtractResult{
			Status:   markdown.ResultStatusLocalUnreadable,
			Provider: markdown.ProviderGoReadability,
		},
	}
	jina := &fakeMarkdownExtractor{
		result: markdown.ExtractResult{
			Status:        markdown.ResultStatusFailure,
			Provider:      markdown.ProviderJina,
			ErrorCode:     markdown.ErrorCodeJinaInsufficientCredit,
			FailureReason: "jina reader insufficient balance",
		},
	}

	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, nil), store, local, jina)

	err = handoff.Handoff(t.Context(), testJob(), "https://example.com/article")
	require.ErrorIs(t, err, pipeline.ErrJinaInsufficientBalance)
}

func TestMarkdownExtractionHandoffMarkdownWriteFailureReturnsARC012(t *testing.T) {
	dataDir := t.TempDir()

	store, err := artifacts.NewStore(dataDir)
	require.NoError(t, err)
	defer store.Close()

	require.NoError(t, store.WriteSnapshot(articleID, strings.NewReader("<html>saved</html>")))
	require.NoError(t, os.Mkdir(filepath.Join(dataDir, "articles", articleID, "content.md"), 0o700))

	local := &fakeMarkdownExtractor{
		result: markdown.ExtractResult{
			Status:   markdown.ResultStatusSuccess,
			Provider: markdown.ProviderGoReadability,
			Markdown: "# Article",
		},
	}

	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, nil), store, local, nil)

	err = handoff.Handoff(t.Context(), testJob(), "https://example.com/article")
	require.ErrorIs(t, err, pipeline.ErrMarkdownWrite)
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
		result: markdown.ExtractResult{
			Status:   markdown.ResultStatusLocalUnreadable,
			Provider: markdown.ProviderGoReadability,
		},
	}
	jina := &fakeMarkdownExtractor{
		result: markdown.ExtractResult{
			Status:        markdown.ResultStatusFailure,
			Provider:      markdown.ProviderJina,
			ErrorCode:     markdown.ErrorCodeJinaFailed,
			FailureReason: "jina reader returned HTTP 500",
		},
	}

	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, nil), store, local, jina)
	p := newTestPipeline(t, database, store, newTestFetcher(srv.URL), handoff)

	err = p.ProcessOne(t.Context())
	require.NoError(t, err)

	assert.Equal(t, "failed", scalarString(t, database, `SELECT status FROM articles WHERE id = ?`, articleID))
	assert.Equal(t, "failed", scalarString(t, database, `SELECT status FROM jobs WHERE id = ?`, jobID))
	assert.Contains(t, scalarNullableString(t, database, `SELECT error_message FROM articles WHERE id = ?`, articleID), "[ARC-010]")
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
		result: markdown.ExtractResult{
			Status:   markdown.ResultStatusLocalUnreadable,
			Provider: markdown.ProviderGoReadability,
		},
	}
	jina := &fakeMarkdownExtractor{
		result: markdown.ExtractResult{
			Status:        markdown.ResultStatusFailure,
			Provider:      markdown.ProviderJina,
			ErrorCode:     markdown.ErrorCodeJinaFailed,
			FailureReason: "jina reader returned HTTP 500",
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
