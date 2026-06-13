# Implementation Diary: OpenTelemetry Observability

This file is an append-only historical log. It records implementation outcomes, validation, decisions, and follow-ups.

It is not the canonical source of requirements. Durable decisions must be promoted to canonical documents.

---

## Log

## 2026-06-03 - OTEL-001..OTEL-010 - done

Summary:

- Created `ROADMAP.md` and canonical `docs/specs/otel-observability/` artifacts.
- Added official OpenTelemetry Collector Contrib configuration with OTLP HTTP, memory limiter, batch processing, and tail sampling that keeps all error traces and 10% of non-error traces.
- Extended dev Compose with Archivist `otelcol` and dev-only Grafana LGTM; extended production Compose and release packaging with Archivist `otelcol`.
- Added Gateway OpenTelemetry foundation, ASP.NET Core and HTTP instrumentation, OTLP traces/logs, trace-aware logging, custom ingestion spans, token-path sanitization, and W3C job carrier injection.
- Added nullable `jobs.traceparent` and `jobs.tracestate` carrier fields with idempotent Gateway and Worker schema upgrades.
- Added Worker OpenTelemetry foundation, W3C extraction/injection, OTLP traces/logs, trace-aware slog output, HTTP instrumentation, Gateway-to-Worker continuation, no-parent CLI support, and fine-grained pipeline spans.
- Added Snapshotter OpenTelemetry foundation, OTLP traces/logs, trace-aware JSON logging, botocore instrumentation, and spans for daemon attempts, archive capture, SQLite backup, artifact copy, manifest write, tarball creation, S3 upload, and cleanup.
- Updated `README.md`, `docs/DESIGN.md`, `docs/ARCHITECTURE.md`, `docs/REBUILD.md`, and `docs/specs/INDEX.md`.

Multi-agent execution:

- Gateway implementation ran in worker agent `019e8e09-01e8-7fd3-93f5-c29dc598f9d6`.
- Worker implementation ran in worker agent `019e8e09-3310-77b3-aebc-8a37dcf2537b`.
- Snapshotter implementation ran in worker agent `019e8e09-63f9-7673-b5b7-e25d8ad0e49b`.
- Coordinator owned roadmap/spec/plan/tasks/diary, Collector, Compose, release packaging, README, and final integration consistency.

Decisions:

- High-cardinality values remain attributes only; they are not Loki labels or metric labels.
- Application SDKs use standard OTEL behavior and continue core work during runtime Collector outages.
- Invalid enabled OTEL configuration may fail startup.
- Worker trace continuation uses persisted W3C carrier fields only; no custom propagation protocol was added.
- Snapshotter backup traces are independent timer-driven root traces.

Validation:

- `cd src/gateway && dotnet format --verify-no-changes && dotnet build && dotnet test` passed with 182 tests.
- `cd src/worker && go test ./...` passed. Worker agent also reported `go tool lefthook run build`, `go tool lefthook run test`, and `git diff --check -- src/worker` passed.
- `cd src/snapshotter && uv run ruff format --check . && uv run ruff check . && uv run ty check . && uv run pytest` passed with 18 tests.
- `scripts/package-compose-release.sh test-version gateway-image worker-image ui-image snapshotter-image` passed and included `otelcol-config.yaml`.
- `otelcol-config.yaml`, `docker-compose.yaml`, and `docker-compose.prod.yaml` parsed as YAML.
- `git diff --check` passed.
- `docker compose --env-file .env.example config --quiet` and packaged production Compose validation were not executed because the local environment has no `docker` CLI.

Follow-ups:

- Run the documented manual Grafana/LGTM validation on a host with Docker.

## 2026-06-03 - OTEL review remediation - done

Summary:

- Ran multi-agent review passes for Gateway, Worker, and Snapshotter/deployment/ALM OTEL changes.
- Gateway review approved without findings.
- Fixed deployment/ALM review findings in release packaging and `docs/DESIGN.md`.
- Fixed Worker review findings in Worker trace-carrier behavior, slog tee filtering, and Go lint compliance.

Decisions:

- Packaged production Compose env files must not inherit the development LGTM endpoint. Release packaging uses `<specify>` for `ARCHIVIST_OTEL_EXPORTER_OTLP_ENDPOINT`.
- Worker direct CLI enqueue leaves `jobs.traceparent` and `jobs.tracestate` null so processing starts from a root trace, as specified.
- `docs/DESIGN.md` keeps OTEL as `DSGN-019`, after the existing stale-job and snapshotter decisions.

Validation:

