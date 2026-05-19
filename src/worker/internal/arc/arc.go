package arc

import (
	"errors"
	"fmt"
)

// Code is a stable ARC article-processing error code.
type Code string

const (
	CodeURLResolutionFailed       Code = "ARC-001"
	CodeURLAccessDenied           Code = "ARC-002"
	CodeURLNotFound               Code = "ARC-003"
	CodeURLFetchTransientFailure  Code = "ARC-004"
	CodeResponseNotHTML           Code = "ARC-005"
	CodeResponseTooLarge          Code = "ARC-006"
	CodeSnapshotWriteFailed       Code = "ARC-007"
	CodeLocalUnreadable           Code = "ARC-008"
	CodeLocalExtractionFailed     Code = "ARC-009"
	CodeJinaReaderFailed          Code = "ARC-010"
	CodeJinaInsufficientBalance   Code = "ARC-011"
	CodeMarkdownWriteFailed       Code = "ARC-012"
	CodeSummarizerProviderFailed  Code = "ARC-013"
	CodeSummarizerRequestTooLarge Code = "ARC-014"
	CodeSummarizerBillingFailed   Code = "ARC-015"
	CodeSummaryWriteFailed        Code = "ARC-016"
	CodeSSRFDetected              Code = "ARC-017"
	CodeUnknownProcessingFailure  Code = "ARC-999"
)

var (
	ErrURLResolutionFailed       = New(CodeURLResolutionFailed)
	ErrURLAccessDenied           = New(CodeURLAccessDenied)
	ErrURLNotFound               = New(CodeURLNotFound)
	ErrURLFetchTransientFailure  = New(CodeURLFetchTransientFailure)
	ErrResponseNotHTML           = New(CodeResponseNotHTML)
	ErrResponseTooLarge          = New(CodeResponseTooLarge)
	ErrSnapshotWrite             = New(CodeSnapshotWriteFailed)
	ErrLocalUnreadable           = New(CodeLocalUnreadable)
	ErrLocalExtractionFailed     = New(CodeLocalExtractionFailed)
	ErrJinaReaderFailure         = New(CodeJinaReaderFailed)
	ErrJinaInsufficientBalance   = New(CodeJinaInsufficientBalance)
	ErrMarkdownWrite             = New(CodeMarkdownWriteFailed)
	ErrSummarizerProviderFailure = New(CodeSummarizerProviderFailed)
	ErrSummarizerRequestTooLarge = New(CodeSummarizerRequestTooLarge)
	ErrSummarizerBillingFailure  = New(CodeSummarizerBillingFailed)
	ErrSummaryWrite              = New(CodeSummaryWriteFailed)
	ErrSSRFDetected              = New(CodeSSRFDetected)
	ErrUnknown                   = New(CodeUnknownProcessingFailure)
)

var publicMessages = map[Code]string{
	CodeURLResolutionFailed:       "The URL could not be resolved.",
	CodeURLAccessDenied:           "The URL requires access Archivist does not have.",
	CodeURLNotFound:               "The URL was not found.",
	CodeURLFetchTransientFailure:  "The URL could not be fetched right now.",
	CodeResponseNotHTML:           "The URL did not return an HTML article page.",
	CodeResponseTooLarge:          "The HTML response is too large to archive.",
	CodeSnapshotWriteFailed:       "Archivist could not store the HTML snapshot.",
	CodeLocalUnreadable:           "Archivist could not read this page locally.",
	CodeLocalExtractionFailed:     "Archivist could not extract this page locally.",
	CodeJinaReaderFailed:          "Archivist could not extract this page with the fallback reader.",
	CodeJinaInsufficientBalance:   "Archivist could not use the fallback reader because the Jina account is out of credit.",
	CodeMarkdownWriteFailed:       "Archivist could not store the Markdown article.",
	CodeSummarizerProviderFailed:  "Archivist could not summarize this article.",
	CodeSummarizerRequestTooLarge: "This article is too large to summarize.",
	CodeSummarizerBillingFailed:   "Archivist could not use the summarizer because the provider account has a billing issue.",
	CodeSummaryWriteFailed:        "Archivist could not store the article summary.",
	CodeSSRFDetected:              "Archivist refused to process suspicious URL.",
	CodeUnknownProcessingFailure:  "Archivist could not process the URL.",
}

// Error is a typed ARC error. It is safe to wrap with fmt.Errorf("%w", err);
// matching and public rendering are based on Code, not the diagnostic string.
type Error struct {
	code Code
}

// New returns an ARC error for code.
func New(code Code) *Error {
	return &Error{code: code}
}

// Code returns the stable ARC code.
func (e *Error) Code() Code {
	if e == nil {
		return ""
	}

	return e.code
}

// Error returns the public ARC message. Persisted user-facing errors should use
// PublicMessage so wrapped diagnostic context is not accidentally stored.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	return Format(e.code)
}

// Is matches ARC errors by code.
func (e *Error) Is(target error) bool {
	targetErr, ok := target.(*Error)
	if !ok {
		return false
	}

	return e.code == targetErr.code
}

// Format returns the public persisted representation for code.
func Format(code Code) string {
	msg, ok := publicMessages[code]
	if !ok {
		msg = publicMessages[CodeUnknownProcessingFailure]
		code = CodeUnknownProcessingFailure
	}

	return fmt.Sprintf("[%s] %s", code, msg)
}

// CodeOf extracts the ARC code from err or any wrapped ARC error.
func CodeOf(err error) (Code, bool) {
	if err == nil {
		return "", false
	}

	arcErr, ok := errors.AsType[*Error](err)
	if !ok {
		return "", false
	}

	return arcErr.Code(), true
}

// CodeString returns the ARC code for structured logs. Unknown non-nil errors
// are reported as ARC-999 because they become unknown public failures.
func CodeString(err error) string {
	code, ok := CodeOf(err)
	if ok {
		return string(code)
	}

	if err != nil {
		return string(CodeUnknownProcessingFailure)
	}

	return ""
}

// PublicMessage returns the clean persisted public message for err.
func PublicMessage(err error) (string, bool) {
	code, ok := CodeOf(err)
	if !ok {
		return "", false
	}

	return Format(code), true
}
