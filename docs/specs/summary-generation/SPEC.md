---
id: SUMGEN
slug: summary-generation
title: Summary Generation
owner: null
depends_on: [markdown-extraction]
impacts: [worker, gateway, filesystem, sqlite]
canonical: true
---
# Feature: Summary Generation

## Intent

Generate a text-only LLM summary from extracted article Markdown, persist it as the summary artifact, and make summary completion the terminal success point for article-processing jobs.

## Motivation

Markdown extraction produces readable article content, but the Telegram completion reply and review surfaces need a concise summary. The system uses Claude through Anthropic while keeping provider-specific APIs behind an Archivist-owned abstraction so provider replacement does not require Worker orchestration changes.

## Scope

In scope:

- Worker summary generation after `content.md` exists.
- Worker-owned `SummarizerService` abstraction.
- Anthropic/Claude provider implementation behind that abstraction.
- Official Anthropic SDK usage when a suitable SDK exists; in Go, prefer `github.com/anthropics/anthropic-sdk-go`.
- Default summarizer configuration with `LLM_PROVIDER=anthropic`, `LLM_MODEL=claude-3-5-haiku-20241022`, and required `LLM_API_KEY`.
- Fixed summary system prompt that requests text-only output.
- Terminal failure when Markdown exceeds context or request limits; the Worker does not chunk or truncate source Markdown.
- Atomic summary artifact writes to `{DATA_DIR}/articles/{article_id}/summary.md`.
- Summary success as the terminal success criterion for `articles.status`, `jobs.status`, and success notification creation.
- ARC-coded public failures for summarizer provider failure, context/request too large, billing failure, and summary artifact write failure.
- Gateway read-only artifact access for `/data`.
- Gateway Telegram success reply using `summary.md` with the prefix `Archived. Summary is:`.
- Summary JSON, SQLite summary columns, artifact path columns, summary tags, key points, template version fields, chunked/map-reduce/truncated summarization, automatic retries, and multi-provider routing beyond the provider abstraction and first Anthropic implementation are not part of this feature's behavior.
- UI rendering is owned by the `ui` feature.

## Users / Actors

- Worker.
- Gateway notification dispatcher.
- Authorized Telegram user.
- SQLite database.
- Filesystem under `DATA_DIR`.
- Anthropic Claude API.

## Requirements

- REQ-001: Summary generation must run after successful Markdown extraction and before final article/job success.
- REQ-002: The Worker must read `{DATA_DIR}/articles/{article_id}/content.md`.
- REQ-003: The Worker must expose summarization through a Worker-owned `SummarizerService` interface.
- REQ-004: The Anthropic implementation must sit behind `SummarizerService` and must not leak Anthropic SDK request or response types into pipeline orchestration.
- REQ-005: The Anthropic implementation must use an official Anthropic SDK when a suitable SDK exists for the implementation language.
- REQ-006: Go implementations must prefer `github.com/anthropics/anthropic-sdk-go` for Anthropic API access.
- REQ-007: The default provider must be `anthropic`.
- REQ-008: The default model must be `claude-3-5-haiku-20241022`.
- REQ-009: `LLM_API_KEY` must be required when `LLM_PROVIDER=anthropic`.
- REQ-010: The Worker must send the extracted Markdown as source content with a fixed system prompt that requests text-only summary output.
- REQ-011: The Worker must not request structured JSON for summaries.
- REQ-012: The Worker must not chunk or truncate Markdown.
- REQ-013: Context-window overflow, request-too-large responses, or preflight size checks that prove the request cannot fit must fail with `ARC-014`.
- REQ-014: Anthropic HTTP 402 `billing_error` must map to `ARC-015`.
- REQ-015: Anthropic HTTP 413 `request_too_large` must map to `ARC-014`.
- REQ-016: Generic Anthropic API, provider, timeout, transport, permission, authentication, rate-limit, overloaded, or malformed-output failures must map to `ARC-013` unless a more specific ARC code applies.
- REQ-017: The Worker must persist text summary to `{DATA_DIR}/articles/{article_id}/summary.md`.
- REQ-018: Summary writes must be atomic: write a temporary file, then rename into place.
- REQ-019: Summary success must set `articles.status = ready`, clear `articles.error_message`, set `jobs.status = succeeded`, set terminal timestamps/TTL, and insert exactly one pending notification in one SQLite transaction.
- REQ-020: Summary failure must set `articles.status = failed`, set `articles.error_message` to an ARC-coded public error, set `jobs.status = failed`, persist job error context, set terminal timestamps/TTL, and insert exactly one pending notification in one SQLite transaction.
- REQ-021: Snapshot and Markdown stages must not mark article/job success once summary generation is present.
- REQ-022: Gateway must read `/data` artifacts through a read-only abstraction.
- REQ-023: Gateway must not expose write, create, rename, or delete operations through its artifact abstraction.
- REQ-024: Gateway success notification must read `summary.md` and reply to the original Telegram message with `Archived. Summary is: <summary>`.
- REQ-025: Gateway failure notifications must continue to use `jobs.error_message` and preserve leading `[ARC-NNN]` prefixes.
- REQ-026: Telegram success replies must fit Telegram message length limits by deterministic truncation when necessary.

## Acceptance Criteria

