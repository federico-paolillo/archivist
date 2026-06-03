package observability

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestSetupAlwaysCreatesTraceAndLogProviders(t *testing.T) {
	t.Setenv("OTEL_SDK_DISABLED", "true")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")
	t.Setenv("OTEL_LOGS_EXPORTER", "none")
	previousPropagator := otel.GetTextMapPropagator()

	providers, err := Setup(t.Context(), slog.NewJSONHandler(io.Discard, nil))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, providers.Shutdown(context.Background()))
		otel.SetTracerProvider(noop.NewTracerProvider())
		otel.SetTextMapPropagator(previousPropagator)
	})

	require.NotNil(t, providers.tracerProvider)
	require.NotNil(t, providers.loggerProvider)
	require.NotNil(t, providers.Logger)
	require.IsType(t, propagation.TraceContext{}, otel.GetTextMapPropagator())
}
