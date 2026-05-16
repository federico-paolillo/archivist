package summary

import "context"

// Provider identifies the summarizer provider used.
type Provider string

const (
	ProviderAnthropic Provider = "anthropic"
)

// SummarizerRequest carries the Markdown source to summarize and optional
// article context metadata that orchestration uses for structured logging.
type SummarizerRequest struct {
	MarkdownSource string

	// Optional metadata for orchestration logging. Populated by the caller;
	// ignored by the adapter.
	ArticleID string
	JobID     string
	URL       string
}

// SummarizerOutput holds a successful summarization attempt.
type SummarizerOutput struct {
	Summary   string
	RequestID string
}

// SummarizerService is the provider-neutral summarization contract.
// Implementations must not expose provider-specific SDK types.
type SummarizerService interface {
	Provider() Provider
	Summarize(ctx context.Context, req SummarizerRequest) (SummarizerOutput, error)
}
