package pipeline

import (
	"context"

	"codeberg.org/federico-paolillo/archivist/pkg/jobs"
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

// NoOpMarkdownHandoff is a test and isolated-stage helper for callers that
// exercise snapshot behavior without running Markdown extraction.
var NoOpMarkdownHandoff MarkdownHandoff = MarkdownHandoffFunc(func(_ context.Context, _ *jobs.Job, _ string) error {
	return nil
})

// SummaryHandoff is called after content.md promotion.
type SummaryHandoff interface {
	Summarize(ctx context.Context, job *jobs.Job, canonicalURL string) error
}

// SummaryHandoffFunc adapts functions to SummaryHandoff.
type SummaryHandoffFunc func(ctx context.Context, job *jobs.Job, canonicalURL string) error

// Summarize implements SummaryHandoff.
func (f SummaryHandoffFunc) Summarize(ctx context.Context, job *jobs.Job, canonicalURL string) error {
	return f(ctx, job, canonicalURL)
}

// NoOpSummaryHandoff preserves isolated Markdown-stage tests.
var NoOpSummaryHandoff SummaryHandoff = SummaryHandoffFunc(func(_ context.Context, _ *jobs.Job, _ string) error {
	return nil
})
