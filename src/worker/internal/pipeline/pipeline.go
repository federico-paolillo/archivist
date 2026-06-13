package pipeline

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
	"codeberg.org/federico-paolillo/archivist/internal/artifacts"
	"codeberg.org/federico-paolillo/archivist/internal/fetcher"
	"codeberg.org/federico-paolillo/archivist/internal/observability"
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// SnapshotPipeline orchestrates article-processing jobs from dequeue through snapshot
// and hands off to the Markdown extraction stage.
type SnapshotPipeline struct {
	logger          *slog.Logger
	repo            jobs.Repository
	store           *artifacts.Store
	fetch           *fetcher.Fetcher
	markdownHandoff MarkdownHandoff
}

// NewSnapshotPipeline constructs a SnapshotPipeline.
func NewSnapshotPipeline(
	logger *slog.Logger,
	repo jobs.Repository,
	store *artifacts.Store,
	fetch *fetcher.Fetcher,
	markdownHandoff MarkdownHandoff,
) *SnapshotPipeline {
	return &SnapshotPipeline{
		logger:          logger,
		repo:            repo,
		store:           store,
		fetch:           fetch,
		markdownHandoff: markdownHandoff,
	}
}

// ProcessOne claims one queued article-processing job and runs it through the pipeline.
// It returns processed=false when no queued job was available. It returns
// processed=true when a job was processed, even if the job itself failed and
// that failure was persisted. It returns a non-nil error only for unexpected
// infrastructure failures.
//
//nolint:funlen // Top-level orchestration is intentionally linear for stage ordering.
func (p *SnapshotPipeline) ProcessOne(ctx context.Context) (bool, error) {
	claimCtx, claimSpan := observability.Tracer().Start(ctx, "worker.pipeline.claim", trace.WithSpanKind(trace.SpanKindConsumer))

	job, claimErr := p.repo.ClaimQueued(claimCtx)
	if claimErr != nil {
		if errors.Is(claimErr, sql.ErrNoRows) {
			claimSpan.End()

			return false, nil
		}

		observability.EndSpan(claimSpan, claimErr)

		return false, fmt.Errorf("pipeline: claim queued job: %w", claimErr)
	}

	claimSpan.SetAttributes(observability.JobUserAttributes(job.ArticleID, job.ID, job.UserID)...)
	claimSpan.End()

	ctx = p.continueJobTrace(ctx, job)

	ctx, processSpan := observability.Tracer().Start(
		ctx,
		"worker.pipeline.process",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(observability.JobUserAttributes(job.ArticleID, job.ID, job.UserID)...),
	)
	defer processSpan.End()

	start := time.Now()

	p.logger.InfoContext(
		ctx,
		"pipeline: job claimed",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("user_id", job.UserID),
		slog.String("stage", "claim"),
		slog.String("status", "claimed"),
	)

	articleURL, urlErr := p.repo.ArticleURL(ctx, job.ArticleID, job.UserID)
	if urlErr != nil {
		processSpan.RecordError(urlErr)
		processSpan.SetStatus(codes.Error, urlErr.Error())

		if errors.Is(urlErr, jobs.ErrOwnershipMismatch) {
			return p.persistOwnershipMismatch(ctx, job, urlErr, time.Since(start))
		}

		return false, fmt.Errorf("pipeline: load article URL for %s: %w", job.ArticleID, urlErr)
	}

	processSpan.SetAttributes(attribute.String("url", articleURL))

	p.logger.InfoContext(
		ctx,
		"pipeline: processing job",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("user_id", job.UserID),
		slog.String("url", articleURL),
		slog.String("stage", "process"),
		slog.String("status", "start"),
	)

	processingErr := p.runStages(ctx, job, articleURL)

	duration := time.Since(start)

	if processingErr != nil {
		processSpan.RecordError(processingErr)
		processSpan.SetStatus(codes.Error, processingErr.Error())

		if _, ok := arc.CodeOf(processingErr); !ok {
			return false, fmt.Errorf("pipeline: run stages for job %s: %w", job.ID, processingErr)
		}

		persistErr := p.persistFailure(ctx, job, articleURL, processingErr, duration)
		if persistErr != nil {
			return false, persistErr
		}

		return true, nil
	}

	p.logger.InfoContext(
		ctx,
		"pipeline: job stages completed",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("user_id", job.UserID),
		slog.Duration("duration", duration),
		slog.String("stage", "process"),
		slog.String("status", "pipeline_done"),
	)

	return true, nil
}

func (p *SnapshotPipeline) continueJobTrace(ctx context.Context, job *jobs.Job) context.Context {
	var traceparent string
	if job.TraceParent != nil {
		traceparent = *job.TraceParent
	}

	var tracestate string
	if job.TraceState != nil {
		tracestate = *job.TraceState
	}

	return observability.ExtractTraceContext(ctx, traceparent, tracestate)
}

// runStages executes the ordered pipeline stages. Returns an ARC-coded public error
// on any stage failure. Unexpected low-level errors are mapped to ErrUnknown.
func (p *SnapshotPipeline) runStages(ctx context.Context, job *jobs.Job, articleURL string) error {
	result, fetchErr := p.fetchHTML(ctx, job, articleURL)
	if fetchErr != nil {
		return fetchErr
	}

	snapshotErr := p.writeSnapshot(ctx, job, result)
	if snapshotErr != nil {
		return snapshotErr
	}

	canonicalErr := p.updateCanonicalURL(ctx, job, articleURL, result.FinalURL)
	if canonicalErr != nil {
		return canonicalErr
	}

	return p.invokeMarkdownHandoff(ctx, job, result.FinalURL)
}

// invokeMarkdownHandoff calls the markdown extraction handoff.
//
//nolint:nonamedreturns // Deferred span completion records the returned error.
func (p *SnapshotPipeline) invokeMarkdownHandoff(ctx context.Context, job *jobs.Job, finalURL string) (err error) {
	ctx, span := observability.Tracer().Start(
		ctx,
		"worker.pipeline.markdown_handoff",
		trace.WithAttributes(append(
			observability.JobUserAttributes(job.ArticleID, job.ID, job.UserID),
			attribute.String("url", finalURL),
		)...),
	)
	defer func() {
		observability.EndSpan(span, err)
	}()

	mdErr := p.markdownHandoff.Handoff(ctx, job, finalURL)
	if mdErr == nil {
		return nil
	}

	if _, ok := arc.CodeOf(mdErr); ok {
		err = pipelineFailure(
			"markdown",
			"handoff",
			mdErr,
			nil,
			withJobContext(job.ArticleID, job.ID),
			withPipelineURL(finalURL),
		)

		return err
	}

	p.logger.ErrorContext(
		ctx,
		"pipeline: unexpected markdown handoff error",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("user_id", job.UserID),
		slog.String("arc_code", arc.CodeString(arc.ErrUnknown)),
		slog.Any("error", mdErr),
	)

	return fmt.Errorf("pipeline: markdown handoff infrastructure failure: %w", mdErr)
}
