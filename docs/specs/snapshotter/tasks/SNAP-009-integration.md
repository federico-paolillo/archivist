---
id: SNAP-009
feature: snapshotter
title: Final integration validation
status: done
depends_on:
  - SNAP-006
  - SNAP-008
blocks: []
parallel: false
exec_plan: null
canonical: true
---

# SNAP-009: Final Integration Validation

## Objective

Integrate reviewed Snapshotter work, run final validation, update ALM status, and record the implementation diary entry.

## Acceptance Criteria

```gherkin
Scenario: Snapshotter feature is complete
  Given all implementation tasks are complete
  When final validation runs
  Then Python, Docker, Compose, release packaging, and diff hygiene checks pass or failures are recorded
  And task statuses, PLAN.md, DIARY.md, and INDEX.md are updated
```

## Done When

- All Snapshotter tasks are `done`.
- Final validation results are recorded.
- Feature status is updated in `PLAN.md`, `SPEC.md`, and `docs/specs/INDEX.md`.

## Validation

```bash
cd src/snapshotter && uv sync --locked --all-extras --dev
cd src/snapshotter && uv run ruff format --check .
cd src/snapshotter && uv run ruff check .
cd src/snapshotter && uv run ty check .
cd src/snapshotter && uv run pytest
docker buildx build --file snapshotter.Dockerfile --platform linux/amd64 --load --tag archivist-snapshotter:test .
docker compose config --quiet
scripts/package-compose-release.sh test-version gateway worker ui snapshotter
docker compose --env-file release/compose/.env --env-file release/compose/.env.images -f release/compose/docker-compose.yml config --quiet
git diff --check
```

## Required Context

- `../SPEC.md`
- `../PLAN.md`
- all Snapshotter task files

## Open Questions

- None.
