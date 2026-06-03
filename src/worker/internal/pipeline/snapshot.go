package pipeline

import (
	"bytes"
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

// MarkdownHandoff is the extension point implemented by Markdown extraction.
// It is called after snapshot.html is atomically written and canonical_url is updated.
//
// Handoff contract:
//   - Input: ctx, job (with ArticleID, ID), canonicalURL (the resolved final URL).
//   - The implementation opens snapshot.html from the artifact store using job.ArticleID.
//   - On success: writes content.md and continues to the summary generation stage,
//     which owns final success after summary.md promotion.
//   - On failure: returns an ARC-coded error (e.g. ARC-008 through ARC-012).
//     The pipeline records terminal failure when this returns non-nil.
type MarkdownHandoff interface {
	Handoff(ctx context.Context, job *jobs.Job, canonicalURL string) error
}

// MarkdownHandoffFunc is an adapter that allows function literals to implement MarkdownHandoff.
type MarkdownHandoffFunc func(ctx context.Context, job *jobs.Job, canonicalURL string) error

// Handoff implements MarkdownHandoff.
func (f MarkdownHandoffFunc) Handoff(ctx context.Context, job *jobs.Job, canonicalURL string) error {
	return f(ctx, job, canonicalURL)
}

// NoOpMarkdownHandoff is the placeholder for the Markdown extraction stage.
// MDEXT-005 replaces this with the real extraction pipeline.
var NoOpMarkdownHandoff MarkdownHandoff = MarkdownHandoffFunc(func(_ context.Context, _ *jobs.Job, _ string) error {
	return nil
})

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
		observability.EndSpan(claimSpan, claimErr)

		if errors.Is(claimErr, sql.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("pipeline: claim queued job: %w", claimErr)
	}

	claimSpan.SetAttributes(observability.JobAttributes(job.ArticleID, job.ID)...)
	claimSpan.End()

	ctx = p.continueJobTrace(ctx, job)

	ctx, processSpan := observability.Tracer().Start(
		ctx,
		"worker.pipeline.process",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(observability.JobAttributes(job.ArticleID, job.ID)...),
	)
	defer processSpan.End()

	start := time.Now()

	p.logger.InfoContext(
		ctx,
		"pipeline: job claimed",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("stage", "claim"),
		slog.String("status", "claimed"),
	)

	articleURL, urlErr := p.repo.ArticleURL(ctx, job.ArticleID)
	if urlErr != nil {
		processSpan.RecordError(urlErr)
		processSpan.SetStatus(codes.Error, urlErr.Error())

		return false, fmt.Errorf("pipeline: load article URL for %s: %w", job.ArticleID, urlErr)
	}

	processSpan.SetAttributes(attribute.String("url", articleURL))

	p.logger.InfoContext(
		ctx,
		"pipeline: processing job",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
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

// persistFailure logs the failure and commits terminal failure state.
//
//nolint:funlen,nonamedreturns // Terminal failure logging and persistence stay adjacent for auditability.
func (p *SnapshotPipeline) persistFailure(
	ctx context.Context,
	job *jobs.Job,
	articleURL string,
	processingErr error,
	duration time.Duration,
) (err error) {
	ctx, span := observability.Tracer().Start(
		ctx,
		"worker.pipeline.terminal_failure",
		trace.WithAttributes(append(
			observability.JobAttributes(job.ArticleID, job.ID),
			attribute.String("url", articleURL),
			attribute.String("arc_code", arc.CodeString(processingErr)),
		)...),
	)
	defer func() {
		observability.EndSpan(span, err)
	}()

	span.SetStatus(codes.Error, arc.CodeString(processingErr))

	p.logger.ErrorContext(
		ctx,
		"pipeline: job failed",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("url", articleURL),
		slog.String("stage", "terminal_failure"),
		slog.String("status", "failed"),
		slog.Duration("duration", duration),
		slog.String("arc_code", arc.CodeString(processingErr)),
		slog.Any("error", processingErr),
	)

	errorMessage, ok := arc.PublicMessage(processingErr)
	if !ok {
		errorMessage = arc.Format(arc.CodeUnknownProcessingFailure)
	}

	p.logger.InfoContext(
		ctx,
		"pipeline: terminal failure persistence started",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("url", articleURL),
		slog.String("stage", "terminal_failure"),
		slog.String("status", "start"),
		slog.String("arc_code", arc.CodeString(processingErr)),
	)

	terminalErr := p.repo.CompleteTerminal(ctx, job, jobs.TerminalOutcome{
		Success:      false,
		ErrorMessage: errorMessage,
	})
	if terminalErr != nil {
		p.logger.ErrorContext(
			ctx,
			"pipeline: terminal failure persistence failed",
			slog.String("article_id", job.ArticleID),
			slog.String("job_id", job.ID),
			slog.String("url", articleURL),
			slog.String("stage", "terminal_failure"),
			slog.String("status", "terminal_persist_failed"),
			slog.String("arc_code", arc.CodeString(processingErr)),
			slog.Any("error", terminalErr),
		)

		return fmt.Errorf("pipeline: persist terminal failure for job %s: %w", job.ID, terminalErr)
	}

	p.logger.InfoContext(
		ctx,
		"pipeline: terminal failure persisted",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("url", articleURL),
		slog.String("stage", "terminal_failure"),
		slog.String("status", "persisted"),
		slog.String("arc_code", arc.CodeString(processingErr)),
	)

	return nil
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

// fetchHTML calls the fetcher and maps unexpected errors to ErrUnknown.
//
//nolint:funlen,nonamedreturns // Fetch telemetry and ARC mapping are intentionally kept together.
func (p *SnapshotPipeline) fetchHTML(ctx context.Context, job *jobs.Job, articleURL string) (result *fetcher.Result, err error) {
	ctx, span := observability.Tracer().Start(
		ctx,
		"worker.pipeline.fetch",
		trace.WithAttributes(append(
			observability.JobAttributes(job.ArticleID, job.ID),
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

// writeSnapshot atomically writes snapshot.html and maps failures to ARC-007.
//
//nolint:nonamedreturns,spancheck // Deferred span completion records the returned error.
func (p *SnapshotPipeline) writeSnapshot(ctx context.Context, job *jobs.Job, result *fetcher.Result) (err error) {
	ctx, span := observability.Tracer().Start(
		ctx,
		"worker.pipeline.snapshot_write",
		trace.WithAttributes(observability.JobAttributes(job.ArticleID, job.ID)...),
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
		slog.String("stage", "snapshot_write"),
		slog.String("status", "success"),
		slog.Duration("duration", time.Since(start)),
		slog.String("artifact_result", "success"),
	)

	return nil
}

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
			observability.JobAttributes(job.ArticleID, job.ID),
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
		slog.String("url", articleURL),
		slog.String("final_url", finalURL),
		slog.String("stage", "canonical_url_update"),
		slog.String("status", "start"),
	)

	canonicalErr := p.repo.UpdateCanonicalURL(ctx, job.ArticleID, finalURL)
	if canonicalErr != nil {
		p.logger.ErrorContext(
			ctx,
			"pipeline: canonical URL update failed",
			slog.String("article_id", job.ArticleID),
			slog.String("job_id", job.ID),
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
		slog.String("url", articleURL),
		slog.String("final_url", finalURL),
		slog.String("stage", "canonical_url_update"),
		slog.String("status", "success"),
		slog.Duration("duration", time.Since(start)),
	)

	return nil
}

// invokeMarkdownHandoff calls the markdown extraction handoff.
// Extension point for MDEXT-005 — see MarkdownHandoff contract above.
//
//nolint:nonamedreturns // Deferred span completion records the returned error.
func (p *SnapshotPipeline) invokeMarkdownHandoff(ctx context.Context, job *jobs.Job, finalURL string) (err error) {
	ctx, span := observability.Tracer().Start(
		ctx,
		"worker.pipeline.markdown_handoff",
		trace.WithAttributes(append(
			observability.JobAttributes(job.ArticleID, job.ID),
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
		slog.String("arc_code", arc.CodeString(arc.ErrUnknown)),
		slog.Any("error", mdErr),
	)

	return fmt.Errorf("pipeline: markdown handoff infrastructure failure: %w", mdErr)
}
