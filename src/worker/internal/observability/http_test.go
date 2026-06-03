package observability_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"codeberg.org/federico-paolillo/archivist/internal/observability"
	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestReqRoundTripWrapperRecordsSafeHTTPAttributes(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	previousPropagator := otel.GetTextMapPropagator()

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() {
		require.NoError(t, tracerProvider.Shutdown(context.Background()))
		otel.SetTracerProvider(noop.NewTracerProvider())
		otel.SetTextMapPropagator(previousPropagator)
	})

	var traceparent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceparent = r.Header.Get("traceparent")
		w.WriteHeader(http.StatusAccepted)
	}))
	t.Cleanup(server.Close)

	client := req.NewClient().WrapRoundTripFunc(observability.ReqRoundTripWrapper())

	resp, err := client.R().
		SetHeader("Authorization", "Bearer secret").
		SetBodyString("secret body").
		Get(server.URL + "/article?token=secret#fragment")

	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp.GetStatusCode())
	assert.NotEmpty(t, traceparent)

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	span := spans[0]
	attrs := attributesByKey(span.Attributes())

	assert.Equal(t, trace.SpanKindClient, span.SpanKind())
	assert.Equal(t, "HTTP GET", span.Name())
	assert.Equal(t, "GET", attrs["http.request.method"].AsString())
	assert.Equal(t, http.StatusAccepted, int(attrs["http.response.status_code"].AsInt64()))
	assert.Equal(t, server.URL+"/article", attrs["url.full"].AsString())
	assert.Equal(t, "/article", attrs["url.path"].AsString())
	assert.NotContains(t, attributeKeys(span.Attributes()), "http.request.header.authorization")
	assert.NotContains(t, attributeKeys(span.Attributes()), "http.request.body.size")
	assert.NotContains(t, attributeKeys(span.Attributes()), "http.response.header.authorization")
	assert.NotContains(t, attrs["url.full"].AsString(), "token=secret")
	assert.NotContains(t, attrs["url.full"].AsString(), "fragment")
}

func attributesByKey(attrs []attribute.KeyValue) map[string]attribute.Value {
	result := make(map[string]attribute.Value, len(attrs))
	for _, attr := range attrs {
		result[string(attr.Key)] = attr.Value
	}

	return result
}

func attributeKeys(attrs []attribute.KeyValue) []string {
	result := make([]string, 0, len(attrs))
	for _, attr := range attrs {
		result = append(result, string(attr.Key))
	}

	return result
}
