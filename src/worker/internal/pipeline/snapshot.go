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
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
)

// MarkdownHandoff is the extension point that MDEXT-005 will implement.
// It is called after snapshot.html is atomically written and canonical_url is updated.
//
// MDEXT-005 handoff contract:
//   - Input: ctx, job (with ArticleID, ID), canonicalURL (the resolved final URL).
//   - The implementation opens snapshot.html from the artifact store using job.ArticleID.
//   - On success: writes content.md and continues to the summary generation stage.
//     The implementation must not set articles.status=ready, jobs.status=succeeded,
//     or insert a success notification — those are owned by SUMGEN-005.
//   - On failure: returns an ARC-coded error (e.g. ARC-008 through ARC-012).
//     The pipeline records terminal failure when this returns non-nil.
//
// Until MDEXT-005 is wired, use NoOpMarkdownHandoff which returns nil.
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
// It returns nil when no queued job was available, and nil when a job was processed
// (even if the job itself failed — job failures are persisted, not surfaced to the caller).
// It returns a non-nil error only for unexpected infrastructure failures.
func (p *SnapshotPipeline) ProcessOne(ctx context.Context) error {
	job, claimErr := p.repo.ClaimQueued(ctx)
	if claimErr != nil {
		if errors.Is(claimErr, sql.ErrNoRows) {
			return nil
		}

		return fmt.Errorf("pipeline: claim queued job: %w", claimErr)
	}

	start := time.Now()

	articleURL, urlErr := p.repo.ArticleURL(ctx, job.ArticleID)
	if urlErr != nil {
		return fmt.Errorf("pipeline: load article URL for %s: %w", job.ArticleID, urlErr)
	}

	p.logger.Info(
		"pipeline: processing job",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("url", articleURL),
	)

	processingErr := p.runStages(ctx, job, articleURL)

	duration := time.Since(start)

	if processingErr != nil {
		return p.persistFailure(ctx, job, articleURL, processingErr, duration)
	}

	p.logger.Info(
		"pipeline: snapshot stored, handed off to markdown extraction",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.Duration("duration", duration),
		slog.String("status", "snapshot_done"),
	)

	return nil
}

// persistFailure logs the failure and commits terminal failure state.
func (p *SnapshotPipeline) persistFailure(
	ctx context.Context,
	job *jobs.Job,
	articleURL string,
	processingErr error,
	duration time.Duration,
) error {
	p.logger.Error(
		"pipeline: job failed",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("url", articleURL),
		slog.String("status", "failed"),
		slog.Duration("duration", duration),
		slog.String("arc_code", arc.CodeString(processingErr)),
		slog.Any("err", processingErr),
	)

	errorMessage, ok := arc.PublicMessage(processingErr)
	if !ok {
		errorMessage = arc.Format(arc.CodeUnknownProcessingFailure)
	}

	terminalErr := p.repo.CompleteTerminal(ctx, job, jobs.TerminalOutcome{
		Success:      false,
		ErrorMessage: errorMessage,
	})
	if terminalErr != nil {
		return fmt.Errorf("pipeline: persist terminal failure for job %s: %w", job.ID, terminalErr)
	}

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
func (p *SnapshotPipeline) fetchHTML(ctx context.Context, job *jobs.Job, articleURL string) (*fetcher.Result, error) {
	result, fetchErr := p.fetch.Fetch(ctx, articleURL)
	if fetchErr == nil {
		return result, nil
	}

	if _, ok := arc.CodeOf(fetchErr); ok {
		return nil, pipelineFailure(
			"fetch",
			"fetch html",
			fetchErr,
			nil,
			withJobContext(job.ArticleID, job.ID),
			withPipelineURL(articleURL),
		)
	}

	p.logger.Error(
		"pipeline: unexpected fetch error",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.Any("err", fetchErr),
	)

	return nil, pipelineFailure(
		"fetch",
		"fetch html",
		arc.ErrUnknown,
		fetchErr,
		withJobContext(job.ArticleID, job.ID),
		withPipelineURL(articleURL),
	)
}

// writeSnapshot atomically writes snapshot.html and maps failures to ARC-007.
func (p *SnapshotPipeline) writeSnapshot(_ context.Context, job *jobs.Job, result *fetcher.Result) error {
	snapshotErr := p.store.WriteSnapshot(job.ArticleID, bytes.NewReader(result.Body))
	if snapshotErr != nil {
		p.logger.Error(
			"pipeline: snapshot write failed",
			slog.String("article_id", job.ArticleID),
			slog.String("job_id", job.ID),
			slog.String("artifact_result", "failure"),
			slog.Any("err", snapshotErr),
		)

		return pipelineFailure(
			"snapshot",
			"write snapshot",
			arc.ErrSnapshotWrite,
			snapshotErr,
			withJobContext(job.ArticleID, job.ID),
		)
	}

	p.logger.Info(
		"pipeline: snapshot written",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("artifact_result", "success"),
	)

	return nil
}

// updateCanonicalURL sets articles.canonical_url to the final redirected URL.
func (p *SnapshotPipeline) updateCanonicalURL(
	ctx context.Context,
	job *jobs.Job,
	articleURL string,
	finalURL string,
) error {
	canonicalErr := p.repo.UpdateCanonicalURL(ctx, job.ArticleID, finalURL)
	if canonicalErr != nil {
		p.logger.Error(
			"pipeline: canonical URL update failed",
			slog.String("article_id", job.ArticleID),
			slog.String("job_id", job.ID),
			slog.String("final_url", finalURL),
			slog.Any("err", canonicalErr),
		)

		return pipelineFailure(
			"snapshot",
			"update canonical URL",
			arc.ErrUnknown,
			canonicalErr,
			withJobContext(job.ArticleID, job.ID),
			withPipelineURL(finalURL),
		)
	}

	p.logger.Info(
		"pipeline: canonical URL updated",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("url", articleURL),
		slog.String("final_url", finalURL),
	)

	return nil
}

// invokeMarkdownHandoff calls the markdown extraction handoff.
// Extension point for MDEXT-005 — see MarkdownHandoff contract above.
func (p *SnapshotPipeline) invokeMarkdownHandoff(ctx context.Context, job *jobs.Job, finalURL string) error {
	mdErr := p.markdownHandoff.Handoff(ctx, job, finalURL)
	if mdErr == nil {
		// Snapshot boundary — NOT a terminal success in final v0.
		// articles.status is NOT set to ready here.
		// jobs.status is NOT set to succeeded here.
		// No success notification is inserted here.
		// Final v0 terminal success is owned by SUMGEN-005 after summary.md is written.
		return nil
	}

	if _, ok := arc.CodeOf(mdErr); ok {
		return pipelineFailure(
			"markdown",
			"handoff",
			mdErr,
			nil,
			withJobContext(job.ArticleID, job.ID),
			withPipelineURL(finalURL),
		)
	}

	p.logger.Error(
		"pipeline: unexpected markdown handoff error",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.Any("err", mdErr),
	)

	return pipelineFailure(
		"markdown",
		"handoff",
		arc.ErrUnknown,
		mdErr,
		withJobContext(job.ArticleID, job.ID),
		withPipelineURL(finalURL),
	)
}
