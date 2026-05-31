package jobs_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"codeberg.org/federico-paolillo/archivist/pkg/db"
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errIDGenerator = errors.New("id generator failed")

type fakeIDGenerator struct {
	ids []string
	err error
}

func newTestIDGenerator(ids ...string) *fakeIDGenerator {
	return &fakeIDGenerator{ids: ids}
}

func (g *fakeIDGenerator) NewID() (string, error) {
	if g.err != nil {
		return "", g.err
	}

	if len(g.ids) == 0 {
		return "", errors.New("test id generator exhausted")
	}

	id := g.ids[0]
	g.ids = g.ids[1:]

	return id, nil
}

// openTestDB returns a temporary in-memory SQLite database for testing.
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

// seedUser inserts the personal Archivist user row required by foreign keys.
func seedUser(t *testing.T, database *sql.DB) {
	t.Helper()

	_, err := database.Exec(
		`INSERT INTO users (id) VALUES (?)`,
		"01ASB2XFCZJY7WHZ2FNRTMQJCT",
	)
	require.NoError(t, err)
}

// seedArticle inserts a queued article row and returns its ID.
func seedArticle(t *testing.T, database *sql.DB, articleID string) {
	t.Helper()

	_, err := database.Exec(
		`INSERT INTO articles (id, user_id, original_url, status, created_at)
		 VALUES (?, ?, ?, 'queued', ?)`,
		articleID,
		"01ASB2XFCZJY7WHZ2FNRTMQJCT",
		"https://example.com/article",
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	require.NoError(t, err)
}

// seedTelegramJob inserts a queued job with Telegram origin metadata.
func seedTelegramJob(t *testing.T, database *sql.DB, jobID, articleID string) {
	t.Helper()

	var (
		telegramUpdateID  int64 = 1001
		telegramChatID    int64 = 2001
		telegramMessageID int64 = 3001
		telegramUserID    int64 = 4001
	)

	_, err := database.Exec(
		`INSERT INTO jobs (id, user_id, article_id, type, status,
		                   telegram_update_id, telegram_chat_id,
		                   telegram_message_id, telegram_user_id, created_at)
		 VALUES (?, ?, ?, ?, 'queued', ?, ?, ?, ?, ?)`,
		jobID,
		"01ASB2XFCZJY7WHZ2FNRTMQJCT",
		articleID,
		jobs.TypeArticleProcessing,
		telegramUpdateID,
		telegramChatID,
		telegramMessageID,
		telegramUserID,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	require.NoError(t, err)
}

// seedNonTelegramJob inserts a queued job without Telegram origin metadata.
func seedNonTelegramJob(t *testing.T, database *sql.DB, jobID, articleID string) {
	t.Helper()

	_, err := database.Exec(
		`INSERT INTO jobs (id, user_id, article_id, type, status, created_at)
		 VALUES (?, ?, ?, ?, 'queued', ?)`,
		jobID,
		"01ASB2XFCZJY7WHZ2FNRTMQJCT",
		articleID,
		jobs.TypeArticleProcessing,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	require.NoError(t, err)
}

func seedTypedJob(t *testing.T, database *sql.DB, jobID, articleID, jobType string) {
	t.Helper()

	_, err := database.Exec(
		`INSERT INTO jobs (id, user_id, article_id, type, status, created_at)
		 VALUES (?, ?, ?, ?, 'queued', ?)`,
		jobID,
		"01ASB2XFCZJY7WHZ2FNRTMQJCT",
		articleID,
		jobType,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	require.NoError(t, err)
}

func TestEnqueueURLCreatesQueuedArticleAndNonTelegramJob(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	repo := jobs.NewSQLiteRepository(database, newTestIDGenerator("ARTICLE001", "JOB001"))

	result, err := repo.EnqueueURL(t.Context(), "https://example.com/article")
	require.NoError(t, err)
	require.Equal(t, &jobs.EnqueueResult{ArticleID: "ARTICLE001", JobID: "JOB001"}, result)

	var articleUserID, originalURL, articleStatus string

	err = database.QueryRow(
		`SELECT user_id, original_url, status FROM articles WHERE id = ?`,
		"ARTICLE001",
	).Scan(&articleUserID, &originalURL, &articleStatus)
	require.NoError(t, err)
	assert.Equal(t, jobs.DefaultUserID, articleUserID)
	assert.Equal(t, "https://example.com/article", originalURL)
	assert.Equal(t, "queued", articleStatus)

	var (
		jobUserID, articleID, jobType, jobStatus string
		telegramUpdateID                         sql.NullInt64
		telegramChatID                           sql.NullInt64
		telegramMessageID                        sql.NullInt64
		telegramUserID                           sql.NullInt64
	)

	err = database.QueryRow(
		`SELECT user_id, article_id, type, status,
		        telegram_update_id, telegram_chat_id, telegram_message_id, telegram_user_id
		 FROM jobs WHERE id = ?`,
		"JOB001",
	).Scan(
		&jobUserID,
		&articleID,
		&jobType,
		&jobStatus,
		&telegramUpdateID,
		&telegramChatID,
		&telegramMessageID,
		&telegramUserID,
	)
	require.NoError(t, err)
	assert.Equal(t, jobs.DefaultUserID, jobUserID)
	assert.Equal(t, "ARTICLE001", articleID)
	assert.Equal(t, jobs.TypeArticleProcessing, jobType)
	assert.Equal(t, jobs.StatusQueued, jobStatus)
	assert.False(t, telegramUpdateID.Valid)
	assert.False(t, telegramChatID.Valid)
	assert.False(t, telegramMessageID.Valid)
	assert.False(t, telegramUserID.Valid)

	var notificationCount int

	err = database.QueryRow(`SELECT COUNT(*) FROM notifications WHERE job_id = ?`, "JOB001").Scan(&notificationCount)
	require.NoError(t, err)
	assert.Equal(t, 0, notificationCount)
}

func TestEnqueueURLFailsClearlyWhenDefaultUserIsMissing(t *testing.T) {
	database := openTestDB(t)
	repo := jobs.NewSQLiteRepository(database, newTestIDGenerator("ARTICLE001", "JOB001"))

	result, err := repo.EnqueueURL(t.Context(), "https://example.com/article")

	require.Error(t, err)
	require.ErrorContains(t, err, "jobs: default user 01ASB2XFCZJY7WHZ2FNRTMQJCT is missing")
	assert.Nil(t, result)
	assertArticleCount(t, database, 0)
	assertJobCount(t, database, 0)
}

func TestEnqueueURLRollsBackWhenJobIDGenerationFails(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	repo := jobs.NewSQLiteRepository(database, newTestIDGenerator("ARTICLE001"))

	result, err := repo.EnqueueURL(t.Context(), "https://example.com/article")

	require.ErrorContains(t, err, "jobs: failed to generate job id")
	assert.Nil(t, result)
	assertArticleCount(t, database, 0)
	assertJobCount(t, database, 0)
}

func TestClaimQueuedChangesJobStatusToRunning(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)
	seedArticle(t, database, "ARTICLE001")
	seedTelegramJob(t, database, "JOB001", "ARTICLE001")

	repo := jobs.NewSQLiteRepository(database, newTestIDGenerator())

	ctx := t.Context()

	claimed, err := repo.ClaimQueued(ctx)

	require.NoError(t, err)
	require.NotNil(t, claimed)

	assert.Equal(t, "JOB001", claimed.ID)
	assert.Equal(t, jobs.StatusRunning, claimed.Status)
	assert.NotNil(t, claimed.StartedAt)

	// Verify the DB row was actually updated.
	var dbStatus string

	err = database.QueryRowContext(ctx, `SELECT status FROM jobs WHERE id = ?`, "JOB001").Scan(&dbStatus)
	require.NoError(t, err)
	assert.Equal(t, jobs.StatusRunning, dbStatus)
}

func assertArticleCount(t *testing.T, database *sql.DB, expected int) {
	t.Helper()

	var count int

	err := database.QueryRow(`SELECT COUNT(*) FROM articles`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, expected, count)
}

func assertJobCount(t *testing.T, database *sql.DB, expected int) {
	t.Helper()

	var count int

	err := database.QueryRow(`SELECT COUNT(*) FROM jobs`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, expected, count)
}

func TestClaimQueuedReturnsErrNoRowsWhenNoJobExists(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	repo := jobs.NewSQLiteRepository(database, newTestIDGenerator())

	ctx := t.Context()

	claimed, err := repo.ClaimQueued(ctx)

	require.ErrorIs(t, err, sql.ErrNoRows)
	assert.Nil(t, claimed)
}

func TestClaimQueuedReturnsErrNoRowsForOrphanQueuedJob(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)

	_, err := database.Exec(`PRAGMA foreign_keys = OFF`)
	require.NoError(t, err)

	seedNonTelegramJob(t, database, "JOB001", "MISSINGARTICLE001")

	_, err = database.Exec(`PRAGMA foreign_keys = ON`)
	require.NoError(t, err)

	repo := jobs.NewSQLiteRepository(database, newTestIDGenerator())

	ctx := t.Context()

	claimed, err := repo.ClaimQueued(ctx)

	require.ErrorIs(t, err, sql.ErrNoRows)
	assert.Nil(t, claimed)

	var dbStatus string

	err = database.QueryRowContext(ctx, `SELECT status FROM jobs WHERE id = ?`, "JOB001").Scan(&dbStatus)
	require.NoError(t, err)
	assert.Equal(t, jobs.StatusQueued, dbStatus)
}

func TestClaimQueuedDoesNotClaimNonArticleProcessingJob(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)
	seedArticle(t, database, "ARTICLE001")
	seedTypedJob(t, database, "JOB001", "ARTICLE001", "summary-generation")

	repo := jobs.NewSQLiteRepository(database, newTestIDGenerator())

	ctx := t.Context()

	claimed, err := repo.ClaimQueued(ctx)

	require.ErrorIs(t, err, sql.ErrNoRows)
	assert.Nil(t, claimed)

	var dbStatus string

	err = database.QueryRowContext(ctx, `SELECT status FROM jobs WHERE id = ?`, "JOB001").Scan(&dbStatus)
	require.NoError(t, err)
	assert.Equal(t, jobs.StatusQueued, dbStatus)
}

func TestClaimQueuedPreservesAllTelegramOriginFields(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)
	seedArticle(t, database, "ARTICLE001")
	seedTelegramJob(t, database, "JOB001", "ARTICLE001")

	repo := jobs.NewSQLiteRepository(database, newTestIDGenerator())

	ctx := t.Context()

	claimed, err := repo.ClaimQueued(ctx)

	require.NoError(t, err)
	require.NotNil(t, claimed)

	require.NotNil(t, claimed.TelegramUpdateID)
	require.NotNil(t, claimed.TelegramChatID)
	require.NotNil(t, claimed.TelegramMessageID)
	require.NotNil(t, claimed.TelegramUserID)

	assert.Equal(t, int64(1001), *claimed.TelegramUpdateID)
	assert.Equal(t, int64(2001), *claimed.TelegramChatID)
	assert.Equal(t, int64(3001), *claimed.TelegramMessageID)
	assert.Equal(t, int64(4001), *claimed.TelegramUserID)
}

