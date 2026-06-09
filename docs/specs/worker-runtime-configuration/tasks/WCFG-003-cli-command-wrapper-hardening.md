---
id: WCFG-003
feature: worker-runtime-configuration
title: CLI command wrapper hardening
status: done
depends_on: [WCFG-002]
blocks: []
parallel: false
exec_plan: null
canonical: true
---

# WCFG-003: CLI command wrapper hardening

## Objective

Refactor Worker CLI command actions so framework registration remains thin and command behavior is testable through plain typed inputs.

## Story / Context

As a maintainer, I want `urfave/cli` usage isolated to the executable registration layer so command validation and behavior can be tested without constructing full framework state or mutating process arguments.

## Scope

This task includes:

- Keep `CliProgram` as the `urfave/cli` registration layer.
- Move command-specific argument validation and behavior dispatch into small command-named wrappers.
- Preserve existing command behavior and public error text.
- Add behavior-focused tests for wrappers where practical.

## Out of Scope

This task does not include:

- Renaming public commands or flags.
- Changing Worker configuration keys.
- Changing processing pipeline behavior.

## Inputs

Required context:

- `../SPEC.md`
- `../PLAN.md`
- `.agents/skills/archivist-worker/SKILL.md`

## Acceptance Criteria

```gherkin
Scenario: Enqueue argument validation is framework-light
  Given command wrapper input with zero or multiple URL arguments
  When enqueue command validation runs
  Then it returns the existing exact-argument error text
  And the behavior can be tested without mutating process arguments
```

## Done When

- CLI registration delegates to command wrappers with plain values.
- Existing executable command behavior is preserved.
- Worker validation passes or failures are recorded.
