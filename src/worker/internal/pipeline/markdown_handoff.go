package pipeline

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
	"codeberg.org/federico-paolillo/archivist/internal/artifacts"
	"codeberg.org/federico-paolillo/archivist/internal/markdown"
	"codeberg.org/federico-paolillo/archivist/internal/observability"
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// MarkdownExtractionHandoff runs the Markdown stage after snapshot promotion.
type MarkdownExtractionHandoff struct {
	logger         *slog.Logger
	repo           jobs.Repository
	store          *artifacts.Store
	local          markdown.MarkdownExtractor
	fallback       markdown.MarkdownExtractor
	summaryHandoff SummaryHandoff
}

var _ MarkdownHandoff = (*MarkdownExtractionHandoff)(nil)

// NewMarkdownExtractionHandoff creates a Markdown handoff backed by provider-neutral extractors.
func NewMarkdownExtractionHandoff(
	logger *slog.Logger,
	repo jobs.Repository,
	store *artifacts.Store,
	local markdown.MarkdownExtractor,
	fallback markdown.MarkdownExtractor,
	summaryHandoff SummaryHandoff,
) *MarkdownExtractionHandoff {
	return &MarkdownExtractionHandoff{
		logger:         logger,
		repo:           repo,
		store:          store,
		local:          local,
		fallback:       fallback,
		summaryHandoff: summaryHandoff,
	}
}

// Handoff extracts Markdown, atomically writes content.md, and leaves terminal
// success for the downstream summary stage.
//
//nolint:nonamedreturns,spancheck // Deferred span completion records the returned error.
func (h *MarkdownExtractionHandoff) Handoff(ctx context.Context, job *jobs.Job, canonicalURL string) (err error) {
	ctx, span := observability.Tracer().Start(
		ctx,
		"worker.pipeline.markdown",
		trace.WithAttributes(append(
			observability.JobAttributes(job.ArticleID, job.ID),
			attribute.String("url", canonicalURL),
		)...),
	)
	defer func() {
		observability.EndSpan(span, err)
	}()

	stageStart := time.Now()

	h.logMarkdownStageStart(ctx, job, canonicalURL)

	snapshot, err := h.readSnapshot(ctx, job)
	if err != nil {
		failure := pipelineFailure(
			"markdown",
			"read snapshot",
			arc.ErrLocalExtractionFailed,
			err,
			withJobContext(job.ArticleID, job.ID),
			withPipelineURL(canonicalURL),
		)
		h.logMarkdownStageFailure(ctx, "pipeline: markdown snapshot read failed", job, canonicalURL, "", stageStart, failure)

		return failure
	}

	input := markdown.ExtractInput{
		HTML:         snapshot,
		CanonicalURL: canonicalURL,
	}

	selected, provider, err := h.extract(ctx, job, canonicalURL, input)
	if err != nil {
		h.logMarkdownStageFailure(ctx, "pipeline: markdown stage completed", job, canonicalURL, "", stageStart, err)

		return err
	}

	err = h.writeMarkdown(ctx, job, canonicalURL, provider, selected.Markdown)
	if err != nil {
		h.logMarkdownStageFailure(ctx, "pipeline: markdown stage completed", job, canonicalURL, provider, stageStart, err)

		return err
	}

	h.persistTitle(ctx, job, canonicalURL, selected)
	h.logMarkdownStageSuccess(ctx, job, canonicalURL, provider, stageStart)

	err = h.summaryHandoff.Summarize(ctx, job, canonicalURL)
	if err != nil {
		return fmt.Errorf("pipeline: summary handoff: %w", err)
	}

	return nil
}

func (h *MarkdownExtractionHandoff) logMarkdownStageStart(ctx context.Context, job *jobs.Job, canonicalURL string) {
	h.logger.LogAttrs(
		ctx,
		slog.LevelInfo,
		"pipeline: markdown stage started",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("url", canonicalURL),
		slog.String("stage", "markdown"),
		slog.String("status", "start"),
	)
}

func (h *MarkdownExtractionHandoff) logMarkdownStageFailure(
	ctx context.Context,
	message string,
	job *jobs.Job,
	canonicalURL string,
	provider markdown.Provider,
	stageStart time.Time,
	err error,
) {
	attrs := []slog.Attr{
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("url", canonicalURL),
		slog.String("stage", "markdown"),
		slog.String("status", "failure"),
		slog.Duration("duration", time.Since(stageStart)),
		slog.String("arc_code", arc.CodeString(err)),
		slog.Any("error", err),
	}
	if provider != "" {
		attrs = append(attrs, slog.String("provider", string(provider)), slog.String("selected_provider", string(provider)))
	}

	h.logger.LogAttrs(ctx, slog.LevelInfo, message, attrs...)
}

