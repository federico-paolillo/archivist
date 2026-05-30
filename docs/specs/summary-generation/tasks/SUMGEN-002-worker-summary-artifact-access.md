---
id: SUMGEN-002
feature: summary-generation
title: Worker Summary Artifact Access
status: done
depends_on: [SUMGEN-001, MDEXT-005, WCFG-001, WCFG-002]
blocks: [SUMGEN-004]
parallel: true
exec_plan: null
canonical: true
---

# SUMGEN-002: Worker Summary Artifact Access

## Objective

Extend the Worker artifact access layer with deterministic reads for `content.md` and traversal-resistant, atomic writes for `summary.md`.

## Story / Context

As the Worker, I need to read extracted Markdown and persist the generated summary artifact before final success is committed.

## Scope

This task includes:

- Deterministic `{DATA_DIR}/articles/{article_id}/content.md` read behavior.
- Deterministic `{DATA_DIR}/articles/{article_id}/summary.md` write behavior.
- Creation of the article artifact directory when needed.
- Atomic summary writes using a temporary file followed by rename.
- Traversal-resistant access using `os.Root` or `os.OpenInRoot` where functionally correct.
- Mapping summary write failures to `ARC-016`.
- Tests for deterministic path behavior, missing `content.md`, atomic writes, failed-write cleanup, and traversal rejection.

## Out of Scope

This task does not include:

- Markdown extraction.
- Summarizer provider calls.
- SQLite job state transitions.
- Gateway artifact reads.
- Writing `summary.json` or metadata artifacts.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- Completed `SUMGEN-001`.
- Completed `MDEXT-005`.
- Completed `WCFG-001`.
- Completed `WCFG-002`.
- `docs/ARTIFACTS.md`
- `docs/ERRORS.md`
- `.agents/skills/archivist-worker/SKILL.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Worker artifact-store behavior for reading `content.md`.
- Worker artifact-store behavior for atomic `summary.md` writes.
- Tests proving path and atomic write behavior.

## Expected Affected Areas

```text
src/worker/
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/ARTIFACTS.md`
- `docs/ERRORS.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-worker/SKILL.md`
- `docs/specs/markdown-extraction/tasks/MDEXT-002-worker-markdown-artifact-access.md`
- `docs/specs/markdown-extraction/tasks/MDEXT-005-worker-markdown-pipeline-integration.md`
- `docs/specs/worker-runtime-configuration/tasks/WCFG-001-canonical-worker-config-loading.md`
- `docs/specs/worker-runtime-configuration/tasks/WCFG-002-non-optional-worker-composition.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Summary is written atomically
  Given an article ID and summary text
  When the Worker stores the summary
  Then the final file is DATA_DIR/articles/{article_id}/summary.md
  And the write uses a temporary file followed by rename
  And no partial temporary file is promoted on write failure

Scenario: Markdown content is read deterministically
  Given an article ID with content.md
  When the Worker loads summary input
  Then it reads DATA_DIR/articles/{article_id}/content.md

Scenario: Artifact access rejects traversal
  Given an invalid article artifact name attempts to escape DATA_DIR
  When the artifact layer resolves or opens it
  Then the operation fails
```

## Done When

- Worker artifact access supports reading `content.md`.
- Worker artifact access supports atomic `summary.md` writes.
- Summary write failures map to `ARC-016`.
- Artifact paths match `docs/ARTIFACTS.md`.
- No placeholder `summary.json` artifact is created.
- Tests cover deterministic path, atomic write, failed-write cleanup, missing input, and traversal rejection.
- Task status and `PLAN.md` are updated if the task is completed.
- `DIARY.md` has an entry if implementation is performed.

## Validation

Required checks:

```bash
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

Manual validation, if any:

- Inspect a test artifact directory to confirm only expected artifacts are written.

Validation completed on 2026-05-20:

- `cd src/worker && go tool lefthook run build` — passed.
- `cd src/worker && go tool lefthook run format` — passed.
- `cd src/worker && go tool lefthook run lint` — passed.
- `cd src/worker && go tool lefthook run test` — passed.

## Dependencies

Depends on:

- `SUMGEN-001`
- `MDEXT-005`
- `WCFG-001`
- `WCFG-002`

Blocks:

- `SUMGEN-004`

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.

## Notes

- This task extends, rather than replaces, the artifact access layer used by article processing and Markdown extraction.
