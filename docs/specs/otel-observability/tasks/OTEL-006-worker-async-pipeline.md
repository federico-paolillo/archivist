---
id: OTEL-006
feature: otel-observability
title: Add Worker async continuation and pipeline spans
status: done
depends_on:
  - OTEL-004
  - OTEL-005
blocks:
parallel: false
exec_plan: null
canonical: true
---

# OTEL-006: Add Worker async continuation and pipeline spans

## Objective

Continue Gateway-originated traces from persisted job carriers and add fine-grained Worker pipeline spans.

## Done When

- Worker extracts `traceparent` and `tracestate` when present.
- Worker starts valid root or consumer traces for jobs with no parent carrier.
- Worker spans cover claim, fetch, artifact writes, Markdown/Jina, Anthropic, DB updates, and terminal persistence.

## Validation

Required checks:

```bash
cd src/worker && go tool lefthook run build && go tool lefthook run test
```
