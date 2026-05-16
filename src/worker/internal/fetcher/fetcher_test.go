package fetcher_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
	"codeberg.org/federico-paolillo/archivist/internal/fetcher"
	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const minimalHTML = `<!DOCTYPE html><html><body>Hello</body></html>`

func TestFetchSuccessWithRedirect(t *testing.T) {
	var finalServer *httptest.Server

	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, finalServer.URL+"/article", http.StatusFound)
	}))
	defer redirectServer.Close()

	finalServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/article" {
			http.NotFound(w, r)

			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(minimalHTML))
	}))
	defer finalServer.Close()

	f := fetcher.New(req.NewClient())

	result, err := f.Fetch(t.Context(), redirectServer.URL)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Contains(t, result.FinalURL, "/article", "FinalURL should be the resolved URL after redirect")
	assert.Equal(t, []byte(minimalHTML), result.Body)
}

func TestFetchHTMLSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(minimalHTML))
	}))
	defer server.Close()

	f := fetcher.New(req.NewClient())

	result, err := f.Fetch(t.Context(), server.URL)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, []byte(minimalHTML), result.Body)
}

func TestFetchXHTMLSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/xhtml+xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(minimalHTML))
	}))
	defer server.Close()

	f := fetcher.New(req.NewClient())

	result, err := f.Fetch(t.Context(), server.URL)

	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestFetch401ReturnsARC002(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	f := fetcher.New(req.NewClient())

	_, err := f.Fetch(t.Context(), server.URL)

	require.ErrorIs(t, err, arc.ErrURLAccessDenied)
	assertFetcherError(t, err, "http status", http.StatusUnauthorized)
}

func TestFetch403ReturnsARC002(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	f := fetcher.New(req.NewClient())

	_, err := f.Fetch(t.Context(), server.URL)

	require.ErrorIs(t, err, arc.ErrURLAccessDenied)
	assertFetcherError(t, err, "http status", http.StatusForbidden)
}

func TestFetch404ReturnsARC003(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	f := fetcher.New(req.NewClient())

	_, err := f.Fetch(t.Context(), server.URL)

	require.ErrorIs(t, err, arc.ErrURLNotFound)
	assertFetcherError(t, err, "http status", http.StatusNotFound)
}

func TestFetchNonHTMLContentTypeReturnsARC005(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("%PDF-1.4 binary content"))
	}))
	defer server.Close()

	f := fetcher.New(req.NewClient())

	_, err := f.Fetch(t.Context(), server.URL)

	require.ErrorIs(t, err, arc.ErrResponseNotHTML)
	fetchErr := assertFetcherError(t, err, "validate content type", 0)
	require.Equal(t, "application/pdf", fetchErr.ContentType)
}

func TestFetchOversizedBodyReturnsARC006(t *testing.T) {
	const oversizedBodySize = 10*1024*1024 + 1

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		chunk := strings.Repeat("a", 4096)
		written := 0

		for written < oversizedBodySize {
			remaining := oversizedBodySize - written
			if remaining < len(chunk) {
				chunk = chunk[:remaining]
			}

			n, err := w.Write([]byte(chunk))
			if err != nil {
				return
			}

			written += n
		}
	}))
	defer server.Close()

	f := fetcher.New(req.NewClient())

	_, err := f.Fetch(t.Context(), server.URL)

	require.ErrorIs(t, err, arc.ErrResponseTooLarge)
	assertFetcherError(t, err, "read body", 0)
}

func TestFetchTimeoutReturnsARC004(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wait until the request context is cancelled (simulates a slow server).
		<-r.Context().Done()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	f := fetcher.New(req.NewClient())

	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()

	_, err := f.Fetch(ctx, server.URL)

	require.ErrorIs(t, err, arc.ErrURLFetchTransientFailure)
	fetchErr := assertFetcherError(t, err, "request", 0)
	require.NotEmpty(t, fetchErr.URL)
}

func TestFetchFTPSchemeReturnsARC001(t *testing.T) {
	f := fetcher.New(req.NewClient())

	_, err := f.Fetch(t.Context(), "ftp://example.com/file.txt")

	require.ErrorIs(t, err, arc.ErrURLResolutionFailed)
	assertFetcherError(t, err, "validate scheme", 0)
}

func TestFetchFileSchemeReturnsARC001(t *testing.T) {
	f := fetcher.New(req.NewClient())

	_, err := f.Fetch(t.Context(), "file:///etc/passwd")

	require.ErrorIs(t, err, arc.ErrURLResolutionFailed)
	assertFetcherError(t, err, "validate scheme", 0)
}

func TestFetchEmptySchemeReturnsARC001(t *testing.T) {
	f := fetcher.New(req.NewClient())

	_, err := f.Fetch(t.Context(), "not-a-url")

	require.ErrorIs(t, err, arc.ErrURLResolutionFailed)
	assertFetcherError(t, err, "validate scheme", 0)
}

func TestFetch5xxReturnsARC004(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	f := fetcher.New(req.NewClient())

	_, err := f.Fetch(t.Context(), server.URL)

	require.ErrorIs(t, err, arc.ErrURLFetchTransientFailure)
	assertFetcherError(t, err, "http status", http.StatusInternalServerError)
}

func assertFetcherError(t *testing.T, err error, op string, statusCode int) *fetcher.FetcherError {
	t.Helper()

	fetchErr, ok := errors.AsType[*fetcher.FetcherError](err)
	require.True(t, ok)
	require.Equal(t, op, fetchErr.Op)
	require.Equal(t, statusCode, fetchErr.StatusCode)

	return fetchErr
}
