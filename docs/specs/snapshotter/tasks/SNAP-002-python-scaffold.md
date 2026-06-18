---
id: SNAP-002
feature: snapshotter
title: Scaffold Python snapshotter project
depends_on: []
blocks:
  - SNAP-003
  - SNAP-004
  - SNAP-007
parallel: true
requires_exec_plan: false
canonical: true
---
# SNAP-002: Scaffold Python Snapshotter Project

## Objective

Create the `src/snapshotter` Python 3.12 `uv` project with locked dependencies, executable entrypoint, and validation tooling.

## Scope

This task includes:

- `pyproject.toml`
- `uv.lock`
- `src/archivist_snapshotter/**`
- `tests/**`
- `ruff`, `ty`, and `pytest` configuration


## Acceptance Criteria

```gherkin
Scenario: Python project validates
  Given the Snapshotter Python project exists
  When validation commands run
  Then uv sync, ruff format check, ruff lint, ty, and pytest pass
```

## Done When

- `uv sync --locked --all-extras --dev` succeeds.
- `uv run ruff format --check .` succeeds.
- `uv run ruff check .` succeeds.
- `uv run ty check .` succeeds.
- `uv run pytest` succeeds.

## Validation

```bash
cd src/snapshotter && uv sync --locked --all-extras --dev
cd src/snapshotter && uv run ruff format --check .
cd src/snapshotter && uv run ruff check .
cd src/snapshotter && uv run ty check .
cd src/snapshotter && uv run pytest
```

## Required Context

- `../SPEC.md`
- `../PLAN.md`
- `.agents/skills/archivist-general/SKILL.md`

## Open Questions

- None.
