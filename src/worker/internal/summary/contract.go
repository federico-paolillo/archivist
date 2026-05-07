package summary

import "context"

// Provider identifies the summarizer provider used.
type Provider string

const (
	ProviderAnthropic Provider = "anthropic"
)

// ResultStatus indicates the outcome of a summarization attempt.
type ResultStatus string

const (
	ResultStatusSuccess ResultStatus = "success"
	ResultStatusFailure ResultStatus = "failure"
)

// ErrorCode maps to ARC error codes for user-facing persisted failures.
type ErrorCode string

const (
	// ErrorCodeProviderFailure maps to ARC-013: generic summarizer provider failure.
	ErrorCodeProviderFailure ErrorCode = "ARC-013"

	// ErrorCodeRequestTooLarge maps to ARC-014: context or request too large.
	ErrorCodeRequestTooLarge ErrorCode = "ARC-014"

	// ErrorCodeBillingFailure maps to ARC-015: billing failure.
	ErrorCodeBillingFailure ErrorCode = "ARC-015"
)

// SummarizerRequest carries the Markdown source to summarize.
type SummarizerRequest struct {
	MarkdownSource string
}

// SummarizerResult holds the outcome of a summarization attempt.
type SummarizerResult struct {
	Status        ResultStatus
	Provider      Provider
	Summary       string
	ErrorCode     ErrorCode
	FailureReason string
}

// SummarizerService is the provider-neutral summarization contract.
// Implementations must not expose provider-specific SDK types.
type SummarizerService interface {
	Summarize(ctx context.Context, req SummarizerRequest) SummarizerResult
}
