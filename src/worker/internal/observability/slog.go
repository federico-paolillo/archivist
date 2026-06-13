package observability

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

type TraceContextHandler struct {
	next slog.Handler
}

func NewTraceContextHandler(next slog.Handler) *TraceContextHandler {
	return &TraceContextHandler{next: next}
}

func (h *TraceContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *TraceContextHandler) Handle(ctx context.Context, rec slog.Record) error {
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		rec.AddAttrs(
			slog.String("trace_id", spanContext.TraceID().String()),
			slog.String("span_id", spanContext.SpanID().String()),
		)
	}

	err := h.next.Handle(ctx, rec)
	if err != nil {
		return fmt.Errorf("observability: handle trace context record: %w", err)
	}

	return nil
}

func (h *TraceContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TraceContextHandler{next: h.next.WithAttrs(attrs)}
}

func (h *TraceContextHandler) WithGroup(name string) slog.Handler {
	return &TraceContextHandler{next: h.next.WithGroup(name)}
}

type MinLevelHandler struct {
	next slog.Handler
	min  slog.Level
}

func NewMinLevelHandler(next slog.Handler, minimum slog.Level) *MinLevelHandler {
	return &MinLevelHandler{next: next, min: minimum}
}

func (h *MinLevelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.min && h.next.Enabled(ctx, level)
}

func (h *MinLevelHandler) Handle(ctx context.Context, rec slog.Record) error {
	if !h.Enabled(ctx, rec.Level) {
		return nil
	}

	err := h.next.Handle(ctx, rec)
	if err != nil {
		return fmt.Errorf("observability: handle min-level record: %w", err)
	}

	return nil
}

func (h *MinLevelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &MinLevelHandler{next: h.next.WithAttrs(attrs), min: h.min}
}

func (h *MinLevelHandler) WithGroup(name string) slog.Handler {
	return &MinLevelHandler{next: h.next.WithGroup(name), min: h.min}
}
