---
feature: snapshotter
status: done
canonical: true
---

# Feature Plan: Snapshotter

## Purpose

This file controls implementation of the Snapshotter service, its backup contract, Python validation surface, Docker image, Compose deployment, and release automation.

---

## Task DAG

```text
SNAP-001 -> SNAP-002
SNAP-001 -> SNAP-006
SNAP-002 -> SNAP-003
SNAP-002 -> SNAP-004
SNAP-003 -> SNAP-005
SNAP-004 -> SNAP-005
SNAP-005 -> SNAP-006
SNAP-002 -> SNAP-007
SNAP-006 -> SNAP-008
SNAP-007 -> SNAP-008
SNAP-008 -> SNAP-009
SNAP-006 -> SNAP-009
```

---

## Execution Phases

### Phase 1: Canonical Contract

- `SNAP-001` records the feature spec, task plan, backup contract, architecture, artifact, design, and rebuild updates.

### Phase 2: Python Service

- `SNAP-002` scaffolds the `uv` project and validation tooling.
- `SNAP-003` implements snapshot staging, SQLite online backup, artifact copy, manifest, tarball creation, and tests.
- `SNAP-004` implements S3-compatible upload and tests.
- `SNAP-005` implements the interval daemon loop and failure behavior.

### Phase 3: Deployment And Automation

- `SNAP-006` adds the Docker image, Compose service, and env examples.
- `SNAP-007` adds local and CI Python validation.
- `SNAP-008` extends CD image build, attestations, release package, and release notes.

### Phase 4: Integration

- `SNAP-009` integrates reviewed slices, runs final validation, and records completion.

---

## Tasks

| ID | Task | Status | Depends On | Blocks | Parallel | ExecPlan |
|---|---|---|---|---|---|---|
| `SNAP-001` | Create canonical snapshotter contract | done | - | `SNAP-002`, `SNAP-006` | no | - |
| `SNAP-002` | Scaffold Python snapshotter project | done | `SNAP-001` | `SNAP-003`, `SNAP-004`, `SNAP-007` | yes | `plans/SNAP-002-python-scaffold.execplan.md` |
| `SNAP-003` | Implement snapshot archive capture | done | `SNAP-002` | `SNAP-005` | no | `plans/SNAP-003-capture.execplan.md` |
| `SNAP-004` | Implement S3-compatible upload | done | `SNAP-002` | `SNAP-005` | yes | `plans/SNAP-004-upload.execplan.md` |
| `SNAP-005` | Implement interval daemon behavior | done | `SNAP-003`, `SNAP-004` | `SNAP-006` | no | `plans/SNAP-005-daemon.execplan.md` |
| `SNAP-006` | Add Docker and Compose integration | done | `SNAP-001`, `SNAP-005` | `SNAP-008`, `SNAP-009` | no | `plans/SNAP-006-deployment.execplan.md` |
| `SNAP-007` | Add Python validation to lefthook and CI | done | `SNAP-002` | `SNAP-008` | yes | - |
| `SNAP-008` | Extend CD release automation | done | `SNAP-006`, `SNAP-007` | `SNAP-009` | no | `plans/SNAP-008-release.execplan.md` |
| `SNAP-009` | Final integration validation | done | `SNAP-006`, `SNAP-008` | - | no | - |

---

## Concurrency Rules

- `SNAP-002` through `SNAP-005` are owned by the Python worker under `src/snapshotter/**`.
- `SNAP-006` through `SNAP-008` are owned by the deployment worker for Docker, Compose, env examples, workflow, release scripts, and `lefthook.yml`.
- Deployment work may prepare scripts and workflow structure after `SNAP-001`, but must not finalize image validation until the Python entrypoint exists.
- Coordinator owns ALM status updates, feature diary, review routing, and final validation.
- Do not run two workers against the same write scope.

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
5. Production Compose config validation with generated image env.
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
docker compose --env-file release/compose/.env --env-file release/compose/.env.images -f release/compose/docker-compose.yaml -f release/compose/docker-compose.prod.yaml config --quiet
git diff --check
```

---

## Open Planning Questions

- None.

---

## Completion Criteria

The feature is complete when:

- all required tasks are `done`;
- acceptance criteria in `SPEC.md` are satisfied;
- validation sequence passes or failures are recorded;
- durable implementation decisions have been promoted to canonical documents;
- `DIARY.md` contains final implementation notes;
- `docs/specs/INDEX.md` reflects the final feature status.
