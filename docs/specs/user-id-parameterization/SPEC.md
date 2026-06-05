---
id: MUSER
slug: user-id-parameterization
title: User ID Parameterization
status: done
owner: null
depends_on:
  - otel-observability
impacts:
  - gateway
  - worker
  - sqlite
canonical: true
---

# Feature: User ID Parameterization

## Intent

Prepare Archivist's existing ownership model for basic multi-user operation by removing runtime dependence on the fixed personal user identifier.

Archivist still ships without user registration, account management, roles, tenant administration, or user-facing user selection. The first personal user is still seeded by authentication bootstrap. Runtime flows must resolve `user_id` from persisted identity, authenticated session state, claimed jobs, or article ownership, with one explicit Worker CLI enqueue exception described below.

## Motivation

The existing schema already stores `users`, `articles.user_id`, `jobs.user_id`, and auth sessions with a user id. The problem is that runtime Gateway and Worker paths still assign the hardcoded personal user identifier directly. That makes the schema look user-aware while key behavior remains single-user-by-constant.

## Scope

In scope:

- Authentication bootstrap remains the primary production code path allowed to hardcode the personal user ULID and the personal Telegram sender id `1559957191`.
- Worker CLI enqueue is an explicit exception: it uses `jobs.DefaultUserID = 01ASB2XFCZJY7WHZ2FNRTMQJCT` only to check for an existing `users.id` row before creating CLI-enqueued articles and jobs.
- Password login resolves the authenticated user id by verifying the submitted password against all non-empty Argon2id PHC hashes and requiring exactly one matching user.
- Telegram webhook authorization is based on existence of a `users.telegram_user_id` mapping.
- Telegram ingestion stores resolved `user_id` on created articles and jobs.
- UI/API article operations continue to scope all reads and deletes by authenticated session `user_id`.
- Worker CLI enqueue uses `jobs.DefaultUserID`, requires a matching `users.id`, never infers ownership from user-table cardinality, and never creates the user.
- Worker processing uses claimed job ownership for article/job reads and mutations.
- Gateway and Worker logs and traces attach `user_id` when the Archivist user is known.

## Out of Scope

Not included:

- User registration or user bootstrap beyond the existing personal user bootstrap path.
- UI-visible user management, user switching, roles, tenants, teams, or invitations.
- Multiple Telegram identity linking flows.
- User-partitioned artifact paths.
- Snapshotter `user_id` attribution.
- Metrics or label changes.

## Users / Actors

- Provisioned Archivist user.
- Telegram sender with a persisted `users.telegram_user_id` mapping.
- Gateway API.
- Worker CLI and process loop.
- SQLite database.

## Requirements

- REQ-001: Authentication bootstrap may hardcode the personal user ULID `01ASB2XFCZJY7WHZ2FNRTMQJCT` and personal Telegram sender id `1559957191`.
- REQ-002: Worker CLI enqueue is the only runtime exception to bootstrap-only hardcoding: it must use `jobs.DefaultUserID = 01ASB2XFCZJY7WHZ2FNRTMQJCT` as the default CLI owner identifier.
- REQ-002A: No other runtime Gateway or Worker path may assign article or job ownership from a literal personal user id.
- REQ-003: Password login must load every user row with a non-empty Argon2id PHC `password_hash`, verify the submitted password against every candidate, and issue a session for the matched row's `id` only when exactly one candidate matches.
- REQ-004: Multiple password-bearing rows are valid. Password login must fail closed when no password-bearing row exists, when no candidate hash matches, or when more than one candidate hash matches the submitted password.
- REQ-005: Telegram webhook processing must authorize a sender only when `users.telegram_user_id` maps to a user row.
- REQ-006: Unknown Telegram senders must not create or mutate users, articles, jobs, or notifications, and must receive no Telegram reply.
- REQ-007: Telegram ingestion must persist the resolved Archivist `user_id` on both the article and job created for an accepted URL.
- REQ-008: Telegram ingestion must not create, upsert, or reassign the `users` row.
- REQ-009: Authentication bootstrap must set the personal user's `users.telegram_user_id` to `1559957191` only when the current value is null, and must preserve an existing non-null `telegram_user_id`.
- REQ-010: Authenticated article list, detail, delete, and force-delete operations must use the session `user_id` and must not affect another user's articles or jobs.
- REQ-011: Worker CLI enqueue must check for a `users.id` row equal to `jobs.DefaultUserID` and must fail if that row does not exist.
- REQ-011A: Worker CLI enqueue must not infer ownership from the number of rows in `users` and must not fail merely because additional user rows exist.
- REQ-011B: Worker CLI enqueue must not create, upsert, or repair the default user row.
- REQ-012: Worker processing must use the claimed job's `user_id` when reading the article URL, updating canonical URL/title, and completing terminal state.
- REQ-013: Worker processing must fail safely when a job references an article owned by a different user.
- REQ-014: Gateway and Worker logs and spans must include `user_id` when the Archivist user has been resolved.
- REQ-015: The canonical log and trace attribute key is exactly `user_id`.
- REQ-016: Snapshotter must not gain a `user_id` requirement.
- REQ-017: `user_id` must remain a trace/log attribute only and must not be promoted to metrics or collector labels.
- REQ-018: Gateway bootstrap must not read `settings.PersonalTelegramUserId`, `Telegram:AllowedUserId`, or equivalent deployment configuration for the personal Telegram sender mapping.
- REQ-019: This feature must not introduce registration, user-management endpoints, or user self-service.

## Acceptance Criteria

