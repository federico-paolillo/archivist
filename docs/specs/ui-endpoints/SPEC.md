---
id: UIEND
slug: ui-endpoints
title: UI Article Endpoints
status: done
owner: null
depends_on: [authn, telegram-ingestion, summary-generation]
impacts: [gateway, sqlite, filesystem, ui]
canonical: true
---

# Feature: UI Article Endpoints

## Intent

Expose the authenticated Gateway article APIs consumed by the web UI.

## Motivation

The UI needs a small API surface for review and administration: list archived articles, load one article with its persisted Markdown artifacts, delete retained article state, and recover abandoned article-processing state when a running job is stale.

## Scope

In scope:

- `GET /articles`
- `GET /articles/{id}`
- `DELETE /articles/{id}`
- `DELETE /articles/{id}/force`
- Cookie-authenticated access through the `app-cookie` scheme.
- Article operations scoped to the authenticated session user id.
- Cursor pagination using article ULID cursors.
- Read-only detail access to `content.md` and `summary.md`.
- Article detail force-delete eligibility metadata through `canForceDelete`.
- Normal hard deletion of article state, associated jobs/notifications, and the deterministic artifact directory when no associated job is `running`.
- Stale force deletion of user-owned articles whose associated running jobs are stale.
- Same-origin unsafe-method checks using the final auth effective-origin contract.
- `/login`, `/logout`, and `/auth/session` are owned by `authn`.
- UI implementation is owned by the `ui` feature.
- Search, filtering, sorting controls, tags, key points, structured summary fields, soft-delete tombstones, restoration, retry, requeue, automatic recovery, heartbeat state, lock owners, and active running-job deletion are excluded from this feature.

## Users / Actors

- Authenticated Archivist user.
- Gateway API.
- Preact/Vite UI.
- SQLite database.
- Filesystem under `DATA_DIR`.

## Requirements

- REQ-001: All UI article endpoints must require the `app-cookie` authenticated session.
- REQ-002: Article list, detail, normal delete, and force delete must be scoped to the authenticated user's `ClaimTypes.NameIdentifier`.
- REQ-003: Article endpoint implementations must not use the personal user ULID constant for runtime ownership.
- REQ-004: `GET /articles` must return article metadata only.
- REQ-005: `GET /articles` must use a fixed page size of 25.
- REQ-006: `GET /articles` must accept optional `after` or `before` query parameters, but not both.
- REQ-007: Pagination cursors must be article ULIDs.
- REQ-008: List ordering must be article ULID descending, newest first.
- REQ-009: `after={id}` must return older articles than the cursor.
- REQ-010: `before={id}` must return newer articles than the cursor.
- REQ-011: Invalid cursors must return `400 Bad Request`.
- REQ-012: `GET /articles/{id}` must normalize valid ULID route values before querying so detail, normal delete, and force delete route semantics agree for canonical and non-canonical casing.
- REQ-013: `GET /articles/{id}` must return `400 Bad Request` for malformed article IDs.
- REQ-014: `GET /articles/{id}` must return `404 Not Found` when the article does not exist for the authenticated user.
- REQ-015: Article detail responses must include metadata plus `summaryMarkdown`, `contentMarkdown`, and `canForceDelete`.
- REQ-016: Queued and failed articles may return `null` for missing or unreadable artifacts.
- REQ-017: Ready articles require both `summary.md` and `content.md`; missing or unreadable required artifacts must return `500 Internal Server Error`.
- REQ-018: `canForceDelete` must be server-computed and true only when the authenticated user owns the article, at least one associated running job exists, and every associated running job is stale.
- REQ-019: A running job is stale when `started_at <= now - 2 hours`.
- REQ-020: A running job with `started_at IS NULL` is stale for force-delete recovery.
- REQ-021: `DELETE /articles/{id}` must normalize valid ULID route values before querying or deleting.
- REQ-022: `DELETE /articles/{id}` must return `400 Bad Request` for malformed article IDs.
- REQ-023: `DELETE /articles/{id}` must return `404 Not Found` when the article does not exist for the authenticated user.
- REQ-024: `DELETE /articles/{id}` must be allowed for `ready`, `failed`, and `queued` articles.
- REQ-025: `DELETE /articles/{id}` must return `409 Conflict` if any associated job is `running`.
- REQ-026: Successful normal delete must remove the article row, associated jobs, associated notifications, and `{DATA_DIR}/articles/{article_id}`.
- REQ-027: `DELETE /articles/{id}/force` must normalize valid ULID route values before querying or deleting.
- REQ-028: `DELETE /articles/{id}/force` must return `400 Bad Request` for malformed article IDs.
- REQ-029: `DELETE /articles/{id}/force` must return `404 Not Found` when the article does not exist for the authenticated user.
- REQ-030: `DELETE /articles/{id}/force` must return `409 Conflict` when the article is not force-delete eligible, including when any associated running job is active rather than stale or when no associated running job exists.
- REQ-031: Successful force delete must remove the article row, associated jobs, associated notifications, and `{DATA_DIR}/articles/{article_id}`.
- REQ-032: Missing artifact directories must not fail normal delete or force delete.
- REQ-033: Artifact cleanup failures must return `500 Internal Server Error` and leave database state intact.
- REQ-034: Normal delete and force delete must enforce same-origin unsafe-method protection.
- REQ-035: JSON response bodies must use lower-camel property names.
- REQ-036: Normal delete, force delete, and worker job claim must serialize through SQLite write transactions. Delete operations must recheck ownership and job status inside the delete transaction, and worker claim must not claim jobs whose article row has been deleted.
- REQ-037: Normal delete and force delete share the documented SQLite/filesystem atomicity limitation: if artifact cleanup succeeds and the subsequent SQLite commit fails, rollback cannot restore the deleted artifact directory.
- REQ-038: Article endpoint logs and spans must attach `user_id` when the authenticated session user id is available.

