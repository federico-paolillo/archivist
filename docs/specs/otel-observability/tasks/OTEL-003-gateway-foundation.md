---
id: OTEL-003
feature: otel-observability
title: Add Gateway OTEL foundation
status: done
depends_on:
  - OTEL-001
blocks:
  - OTEL-004
parallel: true
exec_plan: null
canonical: true
---

# OTEL-003: Add Gateway OTEL foundation

## Objective

Register OpenTelemetry SDK services for Gateway traces and logs while preserving console logging.

## Done When

- Gateway exports OTLP traces and logs when configured.
- ASP.NET Core and outbound HTTP instrumentation are registered.
- Gateway uses `archivist-gateway` resource identity.

## Validation

Required checks:

```bash
cd src/gateway && dotnet format --verify-no-changes && dotnet build && dotnet test
```

Result: passed on 2026-06-03.
