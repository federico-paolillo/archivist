---
id: OTEL-012
feature: otel-observability
title: De-noise Worker heartbeat logs and filter debug OTLP logs
status: done
depends_on:
  - OTEL-011
blocks: []
parallel: false
exec_plan: null
canonical: true
---

# OTEL-012: De-noise Worker Heartbeat Logs And Filter Debug OTLP Logs

## Objective

Demote Worker process-loop heartbeat logs to debug level and make application OTLP log export consistently ignore debug-level records across Gateway, Worker, and Snapshotter.

## Scope

This task includes:

- Worker process-loop iteration start logs.
- Worker idle/no-job process-loop poll-result logs.
- Worker central `slog` OTLP handler filtering at `slog.LevelInfo`.
- Gateway central `OpenTelemetryLoggerProvider` filter at `LogLevel.Information`.
- Snapshotter central JSON logger OTLP emission filtering for info and higher only.
- Focused regression tests for affected logging behavior.

This task does not include:

- Worker pipeline or stage log level changes for claimed jobs and article-processing work.
- Queue state, retry, SQL claim, process-loop sleep, CLI flag, schema, public API, Compose, Collector, or Grafana changes.

## Acceptance Criteria

```gherkin
Scenario: Worker heartbeat logs are debug-only
  Given the Worker polls the queue and no job is claimable
  When debug logging is enabled locally
  Then process-loop start and idle poll-result logs are emitted at debug level
  When only info-level logs are enabled locally
  Then those process-loop heartbeat logs are omitted

Scenario: Debug application logs are not exported through OTLP
  Given Gateway, Worker, or Snapshotter emits a debug-level application log
  When application OTLP log export is configured
  Then the debug-level log is not exported through OTLP
  And info-level and higher application logs remain exportable
```

## Done When

- Worker process-loop heartbeat logs are debug-level.
- Worker stdout runtime log level behavior remains unchanged.
- Worker OTLP `slog` export drops debug records while retaining info and error records.
- Gateway registers a central `OpenTelemetryLoggerProvider` information-level filter.
- Snapshotter debug logs, if emitted locally, do not call the OTEL log emitter.
- Regression tests cover the changed behavior.

## Validation

Required checks:

```bash
cd src/worker && go test ./internal/app ./internal/observability
cd src/gateway && dotnet test --filter OpenTelemetryExtensionsTest
cd src/snapshotter && uv run pytest tests/test_logging.py tests/test_telemetry.py
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
cd src/gateway && dotnet format --verify-no-changes && dotnet build && dotnet test
cd src/snapshotter && uv run ruff format --check . && uv run ruff check . && uv run ty check . && uv run pytest
git diff --check
```

Result: all required checks passed on 2026-06-12.
