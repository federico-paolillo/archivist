package summary

import (
	"errors"
	"fmt"
	"net/http"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
	anthropic "github.com/anthropics/anthropic-sdk-go"
)

// ProviderError carries provider diagnostics while unwrapping to an ARC
// sentinel for classification and public persistence.
type ProviderError struct {
	Provider   Provider
	Reason     string
	RequestID  string
	StatusCode int
	Err        error
}

func (e *ProviderError) Error() string {
	if e == nil {
		return ""
	}

	if e.StatusCode != 0 {
		return fmt.Sprintf("summary: %s: %s (HTTP %d): %v", e.Provider, e.Reason, e.StatusCode, e.Err)
	}

	if e.Reason != "" {
		return fmt.Sprintf("summary: %s: %s: %v", e.Provider, e.Reason, e.Err)
	}

	return fmt.Sprintf("summary: %s: %v", e.Provider, e.Err)
}

func (e *ProviderError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

func providerFailure(err error, reason string, requestID string, statusCode int) error {
	return &ProviderError{
		Provider:   ProviderAnthropic,
		Reason:     reason,
		RequestID:  requestID,
		StatusCode: statusCode,
		Err:        err,
	}
}

func classifyError(err error) error {
	apiErr, ok := errors.AsType[*anthropic.Error](err)
	if ok {
		return classifyAPIError(apiErr)
	}

	return providerFailure(arc.ErrSummarizerProviderFailure, fmt.Sprintf("provider error: %v", err), "", 0)
}

func classifyAPIError(apiErr *anthropic.Error) error {
	reason, arcErr := classifyAnthropicAPIError(apiErr)

	return providerFailure(arcErr, reason, apiErr.RequestID, apiErr.StatusCode)
}

func classifyAnthropicAPIError(apiErr *anthropic.Error) (string, error) {
	errType := apiErr.Type()
	if isBillingError(apiErr) {
		return fmt.Sprintf("billing error: %v", errType), arc.ErrSummarizerBillingFailure
	}

	if isTooLargeError(apiErr) {
		return fmt.Sprintf("request too large: %v", errType), arc.ErrSummarizerRequestTooLarge
	}

	return fmt.Sprintf("provider error: %v", errType), arc.ErrSummarizerProviderFailure
}

func isBillingError(apiErr *anthropic.Error) bool {
	return apiErr.StatusCode == http.StatusPaymentRequired || apiErr.Type() == anthropic.ErrorTypeBillingError
}

func isTooLargeError(apiErr *anthropic.Error) bool {
	return apiErr.StatusCode == http.StatusRequestEntityTooLarge
}
