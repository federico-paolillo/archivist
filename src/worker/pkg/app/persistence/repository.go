package persistence

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const (
	personalUserID      = "01ASB2XFCZJY7WHZ2FNRTMQJCT"
	statusQueued        = "queued"
	statusRunning       = "running"
	statusReady         = "ready"
	statusFailed        = "failed"
	statusSucceeded     = "succeeded"
	notificationPending = "pending"
	terminalJobTTL      = 14 * 24 * time.Hour
)

var (
	ErrNoQueuedJob = errors.New("persistence: no queued job")
	errEmptyError  = errors.New("persistence: terminal failure requires an error message")
)

type IDGenerator interface {
	NewID(now time.Time) (string, error)
}

type SystemIDGenerator struct{}

func (SystemIDGenerator) NewID(now time.Time) (string, error) {
	var bytes [16]byte

	millis := uint64(now.UnixMilli())

	bytes[0] = byte((millis >> 40) & 0xff)
	bytes[1] = byte((millis >> 32) & 0xff)
	bytes[2] = byte((millis >> 24) & 0xff)
	bytes[3] = byte((millis >> 16) & 0xff)
	bytes[4] = byte((millis >> 8) & 0xff)
	bytes[5] = byte(millis & 0xff)

	_, err := rand.Read(bytes[6:])
	if err != nil {
		return "", fmt.Errorf("persistence: generate ulid: %w", err)
	}

	return encodeULID(bytes), nil
}

type Repository struct {
	db  *sql.DB
	ids IDGenerator
}

type Job struct {
	ID                string
	UserID            string
	ArticleID         string
	Type              string
	TelegramUpdateID  sql.NullInt64
	TelegramChatID    sql.NullInt64
	TelegramMessageID sql.NullInt64
	TelegramUserID    sql.NullInt64
}

func NewRepository(db *sql.DB, ids IDGenerator) *Repository {
	return &Repository{
		db:  db,
		ids: ids,
	}
}

