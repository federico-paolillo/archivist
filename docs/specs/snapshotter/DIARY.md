# DIARY.md

Append-only implementation log for the Snapshotter feature.

Historical entries may explain implementation choices, but required behavior must be promoted to canonical docs before being relied on for rebuild.

---

## 2026-06-02: SNAP-001..SNAP-009 Done

- Status outcome: Snapshotter feature completed.
- Summary of changes: added canonical Snapshotter specs/tasks/ExecPlans, Python 3.12 `uv` service under `src/snapshotter`, SQLite-online-backup archive capture, best-effort artifact copying, S3-compatible upload, interval daemon loop, rootless distroless Docker image, Compose service, env examples, lefthook validation, CI/CD validation, image build/smoke checks, and release package updates.
- Decisions made: symlinks under `/data` are skipped during snapshot capture to avoid copying content from outside the backup root; `ARCHIVIST_SNAPSHOTTER_WORK_DIR` is rejected when configured inside `ARCHIVIST_DATA_DIR`; CI/CD now smoke-runs the built Snapshotter image because build-only validation missed missing runtime libraries.
- Validation performed: `uv sync --locked --all-extras --dev`; `uv run ruff format --check .`; `uv run ruff check .`; `uv run ty check .`; `uv run pytest`; `docker buildx build --file snapshotter.Dockerfile --platform linux/amd64 --load --tag archivist-snapshotter:test .`; `docker run --rm --read-only --tmpfs /tmp --entrypoint /app/venv/bin/archivist-snapshotter archivist-snapshotter:test --help`; `docker compose --env-file .env.example config --quiet`; `sh -n scripts/package-compose-release.sh`; `sh -n scripts/create-draft-release.sh`; `scripts/package-compose-release.sh test-version gateway-image worker-image ui-image snapshotter-image`; production Compose validation with generated `docker-compose.images.env`; `git diff --check`.
- Follow-ups: none.
- Canonical documents updated: `docs/ARCHITECTURE.md`, `docs/ARTIFACTS.md`, `docs/DESIGN.md`, `docs/REBUILD.md`, `docs/specs/INDEX.md`, `docs/specs/snapshotter/SPEC.md`, `docs/specs/snapshotter/PLAN.md`, Snapshotter task files, and Snapshotter ExecPlans.

## 2026-06-03: SNAP-008 Release Package Filename Follow-Up

- Status outcome: release package filename contract corrected.
- Summary of changes: updated CD production Compose validation, image display, architecture notes, rebuild ordering, and Snapshotter validation references to use packaged `release/compose/docker-compose.yml`, `release/compose/.env`, and `release/compose/.env.images`.
- Decisions made: the repository source file remains `docker-compose.prod.yaml`; only generated release package filenames are normalized for operator deployment.
- Validation performed: `sh -n scripts/package-compose-release.sh`; `sh -n scripts/create-draft-release.sh`; `scripts/package-compose-release.sh test-version gateway worker ui snapshotter`; `docker compose --env-file release/compose/.env --env-file release/compose/.env.images -f release/compose/docker-compose.yml config --quiet`; `git diff --check`.
- Follow-ups: none.
- Canonical documents updated: `docs/ARCHITECTURE.md`, `docs/REBUILD.md`, `docs/specs/snapshotter/PLAN.md`, `docs/specs/snapshotter/tasks/SNAP-008-release.md`, `docs/specs/snapshotter/tasks/SNAP-009-integration.md`, and `docs/specs/snapshotter/plans/SNAP-008-release.execplan.md`.

## 2026-06-03: Snapshotter Work Directory And Logging Follow-Up

- Status outcome: backup workspace and config logging review findings corrected.
- Summary of changes: updated Compose and env examples to stage Snapshotter work under disk-backed `/work/archivist-snapshotter`, added `snapshotter-work` Compose volumes, kept `/tmp` as tmpfs, and included safe `ConfigError` messages in JSON logs.
- Decisions made: the executable default remains `/tmp/archivist-snapshotter`; Compose deployments explicitly override the work directory to `/work/archivist-snapshotter` to avoid memory-backed backup staging.
- Validation performed: `uv run ruff format --check .`; `uv run ruff check .`; `uv run ty check .`; `uv run pytest`; `docker compose --env-file .env.example config --quiet`; `git diff --check`.
- Follow-ups: none.
- Canonical documents updated: `docs/ARCHITECTURE.md` and `docs/specs/snapshotter/SPEC.md`.
