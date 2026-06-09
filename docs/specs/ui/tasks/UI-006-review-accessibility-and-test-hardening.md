---
id: UI-006
feature: ui
title: Review accessibility and test hardening
status: done
depends_on: [UI-005]
blocks: []
parallel: false
exec_plan: null
canonical: true
---

# UI-006: Review accessibility and test hardening

## Objective

Close UI review findings around app-test cleanup, jsdom setup, and minimal accessible semantics for selected article rows and the user menu.

## Story / Context

As the personal Archivist user, I want the UI tests to be repeatable and the UI controls to expose minimal standards-compliant state without overbuilding full assistive-navigation patterns that the current interactions do not implement.

## Scope

This task includes:

- Unmount the actual Preact roots created by UI app tests.
- Add a focused Vitest setup file for jsdom browser API shims such as `window.scrollTo`.
- Keep article rows as buttons and expose selected state with minimal accessible semantics such as `aria-pressed`.
- Treat the user menu as a simple disclosure/popover with native buttons rather than an ARIA menu.
- Update the UI skill to prefer minimal standards-compliant accessibility semantics over heavyweight ARIA roles unless full patterns are implemented.

## Out of Scope

This task does not include:

- A full screen-reader optimized navigation model.
- Full ARIA menu/listbox/grid keyboard and focus-management patterns.
- Route-link conversion for article rows.
- Visual redesign.

## Inputs

Required context:

- `../SPEC.md`
- `../PLAN.md`
- `.agents/skills/archivist-ui/SKILL.md`

## Acceptance Criteria

```gherkin
Scenario: UI tests clean up mounted roots
  Given a UI test mounts the app into a generated Preact root
  When the test cleanup runs
  Then the actual root is unmounted before document body children are removed
```

```gherkin
Scenario: Selected article row exposes selected state
  Given an article is selected
  Then its row button exposes a programmatic selected state in addition to visual styling
```

```gherkin
Scenario: User menu uses disclosure semantics
  Given the user menu is open
  Then the popover does not claim ARIA menu semantics unless the full menu pattern is implemented
```

## Done When

- UI tests are hermetic and no longer emit `scrollTo` jsdom warnings.
- Selected article and user menu semantics match the minimal canonical contract.
- UI validation passes or failures are recorded.
