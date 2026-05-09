package jobs

import "time"

// Status values for a job. Only the four v0 states are valid.
const (
	StatusQueued    = "queued"
	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
)

// TypeArticleProcessing is the v0 job type for article processing.
const TypeArticleProcessing = "article_processing"

// Job holds the state of a worker job claimed from SQLite.
type Job struct {
	ID        string
	UserID    string
	ArticleID string
	Type      string
	Status    string

	// Telegram origin metadata. All nullable in SQLite; zero value means absent.
	TelegramUpdateID  *int64
	TelegramChatID    *int64
	TelegramMessageID *int64
	TelegramUserID    *int64

	ErrorMessage *string

	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	ExpiresAt   *time.Time
}

// HasTelegramOrigin returns true when the job carries Telegram origin metadata
// (telegram_chat_id and telegram_message_id are both set), meaning the gateway
// expects a terminal Telegram reply through a notification row.
func (j *Job) HasTelegramOrigin() bool {
	return j.TelegramChatID != nil && j.TelegramMessageID != nil
}
