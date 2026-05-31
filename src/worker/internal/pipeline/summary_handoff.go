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
	"codeberg.org/federico-paolillo/archivist/internal/summary"
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
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

func (h *SummaryGenerationHandoff) Summarize(ctx context.Context, job *jobs.Job, canonicalURL string) error {
	markdownSource, err := h.readMarkdown(job.ArticleID)
	if err != nil {
		h.logger.Error(
			"pipeline: summary markdown read failed",
			slog.String("article_id", job.ArticleID),
			slog.String("job_id", job.ID),
			slog.String("url", canonicalURL),
			slog.String("arc_code", arc.CodeString(arc.ErrUnknown)),
			slog.Any("error", err),
		)

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

	err = h.writeSummary(job, canonicalURL, output)
	if err != nil {
		return err
	}

	err = h.repo.CompleteTerminal(ctx, job, jobs.TerminalOutcome{Success: true})
	if err != nil {
		return h.compensateTerminalSuccessFailure(ctx, job, canonicalURL, output, err)
	}

	h.logger.Info(
		"pipeline: summary completed",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("url", canonicalURL),
		slog.String("provider", string(h.summarizer.Provider())),
		slog.String("model", h.summarizer.Model()),
		slog.String("request_id", output.RequestID),
		slog.String("status", "summary_done"),
		slog.String("artifact_result", "success"),
	)

	return nil
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
		slog.String("url", canonicalURL),
		slog.String("provider", string(h.summarizer.Provider())),
		slog.String("model", h.summarizer.Model()),
		slog.String("request_id", output.RequestID),
		slog.String("status", "terminal_persist_failed"),
		slog.String("artifact_result", artifactResult),
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

func (h *SummaryGenerationHandoff) readMarkdown(articleID string) (string, error) {
	rc, err := h.store.OpenMarkdown(articleID)
	if err != nil {
		return "", fmt.Errorf("pipeline: open markdown artifact: %w", err)
	}
	defer func() {
		_ = rc.Close()
	}()

	content, err := io.ReadAll(rc)
	if err != nil {
		return "", fmt.Errorf("pipeline: read markdown artifact: %w", err)
	}

	return string(content), nil
}

func (h *SummaryGenerationHandoff) summarize(
	ctx context.Context,
	job *jobs.Job,
	canonicalURL string,
	markdownSource string,
) (summary.SummarizerOutput, error) {
	start := time.Now()
	output, err := h.summarizer.Summarize(ctx, summary.SummarizerRequest{
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
		slog.String("url", canonicalURL),
		slog.String("provider", string(h.summarizer.Provider())),
		slog.String("model", h.summarizer.Model()),
		slog.String("request_id", requestID),
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

func (h *SummaryGenerationHandoff) writeSummary(
	job *jobs.Job,
	canonicalURL string,
	output summary.SummarizerOutput,
) error {
	err := h.store.WriteSummary(job.ArticleID, bytes.NewBufferString(output.Summary))
	if err != nil {
		h.logger.Error(
			"pipeline: summary write failed",
			slog.String("article_id", job.ArticleID),
			slog.String("job_id", job.ID),
			slog.String("url", canonicalURL),
			slog.String("provider", string(h.summarizer.Provider())),
			slog.String("model", h.summarizer.Model()),
			slog.String("request_id", output.RequestID),
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

	h.logger.Info(
		"pipeline: summary written",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("url", canonicalURL),
		slog.String("provider", string(h.summarizer.Provider())),
		slog.String("model", h.summarizer.Model()),
		slog.String("request_id", output.RequestID),
		slog.String("artifact_result", "success"),
	)

	return nil
}
