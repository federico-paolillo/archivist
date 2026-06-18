---
id: SNAP-006
feature: snapshotter
title: Add Docker and Compose integration
depends_on:
  - SNAP-005
blocks:
  - SNAP-007
  - SNAP-008
parallel: false
requires_exec_plan: false
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
docker compose --env-file .env.local.example -f docker-compose.yaml -f docker-compose.local.yaml config --quiet
ARCHIVIST_GATEWAY_IMAGE=gateway ARCHIVIST_WORKER_IMAGE=worker ARCHIVIST_UI_IMAGE=ui ARCHIVIST_SNAPSHOTTER_IMAGE=snapshotter docker compose --env-file .env.example -f docker-compose.yaml -f docker-compose.prod.yaml config --quiet
```

## Required Context

- `../SPEC.md`
- `../PLAN.md`
- `../../ARCHITECTURE.md`

## Open Questions

- None.
