---
id: UI-004
feature: ui
title: Final UI validation pass
status: done
depends_on: [UI-003]
blocks: []
parallel: false
exec_plan: null
canonical: true
---

# UI-004: Final UI validation pass

## Objective

Validate the completed final browser UI against automated tests, route behavior, and canonical design assets.

## Scope

This task includes:

- Running frontend format, lint, build, and tests.
- Browser validation of `/login`, `/login/failed`, `/articles`, and `/articles/<article_id>`.
- Screenshot comparison against `docs/design/DESIGN.md`, `docs/design/login.png`, and `docs/design/view.png`.
- Recording validation outcomes in task status, `PLAN.md`, and `DIARY.md`.

## Out of Scope

This task does not include:

- New UI features.
- Backend endpoint changes.
- Design changes beyond fixing implementation mismatches with canonical design docs.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- Completed `UI-003`.
- All UI feature acceptance criteria.
- Design assets under `docs/design/`.

## Outputs

Expected outputs, files, behavior, or interfaces:

- Validation results are recorded.
- UI feature status can move to `done` only if validation passes or failures are explicitly documented.

## Expected Affected Areas

```text
docs/specs/ui/tasks/UI-004-final-ui-validation-pass.md
docs/specs/ui/PLAN.md
docs/specs/ui/DIARY.md
src/ui/
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/UI.md`
- `docs/design/DESIGN.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Final validation passes
  Given the UI implementation is complete
  When the validation sequence runs
  Then formatting, lint, build, and tests pass
  And browser captures for required routes match the canonical design constraints
```

## Done When

- Frontend validation commands have run.
- Browser captures have been reviewed against design docs.
- Task status, feature `PLAN.md`, and `DIARY.md` are updated.
- `docs/specs/INDEX.md` reflects final feature status if the feature is complete.

## Validation

Required checks:

```bash
cd src/ui && npm run format
cd src/ui && npm run lint
cd src/ui && npm run build
cd src/ui && npm run test
```

Manual validation:

- Capture and review `/login`.
- Capture and review `/login/failed`.
- Capture and review `/articles`.
- Capture and review `/articles/<article_id>`.

Result:

- `cd src/ui && npm run format`: passed.
- `cd src/ui && npm run lint`: passed.
- `cd src/ui && npm run build`: passed.
- `cd src/ui && npm run test`: passed, 2 test files and 21 tests.
- Gateway-seeded browser validation passed through a local HTTPS same-origin proxy with `/api/*` stripped to Gateway unprefixed routes.
- Browser captures were reviewed for `/login`, `/login/failed`, `/articles`, and `/articles/01HY0000000000000000000003`.
- The seeded ready article rendered title, summary Markdown, content Markdown, `Delete`, and `Original`; `Retry` was not present.
- The `/articles` route rendered the Gateway-backed article list with blank black detail when no article id was selected.
- The `/login/failed` route rendered a blank black page.
- A 1366x768 layout check confirmed no action/title overlap, visible footer, and internal pane scrolling.

## Dependencies

Depends on:

- `UI-003`

Blocks:

- None.

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.
