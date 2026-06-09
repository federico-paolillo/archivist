package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"codeberg.org/federico-paolillo/archivist/internal/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
	UserID    string
}

// Enqueuer defines the worker persistence contract for imperative URL queueing.
type Enqueuer interface {
	// EnqueueURL creates a queued article and non-Telegram article-processing job
	// for DefaultUserID. It does not create users or notifications.
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

	// CompleteJobFailure atomically fails only the job row and inserts one
	// pending notification row when the job has Telegram origin metadata. It is
	// used for corrupt work where mutating the referenced article is unsafe.
	CompleteJobFailure(ctx context.Context, job *Job, errorMessage string) error

	// ArticleURL returns the original_url for an article owned by userID.
	ArticleURL(ctx context.Context, articleID string, userID string) (string, error)

	// UpdateCanonicalURL sets articles.canonical_url to canonicalURL for an article owned by userID.
	UpdateCanonicalURL(ctx context.Context, articleID string, userID string, canonicalURL string) error

	// UpdateArticleTitle sets articles.title to title for an article owned by userID.
	UpdateArticleTitle(ctx context.Context, articleID string, userID string, title string) error
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

// EnqueueURL inserts a queued article and queued non-Telegram job for DefaultUserID.
//
//nolint:funlen,nonamedreturns,spancheck // Deferred span completion records the returned error.
func (r *SQLiteRepository) EnqueueURL(ctx context.Context, rawURL string) (result *EnqueueResult, err error) {
	ctx, span := observability.Tracer().Start(
		ctx,
		"worker.jobs.enqueue_url",
		trace.WithAttributes(attribute.String("url", rawURL)),
	)
	defer func() {
		observability.EndSpan(span, err)
	}()

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

	userID := DefaultUserID
	span.SetAttributes(observability.UserAttribute(userID))

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
		userID,
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

	span.SetAttributes(observability.JobUserAttributes(articleID, jobID, userID)...)

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO jobs (id, user_id, article_id, type, status, created_at)
		 VALUES (?, ?, ?, ?, 'queued', ?)`,
		jobID,
		userID,
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

	return &EnqueueResult{ArticleID: articleID, JobID: jobID, UserID: userID}, nil
}

func ensureDefaultUserExists(ctx context.Context, tx *sql.Tx) error {
	var exists int

	err := tx.QueryRowContext(ctx, `SELECT 1 FROM users WHERE id = ?`, DefaultUserID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf(
			"jobs: default user %s is missing; run Gateway auth bootstrap before enqueueing URLs",
			DefaultUserID,
		)
	}

	if err != nil {
		return fmt.Errorf("jobs: failed to check default user: %w", err)
	}

	return nil
}

// ClaimQueued atomically claims one queued job using UPDATE...RETURNING.
//
//nolint:nonamedreturns // Deferred span completion records the returned error.
func (r *SQLiteRepository) ClaimQueued(ctx context.Context) (job *Job, err error) {
	ctx, span := observability.Tracer().Start(ctx, "worker.jobs.claim", trace.WithSpanKind(trace.SpanKindConsumer))
	defer func() {
		observability.EndSpan(span, err)
	}()

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
		    traceparent,
		    tracestate,
		    created_at,
		    started_at,
		    completed_at,
		    expires_at
	`

	row := r.db.QueryRowContext(ctx, query, now.Format(time.RFC3339Nano))

	job, err = scanJob(row)
	if err == nil {
		span.SetAttributes(observability.JobUserAttributes(job.ArticleID, job.ID, job.UserID)...)
	}

	return job, err
}

// CompleteTerminal persists the terminal state of a job in one transaction.
//
//nolint:nonamedreturns // Deferred span completion records the returned error.
func (r *SQLiteRepository) CompleteTerminal(ctx context.Context, job *Job, outcome TerminalOutcome) (err error) {
	ctx, span := observability.Tracer().Start(
		ctx,
		"worker.jobs.complete_terminal",
		trace.WithAttributes(observability.JobUserAttributes(job.ArticleID, job.ID, job.UserID)...),
	)
	defer func() {
		observability.EndSpan(span, err)
	}()

	span.SetAttributes(attribute.Bool("success", outcome.Success))

	if !outcome.Success {
		span.SetStatus(codes.Error, outcome.ErrorMessage)
	}

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

// CompleteJobFailure persists a job-only terminal failure without mutating the article.
//
//nolint:nonamedreturns,spancheck // Deferred span completion records the returned error.
func (r *SQLiteRepository) CompleteJobFailure(ctx context.Context, job *Job, errorMessage string) (err error) {
	ctx, span := observability.Tracer().Start(
		ctx,
		"worker.jobs.complete_job_failure",
		trace.WithAttributes(observability.JobUserAttributes(job.ArticleID, job.ID, job.UserID)...),
	)
	defer func() {
		observability.EndSpan(span, err)
	}()

	span.SetStatus(codes.Error, errorMessage)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("jobs: failed to begin job failure transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now().UTC()
	expiresAt := now.Add(14 * 24 * time.Hour)

	err = applyJobTerminal(ctx, tx, job, TerminalOutcome{Success: false, ErrorMessage: errorMessage}, now, expiresAt)
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
		return fmt.Errorf("jobs: failed to commit job failure transaction: %w", err)
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

	result, err := tx.ExecContext(
		ctx,
		`UPDATE articles
		 SET    status = ?,
		        error_message = ?
		 WHERE  id = ?
		 AND    user_id = ?`,
		articleStatus,
		articleError,
		job.ArticleID,
		job.UserID,
	)
	if err != nil {
		return fmt.Errorf("jobs: failed to update article terminal state: %w", err)
	}

	return expectOneOwnedRow(result, "update article terminal state", job.ArticleID, job.UserID)
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

	result, err := tx.ExecContext(
		ctx,
		`UPDATE jobs
		 SET    status        = ?,
		        error_message = ?,
		        completed_at  = ?,
		        expires_at    = ?
		 WHERE  id = ?
		 AND    article_id = ?
		 AND    user_id = ?`,
		jobStatus,
		jobError,
		now.Format(time.RFC3339Nano),
		expiresAt.Format(time.RFC3339Nano),
		job.ID,
		job.ArticleID,
		job.UserID,
	)
	if err != nil {
		return fmt.Errorf("jobs: failed to update job terminal state: %w", err)
	}

	return expectOneOwnedRow(result, "update job terminal state", job.ArticleID, job.UserID)
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
	traceParent       sql.NullString
	traceState        sql.NullString

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
		&raw.traceParent,
		&raw.traceState,
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
	applyTraceFields(j, raw)

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

func applyTraceFields(j *Job, raw *jobRaw) {
	if raw.traceParent.Valid {
		j.TraceParent = &raw.traceParent.String
	}

	if raw.traceState.Valid {
		j.TraceState = &raw.traceState.String
	}
}

type jobTimestamps struct {
	createdAt   time.Time
	startedAt   *time.Time
	completedAt *time.Time
	expiresAt   *time.Time
}

const dotnetTimestamp = "2006-01-02 15:04:05.999999999-07:00"

var jobTimestampLayouts = []string{
	time.RFC3339Nano,
	dotnetTimestamp,
}

func parseJobTimestamps(raw *jobRaw) (*jobTimestamps, error) {
	createdAt, err := parseJobTimestamp(raw.createdAtStr, "created_at")
	if err != nil {
		return nil, err
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

func parseJobTimestamp(value string, field string) (time.Time, error) {
	var parseErr error

	for _, layout := range jobTimestampLayouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}

		parseErr = err
	}

	return time.Time{}, fmt.Errorf("jobs: failed to parse %s: %w", field, parseErr)
}

// parseNullableTime parses a supported timestamp string from a nullable column.
// The second return value is false when the column is NULL; the third is non-nil when parsing fails.
func parseNullableTime(ns sql.NullString, field string) (time.Time, bool, error) {
	var zero time.Time

	if !ns.Valid {
		return zero, false, nil
	}

	parsed, parseErr := parseJobTimestamp(ns.String, field)
	if parseErr != nil {
		return zero, false, parseErr
	}

	return parsed, true, nil
}

// ArticleURL returns the original_url for an article with the given ID and user ID.
//
//nolint:nonamedreturns,spancheck // Deferred span completion records the returned error.
func (r *SQLiteRepository) ArticleURL(ctx context.Context, articleID string, userID string) (originalURL string, err error) {
	ctx, span := observability.Tracer().Start(ctx, "worker.jobs.article_url", trace.WithAttributes(
		attribute.String("article_id", articleID),
		observability.UserAttribute(userID),
	))
	defer func() {
		observability.EndSpan(span, err)
	}()

	err = r.db.QueryRowContext(
		ctx,
		`SELECT original_url FROM articles WHERE id = ? AND user_id = ?`,
		articleID,
		userID,
	).Scan(&originalURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("jobs: article %s is not owned by user %s: %w", articleID, userID, ErrOwnershipMismatch)
	}

	if err != nil {
		return "", fmt.Errorf("jobs: failed to load article URL for %s: %w", articleID, err)
	}

	return originalURL, nil
}

// UpdateCanonicalURL sets articles.canonical_url for the given articleID and userID.
//
//nolint:nonamedreturns,spancheck // Deferred span completion records the returned error.
func (r *SQLiteRepository) UpdateCanonicalURL(ctx context.Context, articleID string, userID string, canonicalURL string) (err error) {
	ctx, span := observability.Tracer().Start(ctx, "worker.jobs.update_canonical_url", trace.WithAttributes(
		attribute.String("article_id", articleID),
		observability.UserAttribute(userID),
		attribute.String("url", canonicalURL),
	))
	defer func() {
		observability.EndSpan(span, err)
	}()

	result, err := r.db.ExecContext(
		ctx,
		`UPDATE articles SET canonical_url = ? WHERE id = ? AND user_id = ?`,
		canonicalURL,
		articleID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("jobs: failed to update canonical URL for article %s: %w", articleID, err)
	}

	return expectOneOwnedRow(result, "update canonical URL", articleID, userID)
}

// UpdateArticleTitle sets articles.title for the given articleID and userID.
//
//nolint:nonamedreturns,spancheck // Deferred span completion records the returned error.
func (r *SQLiteRepository) UpdateArticleTitle(ctx context.Context, articleID string, userID string, title string) (err error) {
	ctx, span := observability.Tracer().Start(ctx, "worker.jobs.update_article_title", trace.WithAttributes(
		attribute.String("article_id", articleID),
		observability.UserAttribute(userID),
	))
	defer func() {
		observability.EndSpan(span, err)
	}()

	result, err := r.db.ExecContext(
		ctx,
		`UPDATE articles SET title = ? WHERE id = ? AND user_id = ?`,
		title,
		articleID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("jobs: failed to update title for article %s: %w", articleID, err)
	}

	return expectOneOwnedRow(result, "update article title", articleID, userID)
}

func expectOneOwnedRow(result sql.Result, operation string, articleID string, userID string) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("jobs: failed to inspect %s row count: %w", operation, err)
	}

	// Article and job ids are primary keys, so the practical failure case here is
	// zero rows: the row is missing or the ownership-scoped WHERE clause did not match.
	if affected == 1 {
		return nil
	}

	return fmt.Errorf("jobs: %s for article %s and user %s affected %d rows: %w",
		operation,
		articleID,
		userID,
		affected,
		ErrOwnershipMismatch,
	)
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
