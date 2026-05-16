package markdown

import (
	"fmt"
)

const (
	ProviderJina Provider = "jina"
)

// ExtractionError carries provider diagnostics while unwrapping to an ARC
// sentinel for classification and public persistence.
type ExtractionError struct {
	Provider   Provider
	Reason     string
	StatusCode int
	Err        error
}

func (e *ExtractionError) Error() string {
	if e == nil {
		return ""
	}

	if e.StatusCode != 0 {
		return fmt.Sprintf("markdown: %s: %s (HTTP %d): %v", e.Provider, e.Reason, e.StatusCode, e.Err)
	}

	if e.Reason != "" {
		return fmt.Sprintf("markdown: %s: %s: %v", e.Provider, e.Reason, e.Err)
	}

	return fmt.Sprintf("markdown: %s: %v", e.Provider, e.Err)
}

func (e *ExtractionError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

func extractionFailure(provider Provider, err error, reason string, statusCode int) error {
	return &ExtractionError{
		Provider:   provider,
		Reason:     reason,
		StatusCode: statusCode,
		Err:        err,
	}
}

func localFailure(err error, reason string) error {
	return extractionFailure(ProviderGoReadability, err, reason, 0)
}

func jinaFailure(err error, reason string, statusCode int) error {
	return extractionFailure(ProviderJina, err, reason, statusCode)
}
