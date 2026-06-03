---
id: SNAP-007
feature: snapshotter
title: Add Python validation to lefthook and CI
status: done
depends_on:
  - SNAP-002
blocks:
  - SNAP-008
parallel: true
exec_plan: null
canonical: true
---

# SNAP-007: Add Python Validation To Lefthook And CI

## Objective

Extend local lefthook and GitHub Actions CI so Snapshotter format, lint, type check, tests, and Docker build validation are required.

## Acceptance Criteria

```gherkin
Scenario: CI validates Snapshotter
  Given a push to main
  When CI runs
  Then Snapshotter uv sync, ruff format check, ruff lint, ty, pytest, and Docker build validation run
```

## Done When

- `lefthook.yml` includes Snapshotter validation in relevant groups.
- `.github/workflows/ci.yml` installs uv and runs Snapshotter validation.
- CI includes a Docker build check for `snapshotter.Dockerfile`.

## Validation

```bash
cd src/snapshotter && uv run ruff format --check .
cd src/snapshotter && uv run ruff check .
cd src/snapshotter && uv run ty check .
cd src/snapshotter && uv run pytest
docker buildx build --file snapshotter.Dockerfile --platform linux/amd64 --load --tag archivist-snapshotter:test .
```

## Required Context

- `../SPEC.md`
- `../PLAN.md`
- `.github/workflows/ci.yml`
- `lefthook.yml`

## Open Questions

- None.