func TestUpdateArticleTitlePersistsDiscoveredTitle(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)
	seedArticle(t, database, "ARTICLE001")

	repo := jobs.NewSQLiteRepository(database, newTestIDGenerator())

	err := repo.UpdateArticleTitle(t.Context(), "ARTICLE001", "Readable Article")
	require.NoError(t, err)

	var title sql.NullString

	err = database.QueryRow(`SELECT title FROM articles WHERE id = ?`, "ARTICLE001").Scan(&title)
	require.NoError(t, err)
	require.True(t, title.Valid)
	assert.Equal(t, "Readable Article", title.String)
}

func TestCompleteTerminalSuccessForTelegramJob(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)
	seedArticle(t, database, "ARTICLE001")
	seedTelegramJob(t, database, "JOB001", "ARTICLE001")

	const notificationID = "01ASB2XFCZJY7WHZ2FNRTMQJCA"

	repo := jobs.NewSQLiteRepository(database, newTestIDGenerator(notificationID))

	ctx := t.Context()

	// Claim first to move to running.
	claimed, err := repo.ClaimQueued(ctx)
	require.NoError(t, err)

	outcome := jobs.TerminalOutcome{Success: true}

	err = repo.CompleteTerminal(ctx, claimed, outcome)
	require.NoError(t, err)

	// Article must be ready.
	var articleStatus string
	var articleError sql.NullString

	err = database.QueryRowContext(ctx, `SELECT status, error_message FROM articles WHERE id = ?`, "ARTICLE001").
		Scan(&articleStatus, &articleError)
	require.NoError(t, err)
	assert.Equal(t, "ready", articleStatus)
	assert.False(t, articleError.Valid)

	// Job must be succeeded with completed_at and expires_at set.
	var jobStatus string
	var completedAt, expiresAt sql.NullString
	var jobError sql.NullString

	err = database.QueryRowContext(ctx,
		`SELECT status, error_message, completed_at, expires_at FROM jobs WHERE id = ?`,
		"JOB001",
	).Scan(&jobStatus, &jobError, &completedAt, &expiresAt)
	require.NoError(t, err)

	assert.Equal(t, jobs.StatusSucceeded, jobStatus)
	assert.False(t, jobError.Valid)
	assert.True(t, completedAt.Valid)
	assert.True(t, expiresAt.Valid)

	// Verify expires_at is approximately 14 days from now.
	expAt, parseErr := time.Parse(time.RFC3339Nano, expiresAt.String)
	require.NoError(t, parseErr)
	expectedExpiry := time.Now().UTC().Add(14 * 24 * time.Hour)
	assert.WithinDuration(t, expectedExpiry, expAt, 5*time.Second)

	// One pending notification must exist for the Telegram job.
	var notificationCount int

	err = database.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM notifications WHERE job_id = ? AND status = 'pending'`,
		"JOB001",
	).Scan(&notificationCount)
	require.NoError(t, err)
	assert.Equal(t, 1, notificationCount)

	var persistedNotificationID string

	err = database.QueryRowContext(ctx,
		`SELECT id FROM notifications WHERE job_id = ?`,
		"JOB001",
	).Scan(&persistedNotificationID)
	require.NoError(t, err)
	assert.Equal(t, notificationID, persistedNotificationID)
}

func TestCompleteTerminalFailureForTelegramJob(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)
	seedArticle(t, database, "ARTICLE001")
	seedTelegramJob(t, database, "JOB001", "ARTICLE001")

	const notificationID = "01ASB2XFCZJY7WHZ2FNRTMQJCB"

	repo := jobs.NewSQLiteRepository(database, newTestIDGenerator(notificationID))

	ctx := t.Context()

	claimed, err := repo.ClaimQueued(ctx)
	require.NoError(t, err)

	const errorText = "[ARC-013] Archivist could not summarize this article."

	outcome := jobs.TerminalOutcome{
		Success:      false,
		ErrorMessage: errorText,
	}

	err = repo.CompleteTerminal(ctx, claimed, outcome)
	require.NoError(t, err)

	// Article must be failed with ARC-coded error message.
	var articleStatus string
	var articleError sql.NullString

	err = database.QueryRowContext(ctx, `SELECT status, error_message FROM articles WHERE id = ?`, "ARTICLE001").
		Scan(&articleStatus, &articleError)
	require.NoError(t, err)
	assert.Equal(t, "failed", articleStatus)
	assert.True(t, articleError.Valid)
	assert.Equal(t, errorText, articleError.String)

	// Job must be failed with the same error message.
	var jobStatus string
	var jobError sql.NullString
	var completedAt, expiresAt sql.NullString

	err = database.QueryRowContext(ctx,
		`SELECT status, error_message, completed_at, expires_at FROM jobs WHERE id = ?`,
		"JOB001",
	).Scan(&jobStatus, &jobError, &completedAt, &expiresAt)
	require.NoError(t, err)

	assert.Equal(t, jobs.StatusFailed, jobStatus)
	assert.True(t, jobError.Valid)
	assert.Equal(t, errorText, jobError.String)
	assert.True(t, completedAt.Valid)
	assert.True(t, expiresAt.Valid)

	// Verify expires_at is approximately 14 days from now.
	expAt, parseErr := time.Parse(time.RFC3339Nano, expiresAt.String)
	require.NoError(t, parseErr)
	expectedExpiry := time.Now().UTC().Add(14 * 24 * time.Hour)
	assert.WithinDuration(t, expectedExpiry, expAt, 5*time.Second)

	// One pending notification must exist for the Telegram job.
	var notificationCount int

	err = database.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM notifications WHERE job_id = ? AND status = 'pending'`,
		"JOB001",
	).Scan(&notificationCount)
	require.NoError(t, err)
	assert.Equal(t, 1, notificationCount)

	var persistedNotificationID string

	err = database.QueryRowContext(ctx,
		`SELECT id FROM notifications WHERE job_id = ?`,
		"JOB001",
	).Scan(&persistedNotificationID)
	require.NoError(t, err)
	assert.Equal(t, notificationID, persistedNotificationID)
}

