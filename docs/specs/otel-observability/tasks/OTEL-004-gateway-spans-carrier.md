---
id: OTEL-004
feature: otel-observability
title: Add Gateway spans and job carrier injection
depends_on:
  - OTEL-003
  - OTEL-010
blocks:
  - OTEL-009
  - OTEL-011
parallel: false
requires_exec_plan: false
canonical: true
---
# OTEL-004: Add Gateway spans and job carrier injection

## Objective

Add Gateway custom spans and inject W3C trace context into the shared queued-job carrier fields owned by `OTEL-010`.

## Done When

- Telegram token-bearing HTTP paths are not emitted in spans.
- Accepted Telegram ingestion writes `traceparent` and `tracestate` through the shared repository contract.
- Gateway spans include article/job attributes when available.

## Validation

Required checks:

```bash
cd src/gateway && dotnet build && dotnet test
```
