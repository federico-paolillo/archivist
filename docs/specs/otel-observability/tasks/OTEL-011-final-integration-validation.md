---
id: OTEL-011
feature: otel-observability
title: Final integration and manual validation record
depends_on:
  - OTEL-002
  - OTEL-004
  - OTEL-006
  - OTEL-008
  - OTEL-009
blocks: []
parallel: false
requires_exec_plan: false
canonical: true
---
# OTEL-011: Final Integration And Manual Validation Record

## Objective

Run the final OpenTelemetry integration checks after deployment wiring, Gateway injection, Worker continuation, Snapshotter spans, and README documentation are complete.

This task owns recording Grafana LGTM/manual validation results or explicit reasons those checks could not run.

## Done When

- Local Compose starts with Collector and LGTM configuration valid.
- Gateway-originated article processing shows Worker continuation from the persisted carrier.
- Worker CLI-enqueued jobs produce valid root traces when no parent carrier exists.
- Snapshotter produces independent snapshot attempt traces.
- Logs emitted inside spans include `trace_id` and `span_id`.
- High-cardinality values remain trace/log attributes and are not promoted to Loki labels or metric labels.
- Collector outage is checked as non-fatal for core Gateway, Worker, and Snapshotter behavior.
- LGTM/manual validation results are recorded here, or the reason they could not run is recorded here.

## Validation

Required checks:

```bash
docker compose --env-file .env.local.example -f docker-compose.yaml -f docker-compose.local.yaml config --quiet
git diff --check
```

Manual checks are the canonical expectations listed in `../SPEC.md`.

## Required Context

- `../SPEC.md`
- `../PLAN.md`

## Open Questions

- None.