func TestCompleteTerminalFailurePreservesARCCodedError(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)
	seedArticle(t, database, "ARTICLE001")
	seedTelegramJob(t, database, "JOB001", "ARTICLE001")

	repo := jobs.NewSQLiteRepository(database, newTestIDGenerator("01ASB2XFCZJY7WHZ2FNRTMQJCC"))

	ctx := t.Context()

	claimed, err := repo.ClaimQueued(ctx)
	require.NoError(t, err)

	const arcError = "[ARC-003] The URL was not found."

	outcome := jobs.TerminalOutcome{
		Success:      false,
		ErrorMessage: arcError,
	}

	err = repo.CompleteTerminal(ctx, claimed, outcome)
	require.NoError(t, err)

	var articleError sql.NullString
	var jobError sql.NullString

	err = database.QueryRowContext(ctx, `SELECT error_message FROM articles WHERE id = ?`, "ARTICLE001").
		Scan(&articleError)
	require.NoError(t, err)

	err = database.QueryRowContext(ctx, `SELECT error_message FROM jobs WHERE id = ?`, "JOB001").
		Scan(&jobError)
	require.NoError(t, err)

	// Both article and job must preserve the full ARC-coded error text.
	assert.Equal(t, arcError, articleError.String)
	assert.Equal(t, arcError, jobError.String)
}

