---
id: JREC
slug: job-recovery
title: Job Recovery And Worker Logging
status: done
owner: null
depends_on: [ui, ui-endpoints, summary-generation]
impacts: [gateway, worker, ui, sqlite, filesystem]
canonical: true
---

# Feature: Job Recovery And Worker Logging

## Intent

Allow the authenticated user to clean up abandoned article-processing jobs that are stuck in `running`, while improving Worker logs so claimed jobs and pipeline stages remain diagnosable when infrastructure failures prevent terminal state persistence.

## Motivation

The Worker claims queued jobs by atomically setting `jobs.status = running` before processing. If infrastructure fails after claim but before terminal persistence, the job can remain `running` forever. Normal article deletion intentionally rejects running jobs, so the article, job, notifications, and artifacts become undeletable zombie state.

The fix should preserve v0 simplicity. Archivist should not add automatic retries, lock owners, heartbeats, or new queue states for this recovery slice. It should expose an explicit authenticated force-delete path only after the job is stale, and logs should make the failure point visible.

## Scope

In scope:

- Authenticated Gateway force-delete API for user-owned articles with stale running jobs.
- A fixed stale threshold of 2 hours.
- Server-computed article detail metadata indicating whether force delete is currently available.
- UI force-delete action and separate destructive confirmation when the backend says force delete is available.
- Worker structured logs for process-loop iterations, claim, stage boundaries, terminal outcomes, and terminal-persistence failures.
- Tests for Gateway, UI, and Worker behavior.

## Out of Scope

Not included:

- Automatic retries.
- Requeue, reprocess, or rollback-to-queued behavior.
- New job states, job attempts, lock owners, heartbeats, or schema migrations.
- Force deletion of active running jobs.
- Force deletion without authentication or same-origin unsafe-method protection.
- Persisted processing telemetry columns.

## Users / Actors

- Personal Archivist user.
- Gateway API.
- Preact/Vite UI.
- Worker.
- SQLite database.
- Filesystem artifact store under `DATA_DIR`.

## Requirements

- REQ-001: Normal `DELETE /articles/{id}` must continue to reject any associated `running` job with `409 Conflict`.
- REQ-002: Gateway must expose `DELETE /articles/{id}/force` as an authenticated route.
- REQ-003: Force delete must enforce the same same-origin unsafe-method protection as normal delete.
- REQ-004: Force delete must require article ownership by the authenticated user.
- REQ-005: Force delete must return `400 Bad Request` for malformed article IDs.
- REQ-006: Force delete must return `404 Not Found` when the article does not exist for the authenticated user.
- REQ-007: Force delete must return `409 Conflict` when any associated running job is active rather than stale.
- REQ-008: A running job is stale when `started_at <= now - 2 hours`.
- REQ-009: A running job with `started_at IS NULL` is stale for force-delete recovery.
- REQ-010: Successful force delete must remove the article row, associated jobs, associated notifications, and `{DATA_DIR}/articles/{article_id}`.
- REQ-011: Missing artifact directories must not fail force deletion.
- REQ-012: Artifact cleanup failures must return `500 Internal Server Error` and leave database state intact.
- REQ-013: Force delete must use a SQLite write transaction and recheck ownership and running-job staleness inside the transaction.
- REQ-014: Article detail responses must include server-computed force-delete availability metadata.
- REQ-015: The UI must show a distinct `Force Delete` action only when article detail metadata says it is available.
- REQ-016: The UI must require a separate destructive confirmation before calling the force-delete API.
- REQ-017: Successful force delete in the UI must remove the article from the list, clear the detail pane, and navigate to `/articles`.
- REQ-018: Force-delete failure in the UI must leave the current article selected and display the Gateway error text.
- REQ-019: The Worker must log each process-loop iteration start.
- REQ-020: The Worker must log idle/no-job poll results.
- REQ-021: The Worker must log immediately after a queued job is claimed, before loading the article URL.
- REQ-022: The Worker must log pipeline stage start and result for fetch, snapshot write, canonical URL update, Markdown extraction, summary generation, terminal success, terminal failure, and terminal-persistence failure.
- REQ-023: Worker logs must use stable structured fields including `job_id`, `article_id`, `stage`, `status`, `duration`, `arc_code`, `provider`, `model`, `request_id`, and `artifact_result` where applicable.
- REQ-024: Worker logs must not include secrets, auth material, full article HTML, full Markdown, full summary text, provider payloads, or API keys.

