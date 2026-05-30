---
id: TELING-004
feature: telegram-ingestion
title: Telegram notification dispatcher
status: done
depends_on: [TELING-002, TELING-003]
blocks: []
parallel: false
exec_plan: ../plans/TELING-004-telegram-notification-dispatcher.execplan.md
canonical: true
---

# TELING-004: Telegram Notification Dispatcher

## Objective

Implement the gateway-owned terminal notification dispatcher that polls pending notification rows, derives reply data from jobs/articles, and sends success or failure Telegram replies to the original message.

## Story / Context

As the authorized Telegram user, I want Archivist to reply to my original URL message when processing finishes, so that I can see the summary or final error where I submitted the article.

## Scope

This task includes:

- Gateway background dispatcher for pending terminal notification rows.
- Telegram send-message integration for terminal replies.
- Reply targeting using `jobs.telegram_chat_id` and `jobs.telegram_message_id`.
- Dispatcher infrastructure for succeeded-job notifications without final success content selection.
- Summary-based success notification content remains owned by `SUMGEN-005`.
- Failure reply content loaded from `jobs.error_message`.
- ARC-coded article-processing failure replies preserve `jobs.error_message` unchanged except for deterministic Telegram length truncation.
- Deterministic Telegram message length truncation.
- Delivery failure handling that marks the notification `failed` without retrying.
- Notification cleanup for sent/failed rows after 7 days.
- Dispatcher tests with a fake Telegram client and SQLite-backed notifications.

## Out of Scope

This task does not include:

- Immediate queued acknowledgement dispatch.
- Worker terminal notification writes.
- Webhook ingestion.
- Summary success artifact reads or success reply body construction.
- UI notification surfaces.
- Automatic retry behavior.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- Completed `TELING-002` webhook ingestion.
- Completed `TELING-003` worker terminal notification contract.
- Notification schema from `TELING-001`.

## Outputs

Expected outputs, files, behavior, or interfaces:

- Gateway hosted service or equivalent dispatcher loop.
- Telegram client abstraction used by webhook ingestion and dispatcher where appropriate.
- Tests covering sent, failed delivery, message truncation, and TTL cleanup.

## Expected Affected Areas

```text
src/gateway/Archivist.Gateway.Application/Telegram/
src/gateway/Archivist.Gateway.Api/Telegram/
src/gateway/Archivist.Gateway.Tests/
SQLite repository code
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/ERRORS.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-gateway/SKILL.md`
- `./TELING-002-telegram-webhook-ingestion.md`
- `./TELING-003-worker-terminal-notification-contract.md`
- `../plans/TELING-004-telegram-notification-dispatcher.execplan.md`

Do not load unrelated feature folders unless required by discovered dependencies.

## Acceptance Criteria

```gherkin
Scenario: Dispatcher sends success reply
  Given a pending notification exists for a succeeded job
  When only TELING-004 has been implemented
  Then the reply target is read from the job Telegram metadata
  And final success reply body construction remains unavailable until SUMGEN-005
  And the notification remains pending rather than being sent with snapshot or Markdown completion text

Scenario: Dispatcher sends failure reply
  Given a pending notification exists for a failed job
  When the dispatcher sends the Telegram reply
  Then the reply target is read from the job Telegram metadata
  And the reply body is read from jobs.error_message
  And the notification is marked sent

Scenario: Dispatcher preserves ARC-coded article-processing failure reply
  Given a pending notification exists for a failed article-processing job
  And jobs.error_message is "[ARC-003] The URL was not found."
  When the dispatcher sends the Telegram reply
  Then the reply body is "[ARC-003] The URL was not found."
  And the notification is marked sent

Scenario: Telegram delivery fails
  Given Telegram rejects the reply
  When the dispatcher handles the failure
  Then the notification is marked failed
  And the notification error_message records the delivery error
  And no retry is scheduled
  And terminal article/job state is unchanged

Scenario: Expired sent or failed notifications are cleaned up
  Given a sent or failed notification has expired
  When gateway cleanup runs
  Then the notification is deleted
```

## Done When

- Dispatcher sends terminal replies from pending notification rows.
- Dispatcher leaves succeeded-job success content selection to downstream feature tasks such as `SUMGEN-005`.
- Dispatcher preserves ARC-coded article-processing failure text from `jobs.error_message`, subject only to deterministic Telegram length truncation.
- Dispatcher never changes terminal article/job state as a side effect of Telegram delivery failure.
- Telegram delivery failure marks the notification failed without retrying.
- Sent or failed notifications expire after 7 days and are cleaned up by the gateway.
- Long failure messages are truncated deterministically to Telegram limits.
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

- `TELING-002`
- `TELING-003`

Blocks:

- None.

## ExecPlan

ExecPlan:

```text
../plans/TELING-004-telegram-notification-dispatcher.execplan.md
```

## Open Questions

- None.

## Notes

- The gateway owns this dispatcher because the gateway owns Telegram API integration.
- Snapshot-only success text is an interim bridge for article-processing; later v0 extraction/summarization work must replace it with summary-based completion.
- ARC codes are transported for article-processing terminal failures only. Telegram webhook validation replies and Telegram delivery errors are not ARC-coded.
