---
id: ARTPROC
slug: article-processing
title: URL-To-Article Processing Pipeline
status: draft
owner: null
depends_on: [telegram-ingestion]
impacts: [worker, gateway, sqlite, filesystem]
canonical: true
---

# Feature: URL-To-Article Processing Pipeline

## Intent

Process queued article jobs by resolving the submitted URL, fetching the final HTML response, storing a deterministic raw snapshot, and committing terminal article/job/notification state through SQLite.

## Motivation

Telegram ingestion creates article records and queued jobs but does not process article content. The Worker needs the first reliable processing slice: dequeue a job, resolve redirects, snapshot HTML, persist success or failure state, and produce notification intent for the Gateway.

This feature intentionally stops before extraction, readability comparison, scoring, Markdown generation, and summarization. Those are v0 work but require their own feature specs. Once Markdown extraction and summary generation are present, snapshot completion is an intermediate stage, not final article/job success.

## Scope

In scope:

- Worker dequeue of queued article-processing jobs using the existing SQLite queue contract.
- URL validation for `http` and `https` schemes before fetch.
- Redirect resolution and HTML fetching using `github.com/imroc/req/v3`.
- Conservative fetch limits: 10 redirects, 20 second total timeout, 10 MiB maximum response body.
- HTML-only acceptance for `text/html` and `application/xhtml+xml`.
- Atomic `snapshot.html` writes under `{DATA_DIR}/articles/{article_id}/`.
- Article canonical URL update to the final redirected URL after successful resolution.
- Transactional snapshot-success state update, job completion, and notification creation.
- Transactional failure state update, ARC-coded article error, job failure context, and notification creation.
- Empty pipeline slots for later extraction and rating stages, documented as future replacement points.
- Gateway notification behavior for snapshot-complete success only if implemented before downstream pipeline stages; otherwise snapshot-only notification is superseded by later Markdown and summary completion.

## Out of Scope

Not included:

- Jina.ai extraction.
- go-readability extraction.
- Extraction candidate scoring.
- Markdown generation.
- LLM summarization.
- Summary artifact creation.
- Browser rendering, Playwright, or JavaScript-heavy page handling beyond storing returned HTML.
- Automatic retries.
- New article, job, or notification states.
- Placeholder `content.md`, `summary.json`, or `summary.md` artifacts.

## Users / Actors

- Worker.
- Gateway notification dispatcher.
- Authorized Telegram user.
- SQLite database.
- Article websites.
- Filesystem under `DATA_DIR`.

## Requirements

- REQ-001: The Worker must claim queued article-processing jobs through the existing SQLite queue contract.
- REQ-002: The Worker must process only absolute `http` and `https` URLs.
- REQ-003: The Worker must use `github.com/imroc/req/v3` for article HTTP requests.
- REQ-004: The HTTP fetcher must allow at most 10 redirects.
- REQ-005: The HTTP fetcher must use a 20 second total timeout.
- REQ-006: The HTTP fetcher must reject response bodies larger than 10 MiB.
- REQ-007: The HTTP fetcher must accept only `text/html` and `application/xhtml+xml` responses.
- REQ-008: Successful URL resolution must update `articles.canonical_url` to the final redirected URL.
- REQ-009: Successful snapshot processing must write only `snapshot.html`.
- REQ-010: Snapshot writes must be atomic: write a temporary file, then rename into place.
- REQ-011: Snapshot success is an interim terminal success only when no downstream extraction or summary feature is present. In final v0, snapshot success must continue to Markdown extraction and must not set `articles.status = ready`, set `jobs.status = succeeded`, or insert a success notification.
- REQ-012: Processing failure must set `articles.status = failed`, set `articles.error_message` to an ARC-coded public error, set `jobs.status = failed`, persist job error context, set terminal timestamps/TTL, and insert exactly one pending notification in one SQLite transaction.
- REQ-013: Persisted public article errors must use codes defined in `docs/conventions/ERRORS.md`.
- REQ-014: The Worker must not call Telegram APIs directly.
- REQ-015: The Gateway may send a snapshot-complete success notification only until `markdown-extraction` supersedes snapshot-only completion.
- REQ-016: Extraction and rating pipeline steps must exist only as no-op slots or documentation boundaries in this feature.
- REQ-017: The `markdown-extraction` and `summary-generation` features supersede the snapshot-complete success criterion with summary-complete processing.

## Acceptance Criteria

