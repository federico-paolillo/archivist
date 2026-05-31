package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// TerminalOutcome carries the result of a completed job to be persisted.
type TerminalOutcome struct {
	// Success is true when the job completed successfully, false on failure.
	Success bool

	// ErrorMessage is the user-facing error text on failure.
	// For article-processing failures, this must be an ARC-coded public message
	// e.g. "[ARC-013] Archivist could not summarize this article."
	// It is ignored when Success is true.
	ErrorMessage string
}

// EnqueueResult carries identifiers for records created by EnqueueURL.
type EnqueueResult struct {
	ArticleID string
	JobID     string
}

// Enqueuer defines the worker persistence contract for imperative URL queueing.
type Enqueuer interface {
	// EnqueueURL creates a queued article and non-Telegram article-processing job
	// for the fixed personal user. It does not create users or notifications.
	EnqueueURL(ctx context.Context, rawURL string) (*EnqueueResult, error)
}

// Repository defines the worker persistence contract for jobs.
type Repository interface {
	// ClaimQueued atomically claims one queued job, changes its status to
	// running, and returns the claimed job. Returns sql.ErrNoRows when no
	// queued job is available.
	ClaimQueued(ctx context.Context) (*Job, error)

	// CompleteTerminal atomically:
	//   - updates the article status to ready or failed,
	//   - updates the job status to succeeded or failed,
	//   - sets expires_at on the job to 14 days after now,
	//   - inserts one pending notification row (only when the job has Telegram
	//     origin metadata).
	// All three writes (article, job, notification) happen in one transaction.
	CompleteTerminal(ctx context.Context, job *Job, outcome TerminalOutcome) error

	// ArticleURL returns the original_url for the article associated with the job.
	ArticleURL(ctx context.Context, articleID string) (string, error)

	// UpdateCanonicalURL sets articles.canonical_url to canonicalURL for the given articleID.
	UpdateCanonicalURL(ctx context.Context, articleID string, canonicalURL string) error

	// UpdateArticleTitle sets articles.title to title for the given articleID.
	UpdateArticleTitle(ctx context.Context, articleID string, title string) error
}

// SQLiteRepository is the SQLite-backed implementation of Repository.
type SQLiteRepository struct {
	db  *sql.DB
	ids IDGenerator
}

// NewSQLiteRepository creates a new SQLiteRepository backed by the given database.
func NewSQLiteRepository(database *sql.DB, ids IDGenerator) *SQLiteRepository {
	return &SQLiteRepository{
		db:  database,
		ids: ids,
	}
}

