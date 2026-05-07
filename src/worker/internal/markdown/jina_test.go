package markdown

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Compile-time interface check.
var _ MarkdownExtractor = (*JinaExtractor)(nil)

func TestJinaExtractorIsMarkdownExtractor(t *testing.T) {
	var extractor MarkdownExtractor = NewJinaExtractor(false, "")

	require.NotNil(t, extractor)
}

func TestJinaExtractorDisabledReturnsFallbackFailure(t *testing.T) {
	extractor := NewJinaExtractor(false, "")

	result := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>some content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.Equal(t, ResultStatusFailure, result.Status)
	require.Equal(t, ProviderJina, result.Provider)
	require.Equal(t, ErrorCodeJinaFailed, result.ErrorCode)
	require.NotEmpty(t, result.FailureReason)
	require.Empty(t, result.Markdown)
}

func TestJinaExtractorSuccessfulExtractionReturnsMarkdown(t *testing.T) {
	const expectedMarkdown = "# Article Title\n\nSome content for the article."

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(expectedMarkdown))
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor(true, "", server.URL+"/")

	result := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>some content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.Equal(t, ResultStatusSuccess, result.Status)
	require.Equal(t, ProviderJina, result.Provider)
	require.Equal(t, expectedMarkdown, result.Markdown)
	require.Empty(t, result.ErrorCode)
}

func TestJinaExtractorSuccessfulExtractionPassesAPIKey(t *testing.T) {
	const apiKey = "test-api-key-value"

	var receivedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("# Markdown content"))
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor(true, apiKey, server.URL+"/")

	result := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.Equal(t, ResultStatusSuccess, result.Status)
	require.Equal(t, "Bearer "+apiKey, receivedAuth)
}

func TestJinaExtractorNoAPIKeyOmitsAuthorizationHeader(t *testing.T) {
	var receivedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("# Markdown content"))
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor(true, "", server.URL+"/")

	result := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.Equal(t, ResultStatusSuccess, result.Status)
	require.Empty(t, receivedAuth)
}

func TestJinaExtractorGeneralFailureMapsToARC010(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor(true, "", server.URL+"/")

	result := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.Equal(t, ResultStatusFailure, result.Status)
	require.Equal(t, ProviderJina, result.Provider)
	require.Equal(t, ErrorCodeJinaFailed, result.ErrorCode)
	require.NotEmpty(t, result.FailureReason)
}

func TestJinaExtractorTransportFailureMapsToARC010(t *testing.T) {
	// Use a server that is immediately closed to simulate a transport/connection failure.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	serverURL := server.URL
	server.Close()

	extractor := newTestJinaExtractor(true, "", serverURL+"/")

	result := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.Equal(t, ResultStatusFailure, result.Status)
	require.Equal(t, ProviderJina, result.Provider)
	require.Equal(t, ErrorCodeJinaFailed, result.ErrorCode)
	require.NotEmpty(t, result.FailureReason)
}

func TestJinaExtractorInsufficientBalanceMapsToARC011(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusPaymentRequired)
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor(true, "", server.URL+"/")

	result := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.Equal(t, ResultStatusFailure, result.Status)
	require.Equal(t, ProviderJina, result.Provider)
	require.Equal(t, ErrorCodeJinaInsufficientCredit, result.ErrorCode)
	require.NotEmpty(t, result.FailureReason)
}

func TestJinaExtractorAcceptsTextPlain(t *testing.T) {
	var receivedAccept string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAccept = r.Header.Get("Accept")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("# Content"))
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor(true, "", server.URL+"/")

	result := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.Equal(t, ResultStatusSuccess, result.Status)
	require.Equal(t, "text/plain", receivedAccept)
}

// newTestJinaExtractor creates a JinaExtractor with a custom base URL for testing.
// It trims the trailing slash from baseURL before constructing the extractor so that
// the concatenation baseURL+canonicalURL matches how httptest.Server routes requests.
func newTestJinaExtractor(enabled bool, apiKey, baseURL string) *JinaExtractor {
	baseURL = strings.TrimSuffix(baseURL, "/") + "/"

	extractor := NewJinaExtractor(enabled, apiKey)
	extractor.baseURL = baseURL

	return extractor
}
