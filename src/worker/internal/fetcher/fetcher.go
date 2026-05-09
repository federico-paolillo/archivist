// Package fetcher fetches bounded HTML content and maps failures to ARC error codes.
package fetcher

import (
	"context"
	"fmt"
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
	client *req.Client
}

// New creates a Fetcher using the provided HTTP client.
// The caller is responsible for configuring redirect policy, timeout, and HTTP version settings.
func New(client *req.Client) *Fetcher {
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

