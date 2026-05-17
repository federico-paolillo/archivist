package markdown

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/require"
)

// Compile-time interface check.
var _ MarkdownExtractor = (*JinaExtractor)(nil)

func TestJinaExtractorIsMarkdownExtractor(t *testing.T) {
	var extractor MarkdownExtractor = NewJinaExtractor(req.NewClient(), "test-api-key")

	require.NotNil(t, extractor)
}

func TestJinaExtractorSuccessfulExtractionReturnsMarkdown(t *testing.T) {
	const expectedMarkdown = "# Article Title\n\nSome content for the article."

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(expectedMarkdown))
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor("test-api-key", server.URL+"/")

	output, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>some content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.NoError(t, err)
	require.Equal(t, ProviderJina, extractor.Provider())
	require.Equal(t, expectedMarkdown, output.Markdown)
}

func TestJinaExtractorSuccessfulExtractionPassesAPIKey(t *testing.T) {
	const apiKey = "test-api-key-value"

	var receivedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("# Markdown content"))
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor(apiKey, server.URL+"/")

	_, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.NoError(t, err)
	require.Equal(t, "Bearer "+apiKey, receivedAuth)
}

func TestJinaExtractorNoAPIKeyOmitsAuthorizationHeader(t *testing.T) {
	var receivedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("# Markdown content"))
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor("", server.URL+"/")

	_, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.NoError(t, err)
	require.Empty(t, receivedAuth)
}

func TestJinaExtractorGeneralFailureMapsToARC010(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":{"message":"provider unavailable"}}`))
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor("", server.URL+"/")

	_, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.ErrorIs(t, err, arc.ErrJinaReaderFailure)
	extractionErr, ok := errors.AsType[*ExtractionError](err)
	require.True(t, ok)
	require.Equal(t, ProviderJina, extractionErr.Provider)
	require.Equal(t, http.StatusInternalServerError, extractionErr.StatusCode)
}

func TestJinaExtractorTransportFailureMapsToARC010(t *testing.T) {
	// Use a server that is immediately closed to simulate a transport/connection failure.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	serverURL := server.URL
	server.Close()

	extractor := newTestJinaExtractor("", serverURL+"/")

	_, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.ErrorIs(t, err, arc.ErrJinaReaderFailure)
	extractionErr, ok := errors.AsType[*ExtractionError](err)
	require.True(t, ok)
	require.Equal(t, ProviderJina, extractionErr.Provider)
	require.NotEmpty(t, extractionErr.Reason)
}

func TestJinaExtractorInsufficientBalanceMapsToARC011(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		_, _ = w.Write([]byte(`{"error":{"message":"anything"}}`))
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor("", server.URL+"/")

	_, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.ErrorIs(t, err, arc.ErrJinaInsufficientBalance)
	extractionErr, ok := errors.AsType[*ExtractionError](err)
	require.True(t, ok)
	require.Equal(t, http.StatusPaymentRequired, extractionErr.StatusCode)
}

func TestJinaExtractorAcceptsTextPlain(t *testing.T) {
	var receivedAccept string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAccept = r.Header.Get("Accept")

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("# Content"))
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor("", server.URL+"/")

	_, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.NoError(t, err)
	require.Equal(t, "text/plain", receivedAccept)
}

func TestJinaExtractorAcceptsTextMarkdown(t *testing.T) {
	const expectedMarkdown = "# Markdown content"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(expectedMarkdown))
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor("", server.URL+"/")

	output, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.NoError(t, err)
	require.Equal(t, expectedMarkdown, output.Markdown)
}

