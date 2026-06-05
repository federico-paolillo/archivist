package pipeline

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
	"codeberg.org/federico-paolillo/archivist/internal/artifacts"
	"codeberg.org/federico-paolillo/archivist/internal/observability"
	"codeberg.org/federico-paolillo/archivist/internal/summary"
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// SummaryHandoff is called after content.md promotion.
type SummaryHandoff interface {
	Summarize(ctx context.Context, job *jobs.Job, canonicalURL string) error
}

// SummaryHandoffFunc adapts functions to SummaryHandoff.
type SummaryHandoffFunc func(ctx context.Context, job *jobs.Job, canonicalURL string) error

// Summarize implements SummaryHandoff.
func (f SummaryHandoffFunc) Summarize(ctx context.Context, job *jobs.Job, canonicalURL string) error {
	return f(ctx, job, canonicalURL)
}

// NoOpSummaryHandoff preserves isolated Markdown-stage tests.
var NoOpSummaryHandoff SummaryHandoff = SummaryHandoffFunc(func(_ context.Context, _ *jobs.Job, _ string) error {
	return nil
})

// SummaryGenerationHandoff runs summarization, writes summary.md, and commits
// terminal success only after the summary artifact has been promoted.
type SummaryGenerationHandoff struct {
	logger     *slog.Logger
	repo       jobs.Repository
	store      *artifacts.Store
	summarizer summary.SummarizerService
}

var _ SummaryHandoff = (*SummaryGenerationHandoff)(nil)

func NewSummaryGenerationHandoff(
	logger *slog.Logger,
	repo jobs.Repository,
	store *artifacts.Store,
	summarizer summary.SummarizerService,
) *SummaryGenerationHandoff {
	return &SummaryGenerationHandoff{
		logger:     logger,
		repo:       repo,
		store:      store,
		summarizer: summarizer,
	}
}

//nolint:nonamedreturns,spancheck // Deferred span completion records the returned error.
func (h *SummaryGenerationHandoff) Summarize(ctx context.Context, job *jobs.Job, canonicalURL string) (err error) {
	ctx, span := observability.Tracer().Start(
		ctx,
		"worker.pipeline.summary",
		trace.WithAttributes(append(
			observability.JobUserAttributes(job.ArticleID, job.ID, job.UserID),
			attribute.String("url", canonicalURL),
			attribute.String("provider", string(h.summarizer.Provider())),
			attribute.String("model", h.summarizer.Model()),
		)...),
	)
	defer func() {
		observability.EndSpan(span, err)
	}()

	stageStart := time.Now()

	h.logSummaryStageStart(ctx, job, canonicalURL)

	markdownSource, err := h.readMarkdown(ctx, job)
	if err != nil {
		h.logSummaryReadFailure(ctx, job, canonicalURL, stageStart, err)

		return pipelineFailure(
			"summary",
			"read markdown",
			arc.ErrUnknown,
			err,
			withJobContext(job.ArticleID, job.ID),
			withPipelineURL(canonicalURL),
		)
	}

	output, err := h.summarize(ctx, job, canonicalURL, markdownSource)
	if err != nil {
		return err
	}

	err = h.writeSummary(ctx, job, canonicalURL, output)
	if err != nil {
		return err
	}

	h.logSummaryStageSuccess(ctx, job, canonicalURL, output, stageStart)
	h.logTerminalSuccessStart(ctx, job, canonicalURL, output)

	err = h.completeTerminalSuccess(ctx, job)
	if err != nil {
		return h.compensateTerminalSuccessFailure(ctx, job, canonicalURL, output, err)
	}

	h.logTerminalSuccessCompleted(ctx, job, canonicalURL, output, stageStart)

	return nil
}

func (h *SummaryGenerationHandoff) logSummaryStageStart(ctx context.Context, job *jobs.Job, canonicalURL string) {
	h.logger.LogAttrs(
		ctx,
		slog.LevelInfo,
		"pipeline: summary stage started",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("user_id", job.UserID),
		slog.String("url", canonicalURL),
		slog.String("stage", "summary"),
		slog.String("status", "start"),
		slog.String("provider", string(h.summarizer.Provider())),
		slog.String("model", h.summarizer.Model()),
	)
}

func (h *SummaryGenerationHandoff) logSummaryReadFailure(
	ctx context.Context,
	job *jobs.Job,
	canonicalURL string,
	stageStart time.Time,
	err error,
) {
	h.logger.LogAttrs(
		ctx,
		slog.LevelError,
		"pipeline: summary markdown read failed",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("user_id", job.UserID),
		slog.String("url", canonicalURL),
		slog.String("stage", "summary"),
		slog.String("status", "failure"),
		slog.Duration("duration", time.Since(stageStart)),
		slog.String("arc_code", arc.CodeString(arc.ErrUnknown)),
		slog.Any("error", err),
	)
}

