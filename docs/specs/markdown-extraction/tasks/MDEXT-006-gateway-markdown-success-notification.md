---
id: MDEXT-006
feature: markdown-extraction
title: Gateway Markdown Success Notification
status: blocked
depends_on: [MDEXT-005, TELING-004]
blocks: []
parallel: false
exec_plan: null
canonical: true
---

# MDEXT-006: Gateway Markdown Success Notification

## Objective

Update Gateway terminal notification dispatch so succeeded Markdown-complete jobs produce deterministic success replies.

## Story / Context

As the authorized Telegram user, I need a terminal reply after Markdown extraction succeeds even though summary generation is specified in a later feature.

## Scope

This task includes:

- Success notification content selection for succeeded jobs with `content.md`.
- Deterministic Markdown-complete Telegram reply text.
- Tests covering Markdown-complete success notification.
- Documentation that this bridge must be superseded by the future summary feature.
- Removal or replacement of snapshot-only success behavior introduced by `ARTPROC-006` if it exists.

## Out of Scope

This task does not include:

- Worker processing.
- Summary generation.
- Summary artifact rendering.
- Telegram retry behavior.
- UI notification surfaces.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- Completed `MDEXT-005`.
- Completed `TELING-004`.
- `../SPEC.md`
- `docs/specs/telegram-ingestion/tasks/TELING-004-telegram-notification-dispatcher.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Gateway dispatcher success branch for Markdown-complete jobs.
- Gateway tests for Markdown-complete success and ARC-coded failure notification behavior.

## Expected Affected Areas

```text
src/gateway/
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/conventions/ARTIFACTS.md`
- `docs/conventions/GATEWAY.md`
- `docs/specs/telegram-ingestion/SPEC.md`
- `docs/specs/telegram-ingestion/tasks/TELING-004-telegram-notification-dispatcher.md`
- `./MDEXT-005-worker-markdown-pipeline-integration.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Markdown-complete success notification is sent
  Given a pending notification exists for a succeeded job
  And the article has content.md
  And no summary artifact exists
  When the Gateway dispatcher sends the terminal reply
  Then it sends deterministic Markdown-complete success text
  And marks the notification sent if Telegram accepts the reply

Scenario: ARC-coded failure notification is sent
  Given a pending notification exists for a failed job
  And jobs.error_message contains an ARC-coded failure
  When the Gateway dispatcher sends the terminal reply
  Then it sends that failure text
  And marks the notification sent if Telegram accepts the reply
```

## Done When

- Markdown-complete success notifications are supported.
- Snapshot-only success is no longer the terminal success notification for this sequence.
- Missing summary artifacts for Markdown-complete jobs do not fail notification delivery.
- Failure notifications preserve ARC-coded job error text.
- Tests cover Markdown-complete success and ARC-coded failure.
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

- `MDEXT-005`
- `TELING-004`

Blocks:

- None.

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.

## Notes

- This bridge is not the final v0 success notification contract. The future summary feature must replace it with summary-based completion.
