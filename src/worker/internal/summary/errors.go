package summary

import (
	"errors"
	"fmt"
	"net/http"

	anthropic "github.com/anthropics/anthropic-sdk-go"
)

func classifyError(err error) SummarizerResult {
	apiErr, ok := errors.AsType[*anthropic.Error](err)
	if ok {
		return classifyAPIError(apiErr)
	}

	return SummarizerResult{
		Status:        ResultStatusFailure,
		Provider:      ProviderAnthropic,
		ErrorCode:     ErrorCodeProviderFailure,
		FailureReason: fmt.Sprintf("provider error: %v", err),
	}
}

func classifyAPIError(apiErr *anthropic.Error) SummarizerResult {
	errType := apiErr.Type()

	if isBillingError(apiErr) {
		return SummarizerResult{
			Status:        ResultStatusFailure,
			Provider:      ProviderAnthropic,
			ErrorCode:     ErrorCodeBillingFailure,
			FailureReason: fmt.Sprintf("billing error (HTTP %d): %v", apiErr.StatusCode, errType),
			RequestID:     apiErr.RequestID,
			StatusCode:    apiErr.StatusCode,
		}
	}

	if isTooLargeError(apiErr) {
		return SummarizerResult{
			Status:        ResultStatusFailure,
			Provider:      ProviderAnthropic,
			ErrorCode:     ErrorCodeRequestTooLarge,
			FailureReason: fmt.Sprintf("request too large (HTTP %d): %v", apiErr.StatusCode, errType),
			RequestID:     apiErr.RequestID,
			StatusCode:    apiErr.StatusCode,
		}
	}

	return SummarizerResult{
		Status:        ResultStatusFailure,
		Provider:      ProviderAnthropic,
		ErrorCode:     ErrorCodeProviderFailure,
		FailureReason: fmt.Sprintf("provider error (HTTP %d): %v", apiErr.StatusCode, errType),
		RequestID:     apiErr.RequestID,
		StatusCode:    apiErr.StatusCode,
	}
}

func isBillingError(apiErr *anthropic.Error) bool {
	return apiErr.StatusCode == http.StatusPaymentRequired || apiErr.Type() == anthropic.ErrorTypeBillingError
}

func isTooLargeError(apiErr *anthropic.Error) bool {
	return apiErr.StatusCode == http.StatusRequestEntityTooLarge
}
