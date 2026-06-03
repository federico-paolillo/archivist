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

type TeeHandler struct {
	handlers []slog.Handler
}

func NewTeeHandler(handlers ...slog.Handler) *TeeHandler {
	return &TeeHandler{handlers: handlers}
}

func (h *TeeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}

	return false
}

func (h *TeeHandler) Handle(ctx context.Context, rec slog.Record) error {
	var err error

	for _, handler := range h.handlers {
		if !handler.Enabled(ctx, rec.Level) {
			continue
		}

		handlerErr := handler.Handle(ctx, rec.Clone())
		if handlerErr != nil {
			err = errorsJoin(err, fmt.Errorf("observability: handle tee record: %w", handlerErr))
		}
	}

	return err
}

func (h *TeeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithAttrs(attrs))
	}

	return &TeeHandler{handlers: handlers}
}

func (h *TeeHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithGroup(name))
	}

	return &TeeHandler{handlers: handlers}
}

func errorsJoin(left error, right error) error {
	if left == nil {
		return right
	}

	if right == nil {
		return left
	}

	return &joinedError{left: left, right: right}
}

type joinedError struct {
	left  error
	right error
}

func (e *joinedError) Error() string {
	return e.left.Error() + "; " + e.right.Error()
}

func (e *joinedError) Unwrap() []error {
	return []error{e.left, e.right}
}