```gherkin
Feature: Summary generation

Scenario: Summary generation succeeds
  Given a running article-processing job has stored content.md
  And the configured summarizer returns text
  When the Worker stores summary.md successfully
  Then articles.status is "ready"
  And articles.error_message is null
  And jobs.status is "succeeded"
  And one pending notification exists for the job
  And those database changes commit in one transaction

Scenario: Summary provider fails
  Given a running article-processing job has stored content.md
  And the summarizer provider fails with a generic provider error
  When the Worker records terminal failure
  Then articles.status is "failed"
  And articles.error_message starts with "[ARC-013]"
  And jobs.status is "failed"
  And one pending notification exists for the job

Scenario: Summary source exceeds context or request limits
  Given a running article-processing job has stored content.md
  And the source Markdown cannot fit the configured summarizer request
  When the Worker records terminal failure
  Then articles.status is "failed"
  And articles.error_message starts with "[ARC-014]"
  And jobs.status is "failed"

Scenario: Anthropic billing fails
  Given a running article-processing job has stored content.md
  And Anthropic returns billing_error
  When the Worker records terminal failure
  Then articles.status is "failed"
  And articles.error_message starts with "[ARC-015]"
  And jobs.status is "failed"

Scenario: Summary write fails
  Given the summarizer returns text
  And the artifact store cannot write summary.md
  When the Worker records terminal failure
  Then articles.status is "failed"
  And articles.error_message starts with "[ARC-016]"
  And jobs.status is "failed"
  And no partial summary.md is promoted

Scenario: Gateway sends summary notification
  Given a pending notification exists for a succeeded job
  And the article has summary.md
  When the Gateway dispatches the notification
  Then it replies to the original Telegram message with text starting "Archived. Summary is:"
  And the reply contains the persisted summary
  And the notification is marked sent if Telegram accepts the reply
```

## Data and State

This feature uses the existing `users`, `articles`, `jobs`, and `notifications` schema.

Successful summary generation updates:

- `articles.status`: `ready`.
- `articles.error_message`: `null`.
- `jobs.status`: `succeeded`.
- `jobs.completed_at`: completion timestamp.
- `jobs.expires_at`: 14 days after completion.
- `notifications`: one pending row for the job.

Failed summary generation updates:

- `articles.status`: `failed`.
- `articles.error_message`: ARC-coded public message.
- `jobs.status`: `failed`.
- `jobs.error_message`: job error context suitable for Gateway failure replies and operator diagnosis.
- `jobs.completed_at`: completion timestamp.
- `jobs.expires_at`: 14 days after completion.
- `notifications`: one pending row for the job.

No summary columns, artifact path columns, provider telemetry columns, prompt version columns, or summary JSON artifacts are added by this feature.

## Interfaces

- Worker input artifact: `{DATA_DIR}/articles/{article_id}/content.md`.
- Worker output artifact: `{DATA_DIR}/articles/{article_id}/summary.md`.
- Worker summarizer contract: `SummarizerService`.
- First summarizer provider: Anthropic Claude through an official SDK when suitable.
- Worker terminal state contract: article update, job update, and notification insert in one transaction after `summary.md` is promoted.
- Gateway artifact contract: read-only access to deterministic article artifacts under `DATA_DIR`.
- Gateway notification dispatcher: reads `summary.md` for success replies and uses `jobs.error_message` for failure replies.
- Configuration:
  - `LLM_PROVIDER`
  - `LLM_API_KEY`
  - `LLM_MODEL`

## Dependencies

Depends on:

- `markdown-extraction`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/ARTIFACTS.md`
- `docs/ERRORS.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-worker/SKILL.md`
- `.agents/skills/archivist-gateway/SKILL.md`

Impacts:

- Worker module.
- Gateway notification dispatcher.
- Filesystem artifact access.
- Telegram terminal notification content.

## Rebuild Notes

- Existing code is not authoritative; rebuilds must follow this spec and linked tasks.
- Summary generation is the terminal success point.
- Snapshot and Markdown completion are intermediate stages.
- Do not mark an article `ready`, job `succeeded`, or success notification pending before `summary.md` is promoted.
- Do not create `summary.json` or SQLite summary columns.
- Do not chunk or truncate Markdown for summarization.
- Provider-specific SDK types must stay inside provider adapters.

## Security / Privacy Notes

- `LLM_API_KEY` must be supplied through environment variables or equivalent deployment secret mechanisms.
- The Worker must not log API keys, full Markdown input, full summary output, or full provider responses that may contain article content.
- Persisted article errors must not expose low-level provider, filesystem, transport, or stack details.
- Gateway `/data` access must be read-only.

## Observability / Logging Notes

- Log summarizer provider, model, provider request ID when available, `article_id`, `job_id`, canonical URL, duration, status, ARC code on failure, and artifact write result.
- No dedicated observability stack is required by this feature.

## Open Questions

- None.

## Related Documents

- `./PLAN.md`
- `./tasks/SUMGEN-002-worker-summary-artifact-access.md`
- `./tasks/SUMGEN-003-summarizer-provider-adapter.md`
- `./tasks/SUMGEN-004-worker-summary-pipeline-integration.md`
- `./tasks/SUMGEN-005-gateway-summary-notification.md`