func (h *SummaryGenerationHandoff) logSummaryStageSuccess(
	ctx context.Context,
	job *jobs.Job,
	canonicalURL string,
	output summary.SummarizerOutput,
	stageStart time.Time,
) {
	h.logger.LogAttrs(ctx, slog.LevelInfo, "pipeline: summary stage completed", h.summaryAttrs(job, canonicalURL, output,
		slog.String("stage", "summary"),
		slog.String("status", "success"),
		slog.Duration("duration", time.Since(stageStart)))...)
}

func (h *SummaryGenerationHandoff) logTerminalSuccessStart(
	ctx context.Context,
	job *jobs.Job,
	canonicalURL string,
	output summary.SummarizerOutput,
) {
	h.logger.LogAttrs(ctx, slog.LevelInfo, "pipeline: terminal success persistence started", h.summaryAttrs(job, canonicalURL, output,
		slog.String("stage", "terminal_success"),
		slog.String("status", "start"))...)
}

func (h *SummaryGenerationHandoff) logTerminalSuccessCompleted(
	ctx context.Context,
	job *jobs.Job,
	canonicalURL string,
	output summary.SummarizerOutput,
	stageStart time.Time,
) {
	h.logger.LogAttrs(ctx, slog.LevelInfo, "pipeline: summary completed", h.summaryAttrs(job, canonicalURL, output,
		slog.String("stage", "terminal_success"),
		slog.String("status", "summary_done"),
		slog.Duration("duration", time.Since(stageStart)))...)
}

func (h *SummaryGenerationHandoff) summaryAttrs(
	job *jobs.Job,
	canonicalURL string,
	output summary.SummarizerOutput,
	extra ...slog.Attr,
) []slog.Attr {
	attrs := make([]slog.Attr, 0, 7+len(extra))
	attrs = append(
		attrs,
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("user_id", job.UserID),
		slog.String("url", canonicalURL),
		slog.String("provider", string(h.summarizer.Provider())),
		slog.String("model", h.summarizer.Model()),
		slog.String("request_id", output.RequestID),
		slog.String("artifact_result", "success"),
	)

	return append(attrs, extra...)
}

func (h *SummaryGenerationHandoff) compensateTerminalSuccessFailure(
	ctx context.Context,
	job *jobs.Job,
	canonicalURL string,
	output summary.SummarizerOutput,
	terminalErr error,
) error {
	cleanupErr := h.store.RemoveSummary(job.ArticleID)

	artifactResult := "removed"
	if cleanupErr != nil {
		artifactResult = "remove_failed"
	}

	attrs := []slog.Attr{
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("user_id", job.UserID),
		slog.String("url", canonicalURL),
		slog.String("provider", string(h.summarizer.Provider())),
		slog.String("model", h.summarizer.Model()),
		slog.String("request_id", output.RequestID),
		slog.String("stage", "terminal_success"),
		slog.String("status", "terminal_persist_failed"),
		slog.String("artifact_result", artifactResult),
		slog.Any("error", terminalErr),
		slog.Any("terminal_error", terminalErr),
	}
	if cleanupErr != nil {
		attrs = append(attrs, slog.Any("cleanup_error", cleanupErr))
	}

	h.logger.LogAttrs(
		ctx,
		slog.LevelError,
		"pipeline: summary terminal success persistence failed",
		attrs...,
	)

	if cleanupErr != nil {
		return fmt.Errorf(
			"pipeline: terminal persistence failure after summary promotion for job %s; cleanup failed removing summary.md: %w",
			job.ID,
			errors.Join(terminalErr, cleanupErr),
		)
	}

	return fmt.Errorf(
		"pipeline: terminal persistence failure after summary promotion for job %s; cleanup removed summary.md: %w",
		job.ID,
		terminalErr,
	)
}

//nolint:nonamedreturns,spancheck // Deferred span completion records the returned error.
func (h *SummaryGenerationHandoff) completeTerminalSuccess(ctx context.Context, job *jobs.Job) (err error) {
	ctx, span := observability.Tracer().Start(
		ctx,
		"worker.pipeline.terminal_success",
		trace.WithAttributes(observability.JobUserAttributes(job.ArticleID, job.ID, job.UserID)...),
	)
	defer func() {
		observability.EndSpan(span, err)
	}()

	err = h.repo.CompleteTerminal(ctx, job, jobs.TerminalOutcome{Success: true})
	if err != nil {
		return fmt.Errorf("pipeline: complete terminal success for job %s: %w", job.ID, err)
	}

	return nil
}

//nolint:nonamedreturns,spancheck // Deferred span completion records the returned error.
func (h *SummaryGenerationHandoff) readMarkdown(ctx context.Context, job *jobs.Job) (content string, err error) {
	_, span := observability.Tracer().Start(
		ctx,
		"worker.pipeline.markdown_read",
		trace.WithAttributes(observability.JobUserAttributes(job.ArticleID, job.ID, job.UserID)...),
	)
	defer func() {
		observability.EndSpan(span, err)
	}()

	rc, err := h.store.OpenMarkdown(job.ArticleID)
	if err != nil {
		return "", fmt.Errorf("pipeline: open markdown artifact: %w", err)
	}
	defer func() {
		_ = rc.Close()
	}()

	data, err := io.ReadAll(rc)
	if err != nil {
		return "", fmt.Errorf("pipeline: read markdown artifact: %w", err)
	}

	return string(data), nil
}

