---
id: TELING
slug: telegram-ingestion
title: Telegram Ingestion
status: done
owner: null
depends_on: []
impacts: [gateway, worker, sqlite]
canonical: true
---

# Feature: Telegram Ingestion

## Intent

Accept article URLs from Telegram senders mapped to Archivist users, enqueue them for background processing, record sender identity and resolved ownership, and report processing outcomes back to the original Telegram message.

## Motivation

Telegram is the v0 ingestion channel for Archivist. The user should be able to send a single article URL, receive an immediate acknowledgement once the work is queued, and later receive either the generated summary or the final processing error without checking the web UI.

The data model should stay small. v0 does not implement automatic retries or durable processing telemetry. Failure state must be clear enough that the user can manually re-send the same URL.

## Scope

In scope:

- Telegram webhook ingestion through the gateway.
- Webhook secret validation.
- Sender authorization by `users.telegram_user_id` existence.
- URL-only message validation.
- Atomic article and job creation in SQLite for the resolved user.
- Idempotency for Telegram updates.
- Telegram sender user ID persistence on jobs.
- Preservation of the existing bootstrapped user row and password hash.
- Immediate acknowledgement reply after a valid URL is queued.
- Invalid-message reply for authorized non-URL messages.
- Terminal success/failure Telegram replies using the original message as the reply target.
- Transactional notification creation when worker jobs complete.
- TTL cleanup for terminal jobs and notifications.

## Out of Scope

Not included:

- Article fetching, extraction, Markdown generation, and summarization implementation details.
- Telegram commands, menus, inline keyboards, media messages, captions, or conversation flows.
- User registration, account management, tenant administration, or user-facing user selection.
- Automatic worker retries or automatic Telegram notification retries.
- Persistent extraction observability fields such as selected extractor or extraction score.
- A dedicated observability stack.

## Users / Actors

- Authorized Telegram user.
- Telegram Bot API.
- Gateway API.
- Worker.
- SQLite database.

## Requirements

- REQ-001: The gateway must expose `POST /telegram/webhook` for Telegram update delivery.
- REQ-002: The gateway must validate `X-Telegram-Bot-Api-Secret-Token` against `Telegram:WebhookSecret` before processing a Telegram update.
- REQ-003: The gateway must process messages only from Telegram senders whose id maps to an existing `users.telegram_user_id` row.
- REQ-004: Unauthorized Telegram users must not create users, articles, jobs, notifications, or Telegram replies.
- REQ-005: The gateway must accept only text messages whose trimmed body is exactly one absolute `http` or `https` URL.
- REQ-006: Unsupported schemes, missing schemes, media/captions, extra text, and multiple tokens must be rejected.
- REQ-007: Authorized invalid messages must receive the exact reply `Nope, you must send only an URL`.
- REQ-008: A valid URL from a mapped sender must create one article record and one queued article-processing job for the resolved Archivist user in the same SQLite transaction.
- REQ-009: A valid queued URL must receive the exact acknowledgement reply `Ok, I will have a look` after the enqueue transaction commits.
- REQ-010: Failure to send the queued acknowledgement must not roll back or delete the article/job.
- REQ-011: Telegram `update_id` must be persisted on jobs for idempotency so duplicate updates do not create duplicate jobs.
- REQ-012: Jobs must retain Telegram reply-target metadata: `telegram_chat_id`, `telegram_message_id`, and `telegram_update_id`.
- REQ-013: Jobs must retain Telegram sender identity metadata as `telegram_user_id`, distinct from `telegram_chat_id`.
- REQ-014: Telegram ingestion must not create, upsert, or reassign `users`; user and Telegram identity mapping is owned by bootstrap or future user-provisioning features.
- REQ-015: The worker must claim queued jobs atomically with `UPDATE ... RETURNING`.
- REQ-016: Job states must be limited to `queued`, `running`, `succeeded`, and `failed`.
- REQ-017: Worker completion must update article state, update job state, and insert one pending notification in the same SQLite transaction.
- REQ-018: Successful worker completion must mark the article `ready`, mark the job `succeeded`, set terminal job timestamps/TTL, and create one pending notification.
- REQ-019: Failed worker completion must mark the article `failed`, mark the job `failed`, persist the final error, set terminal job timestamps/TTL, and create one pending notification.
- REQ-020: Automatic worker retries are out of scope for v0.
- REQ-021: The worker must not call Telegram APIs directly.
- REQ-022: The gateway must dispatch pending notifications by joining `notifications -> jobs -> articles`.
- REQ-023: Successful final v0 completion replies must read `summary.md` from the deterministic article artifact path under `DATA_DIR` once summary generation is implemented.
- REQ-023A: Snapshot-only or Markdown-only success replies are interim bridges only before downstream processing is implemented; final v0 success replies are summary-based.
- REQ-024: Failed completion replies must use `jobs.error_message`.
- REQ-024A: Failed article-processing completion replies must preserve ARC-coded public error text from `jobs.error_message`, including the leading `[ARC-NNN]` prefix defined by `docs/ERRORS.md`.
- REQ-025: Terminal Telegram replies must fit within Telegram message length limits by deterministic truncation when necessary.
- REQ-026: Telegram notification delivery errors must mark the notification `failed` with an error message and must not be retried automatically.
- REQ-027: Notification states must be limited to `pending`, `sent`, and `failed`.
- REQ-028: Terminal jobs expire after 14 days.
- REQ-029: Sent or failed notifications expire after 7 days.
- REQ-030: Gateway startup must validate Telegram runtime configuration and fail when `Telegram:BotToken` or `Telegram:WebhookSecret` is blank.

