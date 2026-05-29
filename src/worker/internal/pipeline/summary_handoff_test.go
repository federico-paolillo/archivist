package pipeline_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
	"codeberg.org/federico-paolillo/archivist/internal/artifacts"
	"codeberg.org/federico-paolillo/archivist/internal/markdown"
	"codeberg.org/federico-paolillo/archivist/internal/pipeline"
	"codeberg.org/federico-paolillo/archivist/internal/summary"
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSummarySuccessWritesSummaryAndCommitsTerminalSuccess(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	srv := newHTMLServer("<html><body><article><p>Article text</p></article></body></html>")
	defer srv.Close()

	seedArticle(t, database, articleID, srv.URL+"/article")
	seedTelegramJob(t, database, jobID, articleID)

	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	summarizer := &fakeSummarizer{
		provider: summary.ProviderAnthropic,
		output: summary.SummarizerOutput{
			Summary:   "Concise article summary.",
			RequestID: "req-success",
		},
	}

	p := newPipelineWithSummary(t, database, store, srv.URL, summarizer)

	processed, err := p.ProcessOne(t.Context())
	require.NoError(t, err)
	require.True(t, processed)

	content, err := os.ReadFile(filepath.Join(storeDataDir(t, store), "articles", articleID, "summary.md"))
	require.NoError(t, err)
	assert.Equal(t, "Concise article summary.", string(content))

	assert.Equal(t, 1, summarizer.calls)
	assert.Equal(t, "# Article\n\nArticle text.", summarizer.inputs[0].MarkdownSource)
	assert.Equal(t, articleID, summarizer.inputs[0].ArticleID)
	assert.Equal(t, jobID, summarizer.inputs[0].JobID)

	assert.Equal(t, "ready", scalarString(t, database, `SELECT status FROM articles WHERE id = ?`, articleID))
	assert.Equal(t, "", scalarNullableString(t, database, `SELECT error_message FROM articles WHERE id = ?`, articleID))
	assert.Equal(t, "succeeded", scalarString(t, database, `SELECT status FROM jobs WHERE id = ?`, jobID))
	assert.Equal(t, "", scalarNullableString(t, database, `SELECT error_message FROM jobs WHERE id = ?`, jobID))
	assert.Equal(t, 1, scalarInt(t, database, `SELECT COUNT(*) FROM notifications WHERE job_id = ? AND status = 'pending'`, jobID))
}

func TestMarkdownBoundaryIsNonTerminalBeforeSummaryHandoff(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	srv := newHTMLServer("<html><body><article><p>Article text</p></article></body></html>")
	defer srv.Close()

	seedArticle(t, database, articleID, srv.URL+"/article")
	seedTelegramJob(t, database, jobID, articleID)

	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	local := successfulMarkdownExtractor()
	jina := successfulFallbackExtractor()
	summaryHandoff := pipeline.SummaryHandoffFunc(func(_ context.Context, _ *jobs.Job, _ string) error {
		assert.Equal(t, "queued", scalarString(t, database, `SELECT status FROM articles WHERE id = ?`, articleID))
		assert.Equal(t, "running", scalarString(t, database, `SELECT status FROM jobs WHERE id = ?`, jobID))
		assert.Equal(t, 0, scalarInt(t, database, `SELECT COUNT(*) FROM notifications WHERE job_id = ?`, jobID))

		return arc.ErrSummarizerProviderFailure
	})
	repo := jobs.NewSQLiteRepository(database, jobs.NewULIDGenerator())
	handoff := pipeline.NewMarkdownExtractionHandoff(newBufferLogger(t, nil), repo, store, local, jina, summaryHandoff)
	p := newTestPipeline(t, database, store, newTestFetcher(srv.URL), handoff)

	processed, err := p.ProcessOne(t.Context())
	require.NoError(t, err)
	require.True(t, processed)

	assert.Equal(t, "failed", scalarString(t, database, `SELECT status FROM articles WHERE id = ?`, articleID))
	assert.Contains(t, scalarNullableString(t, database, `SELECT error_message FROM articles WHERE id = ?`, articleID), "[ARC-013]")
}

func TestSummaryProviderFailureCommitsARC013(t *testing.T) {
	assertSummaryFailureCommitsARC(
		t,
		&summary.ProviderError{
			Provider: summary.ProviderAnthropic,
			Reason:   "provider unavailable",
			Err:      arc.ErrSummarizerProviderFailure,
		},
		"[ARC-013]",
	)
}

func TestSummaryRequestTooLargeCommitsARC014(t *testing.T) {
	assertSummaryFailureCommitsARC(
		t,
		&summary.ProviderError{
			Provider: summary.ProviderAnthropic,
			Reason:   "context window exceeded",
			Err:      arc.ErrSummarizerRequestTooLarge,
		},
		"[ARC-014]",
	)
}

func TestSummaryBillingFailureCommitsARC015(t *testing.T) {
	assertSummaryFailureCommitsARC(
		t,
		&summary.ProviderError{
			Provider:   summary.ProviderAnthropic,
			Reason:     "billing error",
			StatusCode: http.StatusPaymentRequired,
			Err:        arc.ErrSummarizerBillingFailure,
		},
		"[ARC-015]",
	)
}

