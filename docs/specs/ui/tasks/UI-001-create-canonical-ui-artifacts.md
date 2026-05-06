---
id: UI-001
feature: ui
title: Create canonical UI artifacts
status: done
depends_on: []
blocks: [UI-002]
parallel: false
exec_plan: null
canonical: true
---

# UI-001: Create canonical UI artifacts

## Objective

Create the canonical ALM artifacts for the final browser UI feature and record required cross-feature ownership decisions.

## Scope

This task includes:

- `docs/specs/ui/SPEC.md`
- `docs/specs/ui/PLAN.md`
- `docs/specs/ui/DIARY.md`
- UI task files.
- UI ExecPlans for complex implementation tasks.
- Updates to `docs/specs/INDEX.md`.
- Updates to affected canonical architecture, convention, auth, and UI endpoint docs.

## Out of Scope

This task does not include:

- Production UI implementation.
- Gateway implementation.
- Worker implementation.
- Browser validation of production UI behavior.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- User-approved Final Browser UI ALM Plan.
- `docs/design/DESIGN.md`
- `docs/design/login.png`
- `docs/design/view.png`
- `docs/specs/authn/SPEC.md`
- `docs/specs/ui-endpoints/SPEC.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Canonical UI feature artifacts exist and are linked from `docs/specs/INDEX.md`.
- Browser UI ownership is separated from auth and article API endpoint ownership.
- `VITE_API_BASE_PATH` and `/api` reverse-proxy expectations are canonical.

## Expected Affected Areas

```text
docs/specs/ui/
docs/specs/INDEX.md
docs/ARCHITECTURE.md
docs/conventions/UI.md
docs/specs/authn/
docs/specs/ui-endpoints/
```

## Required Context

Read before execution:

- `AGENTS.md`
- `docs/BOOKKEEPING.md`
- `docs/REBUILD.md`
- `docs/specs/INDEX.md`
- `docs/ARCHITECTURE.md`
- `docs/conventions/UI.md`
- `docs/specs/authn/SPEC.md`
- `docs/specs/ui-endpoints/SPEC.md`
- `docs/design/DESIGN.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: UI ALM artifacts exist
  Given the repository uses Markdown ALM as the source of truth
  When UI-001 is complete
  Then docs/specs/ui/SPEC.md exists
  And docs/specs/ui/PLAN.md exists
  And docs/specs/ui/tasks/*.md exists
  And docs/specs/ui/plans/*.execplan.md exists for complex tasks
  And docs/specs/INDEX.md links the ui feature
```

## Done When

- UI feature artifacts are created.
- `docs/specs/INDEX.md` includes `ui`.
- Cross-feature canonical docs record auth/UI/API ownership.
- `DIARY.md` contains an entry for this task.

## Validation

Required checks:

```bash
git diff --check
```

Manual validation:

- Confirm no production source files are changed by this task.

## Dependencies

Depends on:

- None.

Blocks:

- `UI-002`

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.
