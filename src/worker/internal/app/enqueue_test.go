package app

import (
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnqueueCommandCreatesQueuedNonTelegramJob(t *testing.T) {
	application, cfg := newProcessTestApp(t)
	seedProcessUser(t, application.DB)

	withArgs(t, "archivist-worker", "enqueue", "https://example.com/article")

	err := CliProgram(t.Context(), application, cfg)
	require.NoError(t, err)

	var (
		articleID       string
		articleUserID   string
		originalURL     string
		articleStatus   string
		jobID           string
		jobUserID       string
		jobArticleID    string
		jobType         string
		jobStatus       string
		telegramChatID  sql.NullInt64
		telegramMsgID   sql.NullInt64
		telegramUserID  sql.NullInt64
		telegramUpdate  sql.NullInt64
		notificationCnt int
	)

	err = application.DB.QueryRow(
		`SELECT id, user_id, original_url, status FROM articles`,
	).Scan(&articleID, &articleUserID, &originalURL, &articleStatus)
	require.NoError(t, err)

	err = application.DB.QueryRow(
		`SELECT id, user_id, article_id, type, status,
		        telegram_update_id, telegram_chat_id, telegram_message_id, telegram_user_id
		 FROM jobs`,
	).Scan(
		&jobID,
		&jobUserID,
		&jobArticleID,
		&jobType,
		&jobStatus,
		&telegramUpdate,
		&telegramChatID,
		&telegramMsgID,
		&telegramUserID,
	)
	require.NoError(t, err)

	err = application.DB.QueryRow(`SELECT COUNT(*) FROM notifications`).Scan(&notificationCnt)
	require.NoError(t, err)

	assert.NotEmpty(t, articleID)
	assert.NotEmpty(t, jobID)
	assert.Equal(t, jobs.DefaultUserID, articleUserID)
	assert.Equal(t, jobs.DefaultUserID, jobUserID)
	assert.Equal(t, articleID, jobArticleID)
	assert.Equal(t, "https://example.com/article", originalURL)
	assert.Equal(t, "queued", articleStatus)
	assert.Equal(t, jobs.TypeArticleProcessing, jobType)
	assert.Equal(t, jobs.StatusQueued, jobStatus)
	assert.False(t, telegramUpdate.Valid)
	assert.False(t, telegramChatID.Valid)
	assert.False(t, telegramMsgID.Valid)
	assert.False(t, telegramUserID.Valid)
	assert.Equal(t, 0, notificationCnt)
}

func TestEnqueueCommandRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "missing URL",
			args: []string{"archivist-worker", "enqueue"},
			want: "enqueue: expected exactly one URL argument, got 0",
		},
		{
			name: "multiple URLs",
			args: []string{"archivist-worker", "enqueue", "https://example.com/a", "https://example.com/b"},
			want: "enqueue: expected exactly one URL argument, got 2",
		},
		{
			name: "relative URL",
			args: []string{"archivist-worker", "enqueue", "/article"},
			want: "worker: enqueue URL must be an absolute http or https URL",
		},
		{
			name: "unsupported scheme",
			args: []string{"archivist-worker", "enqueue", "ftp://example.com/article"},
			want: "worker: enqueue URL must use http or https",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			application, cfg := newProcessTestApp(t)
			seedProcessUser(t, application.DB)
			withArgs(t, tt.args...)

			err := CliProgram(t.Context(), application, cfg)

			require.Error(t, err)
			assert.ErrorContains(t, err, tt.want)
		})
	}
}

func TestEnqueueCommandCreatedJobCanBeProcessed(t *testing.T) {
	application, cfg := newProcessTestApp(t)
	seedProcessUser(t, application.DB)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<!doctype html>
<html>
<head><title>Queued CLI Article</title></head>
<body><article><h1>Queued CLI Article</h1><p>Enough text to snapshot.</p></article></body>
</html>`))
	}))
	defer srv.Close()
	installProcessTestPipeline(t, application, srv.Listener.Addr().String())

	withArgs(t, "archivist-worker", "enqueue", "https://article.example/article")
	require.NoError(t, CliProgram(t.Context(), application, cfg))

	articleID := scalarString(t, application.DB, `SELECT id FROM articles LIMIT 1`)

	withArgs(t, "archivist-worker", "process", "--once")
	require.NoError(t, CliProgram(t.Context(), application, cfg))

	snapshot, err := application.ArtifactStore.OpenSnapshot(articleID)
	require.NoError(t, err)
	defer snapshot.Close()

	_, err = io.ReadAll(snapshot)
	require.NoError(t, err)

	summaryContent, err := os.ReadFile(filepath.Join(cfg.Data.Dir, "articles", articleID, "summary.md"))
	require.NoError(t, err)
	assert.Equal(t, "Process command summary.", string(summaryContent))

	assert.NotEmpty(t, scalarNullableString(t, application.DB, `SELECT canonical_url FROM articles WHERE id = ?`, articleID))
	assert.Equal(t, "ready", scalarString(t, application.DB, `SELECT status FROM articles WHERE id = ?`, articleID))
	assert.Equal(t, jobs.StatusSucceeded, scalarString(t, application.DB, `SELECT status FROM jobs WHERE article_id = ?`, articleID))
	assert.Equal(t, 0, scalarInt(t, application.DB, `SELECT COUNT(*) FROM notifications`))
}

func TestEnqueueCommandFailsWhenDefaultUserIsMissing(t *testing.T) {
	application, cfg := newProcessTestApp(t)
	withArgs(t, "archivist-worker", "enqueue", "https://example.com/article")

	err := CliProgram(t.Context(), application, cfg)

	require.Error(t, err)
	assert.ErrorContains(t, err, "default user 01ASB2XFCZJY7WHZ2FNRTMQJCT is missing")
	assert.Equal(t, 0, scalarInt(t, application.DB, `SELECT COUNT(*) FROM articles`))
	assert.Equal(t, 0, scalarInt(t, application.DB, `SELECT COUNT(*) FROM jobs`))
}

func scalarInt(t *testing.T, database *sql.DB, query string, args ...any) int {
	t.Helper()

	var value int
	require.NoError(t, database.QueryRow(query, args...).Scan(&value))

	return value
}