func TestCompleteTerminalRollsBackWhenNotificationIDGenerationFails(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)
	seedArticle(t, database, "ARTICLE001")
	seedTelegramJob(t, database, "JOB001", "ARTICLE001")

	repo := jobs.NewSQLiteRepository(database, &fakeIDGenerator{err: errIDGenerator})

	ctx := t.Context()

	claimed, err := repo.ClaimQueued(ctx)
	require.NoError(t, err)

	err = repo.CompleteTerminal(ctx, claimed, jobs.TerminalOutcome{Success: true})
	require.ErrorIs(t, err, errIDGenerator)
	require.ErrorContains(t, err, "jobs: failed to generate notification id")

	var articleStatus string
	var articleError sql.NullString

	err = database.QueryRowContext(ctx, `SELECT status, error_message FROM articles WHERE id = ?`, "ARTICLE001").
		Scan(&articleStatus, &articleError)
	require.NoError(t, err)
	assert.Equal(t, "queued", articleStatus)
	assert.False(t, articleError.Valid)

	var jobStatus string
	var jobError, completedAt, expiresAt sql.NullString

	err = database.QueryRowContext(ctx,
		`SELECT status, error_message, completed_at, expires_at FROM jobs WHERE id = ?`,
		"JOB001",
	).Scan(&jobStatus, &jobError, &completedAt, &expiresAt)
	require.NoError(t, err)

	assert.Equal(t, jobs.StatusRunning, jobStatus)
	assert.False(t, jobError.Valid)
	assert.False(t, completedAt.Valid)
	assert.False(t, expiresAt.Valid)

	var notificationCount int

	err = database.QueryRowContext(ctx, `SELECT COUNT(*) FROM notifications WHERE job_id = ?`, "JOB001").
		Scan(&notificationCount)
	require.NoError(t, err)
	assert.Equal(t, 0, notificationCount)
}

