---
id: OTEL-006
feature: otel-observability
title: Add Worker async continuation and pipeline spans
depends_on:
  - OTEL-005
  - OTEL-010
blocks:
  - OTEL-009
  - OTEL-011
parallel: false
requires_exec_plan: false
canonical: true
---
# OTEL-006: Add Worker async continuation and pipeline spans

## Objective

Continue traces from persisted job carriers defined by `OTEL-010` and add fine-grained Worker pipeline spans.

## Done When

- Worker extracts `traceparent` and `tracestate` when present.
- Worker starts valid root or consumer traces for jobs with no parent carrier.
- Worker spans cover claim, fetch, artifact writes, Markdown/Jina, Anthropic, DB updates, and terminal persistence.

## Validation

Required checks:

```bash
cd src/worker && go tool lefthook run build && go tool lefthook run test
```