## Acceptance Criteria

```gherkin
Feature: Telegram ingestion

Scenario: Authorized user submits a valid URL
  Given a Telegram update has a valid webhook secret
  And the message sender maps to an existing users.telegram_user_id row
  And the message text is "https://example.com/article"
  When Telegram posts the update to /telegram/webhook
  Then one article is created with the original URL, resolved user_id, and status "queued"
  And one queued article-processing job is created for that article with the same user_id
  And the job stores telegram_update_id, telegram_chat_id, telegram_message_id, and telegram_user_id
  And the gateway replies to the original Telegram message with "Ok, I will have a look"

Scenario: Authorized user submits non-URL text
  Given a Telegram update has a valid webhook secret
  And the message sender maps to an existing users.telegram_user_id row
  And the message text is "read this please https://example.com/article"
  When Telegram posts the update to /telegram/webhook
  Then no article is created
  And no job is created
  And no notification is created
  And the gateway replies to the original Telegram message with "Nope, you must send only an URL"

Scenario: Unauthorized user sends a valid URL
  Given a Telegram update has a valid webhook secret
  And the message sender does not map to a users.telegram_user_id row
  When Telegram posts the update to /telegram/webhook
  Then no user is created or updated
  And no article is created
  And no job is created
  And no notification is created
  And no Telegram reply is sent

Scenario: Duplicate Telegram update is delivered
  Given a Telegram update has already been processed
  When Telegram posts the same update_id again
  Then no duplicate article is created
  And no duplicate job is created

Scenario: Worker claims a queued job
  Given a queued job exists
  When the worker claims work
  Then the job is atomically changed to "running" using UPDATE RETURNING
  And no locked_at or locked_by fields are required

Scenario: Job succeeds
  Given a running job originated from a Telegram message
  And the worker has written deterministic article artifacts under DATA_DIR
  When the worker completes the job successfully
  Then the article status is "ready"
  And the job status is "succeeded"
  And the job expires_at is 14 days after completion
  And one pending notification exists for the job

Scenario: Job fails
  Given a running job originated from a Telegram message
  And the worker has a final error message
  When the worker completes the job as failed
  Then the article status is "failed"
  And the job status is "failed"
  And the job error_message contains the final error
  And the job expires_at is 14 days after completion
  And one pending notification exists for the job

Scenario: Gateway sends success notification
  Given a pending notification exists for a succeeded job
  And summary generation has implemented success notification content
  When the gateway dispatches the notification through the summary-generation success branch
  Then the gateway replies to the original Telegram message with summary text
  And the notification is marked "sent"

Scenario: Gateway sends failure notification
  Given a pending notification exists for a failed job
  When the gateway dispatches the notification
  Then the gateway replies to the original Telegram message with the job error_message
  And the notification is marked "sent"

Scenario: Gateway preserves ARC-coded article-processing failure notification
  Given a pending notification exists for a failed article-processing job
  And the job error_message is "[ARC-003] The URL was not found."
  When the gateway dispatches the notification
  Then the gateway replies to the original Telegram message with "[ARC-003] The URL was not found."
  And the notification is marked "sent"

Scenario: Telegram notification delivery fails
  Given a pending notification exists
  And Telegram rejects the reply
  When the gateway dispatches the notification
  Then the notification is marked "failed"
  And the notification error_message records the delivery error
  And no automatic retry is scheduled
```

