package observability_test

import (
	"bytes"
	"log/slog"
	"testing"

	"codeberg.org/federico-paolillo/archivist/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestTraceContextHandlerAddsTraceAndSpanIDs(t *testing.T) {
	var logs bytes.Buffer

	logger := slog.New(observability.NewTraceContextHandler(slog.NewJSONHandler(&logs, nil)))
	tracerProvider := sdktrace.NewTracerProvider()
	defer func() {
		require.NoError(t, tracerProvider.Shutdown(t.Context()))
	}()

	ctx, span := tracerProvider.Tracer("test").Start(t.Context(), "test span")
	defer span.End()

	logger.InfoContext(ctx, "test message")

	logText := logs.String()
	assert.Contains(t, logText, `"trace_id":"`)
	assert.Contains(t, logText, `"span_id":"`)
}

func TestTraceContextHandlerOmitsTraceFieldsWithoutSpan(t *testing.T) {
	var logs bytes.Buffer

	logger := slog.New(observability.NewTraceContextHandler(slog.NewJSONHandler(&logs, nil)))

	logger.InfoContext(t.Context(), "test message")

	logText := logs.String()
	assert.NotContains(t, logText, "trace_id")
	assert.NotContains(t, logText, "span_id")
}

func TestTeeHandlerSkipsDisabledHandlers(t *testing.T) {
	var enabledLogs bytes.Buffer
	var disabledLogs bytes.Buffer

	enabledHandler := slog.NewJSONHandler(&enabledLogs, &slog.HandlerOptions{Level: slog.LevelInfo})
	disabledHandler := slog.NewJSONHandler(&disabledLogs, &slog.HandlerOptions{Level: slog.LevelError})
	logger := slog.New(observability.NewTeeHandler(disabledHandler, enabledHandler))

	logger.InfoContext(t.Context(), "test message")

	assert.Contains(t, enabledLogs.String(), `"msg":"test message"`)
	assert.Empty(t, disabledLogs.String())
}
