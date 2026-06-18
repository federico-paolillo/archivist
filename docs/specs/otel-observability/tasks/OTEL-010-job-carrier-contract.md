---
id: OTEL-010
feature: otel-observability
title: Add shared jobs trace carrier contract
depends_on: []
blocks:
  - OTEL-004
  - OTEL-006
parallel: false
requires_exec_plan: false
canonical: true
---
# OTEL-010: Add Shared Jobs Trace Carrier Contract

## Objective

Add the shared SQLite and repository contract that carries W3C trace context across the Gateway-to-Worker asynchronous job boundary.

This task owns the carrier schema and repository surface. Gateway span work consumes the contract to inject carrier values; Worker span work consumes the same contract to extract carrier values.

## Done When

- The `jobs` table has nullable `traceparent` and `tracestate` columns.
- Existing databases are upgraded through an idempotent schema migration.
- Gateway and Worker repository DTOs expose the carrier fields without making them user-visible telemetry history.
- Existing database upgrade tests cover the carrier-column migration path.

## Validation

Required checks:

```bash
cd src/gateway && dotnet build && dotnet test
cd src/worker && go tool lefthook run test
```

## Required Context

- `../SPEC.md`
- `../PLAN.md`
- `../../ARCHITECTURE.md`

## Open Questions

- None.
