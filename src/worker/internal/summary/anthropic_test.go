package summary_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"codeberg.org/federico-paolillo/archivist/internal/summary"
	"github.com/stretchr/testify/require"
)

// buildAnthropicResponse returns a minimal Anthropic Messages API response body.
func buildAnthropicResponse(text string) string {
	resp := map[string]any{
		"id":           "msg_test123",
		"type":         "message",
		"role":         "assistant",
		"model":        "claude-3-5-haiku-20241022",
		"stop_reason":  "end_turn",
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

	logger := slog.Default()

	return summary.NewAnthropicAdapterWithBaseURL(
		"test-api-key",
		"claude-3-5-haiku-20241022",
		srv.URL,
		logger,
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

	result := adapter.Summarize(t.Context(), summary.SummarizerRequest{
		MarkdownSource: "# Article\n\nSome content.",
	})

	require.Equal(t, summary.ResultStatusSuccess, result.Status)
	require.Equal(t, summary.ProviderAnthropic, result.Provider)
	require.Equal(t, expectedSummary, result.Summary)
	require.Empty(t, result.ErrorCode)
	require.Empty(t, result.FailureReason)
}

func TestAnthropicAdapterEmptyOutputIsARC013(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(buildAnthropicResponse("")))
	}))
	t.Cleanup(srv.Close)

	adapter := newTestAdapter(t, srv)

	result := adapter.Summarize(t.Context(), summary.SummarizerRequest{
		MarkdownSource: "# Article\n\nSome content.",
	})

	require.Equal(t, summary.ResultStatusFailure, result.Status)
	require.Equal(t, summary.ProviderAnthropic, result.Provider)
	require.Equal(t, summary.ErrorCodeProviderFailure, result.ErrorCode)
}

func TestAnthropicAdapterGenericErrorIsARC013(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(buildAnthropicErrorResponse("api_error", "Internal server error")))
	}))
	t.Cleanup(srv.Close)

	adapter := newTestAdapter(t, srv)

	result := adapter.Summarize(t.Context(), summary.SummarizerRequest{
		MarkdownSource: "# Article\n\nSome content.",
	})

	require.Equal(t, summary.ResultStatusFailure, result.Status)
	require.Equal(t, summary.ProviderAnthropic, result.Provider)
	require.Equal(t, summary.ErrorCodeProviderFailure, result.ErrorCode)
}

func TestAnthropicAdapterRequestTooLargeIsARC014(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		_, _ = w.Write([]byte(buildAnthropicErrorResponse("request_too_large", "Request too large")))
	}))
	t.Cleanup(srv.Close)

	adapter := newTestAdapter(t, srv)

	result := adapter.Summarize(t.Context(), summary.SummarizerRequest{
		MarkdownSource: "# Article\n\nSome content.",
	})

	require.Equal(t, summary.ResultStatusFailure, result.Status)
	require.Equal(t, summary.ProviderAnthropic, result.Provider)
	require.Equal(t, summary.ErrorCodeRequestTooLarge, result.ErrorCode)
}

func TestAnthropicAdapterBillingErrorIsARC015(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		_, _ = w.Write([]byte(buildAnthropicErrorResponse("billing_error", "Billing error")))
	}))
	t.Cleanup(srv.Close)

	adapter := newTestAdapter(t, srv)

	result := adapter.Summarize(t.Context(), summary.SummarizerRequest{
		MarkdownSource: "# Article\n\nSome content.",
	})

	require.Equal(t, summary.ResultStatusFailure, result.Status)
	require.Equal(t, summary.ProviderAnthropic, result.Provider)
	require.Equal(t, summary.ErrorCodeBillingFailure, result.ErrorCode)
}
