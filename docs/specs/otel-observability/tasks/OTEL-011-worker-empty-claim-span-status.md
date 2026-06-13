---
id: OTEL-011
feature: otel-observability
title: De-noise empty Worker queue claim spans
status: done
depends_on:
  - OTEL-006
blocks: []
parallel: false
exec_plan: null
canonical: true
---

# OTEL-011: De-noise Empty Worker Queue Claim Spans

## Objective

Treat an empty Worker queue claim as a normal idle poll in tracing. The Worker repository may still return `sql.ErrNoRows` for no claimable job, and the pipeline must still return `processed=false, err=nil`, but claim spans must not be marked as `ERROR`.

## Scope

This task includes:

- `worker.jobs.claim` span status for empty queue claims.
- `worker.pipeline.claim` span status for empty queue polls.
- Focused regression tests for both spans.

This task does not include:

- Job state changes.
- Retry, requeue, or backoff behavior.
- Claim SQL changes.
- Process-loop sleep, log, CLI, schema, public API, Collector, or Grafana changes.

## Acceptance Criteria

```gherkin
Scenario: Empty Worker queue poll is not an error trace
  Given no queued article-processing job is claimable
  When the Worker polls for one job
  Then the jobs repository still reports sql.ErrNoRows to its caller
  And the processing pipeline returns processed=false with no error
  And worker.jobs.claim and worker.pipeline.claim spans are not marked as ERROR
```

## Done When

- Empty claim span status does not become `ERROR`.
- Non-empty and unexpected claim errors keep existing behavior.
- Regression tests cover repository and pipeline empty-queue claim spans.

## Validation

Required checks:

```bash
cd src/worker && go test ./internal/pipeline ./pkg/jobs
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

Result: all required checks passed on 2026-06-12.