## Acceptance Criteria

```gherkin
Feature: Job recovery and worker logging

Scenario: Normal delete still rejects running jobs
  Given the authenticated user owns an article with a running job
  When the browser requests DELETE /articles/{id}
  Then the response status is 409
  And the article, job, notifications, and artifact directory remain

Scenario: Force delete removes stale running job state
  Given the authenticated user owns an article with a running job started more than 2 hours ago
  When the browser requests DELETE /articles/{id}/force
  Then the response status is 204
  And the article, jobs, notifications, and artifact directory are removed

Scenario: Force delete rejects active running jobs
  Given the authenticated user owns an article with a running job started less than 2 hours ago
  When the browser requests DELETE /articles/{id}/force
  Then the response status is 409
  And the article, job, notifications, and artifact directory remain

Scenario: Force delete is shown only when stale recovery is available
  Given article detail metadata says canForceDelete is true
  When the UI renders the selected article
  Then a Force Delete action is visible
  And confirming it calls DELETE /articles/{id}/force

Scenario: Active running article does not show force delete
  Given article detail metadata says canForceDelete is false
  When the UI renders the selected article
  Then no Force Delete action is visible

Scenario: Worker logs claimed job before article URL load
  Given a queued article-processing job exists
  When the Worker claims it
  Then the Worker logs job_id, article_id, stage "claim", and status "claimed" before loading the article URL

Scenario: Worker terminal persistence fails
  Given a running job reaches terminal success or failure
  And terminal persistence fails
  When the Worker exits the processing attempt
  Then the Worker logs status "terminal_persist_failed"
  And the log includes job_id, article_id, stage, and the diagnostic error
```

## Data and State

This feature uses the existing `articles`, `jobs`, and `notifications` schema. It does not add columns or states.

Running-job staleness is derived from `jobs.status` and `jobs.started_at`. The fixed threshold is 2 hours.

Article detail adds:

- `canForceDelete`: `true` when the authenticated user owns the article and all associated running jobs are stale according to this feature's policy; otherwise `false`.

## Interfaces

- `DELETE /articles/{id}/force`
  - Success: `204 No Content`.
  - Malformed id: `400 Bad Request`.
  - Not found: `404 Not Found`.
  - Active running job exists: `409 Conflict`.
  - Cross-site unsafe request: `403 Forbidden`.
  - Artifact cleanup failure: `500 Internal Server Error`.
- Article detail JSON adds `canForceDelete`.
- UI API client adds `forceDeleteArticle(id)`.
- Worker process logs remain stdout `slog` entries.

## Dependencies

Depends on:

- `ui`
- `ui-endpoints`
- `summary-generation`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/ARTIFACTS.md`

Impacts:

- Gateway article APIs and persistence services.
- UI article detail actions and API client.
- Worker process loop and pipeline logging.
- SQLite state interpretation for stale running jobs.
- Filesystem artifact deletion through existing artifact cleanup contract.

## Rebuild Notes

- Existing code is not authoritative; rebuilds must follow this spec and linked tasks.
- Force delete is a recovery action for stale running jobs, not a general bypass for active jobs.
- Do not introduce automatic retries or requeue behavior in this feature.
- Do not change normal delete semantics.
- Server code owns stale eligibility; UI must not compute eligibility from raw timestamps.

## Security / Privacy Notes

- Force delete is destructive and must require authentication, ownership, and same-origin unsafe-method protection.
- Worker logging must not disclose secrets, cookies, provider keys, full article content, full Markdown, full summaries, or provider payloads.

## Observability / Logging Notes

- Worker logging must make the claim and every pipeline boundary visible.
- Logs should be structured for grepability and operational diagnosis without a dedicated observability stack.
- Gateway force-delete logs, if added, must not include auth cookie values or private article content.

## Open Questions

- None.

## Related Documents

- `./PLAN.md`
- `./DIARY.md`
- `./tasks/JREC-001-create-canonical-job-recovery-artifacts.md`
- `./tasks/JREC-002-gateway-force-delete-api.md`
- `./tasks/JREC-003-ui-force-delete-workflow.md`
- `./tasks/JREC-004-worker-logging-improvements.md`
- `./tasks/JREC-005-integration-validation.md`