func TestCompleteTerminalSuccessForNonTelegramJobDoesNotCreateNotification(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)
	seedArticle(t, database, "ARTICLE001")
	seedNonTelegramJob(t, database, "JOB001", "ARTICLE001")

	repo := jobs.NewSQLiteRepository(database, newTestIDGenerator())

	ctx := t.Context()

	claimed, err := repo.ClaimQueued(ctx)
	require.NoError(t, err)

	// Non-Telegram jobs have no TelegramChatID or TelegramMessageID.
	assert.Nil(t, claimed.TelegramChatID)
	assert.Nil(t, claimed.TelegramMessageID)
	assert.False(t, claimed.HasTelegramOrigin())

	outcome := jobs.TerminalOutcome{Success: true}

	err = repo.CompleteTerminal(ctx, claimed, outcome)
	require.NoError(t, err)

	// No notification must be created for non-Telegram jobs.
	var notificationCount int

	err = database.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM notifications WHERE job_id = ?`,
		"JOB001",
	).Scan(&notificationCount)
	require.NoError(t, err)
	assert.Equal(t, 0, notificationCount)
}

func TestCompleteTerminalFailureForNonTelegramJobDoesNotCreateNotification(t *testing.T) {
	database := openTestDB(t)
	seedUser(t, database)
	seedArticle(t, database, "ARTICLE001")
	seedNonTelegramJob(t, database, "JOB001", "ARTICLE001")

	repo := jobs.NewSQLiteRepository(database, newTestIDGenerator())

	ctx := t.Context()

	claimed, err := repo.ClaimQueued(ctx)
	require.NoError(t, err)

	outcome := jobs.TerminalOutcome{
		Success:      false,
		ErrorMessage: "[ARC-999] Archivist could not process the URL.",
	}

	err = repo.CompleteTerminal(ctx, claimed, outcome)
	require.NoError(t, err)

	var notificationCount int

	err = database.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM notifications WHERE job_id = ?`,
		"JOB001",
	).Scan(&notificationCount)
	require.NoError(t, err)
	assert.Equal(t, 0, notificationCount)
}