//nolint:funlen,nonamedreturns // Provider telemetry and error mapping are intentionally kept together.
func (h *SummaryGenerationHandoff) summarize(
	ctx context.Context,
	job *jobs.Job,
	canonicalURL string,
	markdownSource string,
) (output summary.SummarizerOutput, err error) {
	ctx, span := observability.Tracer().Start(
		ctx,
		"worker.pipeline.summary_provider",
		trace.WithAttributes(append(
			observability.JobUserAttributes(job.ArticleID, job.ID, job.UserID),
			attribute.String("url", canonicalURL),
			attribute.String("provider", string(h.summarizer.Provider())),
			attribute.String("model", h.summarizer.Model()),
		)...),
	)
	defer func() {
		observability.EndSpan(span, err)
	}()

	start := time.Now()
	output, err = h.summarizer.Summarize(ctx, summary.SummarizerRequest{
		MarkdownSource: markdownSource,
		ArticleID:      job.ArticleID,
		JobID:          job.ID,
		URL:            canonicalURL,
	})
	duration := time.Since(start)
	requestID := output.RequestID

	if err != nil {
		providerErr, ok := errors.AsType[*summary.ProviderError](err)
		if ok {
			requestID = providerErr.RequestID
		}
	}

	attrs := []slog.Attr{
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("user_id", job.UserID),
		slog.String("url", canonicalURL),
		slog.String("provider", string(h.summarizer.Provider())),
		slog.String("model", h.summarizer.Model()),
		slog.String("request_id", requestID),
		slog.String("stage", "summary"),
		slog.Duration("duration", duration),
	}

	if err != nil {
		attrs = append(
			attrs,
			slog.String("status", "failure"),
			slog.String("arc_code", arc.CodeString(err)),
			slog.Any("error", err),
		)
		h.logger.LogAttrs(ctx, slog.LevelInfo, "pipeline: summary provider attempt", attrs...)

		if _, ok := arc.CodeOf(err); ok {
			return summary.SummarizerOutput{}, pipelineFailure(
				"summary",
				"provider "+string(h.summarizer.Provider()),
				err,
				nil,
				withJobContext(job.ArticleID, job.ID),
				withPipelineURL(canonicalURL),
			)
		}

		return summary.SummarizerOutput{}, pipelineFailure(
			"summary",
			"provider "+string(h.summarizer.Provider()),
			arc.ErrUnknown,
			err,
			withJobContext(job.ArticleID, job.ID),
			withPipelineURL(canonicalURL),
		)
	}

	attrs = append(attrs, slog.String("status", "success"))
	h.logger.LogAttrs(ctx, slog.LevelInfo, "pipeline: summary provider attempt", attrs...)

	return output, nil
}

//nolint:nonamedreturns,spancheck // Deferred span completion records the returned error.
func (h *SummaryGenerationHandoff) writeSummary(
	ctx context.Context,
	job *jobs.Job,
	canonicalURL string,
	output summary.SummarizerOutput,
) (err error) {
	ctx, span := observability.Tracer().Start(
		ctx,
		"worker.pipeline.summary_write",
		trace.WithAttributes(append(
			observability.JobUserAttributes(job.ArticleID, job.ID, job.UserID),
			attribute.String("url", canonicalURL),
			attribute.String("provider", string(h.summarizer.Provider())),
			attribute.String("model", h.summarizer.Model()),
			attribute.String("request_id", output.RequestID),
		)...),
	)
	defer func() {
		observability.EndSpan(span, err)
	}()

	err = h.store.WriteSummary(job.ArticleID, bytes.NewBufferString(output.Summary))
	if err != nil {
		h.logger.ErrorContext(
			ctx,
			"pipeline: summary write failed",
			slog.String("article_id", job.ArticleID),
			slog.String("job_id", job.ID),
			slog.String("user_id", job.UserID),
			slog.String("url", canonicalURL),
			slog.String("provider", string(h.summarizer.Provider())),
			slog.String("model", h.summarizer.Model()),
			slog.String("request_id", output.RequestID),
			slog.String("stage", "summary_write"),
			slog.String("status", "failure"),
			slog.String("artifact_result", "failure"),
			slog.String("arc_code", arc.CodeString(arc.ErrSummaryWrite)),
			slog.Any("error", err),
		)

		return pipelineFailure(
			"summary",
			"write summary",
			arc.ErrSummaryWrite,
			err,
			withJobContext(job.ArticleID, job.ID),
			withPipelineURL(canonicalURL),
		)
	}

	h.logger.InfoContext(
		ctx,
		"pipeline: summary written",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("user_id", job.UserID),
		slog.String("url", canonicalURL),
		slog.String("provider", string(h.summarizer.Provider())),
		slog.String("model", h.summarizer.Model()),
		slog.String("request_id", output.RequestID),
		slog.String("stage", "summary_write"),
		slog.String("status", "success"),
		slog.String("artifact_result", "success"),
	)

	return nil
}
