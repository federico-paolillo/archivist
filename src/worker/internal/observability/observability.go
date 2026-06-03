package observability

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	ServiceName         = "archivist-worker"
	InstrumentationName = "codeberg.org/federico-paolillo/archivist/src/worker"
)

type Providers struct {
	Logger         *slog.Logger
	tracerProvider *sdktrace.TracerProvider
	loggerProvider *sdklog.LoggerProvider
}

func Setup(ctx context.Context, stdoutHandler slog.Handler) (*Providers, error) {
	otel.SetTextMapPropagator(propagation.TraceContext{})

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithAttributes(semconv.ServiceName(ServiceName)),
	)
	if err != nil {
		return nil, fmt.Errorf("observability: create resource: %w", err)
	}

	handlers := make([]slog.Handler, 0, 2)
	handlers = append(handlers, NewTraceContextHandler(stdoutHandler))
	providers := &Providers{}

	traceExporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("observability: create trace exporter: %w", err)
	}

	providers.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExporter),
	)
	otel.SetTracerProvider(providers.tracerProvider)

	logExporter, err := otlploghttp.New(ctx)
	if err != nil {
		shutdownErr := providers.Shutdown(ctx)

		return nil, errors.Join(fmt.Errorf("observability: create log exporter: %w", err), shutdownErr)
	}

	providers.loggerProvider = sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
	)

	handlers = append(handlers, otelslog.NewHandler(
		ServiceName,
		otelslog.WithLoggerProvider(providers.loggerProvider),
	))

	providers.Logger = slog.New(NewTeeHandler(handlers...))

	return providers, nil
}

func (p *Providers) Shutdown(ctx context.Context) error {
	if p == nil {
		return nil
	}

	var errs []error

	if p.tracerProvider != nil {
		errs = append(errs, p.tracerProvider.Shutdown(ctx))
	}

	if p.loggerProvider != nil {
		errs = append(errs, p.loggerProvider.Shutdown(ctx))
	}

	return errors.Join(errs...)
}

//nolint:ireturn // OpenTelemetry exposes tracers through the trace.Tracer interface.
func Tracer() trace.Tracer {
	return otel.Tracer(InstrumentationName)
}

func EndSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	span.End()
}

func JobAttributes(articleID, jobID string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("archivist.article.id", articleID),
		attribute.String("archivist.job.id", jobID),
	}
}

func ExtractTraceContext(ctx context.Context, traceparent string, tracestate string) context.Context {
	if traceparent == "" {
		return ctx
	}

	carrier := propagation.MapCarrier{"traceparent": traceparent}
	if tracestate != "" {
		carrier["tracestate"] = tracestate
	}

	return propagation.TraceContext{}.Extract(ctx, carrier)
}

func InjectTraceContext(ctx context.Context) (string, string) {
	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() {
		return "", ""
	}

	carrier := propagation.MapCarrier{}
	propagation.TraceContext{}.Inject(ctx, carrier)

	return carrier.Get("traceparent"), carrier.Get("tracestate")
}
