---
feature: otel-observability
status: done
canonical: true
---

# Feature Plan: OpenTelemetry Observability

## Purpose

This file controls implementation of OpenTelemetry traces/logs, async trace context propagation, Collector deployment, local Grafana LGTM validation, and deployment documentation.

## Task DAG

```text
OTEL-001 -> OTEL-002
OTEL-001 -> OTEL-003 -> OTEL-004
OTEL-001 -> OTEL-005 -> OTEL-006
OTEL-001 -> OTEL-007 -> OTEL-008
OTEL-002 -> OTEL-009 -> OTEL-010
OTEL-004 -> OTEL-006 -> OTEL-010
OTEL-008 -> OTEL-010
OTEL-006 -> OTEL-011
OTEL-011 -> OTEL-012
```

## Tasks

| ID | Task | Status | Depends On | Blocks | Parallel | ExecPlan |
|---|---|---|---|---|---|---|
| `OTEL-001` | Create roadmap and canonical docs | done | - | `OTEL-002`, `OTEL-003`, `OTEL-005`, `OTEL-007` | no | - |
| `OTEL-002` | Add Collector and local LGTM deployment | done | `OTEL-001` | `OTEL-009` | yes | - |
| `OTEL-003` | Add Gateway OTEL foundation | done | `OTEL-001` | `OTEL-004` | yes | - |
| `OTEL-004` | Add Gateway spans and job carrier injection | done | `OTEL-003` | `OTEL-006` | no | - |
| `OTEL-005` | Add Worker OTEL foundation | done | `OTEL-001` | `OTEL-006` | yes | - |
| `OTEL-006` | Add Worker async continuation and pipeline spans | done | `OTEL-004`, `OTEL-005` | `OTEL-010` | no | - |
| `OTEL-007` | Add Snapshotter OTEL foundation | done | `OTEL-001` | `OTEL-008` | yes | - |
| `OTEL-008` | Add Snapshotter backup and upload spans | done | `OTEL-007` | `OTEL-010` | no | - |
| `OTEL-009` | Update README deployment instructions | done | `OTEL-002` | `OTEL-010` | yes | - |
| `OTEL-010` | Document manual OTEL validation | done | `OTEL-006`, `OTEL-008`, `OTEL-009` | - | no | - |
| `OTEL-011` | De-noise empty Worker queue claim spans | done | `OTEL-006` | - | no | - |
| `OTEL-012` | De-noise Worker heartbeat logs and filter debug OTLP logs | done | `OTEL-011` | - | no | - |

## Concurrency Rules

- Gateway, Worker, Snapshotter, and deployment slices can be developed in parallel after `OTEL-001`.
- The SQLite carrier schema and Gateway/Worker queue code must be sequenced because both modules share the same `jobs` contract.
- Release packaging and README deployment instructions depend on the final Compose shape.

## Blocking Interfaces or Schemas

- `jobs.traceparent` and `jobs.tracestate`.
- Private `otelcol` OTLP HTTP endpoint.
- Always-configured application traces/logs with no app-side trace/log exporter disable switches.
- Log correlation fields `trace_id` and `span_id`.
- Attribute naming for `article_id`, `job_id`, stage, provider, model, request id, and ARC code.

## Validation Sequence

Validation commands:

```bash
cd src/gateway && dotnet format --verify-no-changes && dotnet build && dotnet test
cd src/worker && go tool lefthook run build && go tool lefthook run format && go tool lefthook run lint && go tool lefthook run test
cd src/snapshotter && uv run ruff format --check . && uv run ruff check . && uv run ty check . && uv run pytest
docker compose --env-file .env.local.example -f docker-compose.yaml -f docker-compose.local.yaml config --quiet
ARCHIVIST_GATEWAY_IMAGE=gateway ARCHIVIST_WORKER_IMAGE=worker ARCHIVIST_UI_IMAGE=ui ARCHIVIST_SNAPSHOTTER_IMAGE=snapshotter docker compose --env-file .env.example -f docker-compose.yaml -f docker-compose.prod.yaml config --quiet
scripts/package-compose-release.sh test-version gateway worker ui snapshotter
docker compose --env-file release/compose/.env --env-file release/compose/.env.images -f release/compose/docker-compose.yaml -f release/compose/docker-compose.prod.yaml config --quiet
git diff --check
```

Manual validation is described in `ROADMAP.md` and `README.md`.

Docker-based Compose validation was not executed in this environment because the `docker` CLI is unavailable. Release packaging and focused non-Docker validation passed. The generated production release `.env`, Compose files, and `otelcol-config.yaml` were inspected for no development LGTM backend default, removed OTEL environment key, app-side trace/log exporter toggles, or deployment environment resource attribute, then `release/` was removed.

## Completion Criteria

The feature is complete when Gateway, Worker, Snapshotter, Compose, release packaging, and documentation changes are implemented; automated validation passes or failures are recorded; and manual validation steps are documented.
