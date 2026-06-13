---
id: OTEL-005
feature: otel-observability
title: Add Worker OTEL foundation
status: done
depends_on:
blocks:
  - OTEL-006
parallel: true
exec_plan: null
canonical: true
---

# OTEL-005: Add Worker OTEL foundation

## Objective

Add Worker OpenTelemetry SDK setup, OTLP traces/logs, W3C propagator, HTTP instrumentation, and trace-aware logs.

## Done When

- Worker uses `archivist-worker` resource identity.
- Worker logger emits trace context attributes when context contains a span.
- Worker remains functional if the Collector is unreachable at runtime.

## Validation

Required checks:

```bash
cd src/worker && go tool lefthook run build && go tool lefthook run format && go tool lefthook run lint && go tool lefthook run test
```
