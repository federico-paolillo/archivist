package fetcher

import (
	"errors"
	"fmt"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
)

// FetcherError carries fetch operation diagnostics while unwrapping to the
// canonical ARC sentinel when the failure maps to a public processing error.
type FetcherError struct {
	Op          string
	URL         string
	StatusCode  int
	ContentType string
	Reason      string
	Err         error
}

func (e *FetcherError) Error() string {
	if e == nil {
		return ""
	}

	msg := "fetcher"
	if e.Op != "" {
		msg += ": " + e.Op
	}

	if e.URL != "" {
		msg += ": url=" + e.URL
	}

	if e.StatusCode != 0 {
		msg += fmt.Sprintf(": status=%d", e.StatusCode)
	}

	if e.ContentType != "" {
		msg += ": content_type=" + e.ContentType
	}

	if e.Reason != "" {
		msg += ": " + e.Reason
	}

	if e.Err != nil {
		msg += ": " + e.Err.Error()
	}

	return msg
}

func (e *FetcherError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

func fetchFailure(op string, err error, reason string, opts ...func(*FetcherError)) error {
	fetchErr := &FetcherError{
		Op:     op,
		Err:    err,
		Reason: reason,
	}
	for _, opt := range opts {
		opt(fetchErr)
	}

	return fetchErr
}

func withURL(rawURL string) func(*FetcherError) {
	return func(err *FetcherError) {
		err.URL = rawURL
	}
}

func withStatusCode(statusCode int) func(*FetcherError) {
	return func(err *FetcherError) {
		err.StatusCode = statusCode
	}
}

func withContentType(contentType string) func(*FetcherError) {
	return func(err *FetcherError) {
		err.ContentType = contentType
	}
}

// classifyHTTPStatus maps HTTP status codes to ARC errors.
func classifyHTTPStatus(rawURL string, status int) error {
	switch {
	case status == 401 || status == 403:
		return fetchFailure(
			"http status",
			arc.ErrURLAccessDenied,
			"access denied",
			withURL(rawURL),
			withStatusCode(status),
		)
	case status == 404:
		return fetchFailure(
			"http status",
			arc.ErrURLNotFound,
			"not found",
			withURL(rawURL),
			withStatusCode(status),
		)
	case status >= 400:
		// 4xx errors other than 401/403/404 (e.g. 410 Gone, 429 Too Many Requests)
		// and 5xx errors are both treated as transient.
		return fetchFailure(
			"http status",
			arc.ErrURLFetchTransientFailure,
			"transient HTTP failure",
			withURL(rawURL),
			withStatusCode(status),
		)
	default:
		return nil
	}
}

// classifyRequestError maps low-level request errors to ARC errors.
// All request-level errors are transient by policy.
func classifyRequestError(rawURL string, err error) error {
	return fetchFailure(
		"request",
		errors.Join(arc.ErrURLFetchTransientFailure, err),
		"request failed",
		withURL(rawURL),
	)
}

func unsupportedScheme(rawURL string) error {
	return fetchFailure("validate scheme", arc.ErrURLResolutionFailed, "unsupported URL scheme", withURL(rawURL))
}

func notHTML(rawURL, contentType string) error {
	return fetchFailure(
		"validate content type",
		arc.ErrResponseNotHTML,
		"response is not HTML",
		withURL(rawURL),
		withContentType(contentType),
	)
}

func bodyTooLarge(rawURL string) error {
	return fetchFailure("read body", arc.ErrResponseTooLarge, "response body exceeds size limit", withURL(rawURL))
}
