---
id: SNAP-006
feature: snapshotter
title: Add Docker and Compose integration
status: done
depends_on:
  - SNAP-001
  - SNAP-005
blocks:
  - SNAP-008
  - SNAP-009
parallel: false
exec_plan: ../plans/SNAP-006-deployment.execplan.md
canonical: true
---

# SNAP-006: Add Docker And Compose Integration

## Objective

Add the Snapshotter Docker image and wire the service into local and production Compose with env examples.

## Acceptance Criteria

```gherkin
Scenario: Snapshotter is deployable in Compose
  Given the local and production Compose files
  When Compose config validation runs
  Then Snapshotter is present with the shared data volume, configured env vars, and no public ports
```

## Done When

- `snapshotter.Dockerfile` builds for `linux/amd64`.
- Local Compose uses `build:`.
- Production Compose uses `ARCHIVIST_SNAPSHOTTER_IMAGE`.
- Env examples include Snapshotter configuration.

## Validation

```bash
docker buildx build --file snapshotter.Dockerfile --platform linux/amd64 --load --tag archivist-snapshotter:test .
docker compose config --quiet
```

## Required Context

- `../SPEC.md`
- `../PLAN.md`
- `../../ARCHITECTURE.md`

## Open Questions

- None.