func (r *Repository) ClaimNextQueuedJob(ctx context.Context, now time.Time) (*Job, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE jobs
		SET status = ?, started_at = ?
		WHERE id = (
			SELECT id FROM jobs
			WHERE status = ?
			ORDER BY created_at, id
			LIMIT 1
		)
		RETURNING id, user_id, article_id, type, telegram_update_id, telegram_chat_id, telegram_message_id, telegram_user_id;
	`, statusRunning, formatTime(now), statusQueued)

	var job Job

	err := row.Scan(
		&job.ID,
		&job.UserID,
		&job.ArticleID,
		&job.Type,
		&job.TelegramUpdateID,
		&job.TelegramChatID,
		&job.TelegramMessageID,
		&job.TelegramUserID,
	)
	if err == nil {
		return &job, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoQueuedJob
	}

	return nil, fmt.Errorf("persistence: claim queued job: %w", err)
}

func (r *Repository) CompleteJobSucceeded(ctx context.Context, jobID string, now time.Time) error {
	return r.completeJob(ctx, jobID, statusSucceeded, statusReady, "", now)
}

func (r *Repository) CompleteJobFailed(ctx context.Context, jobID string, errorMessage string, now time.Time) error {
	if errorMessage == "" {
		return errEmptyError
	}

	return r.completeJob(ctx, jobID, statusFailed, statusFailed, errorMessage, now)
}

func (r *Repository) ExpiredTerminalJobIDs(ctx context.Context, before time.Time) ([]string, error) {
	return queryIDs(ctx, r.db, `
		SELECT id FROM jobs
		WHERE status IN ('succeeded', 'failed')
		AND expires_at IS NOT NULL
		AND expires_at <= ?
		ORDER BY expires_at, id;
	`, formatTime(before))
}

func (r *Repository) ExpiredNotificationIDs(ctx context.Context, before time.Time) ([]string, error) {
	return queryIDs(ctx, r.db, `
		SELECT id FROM notifications
		WHERE status IN ('sent', 'failed')
		AND expires_at IS NOT NULL
		AND expires_at <= ?
		ORDER BY expires_at, id;
	`, formatTime(before))
}

func (r *Repository) completeJob(
	ctx context.Context,
	jobID string,
	jobStatus string,
	articleStatus string,
	errorMessage string,
	now time.Time,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("persistence: begin terminal transaction: %w", err)
	}

	defer rollback(tx)

	var articleID string

	err = tx.QueryRowContext(ctx, `
		SELECT article_id FROM jobs
		WHERE id = ? AND status = ?;
	`, jobID, statusRunning).Scan(&articleID)
	if err != nil {
		return noRowsAsNotFound(err, "persistence: find running job")
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE articles
		SET status = ?, error_message = ?
		WHERE id = ?;
	`, articleStatus, nullableString(errorMessage), articleID)
	if err != nil {
		return fmt.Errorf("persistence: update terminal article: %w", err)
	}

	completed := formatTime(now)

	_, err = tx.ExecContext(ctx, `
		UPDATE jobs
		SET status = ?, error_message = ?, completed_at = ?, expires_at = ?
		WHERE id = ?;
	`, jobStatus, nullableString(errorMessage), completed, formatTime(now.Add(terminalJobTTL)), jobID)
	if err != nil {
		return fmt.Errorf("persistence: update terminal job: %w", err)
	}

	notificationID, err := r.ids.NewID(now)
	if err != nil {
		return fmt.Errorf("persistence: generate notification id: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO notifications (id, job_id, status, created_at)
		VALUES (?, ?, ?, ?);
	`, notificationID, jobID, notificationPending, completed)
	if err != nil {
		return fmt.Errorf("persistence: insert terminal notification: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("persistence: commit terminal transaction: %w", err)
	}

	return nil
}

func queryIDs(ctx context.Context, db *sql.DB, query string, args ...any) ([]string, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("persistence: query ids: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var ids []string

	for rows.Next() {
		var id string

		err = rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("persistence: scan id: %w", err)
		}

		ids = append(ids, id)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("persistence: iterate ids: %w", err)
	}

	return ids, nil
}

func rollback(tx *sql.Tx) {
	_ = tx.Rollback()
}

func noRowsAsNotFound(err error, operation string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s: not found", operation)
	}

	return fmt.Errorf("%s: %w", operation, err)
}

func nullableString(value string) sql.NullString {
	return sql.NullString{
		String: value,
		Valid:  value != "",
	}
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func encodeULID(bytes [16]byte) string {
	const alphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

	chars := make([]byte, 26)
	chars[0] = alphabet[(bytes[0]&224)>>5]
	chars[1] = alphabet[bytes[0]&31]
	chars[2] = alphabet[(bytes[1]&248)>>3]
	chars[3] = alphabet[((bytes[1]&7)<<2)|((bytes[2]&192)>>6)]
	chars[4] = alphabet[(bytes[2]&62)>>1]
	chars[5] = alphabet[((bytes[2]&1)<<4)|((bytes[3]&240)>>4)]
	chars[6] = alphabet[((bytes[3]&15)<<1)|((bytes[4]&128)>>7)]
	chars[7] = alphabet[(bytes[4]&124)>>2]
	chars[8] = alphabet[((bytes[4]&3)<<3)|((bytes[5]&224)>>5)]
	chars[9] = alphabet[bytes[5]&31]
	chars[10] = alphabet[(bytes[6]&248)>>3]
	chars[11] = alphabet[((bytes[6]&7)<<2)|((bytes[7]&192)>>6)]
	chars[12] = alphabet[(bytes[7]&62)>>1]
	chars[13] = alphabet[((bytes[7]&1)<<4)|((bytes[8]&240)>>4)]
	chars[14] = alphabet[((bytes[8]&15)<<1)|((bytes[9]&128)>>7)]
	chars[15] = alphabet[(bytes[9]&124)>>2]
	chars[16] = alphabet[((bytes[9]&3)<<3)|((bytes[10]&224)>>5)]
	chars[17] = alphabet[bytes[10]&31]
	chars[18] = alphabet[(bytes[11]&248)>>3]
	chars[19] = alphabet[((bytes[11]&7)<<2)|((bytes[12]&192)>>6)]
	chars[20] = alphabet[(bytes[12]&62)>>1]
	chars[21] = alphabet[((bytes[12]&1)<<4)|((bytes[13]&240)>>4)]
	chars[22] = alphabet[((bytes[13]&15)<<1)|((bytes[14]&128)>>7)]
	chars[23] = alphabet[(bytes[14]&124)>>2]
	chars[24] = alphabet[((bytes[14]&3)<<3)|((bytes[15]&224)>>5)]
	chars[25] = alphabet[bytes[15]&31]

	return string(chars)
}
