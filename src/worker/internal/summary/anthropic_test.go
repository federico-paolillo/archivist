package summary_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
	"codeberg.org/federico-paolillo/archivist/internal/summary"
	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/require"
)

// buildAnthropicResponse returns a minimal Anthropic Messages API response body.
func buildAnthropicResponse(text string) string {
	return buildAnthropicResponseWithStopReason(text, "end_turn")
}

func buildAnthropicResponseWithStopReason(text string, stopReason string) string {
	resp := map[string]any{
		"id":            "msg_test123",
		"type":          "message",
		"role":          "assistant",
		"model":         "claude-3-5-haiku-20241022",
		"stop_reason":   stopReason,
		"stop_sequence": nil,
		"content": []map[string]any{
			{"type": "text", "text": text},
		},
		"usage": map[string]any{
			"input_tokens":  10,
			"output_tokens": 20,
		},
	}

	b, _ := json.Marshal(resp)

	return string(b)
}

// buildAnthropicErrorResponse returns an Anthropic API error body.
func buildAnthropicErrorResponse(errType, message string) string {
	resp := map[string]any{
		"type": "error",
		"error": map[string]any{
			"type":    errType,
			"message": message,
		},
	}

	b, _ := json.Marshal(resp)

	return string(b)
}

func newTestAdapter(t *testing.T, srv *httptest.Server) *summary.AnthropicAdapter {
	t.Helper()

	return summary.NewAnthropicAdapterWithBaseURL(
		req.NewClient(),
		"test-api-key",
		"claude-3-5-haiku-20241022",
		srv.URL,
	)
}

func TestAnthropicAdapterSDKIsolation(t *testing.T) {
	// Compile-time assertion is in anthropic.go via var _ SummarizerService = (*AnthropicAdapter)(nil).
	// This test documents that the assertion is present and passing.
	var _ summary.SummarizerService = (*summary.AnthropicAdapter)(nil)
}

func TestAnthropicAdapterSuccessPath(t *testing.T) {
	expectedSummary := "This is a concise summary of the article."

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(buildAnthropicResponse(expectedSummary)))
	}))
	t.Cleanup(srv.Close)

	adapter := newTestAdapter(t, srv)

	output, err := adapter.Summarize(t.Context(), summary.SummarizerRequest{
		MarkdownSource: "# Article\n\nSome content.",
	})

	require.NoError(t, err)
	require.Equal(t, summary.ProviderAnthropic, adapter.Provider())
	require.Equal(t, expectedSummary, output.Summary)
	require.Equal(t, "msg_test123", output.RequestID)
}

func TestAnthropicAdapterEmptyOutputIsARC013(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(buildAnthropicResponse("")))
	}))
	t.Cleanup(srv.Close)

	adapter := newTestAdapter(t, srv)

	_, err := adapter.Summarize(t.Context(), summary.SummarizerRequest{
		MarkdownSource: "# Article\n\nSome content.",
	})

	require.ErrorIs(t, err, arc.ErrSummarizerProviderFailure)
	providerErr, ok := errors.AsType[*summary.ProviderError](err)
	require.True(t, ok)
	require.Equal(t, summary.ProviderAnthropic, providerErr.Provider)
	require.Equal(t, "msg_test123", providerErr.RequestID)
}

func TestAnthropicAdapterGenericErrorIsARC013(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(buildAnthropicErrorResponse("api_error", "Internal server error")))
	}))
	t.Cleanup(srv.Close)

	adapter := newTestAdapter(t, srv)

	_, err := adapter.Summarize(t.Context(), summary.SummarizerRequest{
		MarkdownSource: "# Article\n\nSome content.",
	})

	require.ErrorIs(t, err, arc.ErrSummarizerProviderFailure)
	providerErr, ok := errors.AsType[*summary.ProviderError](err)
	require.True(t, ok)
	require.Equal(t, summary.ProviderAnthropic, providerErr.Provider)
	require.Equal(t, http.StatusInternalServerError, providerErr.StatusCode)
}