func (h *MarkdownExtractionHandoff) logMarkdownStageSuccess(
	ctx context.Context,
	job *jobs.Job,
	canonicalURL string,
	provider markdown.Provider,
	stageStart time.Time,
) {
	h.logger.LogAttrs(
		ctx,
		slog.LevelInfo,
		"pipeline: markdown stage completed",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("url", canonicalURL),
		slog.String("provider", string(provider)),
		slog.String("selected_provider", string(provider)),
		slog.String("stage", "markdown"),
		slog.String("status", "success"),
		slog.Duration("duration", time.Since(stageStart)),
		slog.String("artifact_result", "success"),
	)
}

func (h *MarkdownExtractionHandoff) persistTitle(
	ctx context.Context,
	job *jobs.Job,
	canonicalURL string,
	output markdown.ExtractOutput,
) {
	title := selectedTitle(output)
	if title == "" {
		return
	}

	err := h.repo.UpdateArticleTitle(ctx, job.ArticleID, title)
	if err != nil {
		h.logger.ErrorContext(
			ctx,
			"pipeline: article title update failed",
			slog.String("article_id", job.ArticleID),
			slog.String("job_id", job.ID),
			slog.String("url", canonicalURL),
			slog.String("stage", "title_update"),
			slog.String("status", "failure"),
			slog.Any("error", err),
		)

		return
	}

	h.logger.InfoContext(
		ctx,
		"pipeline: article title updated",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("url", canonicalURL),
		slog.String("stage", "title_update"),
		slog.String("status", "success"),
	)
}

func selectedTitle(output markdown.ExtractOutput) string {
	if title := strings.TrimSpace(output.Title); title != "" {
		return title
	}

	return firstMarkdownH1(output.Markdown)
}

func firstMarkdownH1(content string) string {
	for line := range strings.SplitSeq(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "# ") {
			continue
		}

		title := strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
		if title != "" {
			return title
		}
	}

	return ""
}

//nolint:nonamedreturns,spancheck // Deferred span completion records the returned error.
func (h *MarkdownExtractionHandoff) writeMarkdown(
	ctx context.Context,
	job *jobs.Job,
	canonicalURL string,
	provider markdown.Provider,
	content string,
) (err error) {
	ctx, span := observability.Tracer().Start(
		ctx,
		"worker.pipeline.markdown_write",
		trace.WithAttributes(append(
			observability.JobAttributes(job.ArticleID, job.ID),
			attribute.String("url", canonicalURL),
			attribute.String("provider", string(provider)),
		)...),
	)
	defer func() {
		observability.EndSpan(span, err)
	}()

	writeErr := h.store.WriteMarkdown(job.ArticleID, strings.NewReader(content))
	if writeErr != nil {
		h.logger.ErrorContext(
			ctx,
			"pipeline: markdown write failed",
			slog.String("article_id", job.ArticleID),
			slog.String("job_id", job.ID),
			slog.String("url", canonicalURL),
			slog.String("provider", string(provider)),
			slog.String("selected_provider", string(provider)),
			slog.String("stage", "markdown_write"),
			slog.String("status", "failure"),
			slog.String("artifact_result", "failure"),
			slog.String("arc_code", arc.CodeString(arc.ErrMarkdownWrite)),
			slog.Any("error", writeErr),
		)

		err = pipelineFailure(
			"markdown",
			"write markdown",
			arc.ErrMarkdownWrite,
			writeErr,
			withJobContext(job.ArticleID, job.ID),
			withPipelineURL(canonicalURL),
		)

		return err
	}

	h.logger.InfoContext(
		ctx,
		"pipeline: markdown written",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("url", canonicalURL),
		slog.String("provider", string(provider)),
		slog.String("selected_provider", string(provider)),
		slog.String("artifact_result", "success"),
		slog.String("stage", "markdown_write"),
		slog.String("status", "markdown_done"),
	)

	return nil
}