## Data and State

SQLite remains the source of truth for users, articles, jobs, and notifications.

### `users`

- `id`: ULID, seeded as `01ASB2XFCZJY7WHZ2FNRTMQJCT`.
- `telegram_user_id`: nullable until bootstrap or another user-provisioning path maps a Telegram sender; unique when present and required for accepted Telegram ingestion behavior.
- `password_hash`: Argon2id PHC string owned by `authn`; Telegram ingestion must preserve it.

The v0 system has one user row. No user timestamps, provisioning state, roles, tenants, or external identity table are required.

### `articles`

- `id`
- `user_id`
- `original_url`
- `canonical_url`, nullable
- `title`, nullable
- `status`: `queued`, `ready`, or `failed`
- `error_message`, nullable
- `created_at`

Article rows are durable archive/product state. Jobs process articles. Notifications derive success content by joining through jobs to articles and then reading deterministic artifacts from disk.

Article artifacts are not represented by path columns. Artifact paths are computed from `DATA_DIR` and `article_id`, for example:

```text
{DATA_DIR}/articles/{article_id}/snapshot.html
{DATA_DIR}/articles/{article_id}/content.md
{DATA_DIR}/articles/{article_id}/summary.md
{DATA_DIR}/articles/{article_id}/summary.json
{DATA_DIR}/articles/{article_id}/metadata.json
```

The table must not include `summary`, `domain`, artifact path columns, `selected_extractor`, `extractor_score`, or `processed_at`.

### `jobs`

- `id`
- `user_id`
- `article_id`
- `type`, initially article processing
- `status`: `queued`, `running`, `succeeded`, or `failed`
- `telegram_update_id`, unique for idempotency
- `telegram_chat_id`
- `telegram_message_id`
- `telegram_user_id`
- `error_message`, nullable
- `created_at`
- `started_at`, nullable
- `completed_at`, nullable
- `expires_at`, nullable

Jobs are temporary worker processing attempts against articles. v0 jobs do not include `attempts`, `run_after`, `locked_at`, `locked_by`, `retrying`, or `dead`.

### `notifications`

- `id`
- `job_id`, unique for terminal completion notifications
- `status`: `pending`, `sent`, or `failed`
- `error_message`, nullable
- `created_at`
- `sent_at`, nullable
- `expires_at`

Notifications are gateway delivery records. They do not copy article IDs, Telegram reply targets, user IDs, or payload text. Gateway dispatch derives reply targets from jobs and derives success content from article artifacts.

## Interfaces

