package fetcher

import "errors"

// Public ARC-coded sentinel errors. Callers can use errors.Is to distinguish failure
// classes. Message text follows the format required by docs/conventions/ERRORS.md and
// must end with a period because the persisted article error format is "[ARC-NNN] Sentence."

//nolint:staticcheck // ARC error messages must end with a period per docs/conventions/ERRORS.md
var ErrUnsupportedScheme = errors.New("[ARC-001] The URL could not be resolved.")

//nolint:staticcheck // ARC error messages must end with a period per docs/conventions/ERRORS.md
var ErrAccessDenied = errors.New("[ARC-002] The URL requires access Archivist does not have.")

//nolint:staticcheck // ARC error messages must end with a period per docs/conventions/ERRORS.md
var ErrNotFound = errors.New("[ARC-003] The URL was not found.")

//nolint:staticcheck // ARC error messages must end with a period per docs/conventions/ERRORS.md
var ErrTransientFailure = errors.New("[ARC-004] The URL could not be fetched right now.")

//nolint:staticcheck // ARC error messages must end with a period per docs/conventions/ERRORS.md
var ErrNotHTML = errors.New("[ARC-005] The URL did not return an HTML article page.")

//nolint:staticcheck // ARC error messages must end with a period per docs/conventions/ERRORS.md
var ErrBodyTooLarge = errors.New("[ARC-006] The HTML response is too large to archive.")

// classifyHTTPStatus maps HTTP status codes to ARC errors.
func classifyHTTPStatus(status int) error {
	switch {
	case status == 401 || status == 403:
		return ErrAccessDenied
	case status == 404:
		return ErrNotFound
	case status >= 400:
		// 4xx errors other than 401/403/404 (e.g. 410 Gone, 429 Too Many Requests)
		// and 5xx errors are both treated as transient.
		return ErrTransientFailure
	default:
		return nil
	}
}

// classifyRequestError maps low-level request errors to ARC errors.
// All request-level errors are transient by policy.
func classifyRequestError(_ error) error {
	return ErrTransientFailure
}
