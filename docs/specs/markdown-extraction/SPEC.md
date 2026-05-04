---
id: MDEXT
slug: markdown-extraction
title: Markdown Extraction With Fallbacks
status: draft
owner: null
depends_on: [article-processing]
impacts: [worker, gateway, filesystem, sqlite]
canonical: true
---

# Feature: Markdown Extraction With Fallbacks

## Intent

Extract readable Markdown from successfully snapshotted article HTML, persist it as the canonical Markdown artifact, and make Markdown extraction the current terminal success point for article-processing jobs until the later summary feature supersedes it.

## Motivation

HTML snapshotting preserves source material but does not produce the readable Markdown content needed by the UI or future summarization. The Worker needs a low-cost extraction stage that starts locally with go-readability v2, pays for Jina Reader only when local readability rejects the document, and fails clearly when neither provider can produce Markdown.

## Scope

In scope:

- Worker Markdown extraction after successful canonical URL resolution and HTML snapshotting.
- Local-first extraction using `codeberg.org/readeck/go-readability/v2`.
- Required local readability gate using `CheckDocument()`.
- HTML-to-Markdown conversion for local extraction using `github.com/JohannesKaufmann/html-to-markdown/v2`.
- Jina Reader fallback only when local extraction cannot produce Markdown.
- Preference for an official Jina Reader Go SDK if one exists at implementation time.
- Small internal Jina Reader adapter only when no suitable official Reader SDK exists.
- Atomic Markdown artifact writes to `{DATA_DIR}/articles/{article_id}/content.md`.
- Worker terminal success only after `content.md` is promoted.
- Worker terminal failure when both local extraction and Jina fallback fail.
- ARC-coded public errors for local extraction, Jina fallback, Jina insufficient balance, and Markdown writes.
- Structured Worker logs for critical extraction decisions and provider fallback.
- Gateway success notification behavior for Markdown-complete jobs.

## Out of Scope

Not included:

- LLM summarization.
- Summary artifact creation.
- Extraction candidate scoring or quality thresholds.
- ReaderLM-v2 use by default.
- Browser rendering, Playwright, or JavaScript execution.
- Automatic retries.
- New article, job, notification, or extraction telemetry schema columns.
- SQLite artifact path columns.

## Users / Actors

- Worker.
- Gateway notification dispatcher.
- Authorized Telegram user.
- SQLite database.
- Filesystem under `DATA_DIR`.
- Jina Reader.

## Requirements

- REQ-001: Markdown extraction must run after successful HTML snapshotting and before terminal success.
- REQ-002: The Worker must read the saved HTML snapshot from `{DATA_DIR}/articles/{article_id}/snapshot.html`.
- REQ-003: The Worker must attempt local extraction first with `codeberg.org/readeck/go-readability/v2`.
- REQ-004: The Worker must call `CheckDocument()` before accepting local readability output.
- REQ-005: If `CheckDocument()` returns false, the Worker must log the fallback decision and call Jina Reader when Jina is enabled.
- REQ-006: If local readability extraction or local Markdown conversion fails, the Worker must call Jina Reader when Jina is enabled.
- REQ-007: Local extraction must convert readable HTML to Markdown with `github.com/JohannesKaufmann/html-to-markdown/v2`.
- REQ-008: Jina fallback must use the article canonical URL by default.
- REQ-009: Jina integration must prefer an official Jina Reader Go SDK if one exists at implementation time.
- REQ-010: If no suitable official Jina Reader Go SDK exists, the Worker may implement a small internal HTTP adapter for the Reader API.
- REQ-011: The Worker must not use untagged or low-adoption third-party Jina Reader wrappers as production dependencies.
- REQ-012: The Worker must persist Markdown to `{DATA_DIR}/articles/{article_id}/content.md`.
- REQ-013: Markdown writes must be atomic: write a temporary file, then rename into place.
- REQ-014: Markdown success must set `articles.status = ready`, clear `articles.error_message`, set `jobs.status = succeeded`, set terminal timestamps/TTL, and insert exactly one pending notification in one SQLite transaction.
- REQ-015: Markdown failure must set `articles.status = failed`, set `articles.error_message` to an ARC-coded public error, set `jobs.status = failed`, persist job error context, set terminal timestamps/TTL, and insert exactly one pending notification in one SQLite transaction.
- REQ-016: Persisted public article errors must use codes defined in `docs/conventions/ERRORS.md`.
- REQ-017: The Worker must log provider attempts, fallback reason, selected provider, ARC code on failure, `article_id`, `job_id`, canonical URL, duration, and artifact write result when available.
- REQ-018: The Gateway must send a deterministic Markdown-complete success reply for succeeded jobs with `content.md`.
- REQ-019: Snapshot-only success notification is superseded by Markdown-complete success.
- REQ-020: This feature must not call an LLM summarizer.

## Acceptance Criteria

