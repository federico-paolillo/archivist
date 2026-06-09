package pipeline

import (
	"context"
	"log/slog"
	"time"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
	"codeberg.org/federico-paolillo/archivist/internal/fetcher"
	"codeberg.org/federico-paolillo/archivist/internal/observability"
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// fetchHTML calls the fetcher and maps unexpected errors to ErrUnknown.
//
//nolint:funlen,nonamedreturns // Fetch telemetry and ARC mapping are intentionally kept together.
func (p *SnapshotPipeline) fetchHTML(ctx context.Context, job *jobs.Job, articleURL string) (result *fetcher.Result, err error) {
	ctx, span := observability.Tracer().Start(
		ctx,
		"worker.pipeline.fetch",
		trace.WithAttributes(append(
			observability.JobUserAttributes(job.ArticleID, job.ID, job.UserID),
			attribute.String("url", articleURL),
		)...),
	)
	defer func() {
		observability.EndSpan(span, err)
	}()

	start := time.Now()

	p.logger.InfoContext(
		ctx,
		"pipeline: fetch started",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("user_id", job.UserID),
		slog.String("url", articleURL),
		slog.String("stage", "fetch"),
		slog.String("status", "start"),
	)

	result, fetchErr := p.fetch.Fetch(ctx, articleURL)
	if fetchErr == nil {
		span.SetAttributes(attribute.String("final_url", result.FinalURL))
		p.logger.InfoContext(
			ctx,
			"pipeline: fetch completed",
			slog.String("article_id", job.ArticleID),
			slog.String("job_id", job.ID),
			slog.String("user_id", job.UserID),
			slog.String("url", articleURL),
			slog.String("final_url", result.FinalURL),
			slog.String("stage", "fetch"),
			slog.String("status", "success"),
			slog.Duration("duration", time.Since(start)),
		)

		return result, nil
	}

	if _, ok := arc.CodeOf(fetchErr); ok {
		p.logger.InfoContext(
			ctx,
			"pipeline: fetch completed",
			slog.String("article_id", job.ArticleID),
			slog.String("job_id", job.ID),
			slog.String("user_id", job.UserID),
			slog.String("url", articleURL),
			slog.String("stage", "fetch"),
			slog.String("status", "failure"),
			slog.Duration("duration", time.Since(start)),
			slog.String("arc_code", arc.CodeString(fetchErr)),
			slog.Any("error", fetchErr),
		)

		err = pipelineFailure(
			"fetch",
			"fetch html",
			fetchErr,
			nil,
			withJobContext(job.ArticleID, job.ID),
			withPipelineURL(articleURL),
		)

		return nil, err
	}

	p.logger.ErrorContext(
		ctx,
		"pipeline: unexpected fetch error",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("user_id", job.UserID),
		slog.String("stage", "fetch"),
		slog.String("status", "failure"),
		slog.Duration("duration", time.Since(start)),
		slog.String("arc_code", arc.CodeString(arc.ErrUnknown)),
		slog.Any("error", fetchErr),
	)

	err = pipelineFailure(
		"fetch",
		"fetch html",
		arc.ErrUnknown,
		fetchErr,
		withJobContext(job.ArticleID, job.ID),
		withPipelineURL(articleURL),
	)

	return nil, err
}
