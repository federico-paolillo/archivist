---
id: SNAP-006-PLAN
task: ../tasks/SNAP-006-deployment.md
status: completed
canonical: true
---

# ExecPlan: SNAP-006 Deployment

## Objective

Package Snapshotter as a rootless distroless Docker image and wire it into local and production Compose.

## Linked Task

- `../tasks/SNAP-006-deployment.md`

## Implementation Sequence

1. Add `snapshotter.Dockerfile` using BuildKit cache mounts for uv.
2. Copy a prebuilt virtual environment and application into a distroless Python final stage.
3. Ensure final application files are read-only and the runtime user is non-root.
4. Add Snapshotter to local Compose with `build:`.
5. Add Snapshotter to production Compose with `ARCHIVIST_SNAPSHOTTER_IMAGE`.
6. Add Snapshotter env values to root env examples.

## Validation Plan

```bash
docker buildx build --file snapshotter.Dockerfile --platform linux/amd64 --load --tag archivist-snapshotter:test .
docker compose config --quiet
```

## Risks

- Distroless Python image availability and venv portability must be verified by the Docker build.

## Completion Criteria

- Snapshotter image builds and Compose includes it without public ingress.
