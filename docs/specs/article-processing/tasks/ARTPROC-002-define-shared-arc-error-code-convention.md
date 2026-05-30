---
id: ARTPROC-002
feature: article-processing
title: Define Shared ARC Error-Code Convention
status: done
depends_on: [ARTPROC-001]
blocks: [ARTPROC-004]
parallel: true
exec_plan: null
canonical: true
---

# ARTPROC-002: Define Shared ARC Error-Code Convention

## Objective

Define a shared `ARC-NNN` error-code catalog for persisted article-processing failures.

## Story / Context

As Worker, Gateway, and UI implementers, we need stable user-facing error codes so components can communicate failures without leaking HTTP, filesystem, library, or provider details into public article state.

## Scope

This task includes:

- `docs/ERRORS.md`.
- Initial `ARC-001` through `ARC-007` and `ARC-999` catalog.
- Convention updates requiring persisted article-processing errors to use ARC codes.

## Out of Scope

This task does not include:

- Worker failure mapping implementation.
- UI rendering of error codes.
- Adding database failure-code columns.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- `../SPEC.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-worker/SKILL.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Shared ARC error-code convention.
- Updated general and Worker conventions.

## Expected Affected Areas

```text
docs/ERRORS.md
.agents/skills/archivist-general/SKILL.md
.agents/skills/archivist-worker/SKILL.md
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-worker/SKILL.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: ARC catalog is canonical
  Given article-processing failures need user-facing codes
  When the shared convention is created
  Then ARC codes and public message rules are documented in docs/ERRORS.md
  And Worker conventions require persisted processing failures to use those codes
```

## Done When

- ARC codes are documented.
- Unknown failures use `ARC-999`.
- Public article errors are required to start with `[ARC-NNN]`.
- No new database columns are required by the convention.

## Validation

Required checks:

```bash
git diff -- docs/ERRORS.md .agents/skills/archivist-general/SKILL.md .agents/skills/archivist-worker/SKILL.md
```

Manual validation, if any:

- Inspect Markdown for unresolved template placeholders.

## Dependencies

Depends on:

- `ARTPROC-001`

Blocks:

- `ARTPROC-004`

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.

## Notes

- This task was completed during feature planning artifact creation.
