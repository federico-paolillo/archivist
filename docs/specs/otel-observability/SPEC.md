---
id: OTEL
slug: otel-observability
title: OpenTelemetry Observability
status: done
owner: null
depends_on:
  - snapshotter
impacts:
  - gateway
  - worker
  - snapshotter
  - deployment
canonical: true
---

# Feature: OpenTelemetry Observability

## Intent

Add traces and logs for Gateway, Worker, and Snapshotter, with a private OpenTelemetry Collector Contrib service that tail-samples traces and exports telemetry to a Grafana-compatible backend.

## Motivation

Archivist needs cross-process diagnosis for URL ingestion, asynchronous article processing, LLM calls, artifact writes, backups, and operator-visible failures. `article_id` is the primary domain link, while W3C trace context links Gateway-originated work to Worker processing across the SQLite queue boundary.

## Scope

In scope:

- Gateway traces and OTLP logs.
- Worker traces and OTLP logs.
- Snapshotter traces and OTLP logs without user attribution.
- Logs correlated to traces with `trace_id` and `span_id`.
- Gateway-to-Worker W3C trace propagation through persisted job carrier fields.
- Private official OpenTelemetry Collector Contrib service.
- Tail sampling that retains all error traces and 10% of non-error traces.
- Development Grafana LGTM container for manual local validation.
- README deployment instructions.

## Out of Scope

Not included:

- Metrics.
- Production Grafana, Tempo, Loki, or Grafana Cloud provisioning.
- Promoting high-cardinality values to Loki labels or metric labels.
- Custom trace id generation or hand-rolled trace propagation.
- Snapshotter linking one backup trace to every historical article trace.

## Requirements

- REQ-001: Services must export traces and logs over OTLP HTTP to a private Collector by default in Compose.
- REQ-002: The Collector must use `otel/opentelemetry-collector-contrib`.
- REQ-003: The Collector must tail-sample traces, keeping all traces with error status and 10% of other traces.
- REQ-004: Runtime Collector outages must not stop Gateway, Worker, or Snapshotter core behavior.
- REQ-005: Logs emitted inside a trace context must carry `trace_id` and `span_id`.
- REQ-006: Gateway must persist W3C `traceparent` and `tracestate` when enqueueing article-processing jobs.
- REQ-007: Worker must extract persisted W3C trace context when present and support root traces when absent.
- REQ-008: Snapshotter must produce independent root traces for scheduled snapshot attempts.
- REQ-009: `user_id`, `article_id`, `job_id`, URLs, and request IDs must be trace/log attributes only, not Loki labels or metric labels.
- REQ-010: Telemetry must not expose secrets, cookies, auth headers, Telegram bot tokens, full article HTML, full Markdown, full summaries, LLM payloads, or S3 credentials.
- REQ-011: Use .NET `Activity`, Go/Python OpenTelemetry SDKs, and W3C Trace Context propagators before custom code.
- REQ-012: Compose must always configure application traces and logs; application-side trace/log exporter disable switches are not part of the deployment contract.
- REQ-013: Gateway and Worker must attach `user_id` to logs and spans when the Archivist user has been resolved.
- REQ-014: Snapshotter must not attach `user_id`.
- REQ-015: Gateway must selectively log security-relevant HTTP `401`/`403` responses and operational `5xx` responses without enabling broad request logging; routine unauthenticated `GET /auth/session` probes are excluded.
- REQ-016: Gateway must mark `5xx` request activities and caught operational failures that return `5xx` as `ERROR` so Collector tail sampling retains those traces.

## Acceptance Criteria

```gherkin
Feature: OpenTelemetry observability

Scenario: Gateway-originated article processing is trace-linked
  Given a Telegram webhook enqueues a valid article URL
  When the Worker claims and processes the queued job
  Then Worker processing spans continue from the Gateway-created trace context
  And logs for both services include trace ids and article/job attributes

Scenario: Worker CLI enqueue has no parent
  Given the Worker CLI enqueues an article directly
  When the Worker processes the queued job
  Then processing emits a valid trace without requiring a parent trace

Scenario: Collector outage is non-fatal
  Given the Collector is unreachable after services have started
  When Gateway, Worker, or Snapshotter performs normal work
  Then core application behavior continues
```

## Data and State

The `jobs` table gains nullable carrier columns:

```text
traceparent TEXT
tracestate TEXT
```

These fields are not user-visible telemetry history. They are queue message carrier fields used to continue distributed traces across the asynchronous SQLite handoff. Existing databases require an idempotent schema upgrade.

## Interfaces

- OTLP HTTP receiver: `http://otelcol:4318`.
- Dev Grafana: `http://localhost:40300`, username `admin`, password `admin`.
- Collector exporter backend: configured through env and not committed with secrets.
- Compose sets standard OTEL SDK env vars for service name, resource attributes, OTLP endpoint, and always-on trace sampling.
- Collector-specific env keys use `ARCHIVIST_OTEL_*`.

## Security / Privacy Notes

Telemetry must redact secrets and avoid content payloads. High-cardinality domain identifiers remain searchable attributes, not labels.

## Observability / Logging Notes

Gateway, Worker, and Snapshotter retain stdout logging behavior. OTLP logs are additive. Trace correlation fields are added to logs when a current span exists. Gateway does not emit broad request logs; it emits selective HTTP failure logs for security-relevant `401`/`403` responses and operational `5xx` responses.

## Related Documents

- `./PLAN.md`
- `./DIARY.md`
- `../../../ROADMAP.md`
