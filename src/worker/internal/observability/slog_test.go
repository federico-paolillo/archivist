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

func TestMinLevelHandlerDropsRecordsBelowMinimum(t *testing.T) {
	var logs bytes.Buffer

	logger := slog.New(observability.NewMinLevelHandler(
		slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}),
		slog.LevelInfo,
	))

	logger.DebugContext(t.Context(), "debug message")
	logger.InfoContext(t.Context(), "info message")
	logger.ErrorContext(t.Context(), "error message")

	logText := logs.String()
	assert.NotContains(t, logText, "debug message")
	assert.Contains(t, logText, "info message")
	assert.Contains(t, logText, "error message")
}