//nolint:nonamedreturns,spancheck // Deferred span completion records the returned error.
func (h *MarkdownExtractionHandoff) readSnapshot(ctx context.Context, job *jobs.Job) (snapshot []byte, err error) {
	_, span := observability.Tracer().Start(
		ctx,
		"worker.pipeline.snapshot_read",
		trace.WithAttributes(observability.JobAttributes(job.ArticleID, job.ID)...),
	)
	defer func() {
		observability.EndSpan(span, err)
	}()

	rc, err := h.store.OpenSnapshot(job.ArticleID)
	if err != nil {
		return nil, fmt.Errorf("pipeline: open snapshot artifact: %w", err)
	}
	defer func() {
		_ = rc.Close()
	}()

	snapshot, err = io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("pipeline: read snapshot artifact: %w", err)
	}

	return snapshot, nil
}

func (h *MarkdownExtractionHandoff) extract(
	ctx context.Context,
	job *jobs.Job,
	canonicalURL string,
	input markdown.ExtractInput,
) (markdown.ExtractOutput, markdown.Provider, error) {
	localOutput, localErr := h.attempt(ctx, job, canonicalURL, h.local, input, "")
	if localErr == nil {
		provider := h.local.Provider()
		h.logSelectedProvider(ctx, job, canonicalURL, provider)

		return localOutput, provider, nil
	}

	fallbackReason := fallbackReason(localErr)

	fallbackOutput, fallbackErr := h.attempt(ctx, job, canonicalURL, h.fallback, input, fallbackReason)
	if fallbackErr == nil {
		provider := h.fallback.Provider()
		h.logSelectedProvider(ctx, job, canonicalURL, provider)

		return fallbackOutput, provider, nil
	}

	return markdown.ExtractOutput{}, "", fallbackErr
}

func (h *MarkdownExtractionHandoff) attempt(
	ctx context.Context,
	job *jobs.Job,
	canonicalURL string,
	extractor markdown.MarkdownExtractor,
	input markdown.ExtractInput,
	fallbackReason string,
) (markdown.ExtractOutput, error) {
	ctx, span := observability.Tracer().Start(
		ctx,
		"worker.pipeline.markdown_provider",
		trace.WithAttributes(append(
			observability.JobAttributes(job.ArticleID, job.ID),
			attribute.String("url", canonicalURL),
			attribute.String("provider", string(extractor.Provider())),
		)...),
	)

	start := time.Now()

	output, err := extractor.ExtractMarkdown(ctx, input)
	defer func() {
		observability.EndSpan(span, err)
	}()

	duration := time.Since(start)

	status := "success"
	if err != nil {
		status = "failure"
		if errors.Is(err, arc.ErrLocalUnreadable) {
			status = "local_unreadable"
		}
	}

	attrs := []slog.Attr{
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("url", canonicalURL),
		slog.String("provider", string(extractor.Provider())),
		slog.String("stage", "markdown_provider"),
		slog.String("status", status),
		slog.Duration("duration", duration),
	}
	if fallbackReason != "" {
		attrs = append(attrs, slog.String("fallback_reason", fallbackReason))
	}

	if err != nil {
		attrs = append(
			attrs,
			slog.String("arc_code", arc.CodeString(err)),
			slog.Any("error", err),
		)
	}

	h.logger.LogAttrs(ctx, slog.LevelInfo, "pipeline: markdown provider attempt", attrs...)

	if err != nil {
		return output, pipelineFailure(
			"markdown",
			"provider "+string(extractor.Provider()),
			err,
			nil,
			withJobContext(job.ArticleID, job.ID),
			withPipelineURL(canonicalURL),
		)
	}

	return output, nil
}

func (h *MarkdownExtractionHandoff) logSelectedProvider(
	ctx context.Context,
	job *jobs.Job,
	canonicalURL string,
	provider markdown.Provider,
) {
	h.logger.InfoContext(
		ctx,
		"pipeline: markdown provider selected",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("url", canonicalURL),
		slog.String("provider", string(provider)),
		slog.String("selected_provider", string(provider)),
		slog.String("stage", "markdown_provider"),
		slog.String("status", "selected"),
	)
}

func fallbackReason(err error) string {
	if errors.Is(err, arc.ErrLocalUnreadable) {
		return "local unreadable"
	}

	extractionErr, ok := errors.AsType[*markdown.ExtractionError](err)
	if ok && extractionErr.Reason != "" {
		return extractionErr.Reason
	}

	return err.Error()
}
