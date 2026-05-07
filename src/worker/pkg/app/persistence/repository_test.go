package persistence

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEnsureSchemaCreatesCanonicalTables(t *testing.T) {
	db := openTestDB(t)

	rows, err := db.Query("PRAGMA table_info(articles);")
	require.NoError(t, err)
	defer rows.Close()

	columns := map[string]bool{}
	for rows.Next() {
		var cid int
		var name string
		var dataType string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		require.NoError(t, rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk))
		columns[name] = true
	}
	require.NoError(t, rows.Err())

	require.True(t, columns["original_url"])
	require.False(t, columns["summary"])
	require.False(t, columns["domain"])
	require.False(t, columns["artifact_path"])
	require.False(t, columns["selected_extractor"])
	require.False(t, columns["extractor_score"])
	require.False(t, columns["processed_at"])
}

func TestClaimNextQueuedJobUsesUpdateReturning(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db, fixedIDGenerator{})
	created := time.Date(2026, 5, 7, 12, 0, 0, 0, time.UTC)
	seedQueuedJob(t, db, "01ASB2XFCZJY7WHZ2FNRTMQJC1", "01ASB2XFCZJY7WHZ2FNRTMQJC2", created)

	job, err := repo.ClaimNextQueuedJob(t.Context(), created.Add(time.Minute))

	require.NoError(t, err)
	require.NotNil(t, job)
	require.Equal(t, "01ASB2XFCZJY7WHZ2FNRTMQJC2", job.ID)
	require.Equal(t, int64(100), job.TelegramUpdateID.Int64)
	require.Equal(t, int64(400), job.TelegramUserID.Int64)
	require.Equal(t, statusRunning, scalarString(t, db, "SELECT status FROM jobs WHERE id = ?", job.ID))
	require.NotEmpty(t, scalarString(t, db, "SELECT started_at FROM jobs WHERE id = ?", job.ID))
}

func TestCompleteJobSucceededUpdatesArticleJobAndNotificationAtomically(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db, fixedIDGenerator{})
	now := time.Date(2026, 5, 7, 12, 0, 0, 0, time.UTC)
	articleID := "01ASB2XFCZJY7WHZ2FNRTMQJC1"
	jobID := "01ASB2XFCZJY7WHZ2FNRTMQJC2"
	seedRunningJob(t, db, articleID, jobID, now)

	err := repo.CompleteJobSucceeded(t.Context(), jobID, now)

	require.NoError(t, err)
	require.Equal(t, statusReady, scalarString(t, db, "SELECT status FROM articles WHERE id = ?", articleID))
	require.Equal(t, statusSucceeded, scalarString(t, db, "SELECT status FROM jobs WHERE id = ?", jobID))
	require.Equal(t, "", scalarNullableString(t, db, "SELECT error_message FROM jobs WHERE id = ?", jobID))
	require.Equal(t, notificationPending, scalarString(t, db, "SELECT status FROM notifications WHERE job_id = ?", jobID))
	require.Equal(t, formatTime(now.Add(terminalJobTTL)), scalarString(t, db, "SELECT expires_at FROM jobs WHERE id = ?", jobID))
}

func TestCompleteJobFailedPersistsErrorAndNotification(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db, fixedIDGenerator{})
	now := time.Date(2026, 5, 7, 12, 0, 0, 0, time.UTC)
	articleID := "01ASB2XFCZJY7WHZ2FNRTMQJC1"
	jobID := "01ASB2XFCZJY7WHZ2FNRTMQJC2"
	seedRunningJob(t, db, articleID, jobID, now)

	err := repo.CompleteJobFailed(t.Context(), jobID, "[ARC-003] The URL was not found.", now)

	require.NoError(t, err)
	require.Equal(t, statusFailed, scalarString(t, db, "SELECT status FROM articles WHERE id = ?", articleID))
	require.Equal(t, statusFailed, scalarString(t, db, "SELECT status FROM jobs WHERE id = ?", jobID))
	require.Equal(t, "[ARC-003] The URL was not found.", scalarString(t, db, "SELECT error_message FROM jobs WHERE id = ?", jobID))
	require.Equal(t, notificationPending, scalarString(t, db, "SELECT status FROM notifications WHERE job_id = ?", jobID))
}

