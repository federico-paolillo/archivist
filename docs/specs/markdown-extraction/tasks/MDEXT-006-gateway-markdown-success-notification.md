---
id: MDEXT-006
feature: markdown-extraction
title: Gateway Markdown Success Notification
status: skipped
depends_on: [MDEXT-005, TELING-004]
blocks: []
parallel: false
exec_plan: null
canonical: true
---

# MDEXT-006: Gateway Markdown Success Notification

## Objective

This task is skipped because `summary-generation` supersedes Markdown-stage terminal success before this bridge is implemented.

## Story / Context

As the authorized Telegram user, I ultimately need the generated summary, so Gateway success notification work continues in `SUMGEN-005`.

## Scope

This task includes:

- Historical Markdown-stage Telegram reply text only if this task is explicitly revived before summary generation.
- Documentation that this bridge is superseded by `summary-generation`.

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

- No production outputs while skipped.

## Expected Affected Areas

```text
src/gateway/
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/ARTIFACTS.md`
- `.agents/skills/archivist-gateway/SKILL.md`
- `docs/specs/telegram-ingestion/SPEC.md`
- `docs/specs/telegram-ingestion/tasks/TELING-004-telegram-notification-dispatcher.md`
- `./MDEXT-005-worker-markdown-pipeline-integration.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Markdown-stage success notification is superseded
  Given a job has produced content.md
  And summary-generation is planned
  When this task is evaluated for implementation
  Then implementation is skipped
  And summary notification behavior is handled by SUMGEN-005

Scenario: ARC-coded failure notification is sent
  Given a pending notification exists for a failed job
  And jobs.error_message contains an ARC-coded failure
  When the Gateway dispatcher sends the terminal reply
  Then it sends that failure text
  And marks the notification sent if Telegram accepts the reply
```

## Done When

- Task remains skipped unless explicitly revived.
- Summary-complete success notifications are implemented by `SUMGEN-005`.
- Failure notification preservation remains covered by `TELING-004` and `SUMGEN-005`.
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

- None while skipped.

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

- Skipped because `summary-generation` supersedes Markdown-stage terminal success with summary-complete success before this bridge is implemented.
