// Package fetcher fetches bounded HTML content and maps failures to ARC error codes.
package fetcher

import (
	"context"
	"io"
	"mime"
	"net/url"
	"strings"

	"github.com/imroc/req/v3"
)

const (
	// maxBodyBytes is the maximum response body size accepted (10 MiB).
	maxBodyBytes = 10 * 1024 * 1024
)

// acceptedContentTypes are the only MIME types the fetcher will process.
var acceptedContentTypes = []string{"text/html", "application/xhtml+xml"}

// Result holds the outcome of a successful fetch.
type Result struct {
	// FinalURL is the resolved URL after all redirects.
	FinalURL string

	// Body contains the HTML response bytes.
	Body []byte
}

// Fetcher resolves URLs and retrieves bounded HTML content.
type Fetcher struct {
	client      *req.Client
	validateURL URLValidator
}

// URLValidator validates a raw URL before the fetcher sends a request.
type URLValidator func(rawURL string) error

// New creates a Fetcher using the provided HTTP client.
// The caller is responsible for configuring redirect policy, timeout, and HTTP version settings.
func New(client *req.Client, validators ...URLValidator) *Fetcher {
	var validator URLValidator
	if len(validators) > 0 {
		validator = validators[0]
	}

	return &Fetcher{client: client, validateURL: validator}
}

// Fetch resolves the given rawURL and returns the final URL and HTML bytes.
// It validates the scheme, follows redirects, enforces size and content-type
// limits, and maps every failure to a public ARC error.
func (f *Fetcher) Fetch(ctx context.Context, rawURL string) (*Result, error) {
	if f.validateURL != nil {
		validateErr := f.validateURL(rawURL)
		if validateErr != nil {
			return nil, validateErr
		}
	} else {
		schemeErr := validateScheme(rawURL)
		if schemeErr != nil {
			return nil, schemeErr
		}
	}

	resp, fetchErr := f.client.R().
		SetContext(ctx).
		Get(rawURL)
	if fetchErr != nil {
		return nil, classifyRequestError(rawURL, fetchErr)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	statusErr := classifyHTTPStatus(rawURL, resp.StatusCode)
	if statusErr != nil {
		return nil, statusErr
	}

	contentTypeErr := validateContentType(rawURL, resp.GetContentType())
	if contentTypeErr != nil {
		return nil, contentTypeErr
	}

	body, readErr := readLimited(rawURL, resp.Body)
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

// validateScheme returns ARC-001 when rawURL is not http or https.
func validateScheme(rawURL string) error {
	parsed, parseErr := url.Parse(rawURL)
	if parseErr != nil {
		return unsupportedScheme(rawURL)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return unsupportedScheme(rawURL)
	}

	return nil
}

// validateContentType returns ARC-005 when the MIME type is not an accepted HTML type.
func validateContentType(rawURL, contentType string) error {
	if contentType == "" {
		return notHTML(rawURL, contentType)
	}

	mediaType, _, parseErr := mime.ParseMediaType(contentType)
	if parseErr != nil {
		return notHTML(rawURL, contentType)
	}

	for _, accepted := range acceptedContentTypes {
		if strings.EqualFold(mediaType, accepted) {
			return nil
		}
	}

	return notHTML(rawURL, contentType)
}

// readLimited reads at most maxBodyBytes from r.
// It returns ARC-006 if the body exceeds that limit.
func readLimited(rawURL string, r io.Reader) ([]byte, error) {
	limited := io.LimitReader(r, maxBodyBytes+1)

	data, readErr := io.ReadAll(limited)
	if readErr != nil {
		return nil, fetchFailure("read body", readErr, "read response body", withURL(rawURL))
	}

	if len(data) > maxBodyBytes {
		return nil, bodyTooLarge(rawURL)
	}

	return data, nil
}
