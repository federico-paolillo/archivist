---
id: TELING-002
feature: telegram-ingestion
title: Telegram webhook ingestion
status: done
depends_on: [TELING-001]
blocks: [TELING-004]
parallel: true
exec_plan: null
canonical: true
---

# TELING-002: Telegram Webhook Ingestion

## Objective

Implement the gateway Telegram webhook endpoint that validates Telegram requests, authorizes the configured user, validates URL-only text messages, records sender identity metadata, enqueues valid URLs, and sends immediate Telegram replies for accepted and invalid authorized messages.

## Story / Context

As the authorized Telegram user, I want to send exactly one URL and receive a queued acknowledgement, so that I know Archivist accepted the URL for processing.

## Scope

This task includes:

- `POST /telegram/webhook`.
- Webhook secret validation using `X-Telegram-Bot-Api-Secret-Token`.
- Allowed-user validation using `TELEGRAM_ALLOWED_USER_ID`.
- Strict absolute `http`/`https` URL-only validation.
- Invalid authorized message reply: `Nope, you must send only an URL`.
- Valid queued acknowledgement reply: `Ok, I will have a look`.
- Extraction of Telegram sender user ID from the message sender, not from `chat_id`.
- Persistence of `telegram_user_id` on the personal user and queued job through the `TELING-001` contract.
- Atomic use of the persistence contract from `TELING-001`.
- Idempotent duplicate `update_id` handling.
- Gateway integration tests for the webhook behavior.

## Out of Scope

This task does not include:

- Terminal success/failure notification dispatch.
- Worker terminal state writes.
- Article processing implementation.
- UI changes.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- Completed `TELING-001` persistence contract.
- Telegram Bot API send-message behavior.
- Gateway configuration for Telegram token, allowed user ID, and webhook secret.

## Outputs

Expected outputs, files, behavior, or interfaces:

- Gateway route and handlers under the Telegram feature area.
- Gateway application services for validating updates, ensuring the personal user row, creating articles/jobs, and sending immediate replies.
- Gateway parsing that keeps Telegram sender user ID distinct from Telegram chat ID.
- Integration tests with a fake Telegram client.

## Expected Affected Areas

```text
src/gateway/Archivist.Gateway.Api/Telegram/
src/gateway/Archivist.Gateway.Application/Telegram/
src/gateway/Archivist.Gateway.Tests/
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/GATEWAY.md`
- `./TELING-001-persistence-contracts.md`

Do not load unrelated feature folders unless required by discovered dependencies.

## Acceptance Criteria

```gherkin
Scenario: Valid authorized URL is queued
  Given the webhook secret is valid
  And the Telegram sender is the configured allowed user
  And the message text is exactly "https://example.com/article"
  When Telegram posts the update
  Then the personal user row exists
  And one article is created with status "queued"
  And one queued job is created for that article
  And telegram_user_id is persisted from the message sender
  And Telegram receives the reply "Ok, I will have a look"

Scenario: Invalid authorized message is rejected
  Given the webhook secret is valid
  And the Telegram sender is the configured allowed user
  And the message text is not exactly one absolute http or https URL
  When Telegram posts the update
  Then no article is created
  And no job is created
  And Telegram receives the reply "Nope, you must send only an URL"

Scenario: Unauthorized user is ignored
  Given the webhook secret is valid
  And the Telegram sender is not the configured allowed user
  When Telegram posts the update
  Then no article is created
  And no job is created
  And no user row is created or updated
  And no Telegram reply is sent

Scenario: Sender ID and chat ID differ
  Given the webhook secret is valid
  And the Telegram sender is the configured allowed user
  And the sender user ID differs from the chat ID
  When Telegram posts a valid URL update
  Then telegram_user_id is persisted from the sender user ID
  And chat_id remains available only as reply-target metadata

Scenario: Acknowledgement send fails after enqueue
  Given a valid authorized URL has been persisted
  When the Telegram acknowledgement send fails
  Then the article remains created
  And the job remains queued
  And the failure is logged or persisted for diagnosis
```

## Done When

- Webhook route exists and follows gateway conventions.
- Valid, invalid, unauthorized, duplicate, and bad-secret paths are covered by tests.
- Tests prove sender user ID is persisted separately from chat ID.
- Valid URL enqueue commits before acknowledgement is sent.
- Acknowledgement failure does not roll back ingestion.
- Task status and `PLAN.md` are updated if the task is completed.
- `DIARY.md` has an entry if implementation is performed.

## Validation

Required checks:

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

Manual validation, if any:

- None.

## Dependencies

Depends on:

- `TELING-001`

Blocks:

- `TELING-004`

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.

## Notes

- Do not accept Telegram captions, media-only updates, commands, or URL entities with extra surrounding text for this task.
