package pipeline

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"codeberg.org/federico-paolillo/archivist/internal/artifacts"
	"codeberg.org/federico-paolillo/archivist/internal/markdown"
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
)

// MarkdownExtractionHandoff runs the Markdown stage after snapshot promotion.
type MarkdownExtractionHandoff struct {
	logger   *slog.Logger
	store    *artifacts.Store
	local    markdown.MarkdownExtractor
	fallback markdown.MarkdownExtractor
}

var _ MarkdownHandoff = (*MarkdownExtractionHandoff)(nil)

// NewMarkdownExtractionHandoff creates a Markdown handoff backed by provider-neutral extractors.
func NewMarkdownExtractionHandoff(
	logger *slog.Logger,
	store *artifacts.Store,
	local markdown.MarkdownExtractor,
	fallback markdown.MarkdownExtractor,
) *MarkdownExtractionHandoff {
	return &MarkdownExtractionHandoff{
		logger:   logger,
		store:    store,
		local:    local,
		fallback: fallback,
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
			slog.Any("err", err),
		)

		return ErrLocalMarkdownExtraction
	}

	input := markdown.ExtractInput{
		HTML:         snapshot,
		CanonicalURL: canonicalURL,
	}

	selected, err := h.extract(ctx, job, canonicalURL, input)
	if err != nil {
		return err
	}

	writeErr := h.store.WriteMarkdown(job.ArticleID, strings.NewReader(selected.Markdown))
	if writeErr != nil {
		h.logger.Error(
			"pipeline: markdown write failed",
			slog.String("article_id", job.ArticleID),
			slog.String("job_id", job.ID),
			slog.String("url", canonicalURL),
			slog.String("provider", string(selected.Provider)),
			slog.String("selected_provider", string(selected.Provider)),
			slog.String("artifact_result", "failure"),
			slog.String("arc_code", arcCode(ErrMarkdownWrite)),
			slog.Any("err", writeErr),
		)

		return ErrMarkdownWrite
	}

	h.logger.Info(
		"pipeline: markdown written",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("url", canonicalURL),
		slog.String("provider", string(selected.Provider)),
		slog.String("selected_provider", string(selected.Provider)),
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
) (markdown.ExtractResult, error) {
	localResult := h.attempt(ctx, job, canonicalURL, h.local, input, "")
	if localResult.Status == markdown.ResultStatusSuccess {
		h.logSelectedProvider(job, canonicalURL, localResult.Provider)

		return localResult, nil
	}

	fallbackReason := fallbackReason(localResult)
	if h.fallback == nil {
		return markdown.ExtractResult{}, terminalMarkdownError(localResult)
	}

	fallbackResult := h.attempt(ctx, job, canonicalURL, h.fallback, input, fallbackReason)
	if fallbackResult.Status == markdown.ResultStatusSuccess {
		h.logSelectedProvider(job, canonicalURL, fallbackResult.Provider)

		return fallbackResult, nil
	}

	return markdown.ExtractResult{}, terminalMarkdownError(fallbackResult)
}

func (h *MarkdownExtractionHandoff) attempt(
	ctx context.Context,
	job *jobs.Job,
	canonicalURL string,
	extractor markdown.MarkdownExtractor,
	input markdown.ExtractInput,
	fallbackReason string,
) markdown.ExtractResult {
	start := time.Now()
	result := extractor.ExtractMarkdown(ctx, input)
	duration := time.Since(start)

	attrs := []slog.Attr{
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("url", canonicalURL),
		slog.String("provider", string(result.Provider)),
		slog.String("status", string(result.Status)),
		slog.Duration("duration", duration),
	}
	if fallbackReason != "" {
		attrs = append(attrs, slog.String("fallback_reason", fallbackReason))
	}

	if result.ErrorCode != "" {
		attrs = append(attrs, slog.String("arc_code", string(result.ErrorCode)))
	}

	h.logger.LogAttrs(ctx, slog.LevelInfo, "pipeline: markdown provider attempt", attrs...)

	return result
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

func fallbackReason(result markdown.ExtractResult) string {
	if result.Status == markdown.ResultStatusLocalUnreadable {
		return "local unreadable"
	}

	if result.FailureReason != "" {
		return result.FailureReason
	}

	return string(result.Status)
}

func terminalMarkdownError(result markdown.ExtractResult) error {
	switch result.ErrorCode {
	case markdown.ErrorCodeJinaInsufficientCredit:
		return ErrJinaInsufficientBalance
	case markdown.ErrorCodeJinaFailed:
		return ErrJinaReaderFailure
	case markdown.ErrorCodeLocalExtractionFailed:
		return ErrLocalMarkdownExtraction
	}

	if result.Status == markdown.ResultStatusLocalUnreadable {
		return ErrLocalUnreadable
	}

	return ErrUnknown
}
