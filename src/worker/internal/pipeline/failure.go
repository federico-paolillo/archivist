package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
	"codeberg.org/federico-paolillo/archivist/internal/observability"
	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

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
			observability.JobUserAttributes(job.ArticleID, job.ID, job.UserID),
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
		slog.String("user_id", job.UserID),
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
		slog.String("user_id", job.UserID),
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
			slog.String("user_id", job.UserID),
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
		slog.String("user_id", job.UserID),
		slog.String("url", articleURL),
		slog.String("stage", "terminal_failure"),
		slog.String("status", "persisted"),
		slog.String("arc_code", arc.CodeString(processingErr)),
	)

	return nil
}

func (p *SnapshotPipeline) persistOwnershipMismatch(
	ctx context.Context,
	job *jobs.Job,
	ownershipErr error,
	duration time.Duration,
) (bool, error) {
	errorMessage := arc.Format(arc.CodeUnknownProcessingFailure)

	p.logger.ErrorContext(
		ctx,
		"pipeline: corrupt job ownership mismatch",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("user_id", job.UserID),
		slog.String("stage", "claim"),
		slog.String("status", "corrupt"),
		slog.Duration("duration", duration),
		slog.String("arc_code", string(arc.CodeUnknownProcessingFailure)),
		slog.Any("error", ownershipErr),
	)

	err := p.repo.CompleteJobFailure(ctx, job, errorMessage)
	if err != nil {
		p.logger.ErrorContext(
			ctx,
			"pipeline: corrupt job failure persistence failed",
			slog.String("article_id", job.ArticleID),
			slog.String("job_id", job.ID),
			slog.String("user_id", job.UserID),
			slog.String("stage", "terminal_failure"),
			slog.String("status", "terminal_persist_failed"),
			slog.String("arc_code", string(arc.CodeUnknownProcessingFailure)),
			slog.Any("error", err),
		)

		return false, fmt.Errorf("pipeline: persist ownership mismatch for job %s: %w", job.ID, err)
	}

	p.logger.InfoContext(
		ctx,
		"pipeline: corrupt job failure persisted",
		slog.String("article_id", job.ArticleID),
		slog.String("job_id", job.ID),
		slog.String("user_id", job.UserID),
		slog.String("stage", "terminal_failure"),
		slog.String("status", "persisted"),
		slog.String("arc_code", string(arc.CodeUnknownProcessingFailure)),
	)

	return true, nil
}
