---
id: ARTPROC-006
feature: article-processing
title: Gateway Snapshot Success Notification Bridge
status: skipped
depends_on: [ARTPROC-005, TELING-004]
blocks: []
parallel: false
exec_plan: null
canonical: true
---

# ARTPROC-006: Gateway Snapshot Success Notification Bridge

## Objective

Update Gateway terminal notification dispatch so succeeded snapshot-only jobs produce a deterministic success reply when summary artifacts do not yet exist.

## Skip Rationale

This task is skipped because downstream processing supersedes snapshot-only terminal success before this bridge is implemented. Final Gateway success notification work continues in `SUMGEN-005`.

## Story / Context

As the authorized Telegram user, I need a terminal reply after snapshot processing succeeds even though extraction and summarization are specified in a later v0 feature.

## Scope

This task includes:

- Success notification content selection for succeeded jobs without summary artifacts.
- Deterministic snapshot-complete Telegram reply text.
- Tests covering snapshot-only success notification.
- Documentation that this bridge must be superseded by the later v0 extraction/summarization feature.

## Out of Scope

This task does not include:

- Worker processing.
- Summary generation.
- Summary artifact rendering.
- Telegram retry behavior.
- UI notification surfaces.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- Completed `ARTPROC-005`.
- Completed `TELING-004` notification dispatcher.
- `../SPEC.md`
- `docs/specs/telegram-ingestion/tasks/TELING-004-telegram-notification-dispatcher.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Gateway dispatcher success branch for snapshot-only completed jobs.
- Gateway tests for snapshot-only success and ARC-coded failure notification behavior.

## Expected Affected Areas

```text
src/gateway/
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/conventions/GATEWAY.md`
- `docs/specs/telegram-ingestion/SPEC.md`
- `docs/specs/telegram-ingestion/tasks/TELING-004-telegram-notification-dispatcher.md`
- `./ARTPROC-005-worker-snapshot-pipeline-orchestration.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Snapshot-only success notification is sent
  Given a pending notification exists for a succeeded job
  And the article has snapshot.html
  And no summary artifact exists
  When the Gateway dispatcher sends the terminal reply
  Then it sends deterministic snapshot-complete success text
  And marks the notification sent if Telegram accepts the reply

Scenario: ARC-coded failure notification is sent
  Given a pending notification exists for a failed job
  And jobs.error_message contains an ARC-coded failure
  When the Gateway dispatcher sends the terminal reply
  Then it sends that failure text
  And marks the notification sent if Telegram accepts the reply
```

## Done When

- Snapshot-only success notifications are supported.
- Missing summary artifacts for snapshot-only completed jobs do not fail notification delivery.
- Failure notifications preserve ARC-coded job error text.
- Tests cover snapshot-only success and ARC-coded failure.
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

- `ARTPROC-005`
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

- Skipped because `summary-generation` supersedes snapshot-only terminal success with summary-complete success before this bridge is implemented.