## Acceptance Criteria

```gherkin
Feature: UI article endpoints

Scenario: Unauthenticated article list request
  Given no valid auth cookie is present
  When the browser requests GET /articles
  Then the response status is 401

Scenario: Article list first page
  Given the authenticated user has more than 25 articles
  When the browser requests GET /articles
  Then the response contains the newest 25 articles ordered by ULID descending
  And the response contains pagination cursors

Scenario: Article detail returns artifacts and force-delete metadata
  Given the authenticated user owns a ready article
  And summary.md and content.md exist
  When the browser requests GET /articles/{id}
  Then the response contains article metadata
  And summaryMarkdown contains summary.md
  And contentMarkdown contains content.md
  And canForceDelete is false

Scenario: Article detail reports stale force-delete eligibility
  Given the authenticated user owns an article with only stale running jobs
  When the browser requests GET /articles/{id}
  Then the response status is 200
  And canForceDelete is true

Scenario: Queued article detail allows null artifacts
  Given the authenticated user owns a queued article
  And no article artifacts exist
  When the browser requests GET /articles/{id}
  Then the response status is 200
  And summaryMarkdown is null
  And contentMarkdown is null

Scenario: Article ownership is enforced
  Given user "U1" owns an article
  And user "U2" is authenticated
  When user "U2" lists, reads, normally deletes, or force-deletes articles
  Then user "U1"'s article is not returned or mutated

Scenario: Delete queued article
  Given the authenticated user owns a queued article with queued jobs
  When the browser requests DELETE /articles/{id}
  Then the response status is 204
  And the article, jobs, notifications, and artifact directory are removed
  And a subsequent worker claim cannot claim the deleted job

Scenario: Normal delete rejects running article
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

Scenario: Delete rejects cross-site request
  Given a valid auth cookie is present
  And the request Origin is not same-origin
  When the browser requests DELETE /articles/{id}
  Then the response status is 403

Scenario: Force delete rejects cross-site request
  Given a valid auth cookie is present
  And the request Origin is not same-origin
  When the browser requests DELETE /articles/{id}/force
  Then the response status is 403
```

## Data and State

This feature uses the existing `users`, `articles`, `jobs`, and `notifications` schema from `telegram-ingestion`.

Article list rows expose:

- `id`
- `title`
- `originalUrl`
- `canonicalUrl`
- `status`
- `errorMessage`
- `createdAt`

Article detail adds:

- `summaryMarkdown`: content of `{DATA_DIR}/articles/{article_id}/summary.md`, nullable for non-ready articles.
- `contentMarkdown`: content of `{DATA_DIR}/articles/{article_id}/content.md`, nullable for non-ready articles.
- `canForceDelete`: `true` when the authenticated user owns the article, at least one associated running job exists, and all associated running jobs are stale; otherwise `false`.

Normal delete and force delete are hard deletes. Deleted articles disappear from list and detail responses and do not leave tombstone state.

## Interfaces

- `GET /articles`
  - Query: optional `after` or `before`.
  - Success: `200 OK`.
  - Cursor errors: `400 Bad Request`.
- `GET /articles/{id}`
  - Success: `200 OK`.
  - Malformed id: `400 Bad Request`.
  - Not found: `404 Not Found`.
  - Required artifact unavailable for ready article: `500 Internal Server Error`.
- `DELETE /articles/{id}`
  - Success: `204 No Content`.
  - Malformed id: `400 Bad Request`.
  - Not found: `404 Not Found`.
  - Running job exists: `409 Conflict`.
  - Cross-site unsafe request: `403 Forbidden`.
  - Artifact cleanup failure: `500 Internal Server Error`.