- Telegram webhook: `POST /telegram/webhook`.
- Telegram webhook secret header: `X-Telegram-Bot-Api-Secret-Token`.
- Telegram send API: gateway sends replies using `Telegram:BotToken`.
- SQLite user contract: accepted Telegram messages resolve an existing `users` row by `telegram_user_id`.
- SQLite queue contract: gateway inserts article and queued job records; worker claims queued jobs atomically with `UPDATE ... RETURNING`.
- SQLite notification contract: worker inserts one pending notification when a job reaches `succeeded` or `failed`; gateway dispatches pending notifications.
- Filesystem artifact contract: worker writes deterministic article artifacts under `DATA_DIR`; summary-generation owns Gateway summary artifact reads for success replies, and UI endpoints own UI artifact reads.
- Error convention contract: `docs/ERRORS.md` defines ARC-coded public article-processing failures that Telegram notification dispatch must preserve when transported through `jobs.error_message`.
- Configuration:
  - `DATA_DIR`
  - `SQLITE_PATH`
  - `Telegram:BotToken`
  - `Telegram:WebhookSecret`
  - Gateway reads these hierarchical keys from configuration sections or `ARCHIVIST_`-prefixed environment variables, for example `ARCHIVIST_Telegram__BotToken` and `ARCHIVIST_Telegram__WebhookSecret`.

## Dependencies

Depends on:

- `docs/ARCHITECTURE.md` gateway, worker, SQLite, filesystem, and Telegram boundaries.
- `docs/DESIGN.md` decisions DSGN-002, DSGN-003, DSGN-005, DSGN-011, and DSGN-014.
- `docs/ERRORS.md` for ARC-coded public article-processing failure text transported by terminal Telegram notifications.
- `docs/ARTIFACTS.md` for deterministic article artifact paths used by downstream success notification features.

Impacts:

- Gateway module.
- Worker module.
- SQLite schema and repository contracts.
- Telegram Bot API integration.
- Filesystem artifact lookup conventions.

## Rebuild Notes

- Gateway owns all Telegram API calls.
- Worker must communicate terminal Telegram reply intent through SQLite, not direct Telegram calls or gateway RPC.
- Valid URL acknowledgement is sent synchronously after the enqueue transaction commits.
- Terminal completion replies are sent asynchronously from persisted notifications.
- `telegram_user_id` is sender identity metadata; `telegram_chat_id` and `telegram_message_id` are reply-target metadata. These fields must not be conflated.
- `users` is canonical storage for the personal user and Telegram user mapping in v0.
- Article artifact paths are computed from `DATA_DIR` and `article_id`; do not add artifact path columns unless a future spec changes that decision.
- Snapshot-only and Markdown-only success notifications are interim bridges. The final v0 extraction/summarization pipeline replaces them with summary-based completion.
- ARC error codes apply only to persisted user-facing article-processing failures. Telegram webhook validation replies, authorization failures, acknowledgement failures, and Telegram delivery errors are not ARC-coded.
- Extraction telemetry is logged, not stored in durable schema, for v0.
- Existing code is not authoritative; rebuilds must follow this spec and linked tasks.

## Security / Privacy Notes

- Webhook secret validation must run before processing update content.
- Allowed-user validation must run before any user, article, job, notification, or reply side effect except ordinary request logging.
- Telegram bot token and webhook secret are secrets and must never be committed.
- Logs must not include secret values.

## Observability / Logging Notes

- Logs for ingestion should include `telegram_update_id`, `telegram_chat_id`, `telegram_message_id`, `telegram_user_id`, accepted/rejected outcome, and article/job IDs when available.
- Logs for article processing should include `article_id`, `job_id`, URL, status, duration, and error when available.
- Logs may include extraction warnings or suspicious extraction behavior, but v0 does not persist extractor telemetry columns.
- Logs for notification dispatch should include notification ID, job ID, status, and error text when delivery fails.
- No dedicated observability stack is required for v0.

## Open Questions

- None.

## Related Documents

- `./PLAN.md`
- `./DIARY.md`
- `./tasks/TELING-001-persistence-contracts.md`
- `./tasks/TELING-002-telegram-webhook-ingestion.md`
- `./tasks/TELING-003-worker-terminal-notification-contract.md`
- `./tasks/TELING-004-telegram-notification-dispatcher.md`
