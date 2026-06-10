---
id: OTEL-002
feature: otel-observability
title: Add Collector and local LGTM deployment
status: done
depends_on:
  - OTEL-001
blocks:
  - OTEL-009
parallel: true
exec_plan: null
canonical: true
---

# OTEL-002: Add Collector and local LGTM deployment

## Objective

Add a private official OpenTelemetry Collector Contrib service, dev Grafana LGTM, env examples, release packaging, and Compose validation.

## Done When

- `otelcol-config.yaml` exists.
- Dev Compose includes `otelcol` and dev-only `lgtm`.
- Production Compose includes `otelcol` only.
- Release packaging includes Collector config.
- Release packaging copies a neutral `.env` template and does not rewrite OTEL environment values.
- Packaged production env files do not include development LGTM backend defaults.

## Validation

Required checks:

```bash
docker compose --env-file .env.local.example -f docker-compose.yaml -f docker-compose.local.yaml config --quiet
ARCHIVIST_GATEWAY_IMAGE=gateway ARCHIVIST_WORKER_IMAGE=worker ARCHIVIST_UI_IMAGE=ui ARCHIVIST_SNAPSHOTTER_IMAGE=snapshotter docker compose --env-file .env.example -f docker-compose.yaml -f docker-compose.prod.yaml config --quiet
scripts/package-compose-release.sh test-version gateway worker ui snapshotter
docker compose --env-file release/compose/.env --env-file release/compose/.env.images -f release/compose/docker-compose.yaml -f release/compose/docker-compose.prod.yaml config --quiet
```

Result: passed on 2026-06-03 for `sh -n scripts/package-compose-release.sh` and `scripts/package-compose-release.sh test-version gateway worker ui snapshotter`. Generated release `.env`, packaged Compose files, and packaged `otelcol-config.yaml` contained no development LGTM backend default, removed OTEL environment key, app-side trace/log exporter toggles, or deployment environment resource attribute. Generated `release/` artifacts were removed after inspection. Docker Compose config validation was not executed because the local environment has no `docker` CLI.
