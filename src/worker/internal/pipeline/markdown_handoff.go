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
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
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
func (h *MarkdownExtractionHandoff) Handoff(ctx context.Context, job *jobs.Job, canonicalURL string) error {
	snapshot, err := h.readSnapshot(job.ArticleID)
	if err != nil {
		h.logger.Error(
			"pipeline: markdown snapshot read failed",
			slog.String("article_id", job.ArticleID),
			slog.String("job_id", job.ID),
			slog.String("url", canonicalURL),
			slog.String("arc_code", arc.CodeString(arc.ErrLocalExtractionFailed)),
			slog.Any("error", err),
		)

		return pipelineFailure(
			"markdown",
			"read snapshot",
			arc.ErrLocalExtractionFailed,
			err,
			withJobContext(job.ArticleID, job.ID),
			withPipelineURL(canonicalURL),
		)
	}

	input := markdown.ExtractInput{
		HTML:         snapshot,
		CanonicalURL: canonicalURL,
	}

	selected, provider, err := h.extract(ctx, job, canonicalURL, input)
	if err != nil {
		return err
	}

	err = h.writeMarkdown(job, canonicalURL, provider, selected.Markdown)
	if err != nil {
		return err
	}

	h.persistTitle(ctx, job, canonicalURL, selected)

	err = h.summaryHandoff.Summarize(ctx, job, canonicalURL)
	if err != nil {
		return fmt.Errorf("pipeline: summary handoff: %w", err)
	}

	return nil
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
		h.logger.Error(
			"pipeline: article title update failed",
			slog.String("article_id", job.ArticleID),
			slog.String("job_id", job.ID),
			slog.String("url", canonicalURL),
			slog.Any("error", err),
		)

		return
	}

	h.logger.Info(
		"pipeline: article title updated",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("url", canonicalURL),
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

func (h *MarkdownExtractionHandoff) writeMarkdown(
	job *jobs.Job,
	canonicalURL string,
	provider markdown.Provider,
	content string,
) error {
	writeErr := h.store.WriteMarkdown(job.ArticleID, strings.NewReader(content))
	if writeErr != nil {
		h.logger.Error(
			"pipeline: markdown write failed",
			slog.String("article_id", job.ArticleID),
			slog.String("job_id", job.ID),
			slog.String("url", canonicalURL),
			slog.String("provider", string(provider)),
			slog.String("selected_provider", string(provider)),
			slog.String("artifact_result", "failure"),
			slog.String("arc_code", arc.CodeString(arc.ErrMarkdownWrite)),
			slog.Any("error", writeErr),
		)

		return pipelineFailure(
			"markdown",
			"write markdown",
			arc.ErrMarkdownWrite,
			writeErr,
			withJobContext(job.ArticleID, job.ID),
			withPipelineURL(canonicalURL),
		)
	}

	h.logger.Info(
		"pipeline: markdown written",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("url", canonicalURL),
		slog.String("provider", string(provider)),
		slog.String("selected_provider", string(provider)),
		slog.String("artifact_result", "success"),
		slog.String("status", "markdown_done"),
	)

	return nil
}

func (h *MarkdownExtractionHandoff) readSnapshot(articleID string) ([]byte, error) {
	rc, err := h.store.OpenSnapshot(articleID)
	if err != nil {
		return nil, fmt.Errorf("pipeline: open snapshot artifact: %w", err)
	}
	defer func() {
		_ = rc.Close()
	}()

	snapshot, err := io.ReadAll(rc)
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
		h.logSelectedProvider(job, canonicalURL, provider)

		return localOutput, provider, nil
	}

	fallbackReason := fallbackReason(localErr)

	fallbackOutput, fallbackErr := h.attempt(ctx, job, canonicalURL, h.fallback, input, fallbackReason)
	if fallbackErr == nil {
		provider := h.fallback.Provider()
		h.logSelectedProvider(job, canonicalURL, provider)

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
	start := time.Now()
	output, err := extractor.ExtractMarkdown(ctx, input)
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
	job *jobs.Job,
	canonicalURL string,
	provider markdown.Provider,
) {
	h.logger.Info(
		"pipeline: markdown provider selected",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("url", canonicalURL),
		slog.String("provider", string(provider)),
		slog.String("selected_provider", string(provider)),
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