- Gateway reviewer reported `dotnet build`, `dotnet test`, and `dotnet format --verify-no-changes --no-restore` passed.
- Coordinator ran `cd src/gateway && dotnet format --verify-no-changes && dotnet build && dotnet test`; passed with 182 tests.
- Worker reviewer initially requested changes, then re-reviewed fixes and approved.
- Coordinator ran `cd src/worker && go tool lefthook run build && go tool lefthook run format && go tool lefthook run lint && go tool lefthook run test`; passed.
- Snapshotter/deployment/ALM reviewer initially requested changes, then re-reviewed fixes and approved.
- Coordinator ran `scripts/package-compose-release.sh test-version gateway worker ui snapshotter`; passed. Generated release `.env` and tarball contain `ARCHIVIST_OTEL_EXPORTER_OTLP_ENDPOINT=<specify>` and no development LGTM backend endpoint.
- `sh -n scripts/package-compose-release.sh`, `git diff --check`, and duplicate design-id checks passed.
- Docker Compose config validation was not executed because the local environment has no `docker` CLI.

Follow-ups:

- Run Docker Compose config validation and documented manual Grafana/LGTM validation on a host with Docker.

## 2026-06-12 - OTEL-011 empty Worker queue claim span de-noising - done

Summary:

- Added canonical task `OTEL-011` for treating empty Worker queue polls as normal idle telemetry.
- Updated the OTEL spec so empty queued-job claims must not mark Worker claim spans as `ERROR`.
- Changed Worker job-claim span finalization so `sql.ErrNoRows` ends `worker.jobs.claim` without recording an error.
- Changed pipeline claim span finalization so an empty claim ends `worker.pipeline.claim` cleanly while preserving `processed=false, err=nil`.
- Added span-recorder regression tests for empty repository claims and empty pipeline polls.

Decisions:

- `sql.ErrNoRows` from the claim query remains the repository contract for no claimable job.
- Empty queue polls are not infrastructure failures and must not drive tail-sampled error traces.
- Unexpected claim errors still record error spans and still return as pipeline failures.
- No claim SQL, job state, retry, requeue, sleep, logging, schema, API, Collector, or Grafana behavior changed.

Validation:

- `cd src/worker && go test ./internal/pipeline ./pkg/jobs` passed.
- `cd src/worker && go tool lefthook run build` passed.
- `cd src/worker && go tool lefthook run format` passed.
- `cd src/worker && go tool lefthook run lint` passed.
- `cd src/worker && go tool lefthook run test` passed.

Follow-ups:

- None.

## 2026-06-03 - OTEL user-feedback remediation - done

Summary:

- Addressed user review findings with Worker, Gateway, Snapshotter, and deployment/docs agents.
- Removed application-side trace/log exporter toggles from module setup and deployment surfaces.
- Switched Worker outbound HTTP tracing from direct `http.Client.Transport` mutation to req `WrapRoundTripFunc`.
- Removed the Worker direct `otelhttp` dependency while retaining useful OpenTelemetry contrib usage such as `otelslog`.
- Kept `otelslog` as the Worker OTLP log bridge and removed redundant trace/span attribute injection around the OTLP handler; stdout JSON logs still use a small trace-context wrapper because `otelslog` does not replace stdout logging.
- Kept release packaging simple: `scripts/package-compose-release.sh` directly copies `.env.example`, `docker-compose.prod.yaml`, `rp.Caddyfile`, and `otelcol-config.yaml`.
- Fixed final review nits: Snapshotter test no longer asserts `OTEL_SDK_DISABLED` is ignored, and `docs/ARCHITECTURE.md` explicitly lists `otelcol-config.yaml` in release package contents.

Decisions:

- Accepted: this repo has no useful deployment environment dimension, so no `deployment.environment` resource attribute or `ARCHIVIST_OTEL_ENVIRONMENT` key is part of the contract.
- Accepted: Archivist does not expose app-side trace/log exporter switches in Compose or canonical docs. Telemetry is configured by default; Collector runtime outage remains non-fatal.
- Accepted: Worker should use req's round-trip wrapper surface for outbound HTTP tracing.
- Clarified: Worker already used OpenTelemetry Go contrib modules; the useful correction was replacing the transport mutation with req middleware, not adopting contrib generally.
- Rejected: product code should not neutralize the standard Python SDK `OTEL_SDK_DISABLED` behavior. The contract is that Archivist does not expose or document that app-side disable switch.

Validation:

- `cd src/gateway && dotnet format --verify-no-changes && dotnet build && dotnet test` passed with 183 tests.
- `cd src/worker && go tool lefthook run build && go tool lefthook run format && go tool lefthook run lint && go tool lefthook run test` passed.
- Worker reviewer approved after `go test ./...`, `go mod tidy -diff`, and `go tool lefthook run lint`.
- `cd src/snapshotter && uv run ruff format --check . && uv run ruff check . && uv run ty check . && uv run pytest` passed with 20 tests.
- Focused final checks passed: `dotnet test --filter OpenTelemetryExtensionsTest`, `uv run pytest tests/test_telemetry.py tests/test_logging.py`, `go mod tidy -diff`, `go tool lefthook run lint`, `sh -n scripts/package-compose-release.sh`, `scripts/package-compose-release.sh test-version gateway worker ui snapshotter`, `git diff --check`, and duplicate design-id check.
- Generated release artifacts were removed before finishing.
- Docker Compose config validation was not executed because the `docker` CLI is unavailable in this environment.