func TestSummaryUnknownFailureCommitsARC999AndLogsModel(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	srv := newHTMLServer("<html><body><article><p>Article text</p></article></body></html>")
	defer srv.Close()

	seedArticle(t, database, articleID, srv.URL+"/article")
	seedTelegramJob(t, database, jobID, articleID)

	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	var logs bytes.Buffer
	summarizer := &fakeSummarizer{
		provider: summary.ProviderAnthropic,
		err:      errors.New("unexpected summarizer failure"),
	}
	p := newPipelineWithSummaryLogger(t, database, store, srv.URL, summarizer, &logs)

	processed, err := p.ProcessOne(t.Context())
	require.NoError(t, err)
	require.True(t, processed)

	assert.Equal(t, "failed", scalarString(t, database, `SELECT status FROM articles WHERE id = ?`, articleID))
	assert.Equal(t, "failed", scalarString(t, database, `SELECT status FROM jobs WHERE id = ?`, jobID))
	assert.Equal(
		t,
		"[ARC-999] Archivist could not process the URL.",
		scalarNullableString(t, database, `SELECT error_message FROM articles WHERE id = ?`, articleID),
	)
	assert.Equal(t, 1, scalarInt(t, database, `SELECT COUNT(*) FROM notifications WHERE job_id = ? AND status = 'pending'`, jobID))

	logText := logs.String()
	assert.Contains(t, logText, `"provider":"anthropic"`)
	assert.Contains(t, logText, `"model":"claude-test"`)
	assert.Contains(t, logText, `"status":"failure"`)
	assert.Contains(t, logText, `"arc_code":"ARC-999"`)
	assert.NotContains(t, logText, "# Article")
}

func TestSummaryWriteFailureCommitsARC016AndDoesNotPromotePartialSummary(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	srv := newHTMLServer("<html><body><article><p>Article text</p></article></body></html>")
	defer srv.Close()

	seedArticle(t, database, articleID, srv.URL+"/article")
	seedTelegramJob(t, database, jobID, articleID)

	dataDir := t.TempDir()
	store, err := artifacts.NewStore(dataDir)
	require.NoError(t, err)
	defer store.Close()

	require.NoError(t, os.MkdirAll(filepath.Join(dataDir, "articles", articleID, "summary.md"), 0o700))

	summarizer := &fakeSummarizer{
		provider: summary.ProviderAnthropic,
		output: summary.SummarizerOutput{
			Summary:   "Summary that cannot be promoted.",
			RequestID: "req-write-failure",
		},
	}

	p := newPipelineWithSummary(t, database, store, srv.URL, summarizer)

	processed, err := p.ProcessOne(t.Context())
	require.NoError(t, err)
	require.True(t, processed)

	assert.Equal(t, "failed", scalarString(t, database, `SELECT status FROM articles WHERE id = ?`, articleID))
	assert.Contains(t, scalarNullableString(t, database, `SELECT error_message FROM articles WHERE id = ?`, articleID), "[ARC-016]")

	info, statErr := os.Stat(filepath.Join(dataDir, "articles", articleID, "summary.md"))
	require.NoError(t, statErr)
	assert.True(t, info.IsDir())
}

func TestSummarySuccessRollsBackWhenTerminalTransactionFails(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	srv := newHTMLServer("<html><body><article><p>Article text</p></article></body></html>")
	defer srv.Close()

	seedArticle(t, database, articleID, srv.URL+"/article")
	seedTelegramJob(t, database, jobID, articleID)

	_, err := database.Exec(
		`INSERT INTO notifications (id, job_id, status, created_at, expires_at) VALUES (?, ?, 'pending', datetime('now'), datetime('now', '+7 days'))`,
		"01ASB2XFCZJY7WHZ2FNRTMQJZ9",
		jobID,
	)
	require.NoError(t, err)

	store, storeErr := artifacts.NewStore(t.TempDir())
	require.NoError(t, storeErr)
	defer store.Close()

	summarizer := &fakeSummarizer{
		provider: summary.ProviderAnthropic,
		output: summary.SummarizerOutput{
			Summary:   "Summary promoted before terminal transaction.",
			RequestID: "req-terminal-failure",
		},
	}
	p := newPipelineWithSummary(t, database, store, srv.URL, summarizer)

	processed, processErr := p.ProcessOne(t.Context())
	require.Error(t, processErr)
	require.False(t, processed)

	assert.Equal(t, "queued", scalarString(t, database, `SELECT status FROM articles WHERE id = ?`, articleID))
	assert.Equal(t, "running", scalarString(t, database, `SELECT status FROM jobs WHERE id = ?`, jobID))

	assert.FileExists(t, filepath.Join(storeDataDir(t, store), "articles", articleID, "summary.md"))
}