func TestCleanupEligibilityQueriesTerminalRows(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db, fixedIDGenerator{})
	now := time.Date(2026, 5, 7, 12, 0, 0, 0, time.UTC)
	seedExpiredRows(t, db, now)

	jobs, err := repo.ExpiredTerminalJobIDs(t.Context(), now)
	require.NoError(t, err)
	require.Equal(t, []string{"01ASB2XFCZJY7WHZ2FNRTMQJC2"}, jobs)

	notifications, err := repo.ExpiredNotificationIDs(t.Context(), now)
	require.NoError(t, err)
	require.Equal(t, []string{"01ASB2XFCZJY7WHZ2FNRTMQJC3"}, notifications)
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "archive.db"))
	require.NoError(t, err)
	require.NoError(t, EnsureSchema(t.Context(), db))
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return db
}

func seedQueuedJob(t *testing.T, db *sql.DB, articleID string, jobID string, created time.Time) {
	t.Helper()
	seedUserAndArticle(t, db, articleID, created)
	_, err := db.Exec(`
		INSERT INTO jobs (
			id, user_id, article_id, type, status, telegram_update_id, telegram_chat_id, telegram_message_id, telegram_user_id, created_at
		)
		VALUES (?, ?, ?, 'article-processing', ?, 100, 200, 300, 400, ?);
	`, jobID, personalUserID, articleID, statusQueued, formatTime(created))
	require.NoError(t, err)
}

func seedRunningJob(t *testing.T, db *sql.DB, articleID string, jobID string, created time.Time) {
	t.Helper()
	seedQueuedJob(t, db, articleID, jobID, created)
	_, err := db.Exec("UPDATE jobs SET status = ?, started_at = ? WHERE id = ?;", statusRunning, formatTime(created), jobID)
	require.NoError(t, err)
}

func seedUserAndArticle(t *testing.T, db *sql.DB, articleID string, created time.Time) {
	t.Helper()
	_, err := db.Exec("INSERT INTO users (id, telegram_user_id) VALUES (?, 400);", personalUserID)
	require.NoError(t, err)
	_, err = db.Exec(`
		INSERT INTO articles (id, user_id, original_url, status, created_at)
		VALUES (?, ?, 'https://example.com/article', ?, ?);
	`, articleID, personalUserID, statusQueued, formatTime(created))
	require.NoError(t, err)
}

func seedExpiredRows(t *testing.T, db *sql.DB, now time.Time) {
	t.Helper()
	articleID := "01ASB2XFCZJY7WHZ2FNRTMQJC1"
	jobID := "01ASB2XFCZJY7WHZ2FNRTMQJC2"
	notificationID := "01ASB2XFCZJY7WHZ2FNRTMQJC3"
	seedRunningJob(t, db, articleID, jobID, now)
	_, err := db.Exec(`
		UPDATE jobs SET status = ?, completed_at = ?, expires_at = ? WHERE id = ?;
	`, statusSucceeded, formatTime(now.Add(-time.Hour)), formatTime(now.Add(-time.Minute)), jobID)
	require.NoError(t, err)
	_, err = db.Exec(`
		INSERT INTO notifications (id, job_id, status, created_at, sent_at, expires_at)
		VALUES (?, ?, 'sent', ?, ?, ?);
	`, notificationID, jobID, formatTime(now.Add(-time.Hour)), formatTime(now.Add(-time.Hour)), formatTime(now.Add(-time.Minute)))
	require.NoError(t, err)
}

func scalarString(t *testing.T, db *sql.DB, query string, args ...any) string {
	t.Helper()

	var value string
	require.NoError(t, db.QueryRow(query, args...).Scan(&value))

	return value
}

func scalarNullableString(t *testing.T, db *sql.DB, query string, args ...any) string {
	t.Helper()

	var value sql.NullString
	require.NoError(t, db.QueryRow(query, args...).Scan(&value))

	return value.String
}

type fixedIDGenerator struct{}

func (fixedIDGenerator) NewID(time.Time) (string, error) {
	return "01ASB2XFCZJY7WHZ2FNRTMQJCN", nil
}