```gherkin
Feature: Markdown extraction with fallbacks

Scenario: go-readability extracts Markdown successfully
  Given a queued article-processing job has a stored snapshot.html
  And go-readability v2 CheckDocument returns true
  When the Worker extracts Markdown
  Then the Worker stores content.md under DATA_DIR/articles/{article_id}/
  And articles.status is "ready"
  And articles.error_message is null
  And jobs.status is "succeeded"
  And one pending notification exists for the job

Scenario: go-readability rejects the document and Jina succeeds
  Given a queued article-processing job has a stored snapshot.html
  And go-readability v2 CheckDocument returns false
  And Jina Reader returns Markdown
  When the Worker extracts Markdown
  Then the Worker logs that it switched from go-readability to Jina
  And the Worker stores content.md under DATA_DIR/articles/{article_id}/
  And jobs.status is "succeeded"

Scenario: go-readability fails and Jina succeeds
  Given a queued article-processing job has a stored snapshot.html
  And go-readability extraction or Markdown conversion fails
  And Jina Reader returns Markdown
  When the Worker extracts Markdown
  Then the Worker logs the local failure and fallback provider
  And the Worker stores content.md under DATA_DIR/articles/{article_id}/
  And jobs.status is "succeeded"

Scenario: both local extraction and Jina fail
  Given a queued article-processing job has a stored snapshot.html
  And local extraction cannot produce Markdown
  And Jina Reader fails
  When the Worker records terminal failure
  Then articles.status is "failed"
  And articles.error_message starts with "[ARC-010]"
  And jobs.status is "failed"
  And one pending notification exists for the job

Scenario: Jina reports insufficient balance
  Given local extraction cannot produce Markdown
  And Jina Reader reports insufficient balance
  When the Worker records terminal failure
  Then articles.status is "failed"
  And articles.error_message starts with "[ARC-011]"
  And jobs.status is "failed"

Scenario: Markdown artifact write fails
  Given an extractor returns Markdown
  And the artifact store cannot write content.md
  When the Worker records terminal failure
  Then articles.status is "failed"
  And articles.error_message starts with "[ARC-012]"
  And jobs.status is "failed"
  And no partial content.md is promoted

Scenario: Gateway sends Markdown-complete notification
  Given a pending notification exists for a succeeded job
  And the article has content.md
  When the Gateway dispatches the notification
  Then it sends deterministic Markdown-complete success text
  And marks the notification sent if Telegram accepts the reply
```

## Data and State

This feature uses the existing `users`, `articles`, `jobs`, and `notifications` schema.

Successful Markdown extraction updates:

- `articles.status`: `ready`.
- `articles.error_message`: `null`.
- `jobs.status`: `succeeded`.
- `jobs.completed_at`: completion timestamp.
- `jobs.expires_at`: 14 days after completion.
- `notifications`: one pending row for the job.

Failed Markdown extraction updates:

- `articles.status`: `failed`.
- `articles.error_message`: ARC-coded public message.
- `jobs.status`: `failed`.
- `jobs.error_message`: job error context suitable for Gateway failure replies and operator diagnosis.
- `jobs.completed_at`: completion timestamp.
- `jobs.expires_at`: 14 days after completion.
- `notifications`: one pending row for the job.

No artifact paths, provider telemetry columns, failure-code columns, score columns, or processed timestamp columns are added by this feature.

## Interfaces

- Worker input artifact: `{DATA_DIR}/articles/{article_id}/snapshot.html`.
- Worker output artifact: `{DATA_DIR}/articles/{article_id}/content.md`.
- Local extraction library: `codeberg.org/readeck/go-readability/v2`.
- Local Markdown conversion library: `github.com/JohannesKaufmann/html-to-markdown/v2`.
- Jina fallback: Reader API through an official Reader Go SDK if available, otherwise a small internal adapter.
- Worker terminal state contract: article update, job update, and notification insert in one transaction.
- Gateway notification dispatcher: sends failure replies from job error text and Markdown-complete success replies for succeeded jobs with `content.md`.

## Dependencies

Depends on:

- `article-processing`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/conventions/ARTIFACTS.md`
- `docs/conventions/ERRORS.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/WORKER.md`
- `docs/conventions/GATEWAY.md`

Impacts:

- Worker module.
- Gateway notification dispatcher.
- Filesystem artifact access.
- Telegram terminal notification content.

## Rebuild Notes

- Existing code is not authoritative; rebuilds must follow this spec and linked tasks.
- Markdown completion is the current terminal success point until the future summary feature supersedes it.
- Do not restore candidate scoring for v0 unless a future canonical decision changes the extraction strategy.
- Do not add article artifact path columns.
- Do not call an LLM summarizer from this feature.

## Security / Privacy Notes

- Jina API keys must be supplied through environment variables or equivalent deployment secret mechanisms.
- The Worker must not log secrets, API keys, Telegram tokens, or full provider response bodies that may contain sensitive content.
- Persisted article errors must not expose low-level transport, library, provider, filesystem, or stack details.

## Observability / Logging Notes

- Log local extraction attempt and result.
- Log fallback from go-readability to Jina, including the fallback reason.
- Log selected provider on success.
- Log ARC code and provider failure class on failure.
- Logs should include `article_id`, `job_id`, canonical URL, provider, fallback reason, duration, status, artifact path kind, and artifact write result when available.
- A dedicated observability stack is out of scope for v0.

## Open Questions

- None.

## Related Documents

- `./PLAN.md`
- `./DIARY.md`
- `./tasks/MDEXT-001-create-feature-artifacts-and-contracts.md`
- `./tasks/MDEXT-002-worker-markdown-artifact-access.md`
- `./tasks/MDEXT-003-worker-go-readability-extraction.md`
- `./tasks/MDEXT-004-worker-jina-reader-fallback.md`
- `./tasks/MDEXT-005-worker-markdown-pipeline-integration.md`
- `./tasks/MDEXT-006-gateway-markdown-success-notification.md`
- `./plans/MDEXT-005-worker-markdown-pipeline-integration.execplan.md`
