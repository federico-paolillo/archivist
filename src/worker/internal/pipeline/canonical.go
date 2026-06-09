package pipeline

import (
	"context"
	"log/slog"
	"time"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
	"codeberg.org/federico-paolillo/archivist/internal/observability"
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// updateCanonicalURL sets articles.canonical_url to the final redirected URL.
//
//nolint:funlen,nonamedreturns,spancheck // Canonical URL telemetry and persistence logging stay together.
func (p *SnapshotPipeline) updateCanonicalURL(
	ctx context.Context,
	job *jobs.Job,
	articleURL string,
	finalURL string,
) (err error) {
	ctx, span := observability.Tracer().Start(
		ctx,
		"worker.pipeline.canonical_url_update",
		trace.WithAttributes(append(
			observability.JobUserAttributes(job.ArticleID, job.ID, job.UserID),
			attribute.String("url", articleURL),
			attribute.String("final_url", finalURL),
		)...),
	)
	defer func() {
		observability.EndSpan(span, err)
	}()

	start := time.Now()

	p.logger.InfoContext(
		ctx,
		"pipeline: canonical URL update started",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("user_id", job.UserID),
		slog.String("url", articleURL),
		slog.String("final_url", finalURL),
		slog.String("stage", "canonical_url_update"),
		slog.String("status", "start"),
	)

	canonicalErr := p.repo.UpdateCanonicalURL(ctx, job.ArticleID, job.UserID, finalURL)
	if canonicalErr != nil {
		p.logger.ErrorContext(
			ctx,
			"pipeline: canonical URL update failed",
			slog.String("article_id", job.ArticleID),
			slog.String("job_id", job.ID),
			slog.String("user_id", job.UserID),
			slog.String("final_url", finalURL),
			slog.String("stage", "canonical_url_update"),
			slog.String("status", "failure"),
			slog.Duration("duration", time.Since(start)),
			slog.String("arc_code", arc.CodeString(arc.ErrUnknown)),
			slog.Any("error", canonicalErr),
		)

		err = pipelineFailure(
			"snapshot",
			"update canonical URL",
			arc.ErrUnknown,
			canonicalErr,
			withJobContext(job.ArticleID, job.ID),
			withPipelineURL(finalURL),
		)

		return err
	}

	p.logger.InfoContext(
		ctx,
		"pipeline: canonical URL updated",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("user_id", job.UserID),
		slog.String("url", articleURL),
		slog.String("final_url", finalURL),
		slog.String("stage", "canonical_url_update"),
		slog.String("status", "success"),
		slog.Duration("duration", time.Since(start)),
	)

	return nil
}
