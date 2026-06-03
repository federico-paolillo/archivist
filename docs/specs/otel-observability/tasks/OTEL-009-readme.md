---
id: OTEL-009
feature: otel-observability
title: Update README deployment instructions
status: done
depends_on:
  - OTEL-002
blocks:
  - OTEL-010
parallel: true
exec_plan: null
canonical: true
---

# OTEL-009: Update README deployment instructions

## Objective

Document local and production OTEL deployment, configuration, tail sampling, Collector outage behavior, and manual validation.

## Done When

- README documents dev Grafana LGTM access.
- README documents production Collector/backend configuration.
- README documents that application telemetry is always configured and no app-side trace/log disable switches are exposed.
- README documents manual trace/log correlation validation.

## Validation

Required checks:

```bash
git diff --check
```

Result: passed on 2026-06-03. README now documents always-configured application telemetry, local LGTM validation setup, production Collector backend configuration, and Collector runtime outage behavior without app-side trace/log exporter toggles or deployment environment values.
