---
feature: otel-observability
canonical: true
---
# Feature Plan: OpenTelemetry Observability

## Purpose

This file controls implementation of OpenTelemetry traces/logs, async trace context propagation, Collector deployment, local Grafana LGTM validation, and deployment documentation.

## Task DAG

```text
OTEL-002 -> OTEL-009
OTEL-002 -> OTEL-011
OTEL-003 -> OTEL-004
OTEL-010 -> OTEL-004
OTEL-010 -> OTEL-006
OTEL-005 -> OTEL-006
OTEL-007 -> OTEL-008
OTEL-004 -> OTEL-009
OTEL-006 -> OTEL-009
OTEL-008 -> OTEL-009
OTEL-004 -> OTEL-011
OTEL-006 -> OTEL-011
OTEL-008 -> OTEL-011
OTEL-009 -> OTEL-011
```

## Tasks

| ID | Task | Depends On | Blocks | Parallel | Requires ExecPlan |
|---|---|---|---|---|---|
| `OTEL-002` | Add Collector and local LGTM deployment | - | `OTEL-009`, `OTEL-011` | yes | no |
| `OTEL-003` | Add Gateway OTEL foundation | - | `OTEL-004` | yes | no |
| `OTEL-010` | Add shared jobs trace carrier contract | - | `OTEL-004`, `OTEL-006` | no | no |
| `OTEL-004` | Add Gateway spans and job carrier injection | `OTEL-003`, `OTEL-010` | `OTEL-009`, `OTEL-011` | no | no |
| `OTEL-005` | Add Worker OTEL foundation | - | `OTEL-006` | yes | no |
| `OTEL-006` | Add Worker async continuation and pipeline spans | `OTEL-005`, `OTEL-010` | `OTEL-009`, `OTEL-011` | no | no |
| `OTEL-007` | Add Snapshotter OTEL foundation | - | `OTEL-008` | yes | no |
| `OTEL-008` | Add Snapshotter backup and upload spans | `OTEL-007` | `OTEL-009`, `OTEL-011` | no | no |
| `OTEL-009` | Update README deployment instructions | `OTEL-002`, `OTEL-004`, `OTEL-006`, `OTEL-008` | `OTEL-011` | no | no |
| `OTEL-011` | Final integration and manual validation record | `OTEL-002`, `OTEL-004`, `OTEL-006`, `OTEL-008`, `OTEL-009` | - | no | no |

## Concurrency Rules

- The SQLite carrier schema and Gateway/Worker queue code must be sequenced because both modules share the same `jobs` contract.
- Release packaging and README deployment instructions depend on the final Compose shape.
- `OTEL-010` owns the shared jobs carrier schema and repository contract so Gateway injection and Worker extraction can depend on the same durable contract without either task redefining it.
- `OTEL-011` runs last and records LGTM/manual validation results or explicit reasons those checks could not run.

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

The feature is complete when Gateway, Worker, Snapshotter, Compose, release packaging, and documentation changes are implemented; automated validation requirements pass; and manual validation expectations are documented in canonical feature notes.
