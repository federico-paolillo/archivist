package pipeline

import (
	"bytes"
	"context"
	"log/slog"
	"time"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
	"codeberg.org/federico-paolillo/archivist/internal/fetcher"
	"codeberg.org/federico-paolillo/archivist/internal/observability"
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
	"go.opentelemetry.io/otel/trace"
)

// writeSnapshot atomically writes snapshot.html and maps failures to ARC-007.
//
//nolint:nonamedreturns,spancheck // Deferred span completion records the returned error.
func (p *SnapshotPipeline) writeSnapshot(ctx context.Context, job *jobs.Job, result *fetcher.Result) (err error) {
	ctx, span := observability.Tracer().Start(
		ctx,
		"worker.pipeline.snapshot_write",
		trace.WithAttributes(observability.JobUserAttributes(job.ArticleID, job.ID, job.UserID)...),
	)
	defer func() {
		observability.EndSpan(span, err)
	}()

	start := time.Now()

	p.logger.InfoContext(
		ctx,
		"pipeline: snapshot write started",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("user_id", job.UserID),
		slog.String("stage", "snapshot_write"),
		slog.String("status", "start"),
	)

	snapshotErr := p.store.WriteSnapshot(job.ArticleID, bytes.NewReader(result.Body))
	if snapshotErr != nil {
		p.logger.ErrorContext(
			ctx,
			"pipeline: snapshot write failed",
			slog.String("article_id", job.ArticleID),
			slog.String("job_id", job.ID),
			slog.String("user_id", job.UserID),
			slog.String("stage", "snapshot_write"),
			slog.String("status", "failure"),
			slog.Duration("duration", time.Since(start)),
			slog.String("artifact_result", "failure"),
			slog.String("arc_code", arc.CodeString(arc.ErrSnapshotWrite)),
			slog.Any("error", snapshotErr),
		)

		err = pipelineFailure(
			"snapshot",
			"write snapshot",
			arc.ErrSnapshotWrite,
			snapshotErr,
			withJobContext(job.ArticleID, job.ID),
		)

		return err
	}

	p.logger.InfoContext(
		ctx,
		"pipeline: snapshot written",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("user_id", job.UserID),
		slog.String("stage", "snapshot_write"),
		slog.String("status", "success"),
		slog.Duration("duration", time.Since(start)),
		slog.String("artifact_result", "success"),
	)

	return nil
}
