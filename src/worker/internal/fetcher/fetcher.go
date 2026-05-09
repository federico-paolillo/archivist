// Package fetcher resolves URLs and fetches HTML content with conservative limits.
// It accepts only http and https schemes, follows at most 10 redirects, enforces
// a 20-second total timeout, rejects bodies larger than 10 MiB, and accepts only
// text/html and application/xhtml+xml responses.
//
// Every failure class maps to a public ARC error code from docs/conventions/ERRORS.md.
// The mapping is exported so callers can inspect or present error codes without
// depending on internal error types.
package fetcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/url"
	"strings"
	"time"

	"github.com/imroc/req/v3"
)

const (
	// maxRedirects is the maximum number of redirects the fetcher will follow.
	maxRedirects = 10

	// fetchTimeout is the total time budget for a single fetch operation.
	fetchTimeout = 20 * time.Second

	// maxBodyBytes is the maximum response body size accepted (10 MiB).
	maxBodyBytes = 10 * 1024 * 1024
)

// acceptedContentTypes are the only MIME types the fetcher will process.
var acceptedContentTypes = []string{"text/html", "application/xhtml+xml"}

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

// Result holds the outcome of a successful fetch.
type Result struct {
	// FinalURL is the resolved URL after all redirects.
	FinalURL string

	// Body contains the HTML response bytes.
	Body []byte
}

// Fetcher resolves URLs and retrieves bounded HTML content.
type Fetcher struct {
	client *req.Client
}

// New creates a Fetcher with the project-mandated limits applied.
func New() *Fetcher {
	client := req.C().
		SetRedirectPolicy(req.MaxRedirectPolicy(maxRedirects)).
		SetTimeout(fetchTimeout).
		DisableForceHttpVersion()

	return &Fetcher{client: client}
}

// Fetch resolves the given rawURL and returns the final URL and HTML bytes.
// It validates the scheme, follows redirects, enforces size and content-type
// limits, and maps every failure to a public ARC error.
func (f *Fetcher) Fetch(ctx context.Context, rawURL string) (*Result, error) {
	schemeErr := validateScheme(rawURL)
	if schemeErr != nil {
		return nil, schemeErr
	}

	resp, fetchErr := f.client.R().
		SetContext(ctx).
		Get(rawURL)
	if fetchErr != nil {
		return nil, classifyRequestError(fetchErr)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	statusErr := classifyHTTPStatus(resp.StatusCode)
	if statusErr != nil {
		return nil, statusErr
	}

	contentTypeErr := validateContentType(resp.GetContentType())
	if contentTypeErr != nil {
		return nil, contentTypeErr
	}

	body, readErr := readLimited(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	// resp.Response is the embedded *http.Response. Its Request field is the
	// actual *http.Request that produced this response, which is the final
	// request after all redirects have been followed.
	finalURL := resp.Response.Request.URL.String()

	return &Result{
		FinalURL: finalURL,
		Body:     body,
	}, nil
}

// validateScheme returns ErrUnsupportedScheme when rawURL is not http or https.
func validateScheme(rawURL string) error {
	parsed, parseErr := url.Parse(rawURL)
	if parseErr != nil {
		return ErrUnsupportedScheme
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return ErrUnsupportedScheme
	}

	return nil
}

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

// validateContentType returns ErrNotHTML when the MIME type is not an accepted HTML type.
func validateContentType(contentType string) error {
	if contentType == "" {
		return ErrNotHTML
	}

	mediaType, _, parseErr := mime.ParseMediaType(contentType)
	if parseErr != nil {
		return ErrNotHTML
	}

	for _, accepted := range acceptedContentTypes {
		if strings.EqualFold(mediaType, accepted) {
			return nil
		}
	}

	return ErrNotHTML
}

// readLimited reads at most maxBodyBytes from r.
// It returns ErrBodyTooLarge if the body exceeds that limit.
func readLimited(r io.Reader) ([]byte, error) {
	limited := io.LimitReader(r, maxBodyBytes+1)

	data, readErr := io.ReadAll(limited)
	if readErr != nil {
		return nil, fmt.Errorf("fetcher: reading response body: %w", readErr)
	}

	if len(data) > maxBodyBytes {
		return nil, ErrBodyTooLarge
	}

	return data, nil
}

// classifyRequestError maps low-level request errors to ARC errors.
func classifyRequestError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return ErrTransientFailure
	}

	urlErr, ok := errors.AsType[*url.Error](err)
	if ok && urlErr.Timeout() {
		return ErrTransientFailure
	}

	return ErrTransientFailure
}