- `DELETE /articles/{id}/force`
  - Success: `204 No Content`.
  - Malformed id: `400 Bad Request`.
  - Not found: `404 Not Found`.
  - Not force-delete eligible: `409 Conflict`.
  - Cross-site unsafe request: `403 Forbidden`.
  - Artifact cleanup failure: `500 Internal Server Error`.

JSON response contracts:

```json
{
  "items": [
    {
      "id": "01H00000000000000000000000",
      "title": "Article title",
      "originalUrl": "https://example.com/input",
      "canonicalUrl": "https://example.com/final",
      "status": "ready",
      "errorMessage": null,
      "createdAt": "2026-05-06T12:00:00Z"
    }
  ],
  "nextCursor": "01GZZZZZZZZZZZZZZZZZZZZZZ",
  "previousCursor": null
}
```

```json
{
  "id": "01H00000000000000000000000",
  "title": "Article title",
  "originalUrl": "https://example.com/input",
  "canonicalUrl": "https://example.com/final",
  "status": "ready",
  "errorMessage": null,
  "createdAt": "2026-05-06T12:00:00Z",
  "summaryMarkdown": "Summary text",
  "contentMarkdown": "Article Markdown",
  "canForceDelete": false
}
```

Error responses use the minimal shape below when a response body is needed:

```json
{
  "error": "Short public error message."
}
```

Delete routes return `204 No Content` without a response body on success.

Delete/claim serialization:

- Gateway delete must start a SQLite write transaction, re-read the article and associated job statuses inside that transaction, reject if any job is `running`, delete associated notifications/jobs/article rows, and commit only after artifact cleanup succeeds or is known to be unnecessary.
- Gateway force delete must start a SQLite write transaction, re-read the article and associated job statuses inside that transaction, reject if any running job is active rather than stale, delete associated notifications/jobs/article rows, and commit only after artifact cleanup succeeds or is known to be unnecessary.
- Worker job claim must use a SQLite write transaction and claim only queued jobs whose article row still exists.
- If delete commits first, a subsequent worker claim observes no claimable job.
- If worker claim commits first and sets a job to `running`, normal delete returns `409 Conflict`; force delete returns `409 Conflict` until the running job is stale.
- If artifact cleanup fails, the SQLite transaction rolls back so the user can retry.
- If artifact cleanup succeeds but the subsequent SQLite commit fails, rollback cannot restore the deleted artifact directory; operational repair follows the canonical SQLite/filesystem consistency limitation.

## Dependencies

Depends on:

- `authn` final auth contract for `app-cookie` authentication, authenticated session user id, effective-origin interpretation, and same-origin unsafe-method checks.
- `telegram-ingestion` for SQLite article/job/notification schema.
- `summary-generation` for `summary.md` production and Gateway artifact conventions.
- `docs/ARTIFACTS.md`

Implementation agents should use `.agents/skills/archivist-gateway/SKILL.md` for Gateway coding guidance. The skill is not a feature dependency or rebuild source of truth.

Impacts:

- Gateway article API routes.
- Gateway SQLite read/delete access.
- Gateway artifact read and delete cleanup boundaries.
- UI API client contract.

## Rebuild Notes

- Backend article routes are unprefixed.
- Public browser UI calls reach these routes through the configured UI API base path, default `/api`, with the reverse proxy stripping `/api` before forwarding to Gateway.
- Do not add `/api` to Gateway route definitions to resolve browser page/API path collisions.
- The list endpoint does not read `/data`.
- Detail reads `summary.md` and `content.md`; it does not expose `snapshot.html`.
- Detail response metadata includes `canForceDelete`; the UI must not compute force-delete eligibility from raw job timestamps.
- Delete is a hard admin action and must remove the deterministic artifact directory when article state is removed.
- Force delete is a recovery action for stale running jobs, not a bypass for active work.

## Security / Privacy Notes

- UI article endpoints must not be reachable without the authenticated `app-cookie` session.
- Delete routes are state-changing and must reject cross-site unsafe requests.
- Article queries and mutations must be scoped by authenticated user ID.
- Artifact paths must be derived from validated and normalized ULID article IDs, not raw path segments.

## Observability / Logging Notes

- Gateway logs should include article id, authenticated user id, operation, result status, and artifact cleanup failure details when deletion fails.
- Logs must not include auth cookie values or private article content.

## Open Questions

- None.

## Related Documents

- `./PLAN.md`
- `./tasks/UIEND-002-gateway-article-read-api.md`
- `./tasks/UIEND-003-gateway-article-delete-api.md`