// EnqueueURL inserts a queued article and queued non-Telegram job for the fixed personal user.
func (r *SQLiteRepository) EnqueueURL(ctx context.Context, rawURL string) (*EnqueueResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("jobs: failed to begin enqueue transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	err = ensureDefaultUserExists(ctx, tx)
	if err != nil {
		return nil, err
	}

	articleID, err := r.ids.NewID()
	if err != nil {
		return nil, fmt.Errorf("jobs: failed to generate article id: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO articles (id, user_id, original_url, status, created_at)
		 VALUES (?, ?, ?, 'queued', ?)`,
		articleID,
		DefaultUserID,
		rawURL,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("jobs: failed to insert queued article: %w", err)
	}

	jobID, err := r.ids.NewID()
	if err != nil {
		return nil, fmt.Errorf("jobs: failed to generate job id: %w", err)
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO jobs (id, user_id, article_id, type, status, created_at)
		 VALUES (?, ?, ?, ?, 'queued', ?)`,
		jobID,
		DefaultUserID,
		articleID,
		TypeArticleProcessing,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("jobs: failed to insert queued job: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("jobs: failed to commit enqueue transaction: %w", err)
	}

	return &EnqueueResult{ArticleID: articleID, JobID: jobID}, nil
}

func ensureDefaultUserExists(ctx context.Context, tx *sql.Tx) error {
	var exists int

	err := tx.QueryRowContext(
		ctx,
		`SELECT 1 FROM users WHERE id = ?`,
		DefaultUserID,
	).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("jobs: default user %s is missing; run Gateway auth bootstrap before enqueueing URLs", DefaultUserID)
	}

	if err != nil {
		return fmt.Errorf("jobs: failed to check default user: %w", err)
	}

	return nil
}

// ClaimQueued atomically claims one queued job using UPDATE...RETURNING.
func (r *SQLiteRepository) ClaimQueued(ctx context.Context) (*Job, error) {
	now := time.Now().UTC()

	query := `
		UPDATE jobs
		SET    status     = 'running',
		       started_at = ?
		WHERE  id = (
		    SELECT id FROM jobs
		    WHERE  status = 'queued'
		    AND    type = 'article-processing'
		    AND    EXISTS (
		        SELECT 1
		        FROM   articles
		        WHERE  articles.id = jobs.article_id
		    )
		    ORDER  BY created_at ASC
		    LIMIT  1
		)
		RETURNING
		    id,
		    user_id,
		    article_id,
		    type,
		    status,
		    telegram_update_id,
		    telegram_chat_id,
		    telegram_message_id,
		    telegram_user_id,
		    error_message,
		    created_at,
		    started_at,
		    completed_at,
		    expires_at
	`

	row := r.db.QueryRowContext(ctx, query, now.Format(time.RFC3339Nano))

	return scanJob(row)
}

// CompleteTerminal persists the terminal state of a job in one transaction.
func (r *SQLiteRepository) CompleteTerminal(ctx context.Context, job *Job, outcome TerminalOutcome) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("jobs: failed to begin terminal transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now().UTC()
	expiresAt := now.Add(14 * 24 * time.Hour)

	err = applyArticleTerminal(ctx, tx, job, outcome)
	if err != nil {
		return err
	}

	err = applyJobTerminal(ctx, tx, job, outcome, now, expiresAt)
	if err != nil {
		return err
	}

	if job.HasTelegramOrigin() {
		err = r.insertPendingNotification(ctx, tx, job, now)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("jobs: failed to commit terminal transaction: %w", err)
	}

	return nil
}

func applyArticleTerminal(ctx context.Context, tx *sql.Tx, job *Job, outcome TerminalOutcome) error {
	var articleStatus string

	var articleError *string

	if outcome.Success {
		articleStatus = "ready"
	} else {
		articleStatus = "failed"
		articleError = &outcome.ErrorMessage
	}

	_, err := tx.ExecContext(
		ctx,
		`UPDATE articles SET status = ?, error_message = ? WHERE id = ?`,
		articleStatus,
		articleError,
		job.ArticleID,
	)
	if err != nil {
		return fmt.Errorf("jobs: failed to update article terminal state: %w", err)
	}

	return nil
}

func applyJobTerminal(
	ctx context.Context,
	tx *sql.Tx,
	job *Job,
	outcome TerminalOutcome,
	now time.Time,
	expiresAt time.Time,
) error {
	var jobStatus string

	var jobError *string

	if outcome.Success {
		jobStatus = "succeeded"
	} else {
		jobStatus = "failed"
		jobError = &outcome.ErrorMessage
	}

	_, err := tx.ExecContext(
		ctx,
		`UPDATE jobs
		 SET    status        = ?,
		        error_message = ?,
		        completed_at  = ?,
		        expires_at    = ?
		 WHERE  id = ?`,
		jobStatus,
		jobError,
		now.Format(time.RFC3339Nano),
		expiresAt.Format(time.RFC3339Nano),
		job.ID,
	)
	if err != nil {
		return fmt.Errorf("jobs: failed to update job terminal state: %w", err)
	}

	return nil
}

// scanJob scans one row from an UPDATE...RETURNING or SELECT of the jobs table.
func scanJob(row *sql.Row) (*Job, error) {
	raw, err := scanJobRaw(row)
	if err != nil {
		return nil, err
	}

	return parseJobRaw(raw)
}

type jobRaw struct {
	id        string
	userID    string
	articleID string
	jobType   string
	status    string

	telegramUpdateID  sql.NullInt64
	telegramChatID    sql.NullInt64
	telegramMessageID sql.NullInt64
	telegramUserID    sql.NullInt64
	errorMessage      sql.NullString

	createdAtStr   string
	startedAtStr   sql.NullString
	completedAtStr sql.NullString
	expiresAtStr   sql.NullString
}

func scanJobRaw(row *sql.Row) (*jobRaw, error) {
	var raw jobRaw

	err := row.Scan(
		&raw.id,
		&raw.userID,
		&raw.articleID,
		&raw.jobType,
		&raw.status,
		&raw.telegramUpdateID,
		&raw.telegramChatID,
		&raw.telegramMessageID,
		&raw.telegramUserID,
		&raw.errorMessage,
		&raw.createdAtStr,
		&raw.startedAtStr,
		&raw.completedAtStr,
		&raw.expiresAtStr,
	)
	if err != nil {
		return nil, fmt.Errorf("jobs: failed to scan job row: %w", err)
	}

	return &raw, nil
}

func parseJobRaw(raw *jobRaw) (*Job, error) {
	j := &Job{
		ID:        raw.id,
		UserID:    raw.userID,
		ArticleID: raw.articleID,
		Type:      raw.jobType,
		Status:    raw.status,
	}

	applyTelegramFields(j, raw)
	applyErrorMessage(j, raw)

	timestamps, err := parseJobTimestamps(raw)
	if err != nil {
		return nil, err
	}

	j.CreatedAt = timestamps.createdAt
	j.StartedAt = timestamps.startedAt
	j.CompletedAt = timestamps.completedAt
	j.ExpiresAt = timestamps.expiresAt

	return j, nil
}

func applyTelegramFields(j *Job, raw *jobRaw) {
	if raw.telegramUpdateID.Valid {
		j.TelegramUpdateID = &raw.telegramUpdateID.Int64
	}

	if raw.telegramChatID.Valid {
		j.TelegramChatID = &raw.telegramChatID.Int64
	}

	if raw.telegramMessageID.Valid {
		j.TelegramMessageID = &raw.telegramMessageID.Int64
	}

	if raw.telegramUserID.Valid {
		j.TelegramUserID = &raw.telegramUserID.Int64
	}
}

func applyErrorMessage(j *Job, raw *jobRaw) {
	if raw.errorMessage.Valid {
		j.ErrorMessage = &raw.errorMessage.String
	}
}

type jobTimestamps struct {
	createdAt   time.Time
	startedAt   *time.Time
	completedAt *time.Time
	expiresAt   *time.Time
}

func parseJobTimestamps(raw *jobRaw) (*jobTimestamps, error) {
	createdAt, err := time.Parse(time.RFC3339Nano, raw.createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("jobs: failed to parse created_at: %w", err)
	}

	ts := &jobTimestamps{createdAt: createdAt}

	startedAt, startedOk, startedErr := parseNullableTime(raw.startedAtStr, "started_at")
	if startedErr != nil {
		return nil, startedErr
	}

	if startedOk {
		ts.startedAt = &startedAt
	}

	completedAt, completedOk, completedErr := parseNullableTime(raw.completedAtStr, "completed_at")
	if completedErr != nil {
		return nil, completedErr
	}

	if completedOk {
		ts.completedAt = &completedAt
	}

	expiresAt, expiresOk, expiresErr := parseNullableTime(raw.expiresAtStr, "expires_at")
	if expiresErr != nil {
		return nil, expiresErr
	}

	if expiresOk {
		ts.expiresAt = &expiresAt
	}

	return ts, nil
}

// parseNullableTime parses an RFC3339Nano string from a nullable column.
// The second return value is false when the column is NULL; the third is non-nil when parsing fails.
func parseNullableTime(ns sql.NullString, field string) (time.Time, bool, error) {
	var zero time.Time

	if !ns.Valid {
		return zero, false, nil
	}

	parsed, parseErr := time.Parse(time.RFC3339Nano, ns.String)
	if parseErr != nil {
		return zero, false, fmt.Errorf("jobs: failed to parse %s: %w", field, parseErr)
	}

	return parsed, true, nil
}

// ArticleURL returns the original_url for the article with the given ID.
func (r *SQLiteRepository) ArticleURL(ctx context.Context, articleID string) (string, error) {
	var originalURL string

	err := r.db.QueryRowContext(
		ctx,
		`SELECT original_url FROM articles WHERE id = ?`,
		articleID,
	).Scan(&originalURL)
	if err != nil {
		return "", fmt.Errorf("jobs: failed to load article URL for %s: %w", articleID, err)
	}

	return originalURL, nil
}

// UpdateCanonicalURL sets articles.canonical_url for the given articleID.
func (r *SQLiteRepository) UpdateCanonicalURL(ctx context.Context, articleID string, canonicalURL string) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE articles SET canonical_url = ? WHERE id = ?`,
		canonicalURL,
		articleID,
	)
	if err != nil {
		return fmt.Errorf("jobs: failed to update canonical URL for article %s: %w", articleID, err)
	}

	return nil
}

// UpdateArticleTitle sets articles.title for the given articleID.
func (r *SQLiteRepository) UpdateArticleTitle(ctx context.Context, articleID string, title string) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE articles SET title = ? WHERE id = ?`,
		title,
		articleID,
	)
	if err != nil {
		return fmt.Errorf("jobs: failed to update title for article %s: %w", articleID, err)
	}

	return nil
}

func (r *SQLiteRepository) insertPendingNotification(ctx context.Context, tx *sql.Tx, job *Job, now time.Time) error {
	id, err := r.ids.NewID()
	if err != nil {
		return fmt.Errorf("jobs: failed to generate notification id: %w", err)
	}

	// Notifications expire 7 days after completion per REQ-029.
	notificationExpiresAt := now.Add(7 * 24 * time.Hour)

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO notifications (id, job_id, status, created_at, expires_at)
		 VALUES (?, ?, 'pending', ?, ?)`,
		id,
		job.ID,
		now.Format(time.RFC3339Nano),
		notificationExpiresAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("jobs: failed to insert pending notification: %w", err)
	}

	return nil
}