```gherkin
Feature: URL-to-article processing pipeline

Scenario: Worker snapshots an HTML article
  Given a queued article-processing job exists
  And the article original_url redirects to a 200 HTML response
  When the Worker processes the job
  Then the Worker stores snapshot.html under DATA_DIR/articles/{article_id}/
  And articles.canonical_url is set to the final redirected URL
  And articles.status is "ready"
  And articles.error_message is null
  And jobs.status is "succeeded"
  And one pending notification exists for the job

Scenario: URL returns not found
  Given a queued article-processing job exists
  And the article original_url resolves to a 404 response
  When the Worker processes the job
  Then articles.status is "failed"
  And articles.error_message starts with "[ARC-003]"
  And jobs.status is "failed"
  And one pending notification exists for the job

Scenario: URL requires unavailable access
  Given a queued article-processing job exists
  And the article original_url resolves to a 401 or 403 response
  When the Worker processes the job
  Then articles.status is "failed"
  And articles.error_message starts with "[ARC-002]"
  And jobs.status is "failed"
  And one pending notification exists for the job

Scenario: URL returns non-HTML content
  Given a queued article-processing job exists
  And the article original_url returns a successful PDF response
  When the Worker processes the job
  Then snapshot.html is not written
  And articles.status is "failed"
  And articles.error_message starts with "[ARC-005]"
  And jobs.status is "failed"

Scenario: Snapshot write fails
  Given a queued article-processing job exists
  And the article original_url returns valid HTML
  And the artifact store cannot write snapshot.html
  When the Worker processes the job
  Then articles.status is "failed"
  And articles.error_message starts with "[ARC-007]"
  And jobs.status is "failed"
  And no partial snapshot is promoted as snapshot.html

Scenario: Gateway sends interim snapshot success notification before Markdown extraction
  Given a pending notification exists for a succeeded snapshot-only job
  And no summary artifact exists for the article
  When the Gateway dispatches the notification
  Then it sends a deterministic snapshot-complete success reply unless markdown-extraction has superseded this behavior
  And marks the notification sent if Telegram accepts the reply
```

## Data and State

This feature uses the existing `users`, `articles`, `jobs`, and `notifications` schema from `telegram-ingestion`.

Successful snapshot processing updates:

- `articles.canonical_url`: final URL after redirects.
- `articles.status`: `ready`.
- `articles.error_message`: `null`.
- `jobs.status`: `succeeded`.
- `jobs.completed_at`: completion timestamp.
- `jobs.expires_at`: 14 days after completion.
- `notifications`: one pending row for the job.

Failed processing updates:

- `articles.status`: `failed`.
- `articles.error_message`: ARC-coded public message.
- `jobs.status`: `failed`.
- `jobs.error_message`: job error context suitable for Gateway failure replies and operator diagnosis.
- `jobs.completed_at`: completion timestamp.
- `jobs.expires_at`: 14 days after completion.
- `notifications`: one pending row for the job.

No artifact paths, HTTP status columns, failure-code columns, extraction telemetry columns, or processed timestamp columns are added by this feature.

## Interfaces

- Worker job source: SQLite queued article-processing jobs.
- Worker HTTP client: `github.com/imroc/req/v3`.
- Worker artifact store: `{DATA_DIR}/articles/{article_id}/snapshot.html`.
- Worker terminal state contract: article update, job update, and notification insert in one transaction.
- Gateway notification dispatcher: sends failure replies from job error text and interim success replies for snapshot-complete jobs without summary artifacts.

## Dependencies

Depends on:

- `telegram-ingestion`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/WORKER.md`
- `docs/conventions/GATEWAY.md`
- `docs/conventions/ERRORS.md`

Impacts:

- Worker module.
- Gateway notification dispatcher.
- SQLite repository contracts.
- Filesystem artifact access.
- Telegram terminal notification content.

## Rebuild Notes

- Existing code is not authoritative; rebuilds must follow this spec and linked tasks.
- Snapshot success is an interim completion point only while extraction/summarization are not implemented. Final v0 success is summary-complete.
- The `markdown-extraction` and `summary-generation` features replace snapshot-complete success with summary-complete processing.
- `snapshot.html` is the only artifact written by this feature.
- Do not add article artifact path columns.
- Do not add retry states or automatic retry scheduling.

## Security / Privacy Notes

- The Worker must not log secrets or Telegram tokens.
- Persisted article errors must not expose low-level transport, filesystem, library, or stack details.
- The Worker must not fetch non-HTTP(S) schemes.

## Observability / Logging Notes

- Worker logs should include `article_id`, `job_id`, original URL, final URL when known, status, duration, and ARC code on failure.
- Low-level HTTP and filesystem diagnostics belong in logs or job diagnostic context, not in public article error messages.
- No dedicated observability stack is required.

## Open Questions

- None.

## Related Documents

- `./PLAN.md`
- `./DIARY.md`
- `./tasks/ARTPROC-001-create-feature-spec-and-plan-artifacts.md`
- `./tasks/ARTPROC-002-define-shared-arc-error-code-convention.md`
- `./tasks/ARTPROC-003-worker-filesystem-artifact-access-layer.md`
- `./tasks/ARTPROC-004-worker-url-resolver-and-html-fetcher.md`
- `./tasks/ARTPROC-005-worker-snapshot-pipeline-orchestration.md`
- `./tasks/ARTPROC-006-gateway-snapshot-success-notification-bridge.md`
- `../markdown-extraction/SPEC.md`
- `./plans/ARTPROC-005-worker-snapshot-pipeline-orchestration.execplan.md`
