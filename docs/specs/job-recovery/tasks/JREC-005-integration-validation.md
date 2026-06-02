---
id: JREC-005
feature: job-recovery
title: Integration validation
status: done
depends_on: [JREC-002, JREC-003, JREC-004]
blocks: []
parallel: false
exec_plan: null
canonical: true
---

# JREC-005: Integration Validation

## Objective

Integrate reviewed Gateway, UI, and Worker slices, run final validation, and complete ALM bookkeeping.

## Scope

This task includes:

- Integrating module worker branches.
- Resolving conflicts narrowly.
- Running final validation.
- Updating task statuses, `PLAN.md`, `DIARY.md`, and `docs/specs/INDEX.md`.

## Out of Scope

This task does not include:

- New feature behavior beyond preserving approved worker changes.
- Cleanup of unrelated untracked files.

## Inputs

- Completed and reviewed `JREC-002`.
- Completed and reviewed `JREC-003`.
- Completed and reviewed `JREC-004`.

## Outputs

- Integrated feature implementation.
- Validation record.
- ALM completion updates.

## Expected Affected Areas

```text
docs/specs/job-recovery/**
docs/specs/INDEX.md
src/gateway/**
src/ui/**
src/worker/**
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `./JREC-005-integration-validation.md`
- Worker reports and reviewer reports.
- `.agents/skills/archivist-integrator/SKILL.md`

## Acceptance Criteria

```gherkin
Scenario: Feature validates after integration
  Given Gateway, UI, and Worker slices are approved
  When the coordinator integrates them
  Then final validation passes
  And ALM artifacts record completion
```

## Done When

- Module branches are integrated.
- Final validation results are recorded.
- Feature status is updated.
- Diary records implementation outcome.

## Validation

Required checks:

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
cd src/ui && npm run format
cd src/ui && npm run lint
cd src/ui && npm run build
cd src/ui && npm run test
git diff --check
```

Manual validation, if any:

- Browser validation for article detail force-delete visibility and success when a local API/UI fixture is available.

## Dependencies

Depends on:

- `JREC-002`
- `JREC-003`
- `JREC-004`

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

- Integrated approved Gateway, UI, and Worker slices in the coordinator worktree.
- Validation: `git diff --check` passed; `go tool lefthook run build`, `format`, `lint`, and `test` passed; `npm run format`, `lint`, `build`, and `test` passed; `dotnet format` and `dotnet build` passed; focused Gateway rollback test passed.
- Direct Gateway `dotnet test` stalled after test discovery and was terminated; integrated hook Gateway tests passed with 179 tests.
