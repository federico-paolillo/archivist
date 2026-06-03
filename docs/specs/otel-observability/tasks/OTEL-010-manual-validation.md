---
id: OTEL-010
feature: otel-observability
title: Document manual OTEL validation
status: done
depends_on:
  - OTEL-006
  - OTEL-008
  - OTEL-009
blocks: []
parallel: false
exec_plan: null
canonical: true
---

# OTEL-010: Document manual OTEL validation

## Objective

Record manual Grafana/LGTM validation steps without adding brittle automated span-shape tests.

## Done When

- `ROADMAP.md` contains local manual validation steps.
- README contains deployment validation steps.

## Validation

Required checks:

```bash
git diff --check
```

Result: passed on 2026-06-03. Manual runtime validation steps are documented with a local LGTM backend override. Manual runtime validation and Docker Compose config validation were not executed in this environment because the `docker` CLI is unavailable.
