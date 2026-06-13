---
id: OTEL-004
feature: otel-observability
title: Add Gateway spans and job carrier injection
status: done
depends_on:
  - OTEL-003
blocks:
  - OTEL-006
parallel: false
exec_plan: null
canonical: true
---

# OTEL-004: Add Gateway spans and job carrier injection

## Objective

Add Gateway custom spans and persist W3C trace context on queued jobs.

## Done When

- Telegram token-bearing HTTP paths are not emitted in spans.
- Accepted Telegram ingestion persists `traceparent` and `tracestate`.
- Gateway spans include article/job attributes when available.

## Validation

Required checks:

```bash
cd src/gateway && dotnet build && dotnet test
```