func TestAnthropicAdapterRequestTooLargeIsARC014(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		_, _ = w.Write([]byte(buildAnthropicErrorResponse("request_too_large", "Request too large")))
	}))
	t.Cleanup(srv.Close)

	adapter := newTestAdapter(t, srv)

	_, err := adapter.Summarize(t.Context(), summary.SummarizerRequest{
		MarkdownSource: "# Article\n\nSome content.",
	})

	require.ErrorIs(t, err, arc.ErrSummarizerRequestTooLarge)
	providerErr, ok := errors.AsType[*summary.ProviderError](err)
	require.True(t, ok)
	require.Equal(t, http.StatusRequestEntityTooLarge, providerErr.StatusCode)
}

func TestAnthropicAdapterInvalidRequestContextOverflowIsARC014(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(buildAnthropicErrorResponse(
			"invalid_request_error",
			"input tokens and max_tokens exceed the model context window",
		)))
	}))
	t.Cleanup(srv.Close)

	adapter := newTestAdapter(t, srv)

	_, err := adapter.Summarize(t.Context(), summary.SummarizerRequest{
		MarkdownSource: "# Article\n\nSome content.",
	})

	require.ErrorIs(t, err, arc.ErrSummarizerRequestTooLarge)
	providerErr, ok := errors.AsType[*summary.ProviderError](err)
	require.True(t, ok)
	require.Equal(t, http.StatusBadRequest, providerErr.StatusCode)
}

func TestAnthropicAdapterInvalidRequestNonSizeErrorIsARC013(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(buildAnthropicErrorResponse(
			"invalid_request_error",
			"Prefilling assistant messages is not supported for this model.",
		)))
	}))
	t.Cleanup(srv.Close)

	adapter := newTestAdapter(t, srv)

	_, err := adapter.Summarize(t.Context(), summary.SummarizerRequest{
		MarkdownSource: "# Article\n\nSome content.",
	})

	require.ErrorIs(t, err, arc.ErrSummarizerProviderFailure)
	require.NotErrorIs(t, err, arc.ErrSummarizerRequestTooLarge)
	providerErr, ok := errors.AsType[*summary.ProviderError](err)
	require.True(t, ok)
	require.Equal(t, http.StatusBadRequest, providerErr.StatusCode)
}

func TestAnthropicAdapterModelContextWindowStopReasonIsARC014(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(buildAnthropicResponseWithStopReason(
			"partial summary",
			"model_context_window_exceeded",
		)))
	}))
	t.Cleanup(srv.Close)

	adapter := newTestAdapter(t, srv)

	_, err := adapter.Summarize(t.Context(), summary.SummarizerRequest{
		MarkdownSource: "# Article\n\nSome content.",
	})

	require.ErrorIs(t, err, arc.ErrSummarizerRequestTooLarge)
	providerErr, ok := errors.AsType[*summary.ProviderError](err)
	require.True(t, ok)
	require.Equal(t, "msg_test123", providerErr.RequestID)
}

func TestAnthropicAdapterBillingErrorIsARC015(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		_, _ = w.Write([]byte(buildAnthropicErrorResponse("billing_error", "Billing error")))
	}))
	t.Cleanup(srv.Close)

	adapter := newTestAdapter(t, srv)

	_, err := adapter.Summarize(t.Context(), summary.SummarizerRequest{
		MarkdownSource: "# Article\n\nSome content.",
	})

	require.ErrorIs(t, err, arc.ErrSummarizerBillingFailure)
	providerErr, ok := errors.AsType[*summary.ProviderError](err)
	require.True(t, ok)
	require.Equal(t, http.StatusPaymentRequired, providerErr.StatusCode)
}

func TestAnthropicAdapterContextCancellationIsARC013(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(buildAnthropicResponse("this should not be reached")))
	}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	adapter := newTestAdapter(t, srv)

	_, err := adapter.Summarize(ctx, summary.SummarizerRequest{
		MarkdownSource: "# Article\n\nSome content.",
	})

	require.ErrorIs(t, err, arc.ErrSummarizerProviderFailure)
}
