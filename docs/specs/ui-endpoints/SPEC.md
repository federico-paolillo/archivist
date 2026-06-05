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

The UI needs a small API surface for review and administration: list archived articles, load one article with its persisted Markdown artifacts, and delete articles that no longer need to be retained.

## Scope

In scope:

- `GET /articles`
- `GET /articles/{id}`
- `DELETE /articles/{id}`
- Cookie-authenticated access through the `app-cookie` scheme.
- Cursor pagination using article ULID cursors.
- Read-only detail access to `content.md` and `summary.md`.
- Hard deletion of article state, associated jobs/notifications, and the deterministic artifact directory.

## Out of Scope

Not included:

- `/login`, `/logout`, or `/auth/session`, which are owned by `authn`.
- UI implementation.
- Search, filtering, sorting controls, tags, key points, or structured summary fields.
- Soft-delete tombstones or restoration.
- Deleting running jobs.

## Users / Actors

- Authenticated Archivist user.
- Gateway API.
- Preact/Vite UI.
- SQLite database.
- Filesystem under `DATA_DIR`.

## Requirements

- REQ-001: All UI article endpoints must require the `app-cookie` authenticated session.
- REQ-002: Article queries must be scoped to the authenticated user's `ClaimTypes.NameIdentifier`.
- REQ-003: `GET /articles` must return article metadata only.
- REQ-004: `GET /articles` must use a fixed page size of 25.
- REQ-005: `GET /articles` must accept optional `after` or `before` query parameters, but not both.
- REQ-006: Pagination cursors must be article ULIDs.
- REQ-007: List ordering must be article ULID descending, newest first.
- REQ-008: `after={id}` must return older articles than the cursor.
- REQ-009: `before={id}` must return newer articles than the cursor.
- REQ-010: Invalid cursors must return `400 Bad Request`.
- REQ-011: `GET /articles/{id}` must return `400 Bad Request` for malformed article IDs.
- REQ-012: `GET /articles/{id}` must return `404 Not Found` when the article does not exist for the authenticated user.
- REQ-013: Article detail responses must include metadata plus `summaryMarkdown` and `contentMarkdown`.
- REQ-014: Queued and failed articles may return `null` for missing or unreadable artifacts.
- REQ-015: Ready articles require both `summary.md` and `content.md`; missing or unreadable required artifacts must return `500 Internal Server Error`.
- REQ-016: `DELETE /articles/{id}` must return `400 Bad Request` for malformed article IDs.
- REQ-017: `DELETE /articles/{id}` must return `404 Not Found` when the article does not exist for the authenticated user.
- REQ-018: `DELETE /articles/{id}` must be allowed for `ready`, `failed`, and `queued` articles.
- REQ-019: `DELETE /articles/{id}` must return `409 Conflict` if any associated job is `running`.
- REQ-020: Successful delete must remove the article row, associated jobs, associated notifications, and `{DATA_DIR}/articles/{article_id}`.
- REQ-021: Missing artifact directories must not fail deletion.
- REQ-022: Artifact cleanup failures must return `500 Internal Server Error`.
- REQ-023: `DELETE /articles/{id}` must enforce same-origin unsafe-method protection.
- REQ-024: JSON response bodies must use lower-camel property names.
- REQ-025: Delete and worker job claim must serialize through SQLite write transactions. Delete must recheck associated job status inside the delete transaction, and worker claim must not claim jobs whose article row has been deleted.
- REQ-026: Article endpoint logs and spans must attach `user_id` when the authenticated session user id is available.

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

Scenario: Article list with after cursor
  Given the authenticated user has older articles after a cursor
  When the browser requests GET /articles?after={article_id}
  Then the response contains articles older than the cursor

Scenario: Article list rejects ambiguous cursors
  Given both after and before are supplied
  When the browser requests GET /articles?after={a}&before={b}
  Then the response status is 400

Scenario: Article detail returns artifacts
  Given the authenticated user owns a ready article
  And summary.md and content.md exist
  When the browser requests GET /articles/{id}
  Then the response contains article metadata
  And summaryMarkdown contains summary.md
  And contentMarkdown contains content.md

Scenario: Ready article artifact is missing
  Given the authenticated user owns a ready article
  And summary.md is missing
  When the browser requests GET /articles/{id}
  Then the response status is 500

Scenario: Queued article detail allows null artifacts
  Given the authenticated user owns a queued article
  And no article artifacts exist
  When the browser requests GET /articles/{id}
  Then the response status is 200
  And summaryMarkdown is null
  And contentMarkdown is null

Scenario: Delete queued article
  Given the authenticated user owns a queued article with queued jobs
  When the browser requests DELETE /articles/{id}
  Then the response status is 204
  And the article, jobs, notifications, and artifact directory are removed
  And a later worker claim cannot claim the deleted job

Scenario: Delete rejects running article
  Given the authenticated user owns an article with a running job
  When the browser requests DELETE /articles/{id}
  Then the response status is 409

Scenario: Delete rejects cross-site request
  Given a valid auth cookie is present
  And the request Origin is not same-origin
  When the browser requests DELETE /articles/{id}
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

Delete behavior is hard delete. Deleted articles disappear from list and detail responses and do not leave tombstone state.

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
  "contentMarkdown": "Article Markdown"
}
```

Error responses use the minimal shape below when a response body is needed:

```json
{
  "error": "Short public error message."
}
```

`DELETE /articles/{id}` returns `204 No Content` without a response body on success.

Delete/claim serialization:

- Gateway delete must start a SQLite write transaction, re-read the article and associated job statuses inside that transaction, reject if any job is `running`, delete associated notifications/jobs/article rows, and commit only after artifact cleanup succeeds or is known to be unnecessary.
- Worker job claim must use a SQLite write transaction and claim only queued jobs whose article row still exists.
- If delete commits first, later worker claim observes no claimable job.
- If worker claim commits first and sets a job to `running`, delete returns `409 Conflict`.

## Dependencies

Depends on:

- `authn` for `app-cookie` authentication and same-origin unsafe-method checks.
- `telegram-ingestion` for SQLite article/job/notification schema.
- `summary-generation` for final v0 `summary.md` production and Gateway artifact conventions.
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
- Do not add structured summary fields until a future summary contract reintroduces them.
- Delete is a hard admin action and must remove the deterministic artifact directory when article state is removed.

## Security / Privacy Notes

- UI article endpoints must not be reachable without the authenticated `app-cookie` session.
- Delete is state-changing and must reject cross-site unsafe requests.
- Article queries must be scoped by authenticated user ID.
- Artifact paths must be derived from validated ULID article IDs, not raw path segments.

## Observability / Logging Notes

- Gateway logs should include article id, authenticated user id, operation, result status, and artifact cleanup failure details when deletion fails.
- Logs must not include auth cookie values.

## Open Questions

- None.

## Related Documents

- `./PLAN.md`
- `./DIARY.md`
- `./tasks/UIEND-001-create-canonical-artifacts.md`
- `./tasks/UIEND-002-gateway-article-read-api.md`
- `./tasks/UIEND-003-gateway-article-delete-api.md`
- `./plans/UIEND-002-gateway-article-read-api.execplan.md`
- `./plans/UIEND-003-gateway-article-delete-api.execplan.md`