func TestSummaryLogsRequiredFieldsWithoutContent(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	srv := newHTMLServer("<html><body><article><p>Private article text</p></article></body></html>")
	defer srv.Close()

	seedArticle(t, database, articleID, srv.URL+"/article")
	seedTelegramJob(t, database, jobID, articleID)

	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	var logs bytes.Buffer
	summarizer := &fakeSummarizer{
		provider: summary.ProviderAnthropic,
		output: summary.SummarizerOutput{
			Summary:   "Private generated summary.",
			RequestID: "req-log",
		},
	}
	p := newPipelineWithSummaryLogger(t, database, store, srv.URL, summarizer, &logs)

	processed, err := p.ProcessOne(t.Context())
	require.NoError(t, err)
	require.True(t, processed)

	logText := logs.String()
	assert.Contains(t, logText, `"article_id":"`+articleID+`"`)
	assert.Contains(t, logText, `"job_id":"`+jobID+`"`)
	assert.Contains(t, logText, `"url":"`+srv.URL+`/article"`)
	assert.Contains(t, logText, `"provider":"anthropic"`)
	assert.Contains(t, logText, `"model":"claude-test"`)
	assert.Contains(t, logText, `"request_id":"req-log"`)
	assert.Contains(t, logText, `"duration":`)
	assert.Contains(t, logText, `"status":"success"`)
	assert.Contains(t, logText, `"artifact_result":"success"`)
	assert.NotContains(t, logText, "# Article")
	assert.NotContains(t, logText, "Private generated summary.")
}

func assertSummaryFailureCommitsARC(t *testing.T, summaryErr error, prefix string) {
	t.Helper()

	database := openTestDB(t)
	seedUser(t, database)

	srv := newHTMLServer("<html><body><article><p>Article text</p></article></body></html>")
	defer srv.Close()

	seedArticle(t, database, articleID, srv.URL+"/article")
	seedTelegramJob(t, database, jobID, articleID)

	store, err := artifacts.NewStore(t.TempDir())
	require.NoError(t, err)
	defer store.Close()

	summarizer := &fakeSummarizer{
		provider: summary.ProviderAnthropic,
		err:      summaryErr,
	}
	p := newPipelineWithSummary(t, database, store, srv.URL, summarizer)

	processed, err := p.ProcessOne(t.Context())
	require.NoError(t, err)
	require.True(t, processed)

	assert.Equal(t, "failed", scalarString(t, database, `SELECT status FROM articles WHERE id = ?`, articleID))
	assert.Equal(t, "failed", scalarString(t, database, `SELECT status FROM jobs WHERE id = ?`, jobID))
	assert.True(t, strings.HasPrefix(scalarNullableString(t, database, `SELECT error_message FROM articles WHERE id = ?`, articleID), prefix))
	assert.True(t, strings.HasPrefix(scalarNullableString(t, database, `SELECT error_message FROM jobs WHERE id = ?`, jobID), prefix))
	assert.Equal(t, 1, scalarInt(t, database, `SELECT COUNT(*) FROM notifications WHERE job_id = ? AND status = 'pending'`, jobID))

	assert.NoFileExists(t, filepath.Join(storeDataDir(t, store), "articles", articleID, "summary.md"))
}

func newPipelineWithSummary(
	t *testing.T,
	database *sql.DB,
	store *artifacts.Store,
	serverURL string,
	summarizer summary.SummarizerService,
) *pipeline.SnapshotPipeline {
	t.Helper()

	return newPipelineWithSummaryLogger(t, database, store, serverURL, summarizer, nil)
}

func newPipelineWithSummaryLogger(
	t *testing.T,
	database *sql.DB,
	store *artifacts.Store,
	serverURL string,
	summarizer summary.SummarizerService,
	logs *bytes.Buffer,
) *pipeline.SnapshotPipeline {
	t.Helper()

	repo := jobs.NewSQLiteRepository(database, jobs.NewULIDGenerator())
	summaryHandoff := pipeline.NewSummaryGenerationHandoff(newBufferLogger(t, logs), repo, store, summarizer)
	markdownHandoff := pipeline.NewMarkdownExtractionHandoff(
		newBufferLogger(t, logs),
		repo,
		store,
		successfulMarkdownExtractor(),
		successfulFallbackExtractor(),
		summaryHandoff,
	)

	return pipeline.NewSnapshotPipeline(newBufferLogger(t, logs), repo, store, newTestFetcher(serverURL), markdownHandoff)
}

func successfulMarkdownExtractor() *fakeMarkdownExtractor {
	return &fakeMarkdownExtractor{
		provider: markdown.ProviderGoReadability,
		output: markdown.ExtractOutput{
			Markdown: "# Article\n\nArticle text.",
		},
	}
}

func successfulFallbackExtractor() *fakeMarkdownExtractor {
	return &fakeMarkdownExtractor{
		provider: markdown.ProviderJina,
		output: markdown.ExtractOutput{
			Markdown: "# Fallback",
		},
	}
}

func storeDataDir(t *testing.T, store *artifacts.Store) string {
	t.Helper()

	value := reflect.ValueOf(store).Elem().FieldByName("dataDir")
	require.True(t, value.IsValid())

	return value.String()
}