Follow-ups:

- Run Docker Compose config validation and documented manual Grafana/LGTM validation on a host with Docker.

## 2026-06-03 - OTEL deployment simplification - done

Summary:

- Removed the deployment environment dimension from service OTEL resource attributes, Collector processing, env templates, README, and canonical OTEL docs.
- Removed app-side trace/log exporter toggles from Compose, `.env.example`, README, and canonical docs.
- Kept application telemetry always configured in Compose through the private OTLP endpoint and fixed always-on trace sampling.
- Simplified `scripts/package-compose-release.sh` so it copies the neutral env template without OTEL-specific post-processing.
- Removed development LGTM backend defaults from the env template copied into production release packages.
- Removed generated `release/` artifacts after package inspection.

Decisions:

- This repo has one deployment environment, so no deployment environment resource attribute or environment env key is part of the rebuild contract.
- Compose does not expose application-side trace/log disable switches. Collector runtime outages remain non-fatal for Gateway, Worker, and Snapshotter core behavior.
- `.env.example` uses `<specify>` for the Collector exporter backend. Local LGTM validation sets that value to the dev `lgtm` service; production releases inherit the neutral placeholder.

Validation:

- `git diff --check` passed.
- `sh -n scripts/package-compose-release.sh` passed.
- `scripts/package-compose-release.sh test-version gateway worker ui snapshotter` passed.
- Generated release `.env`, packaged `docker-compose.yml`, and packaged `otelcol-config.yaml` were inspected for no development LGTM backend default, removed OTEL environment key, app-side trace/log exporter toggles, or deployment environment resource attribute.
- Generated `release/` artifacts were removed before finishing.
- Docker Compose config validation was not executed because the `docker` CLI is unavailable in this environment.

Follow-ups:

- Run Docker Compose config validation and documented manual Grafana/LGTM validation on a host with Docker.

## 2026-06-12 - OTEL-012 Worker heartbeat and debug OTLP log de-noising - done

Summary:

- Added canonical task `OTEL-012` for demoting Worker process-loop heartbeat logs and filtering debug logs from application OTLP export.
- Updated the OTEL spec and design decision so Gateway, Worker, and Snapshotter application OTLP log export is limited to `Info`/`Information` and higher.
- Updated job-recovery docs so Worker process-loop iteration start and idle/no-job poll-result logs are debug-level heartbeat diagnostics.
- Changed Worker process-loop start and idle poll-result logs from info to debug.
- Added a central Worker `slog` minimum-level handler and applied it only to the `otelslog` handler so stdout keeps existing runtime level behavior while OTLP export drops debug records.
- Added a Gateway central logging extension that filters `OpenTelemetryLoggerProvider` at `LogLevel.Information`.
- Added Snapshotter debug logging for local JSON output while keeping the central OTEL emission path limited to info and error levels.
- Added regression tests for Worker heartbeat log levels, Worker OTLP handler filtering, Gateway provider filtering, and Snapshotter debug-local/no-OTLP behavior.

Decisions:

- Worker job and article-processing pipeline logs with claim, stage, job, and article context remain info-level operational telemetry.
- Worker process-loop start and idle/no-job poll-result logs are debug-level heartbeat diagnostics.
- Debug logs are local diagnostics only for OTLP purposes across Gateway, Worker, and Snapshotter.
- No queue state, retry behavior, SQL claim semantics, process-loop sleep behavior, public APIs, schema, CLI flags, Compose, Collector, or Grafana configuration changed.

Validation:

- `cd src/worker && go test ./internal/app ./internal/observability` passed.
- `cd src/gateway && dotnet test --filter OpenTelemetryExtensionsTest` passed with 2 tests.
- `cd src/snapshotter && uv run pytest tests/test_logging.py tests/test_telemetry.py` passed with 6 tests.
- `cd src/worker && go tool lefthook run build` passed.
- `cd src/worker && go tool lefthook run format` passed.
- `cd src/worker && go tool lefthook run lint` passed.
- `cd src/worker && go tool lefthook run test` passed.
- `cd src/gateway && dotnet format --verify-no-changes` passed after normalizing formatting with `dotnet format`.
- `cd src/gateway && dotnet build` passed.
- `cd src/gateway && dotnet test` passed with 209 tests.
- `cd src/snapshotter && uv run ruff format --check .` passed.
- `cd src/snapshotter && uv run ruff check .` passed.
- `cd src/snapshotter && uv run ty check .` passed.
- `cd src/snapshotter && uv run pytest` passed with 21 tests.

Follow-ups:

- None.