func TestJinaExtractorRejectsJSONSuccessContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"content":"# Not markdown"}`))
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor("", server.URL+"/")

	_, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.ErrorIs(t, err, arc.ErrJinaReaderFailure)
	extractionErr, ok := errors.AsType[*ExtractionError](err)
	require.True(t, ok)
	require.Equal(t, "jina reader returned unexpected content type", extractionErr.Reason)
}

func TestJinaExtractorRejectsMissingSuccessContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor("", server.URL+"/")

	_, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.ErrorIs(t, err, arc.ErrJinaReaderFailure)
	extractionErr, ok := errors.AsType[*ExtractionError](err)
	require.True(t, ok)
	require.Equal(t, "jina reader returned unexpected content type", extractionErr.Reason)
}

func TestJinaExtractorRejectsMalformedSuccessContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("# Content"))
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor("", server.URL+"/")

	_, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.ErrorIs(t, err, arc.ErrJinaReaderFailure)
}

func TestJinaExtractorOversizedSuccessBodyMapsToARC010(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(strings.Repeat("a", maxJinaMarkdownBytes+1)))
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor("", server.URL+"/")

	_, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.ErrorIs(t, err, arc.ErrJinaReaderFailure)
	extractionErr, ok := errors.AsType[*ExtractionError](err)
	require.True(t, ok)
	require.Equal(t, jinaReadLimitExceededMsg, extractionErr.Reason)
}

func TestJinaExtractorOversizedErrorBodyMapsToARC010(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(strings.Repeat("a", maxJinaDiagnosticBytes+1)))
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor("", server.URL+"/")

	_, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.ErrorIs(t, err, arc.ErrJinaReaderFailure)
	require.NotErrorIs(t, err, arc.ErrJinaInsufficientBalance)
	extractionErr, ok := errors.AsType[*ExtractionError](err)
	require.True(t, ok)
	require.Equal(t, jinaReadLimitExceededMsg, extractionErr.Reason)
	require.Equal(t, http.StatusForbidden, extractionErr.StatusCode)
}

func TestJinaExtractorNon402JSONInsufficientBalanceMapsToARC011(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"code":"invalid_request_error","message":"Insufficient Balance"}}`))
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor("", server.URL+"/")

	_, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.ErrorIs(t, err, arc.ErrJinaInsufficientBalance)
	extractionErr, ok := errors.AsType[*ExtractionError](err)
	require.True(t, ok)
	require.Equal(t, http.StatusForbidden, extractionErr.StatusCode)
}

func TestJinaExtractorNon402TextInsufficientBalanceMapsToARC011(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("request failed: out_of_tokens"))
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor("", server.URL+"/")

	_, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.ErrorIs(t, err, arc.ErrJinaInsufficientBalance)
	extractionErr, ok := errors.AsType[*ExtractionError](err)
	require.True(t, ok)
	require.Equal(t, http.StatusTooManyRequests, extractionErr.StatusCode)
}

func TestJinaExtractorNon402UnrelatedErrorMapsToARC010(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"message":"invalid API key"}}`))
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor("", server.URL+"/")

	_, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.ErrorIs(t, err, arc.ErrJinaReaderFailure)
	require.NotErrorIs(t, err, arc.ErrJinaInsufficientBalance)
}

func TestJinaExtractorEmptyResponseMapsToARC010(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	extractor := newTestJinaExtractor("", server.URL+"/")

	_, err := extractor.ExtractMarkdown(t.Context(), ExtractInput{
		HTML:         []byte(`<html><body><p>content</p></body></html>`),
		CanonicalURL: "https://example.com/article",
	})

	require.ErrorIs(t, err, arc.ErrJinaReaderFailure)
}

// newTestJinaExtractor creates a JinaExtractor with a custom base URL for testing.
// It trims the trailing slash from baseURL before constructing the extractor so that
// the concatenation baseURL+canonicalURL matches how httptest.Server routes requests.
func newTestJinaExtractor(apiKey, baseURL string) *JinaExtractor {
	baseURL = strings.TrimSuffix(baseURL, "/") + "/"

	extractor := NewJinaExtractor(req.NewClient(), apiKey)
	extractor.baseURL = baseURL

	return extractor
}
