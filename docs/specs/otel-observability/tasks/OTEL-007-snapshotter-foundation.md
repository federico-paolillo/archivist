---
id: OTEL-007
feature: otel-observability
title: Add Snapshotter OTEL foundation
status: done
depends_on:
blocks:
  - OTEL-008
parallel: true
exec_plan: null
canonical: true
---

# OTEL-007: Add Snapshotter OTEL foundation

## Objective

Add Python OpenTelemetry SDK setup, OTLP traces/logs, trace-aware JSON logs, and botocore instrumentation.

## Done When

- Snapshotter uses `archivist-snapshotter` resource identity.
- Snapshotter logs include `trace_id` and `span_id` when a span is active.
- Existing redaction behavior remains intact.

## Validation

Required checks:

```bash
cd src/snapshotter && uv run ruff format --check . && uv run ruff check . && uv run ty check . && uv run pytest
```