```gherkin
Feature: User ID parameterization

Scenario: Password login creates a session for the password owner
  Given multiple users may have password_hash values
  And exactly one password_hash verifies the submitted password
  When the user logs in with the correct password
  Then the session entry contains the matching row's user_id

Scenario: Duplicate password matches fail closed
  Given two password-bearing users have password_hash values that verify the submitted password
  When the user logs in with that password
  Then the response is unauthorized
  And no session is issued

Scenario: Auth bootstrap sets the personal Telegram sender id
  Given the personal user row exists with telegram_user_id null
  When auth bootstrap runs
  Then telegram_user_id is set to 1559957191

Scenario: Auth bootstrap preserves an existing Telegram sender id
  Given the personal user row exists with a non-null telegram_user_id
  When auth bootstrap runs
  Then telegram_user_id is unchanged

Scenario: Telegram sender is mapped to a user
  Given users.telegram_user_id maps the sender to user "U1"
  When the sender submits a valid URL
  Then one article and one job are created with user_id "U1"
  And the queued acknowledgement is sent

Scenario: Telegram sender is unknown
  Given no users.telegram_user_id row matches the sender
  When the sender submits a valid URL
  Then no user, article, job, or notification row is created or changed
  And no Telegram reply is sent

Scenario: Article ownership is enforced
  Given user "U1" has an article
  And user "U2" is authenticated
  When user "U2" lists, reads, deletes, or force-deletes articles
  Then user "U1" article state is not returned or changed

Scenario: Worker CLI enqueue uses the configured default user
  Given a user row exists with id "01ASB2XFCZJY7WHZ2FNRTMQJCT"
  When the Worker CLI enqueues a URL
  Then the created article and job use "01ASB2XFCZJY7WHZ2FNRTMQJCT"

Scenario: Worker CLI enqueue does not infer from user count
  Given a user row exists with id "01ASB2XFCZJY7WHZ2FNRTMQJCT"
  And another user row also exists
  When the Worker CLI enqueues a URL
  Then the created article and job use "01ASB2XFCZJY7WHZ2FNRTMQJCT"

Scenario: Worker CLI enqueue fails when the default user is missing
  Given no user row exists with id "01ASB2XFCZJY7WHZ2FNRTMQJCT"
  When the Worker CLI enqueues a URL
  Then no user, article, or job row is created
  And enqueue fails

Scenario: Worker detects mismatched job and article ownership
  Given a job belongs to user "U1"
  And the referenced article belongs to user "U2"
  When the Worker tries to process the job
  Then processing fails safely without mutating user "U2" article as user "U1"
```

## Data and State

No new tables are required.

Existing ownership columns remain authoritative:

- `users.id`
- `users.telegram_user_id`
- `users.password_hash`
- `articles.user_id`
- `jobs.user_id`

The existing `users` row for the personal account remains bootstrapped as `01ASB2XFCZJY7WHZ2FNRTMQJCT`. Auth bootstrap sets that row's `telegram_user_id` to `1559957191` only when it is null and preserves any existing non-null mapping. Runtime code must treat the personal account as data read from SQLite after bootstrap, except Worker CLI enqueue may use `jobs.DefaultUserID` with that value to check that the bootstrapped row exists before creating CLI-owned article and job rows.

## Interfaces

- `POST /login` remains password-only. The response shape is unchanged. Gateway verifies the submitted password against all non-empty Argon2id PHC hashes and succeeds only when exactly one user matches.
- `POST /telegram/webhook` remains unchanged externally. Runtime authorization is now database mapping by Telegram sender id.
- Worker `enqueue` command keeps the same CLI surface and checks SQLite for `users.id = jobs.DefaultUserID`.
- Logs and spans use `user_id` for the resolved Archivist user id.

## Dependencies

Depends on:

- `otel-observability`
- Existing `authn`, `telegram-ingestion`, `ui-endpoints`, `job-recovery`, and Worker processing contracts.

Impacts:

- Gateway auth and Telegram persistence.
- Worker job repository and pipeline.
- Observability attributes.
- Canonical architecture and design docs.

## Rebuild Notes

- Do not infer ownership from the personal user constant outside auth bootstrap, except for Worker CLI enqueue's explicit `jobs.DefaultUserID` existence check.
- Do not use `Telegram:AllowedUserId`, `settings.PersonalTelegramUserId`, or equivalent deployment configuration for auth bootstrap. The personal Telegram sender id is the fixed value `1559957191` unless an existing non-null database value is preserved.
- Multiple password-bearing user rows are valid. Password-only login ambiguity is resolved by the submitted password: exactly one matching hash succeeds; duplicate matches fail closed.
- Worker CLI enqueue must never infer ownership from user-table cardinality and must never create the default user row.
- Runtime ownership must be propagated through explicit values read from SQLite or session state.
- Unknown Telegram users are unauthorized by absence of a database mapping and receive no reply.
- Artifact paths remain based on `article_id` only.

## Security / Privacy Notes

- `user_id` is not secret, but it is identifying metadata. It may be used in logs and traces only as required for diagnosis and must not be promoted to labels or metrics.
- Failed login and unknown Telegram requests cannot attach `user_id` because no user has been resolved.

## Observability / Logging Notes

- Gateway logs/spans attach `user_id` after password, session, or Telegram mapping resolution.
- Worker logs/spans attach `user_id` from the CLI enqueue default-user existence check or claimed jobs.
- Snapshotter remains intentionally user-agnostic.

## Open Questions

- None.

## Related Documents

- `./PLAN.md`
- `./DIARY.md`
