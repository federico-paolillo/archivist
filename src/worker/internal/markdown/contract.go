package markdown

import "context"

type Provider string

const (
	ProviderGoReadability Provider = "go-readability"
)

type ResultStatus string

const (
	ResultStatusSuccess         ResultStatus = "success"
	ResultStatusLocalUnreadable ResultStatus = "local_unreadable"
	ResultStatusFailure         ResultStatus = "failure"
)

type ErrorCode string

const (
	ErrorCodeLocalExtractionFailed ErrorCode = "ARC-009"
)

type ExtractInput struct {
	HTML         []byte
	CanonicalURL string
}

type ExtractResult struct {
	Status        ResultStatus
	Provider      Provider
	Markdown      string
	Title         string
	ErrorCode     ErrorCode
	FailureReason string
}

type MarkdownExtractor interface {
	ExtractMarkdown(ctx context.Context, input ExtractInput) ExtractResult
}
