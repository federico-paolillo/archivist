---
feature: snapshotter
canonical: true
---
# Feature Plan: Snapshotter

## Purpose

This file controls implementation of the Snapshotter service, its backup contract, Python validation surface, Docker image, Compose deployment, and release automation.

---

## Task DAG

```text
SNAP-002 -> SNAP-003
SNAP-002 -> SNAP-004
SNAP-003 -> SNAP-005
SNAP-004 -> SNAP-005
SNAP-005 -> SNAP-006
SNAP-002 -> SNAP-007
SNAP-006 -> SNAP-007
SNAP-006 -> SNAP-008
SNAP-007 -> SNAP-008
```

---

## Execution Phases

### Phase 1: Python Service

- `SNAP-002` scaffolds the `uv` project and validation tooling.
- `SNAP-003` implements snapshot staging, SQLite online backup, artifact copy, manifest, tarball creation, and tests.
- `SNAP-004` implements S3-compatible upload and tests.
- `SNAP-005` implements the interval daemon loop and failure behavior.

### Phase 2: Deployment And Automation

- `SNAP-006` adds the Docker image, Compose service, and env examples.
- `SNAP-007` adds local and CI Python validation plus Docker build validation after the Dockerfile exists.
- `SNAP-008` extends CD image build, attestations, release package, and release notes.

---

## Tasks

| ID | Task | Depends On | Blocks | Parallel | Requires ExecPlan |
|---|---|---|---|---|---|
| `SNAP-002` | Scaffold Python snapshotter project | - | `SNAP-003`, `SNAP-004`, `SNAP-007` | yes | no |
| `SNAP-003` | Implement snapshot archive capture | `SNAP-002` | `SNAP-005` | no | no |
| `SNAP-004` | Implement S3-compatible upload | `SNAP-002` | `SNAP-005` | yes | no |
| `SNAP-005` | Implement interval daemon behavior | `SNAP-003`, `SNAP-004` | `SNAP-006` | no | no |
| `SNAP-006` | Add Docker and Compose integration | `SNAP-005` | `SNAP-007`, `SNAP-008` | no | no |
| `SNAP-007` | Add Python validation to lefthook and CI | `SNAP-002`, `SNAP-006` | `SNAP-008` | no | no |
| `SNAP-008` | Extend CD release automation | `SNAP-006`, `SNAP-007` | - | no | no |

---

## Concurrency Rules

- `SNAP-002` through `SNAP-005` modify the Snapshotter Python service and tests under `src/snapshotter/**`.
- `SNAP-006` through `SNAP-008` modify deployment, Compose, env examples, CI/CD, release scripts, and validation wiring. `SNAP-007` is sequenced after `SNAP-006` because it requires the Dockerfile before CI can validate Docker builds.
- Do not run concurrent task work against the same write scope.

---

## Blocking Interfaces or Schemas

- Snapshotter runtime command: `archivist-snapshotter`.
- Archive name: `archivist-<yyyy-mm-dd>-<unix-timestamp>.tar.gz`.
- Archive root entries: `manifest.json` and `data/`.
- S3 env vars listed in `SPEC.md`.
- Production image variable: `ARCHIVIST_SNAPSHOTTER_IMAGE`.
- No database schema changes.

---

## Validation Sequence

1. Python project validation.
2. Snapshot archive unit tests and upload stub tests.
3. Docker image build smoke.
4. Local Compose config validation.
5. Production Compose config validation with explicit image env, including `ARCHIVIST_SNAPSHOTTER_IMAGE`.
6. Full diff hygiene.

Validation commands:

```bash
cd src/snapshotter && uv sync --locked --all-extras --dev
cd src/snapshotter && uv run ruff format --check .
cd src/snapshotter && uv run ruff check .
cd src/snapshotter && uv run ty check .
cd src/snapshotter && uv run pytest
docker buildx build --file snapshotter.Dockerfile --platform linux/amd64 --load --tag archivist-snapshotter:test .
docker compose --env-file .env.local.example -f docker-compose.yaml -f docker-compose.local.yaml config --quiet
ARCHIVIST_GATEWAY_IMAGE=gateway ARCHIVIST_WORKER_IMAGE=worker ARCHIVIST_UI_IMAGE=ui ARCHIVIST_SNAPSHOTTER_IMAGE=snapshotter docker compose --env-file .env.example -f docker-compose.yaml -f docker-compose.prod.yaml config --quiet
docker compose --env-file release/compose/.env --env-file release/compose/.env.images -f release/compose/docker-compose.yaml -f release/compose/docker-compose.prod.yaml config --quiet
git diff --check
```

---

## Open Planning Questions

- None.

---

## Completion Criteria

The feature is complete when:

- all task acceptance criteria are satisfied;
- acceptance criteria in `SPEC.md` are satisfied;
- validation sequence passes or failures are recorded;
- durable implementation decisions have been promoted to canonical documents;
