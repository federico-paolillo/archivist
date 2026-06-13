---
feature: otel-observability
status: done
canonical: true
---

# Feature Plan: OpenTelemetry Observability

## Purpose

This file controls implementation of OpenTelemetry traces/logs, async trace context propagation, Collector deployment, local Grafana LGTM validation, and deployment documentation.

## Task DAG

No remaining intra-feature task edges are required beyond the task table and cross-feature dependencies.

## Tasks

| ID | Task | Status | Depends On | Blocks | Parallel | ExecPlan |
|---|---|---|---|---|---|---|
| `OTEL-002` | Add Collector and local LGTM deployment | done | - | `OTEL-009` | yes | - |
| `OTEL-003` | Add Gateway OTEL foundation | done | - | `OTEL-004` | yes | - |
| `OTEL-004` | Add Gateway spans and job carrier injection | done | `OTEL-003` | `OTEL-006` | no | - |
| `OTEL-005` | Add Worker OTEL foundation | done | - | `OTEL-006` | yes | - |
| `OTEL-006` | Add Worker async continuation and pipeline spans | done | `OTEL-004`, `OTEL-005` | - | no | - |
| `OTEL-007` | Add Snapshotter OTEL foundation | done | - | `OTEL-008` | yes | - |
| `OTEL-008` | Add Snapshotter backup and upload spans | done | `OTEL-007` | - | no | - |
| `OTEL-009` | Update README deployment instructions | done | `OTEL-002` | - | yes | - |

## Concurrency Rules

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

Manual validation expectations are defined in `SPEC.md`. README may summarize the operator workflow but does not define canonical validation behavior.

## Completion Criteria

The feature is complete when Gateway, Worker, Snapshotter, Compose, release packaging, and documentation changes are implemented; automated validation passes or failures are recorded in the task being executed; and manual validation expectations are documented in canonical feature notes.
