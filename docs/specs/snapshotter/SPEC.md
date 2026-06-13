---
id: SNAP
slug: snapshotter
title: Snapshotter
status: done
owner: null
depends_on:
  - summary-generation
  - ui-endpoints
impacts:
  - architecture
  - artifacts
  - deployment
  - ci-cd
canonical: true
---

# Feature: Snapshotter

## Intent

Add a fourth Archivist service that periodically archives `/data` and uploads the archive to an S3-compatible object storage bucket.

## Motivation

Archivist stores durable state in SQLite and retained article artifacts on the filesystem under `/data`. The deployment needs an automated backup mechanism that can preserve that data outside the VPS without introducing another queue, database, or orchestration service.

## Scope

In scope:

- Python 3.12 Snapshotter executable under `src/snapshotter`.
- `uv` project management with locked dependencies.
- `ruff` formatting/linting, `ty` type checking, and `pytest` tests.
- Periodic daemon behavior with a simple sleep interval.
- SQLite online backup into a staging directory.
- Best-effort copy of non-database artifact files from `/data`.
- Single `.tar.gz` archive upload to S3-compatible Object Storage.
- Rootless distroless Docker image.
- Docker Compose, env example, CI, CD, release packaging, and validation updates.
- Manual restore documentation in canonical backup notes.
- Gateway, Worker, and UI continue running during artifact copy; Snapshotter does not coordinate writers.
- Snapshotter backup behavior does not define metrics or alerting. Trace and log export for Snapshotter is governed by `otel-observability`.
- Remote snapshot retention/pruning is delegated to object-storage policy. Restore remains manual. Telegram/UI-visible backup status, automated destructive restore, and client-side archive encryption require separate canonical behavior before implementation.

## Users / Actors

- Operator: configures Scaleway Object Storage credentials, deploys the Snapshotter, and restores backups manually when needed.
- Snapshotter service: captures and uploads scheduled backups.

## Requirements

- REQ-001: Snapshotter must sleep first, then run a snapshot every configured interval.
- REQ-002: The default interval must be 86,400 seconds.
- REQ-003: Snapshotter must capture SQLite through the SQLite online backup API rather than raw-copying the live database file.
- REQ-004: Snapshotter must copy non-database `/data` files best-effort and must not require Gateway or Worker to pause writes.
- REQ-005: Snapshotter must omit live SQLite sidecar files for the configured database path from the archive.
- REQ-006: Snapshotter archives must be named `archivist-<yyyy-mm-dd>-<unix-timestamp>.tar.gz`.
- REQ-007: Snapshotter must include `manifest.json` at the archive root.
- REQ-008: Snapshotter must upload to an S3-compatible bucket using explicit endpoint URL, region, bucket, key, access key, and secret key configuration.
- REQ-009: Snapshotter must never delete remote objects.
- REQ-010: Snapshotter must log failures, clean temporary work files, and continue to the next interval.
- REQ-011: Snapshotter logs must not include S3 secret values.
- REQ-012: Snapshotter application files in the final Docker image must be read-only and the container must run as a non-root user.

## Acceptance Criteria

```gherkin
Feature: Snapshotter

Scenario: Archive name and upload key are deterministic
  Given an archive timestamp for 2026-06-02 10:30:00 UTC
  And the configured object prefix is "prod"
  When Snapshotter builds the upload key
  Then the archive name is "archivist-2026-06-02-1780396200.tar.gz"
  And the object key is "prod/archivist-2026-06-02-1780396200.tar.gz"

Scenario: SQLite is captured safely
  Given /data/archive.db exists
  When Snapshotter creates a staged snapshot
  Then staged data/archive.db is produced through the SQLite online backup API
  And live archive.db-wal and archive.db-shm files are not copied into the archive

Scenario: Artifact files are copied best-effort
  Given /data contains article artifact files
  When Snapshotter creates a staged snapshot
  Then artifact files observed during traversal are copied into staged data
  And files that disappear during traversal are logged as best-effort copy misses without failing the entire snapshot

Scenario: Upload failure does not terminate the daemon
  Given the configured S3 endpoint rejects an upload
  When Snapshotter runs a scheduled snapshot
  Then it logs the failure without secrets
  And it waits for the next interval
```

## Data and State

Snapshotter does not create database tables or persist local state across runs. It uses a temporary work directory, default `/tmp/archivist-snapshotter`, to stage snapshot contents and tarballs. Work directories for failed or completed attempts are removed before the service waits for the next interval.

The archive contains:

```text
manifest.json
data/
  archive.db
  articles/
    ...
```

`manifest.json` records:

- `archive_name`
- `object_key`
- `created_at`
- `unix_timestamp`
- `source_data_dir`
- `source_sqlite_path`
- `snapshotter_version`
- `consistency`

`consistency` must state that SQLite is captured with the online backup API and non-database artifacts are best-effort live copies.

## Interfaces

Runtime command:

```bash
archivist-snapshotter
```

Configuration environment variables:

```text
ARCHIVIST_DATA_DIR
ARCHIVIST_SQLITE_PATH
ARCHIVIST_SNAPSHOTTER_INTERVAL_SECONDS
ARCHIVIST_SNAPSHOTTER_WORK_DIR
ARCHIVIST_SNAPSHOTTER_S3_ENDPOINT_URL
ARCHIVIST_SNAPSHOTTER_S3_REGION
ARCHIVIST_SNAPSHOTTER_S3_BUCKET
ARCHIVIST_SNAPSHOTTER_S3_ACCESS_KEY_ID
ARCHIVIST_SNAPSHOTTER_S3_SECRET_ACCESS_KEY
ARCHIVIST_SNAPSHOTTER_OBJECT_PREFIX
```

Defaults:

- `ARCHIVIST_DATA_DIR=/data`
- `ARCHIVIST_SQLITE_PATH=/data/archive.db`
- `ARCHIVIST_SNAPSHOTTER_INTERVAL_SECONDS=86400`
- `ARCHIVIST_SNAPSHOTTER_WORK_DIR=/tmp/archivist-snapshotter`
- `ARCHIVIST_SNAPSHOTTER_OBJECT_PREFIX=` empty

Compose deployments override `ARCHIVIST_SNAPSHOTTER_WORK_DIR` to `/work/archivist-snapshotter` on a disk-backed `snapshotter-work` volume while keeping `/tmp` as tmpfs.

## Dependencies

Depends on:

- Summary-complete article artifacts from `summary-generation`
- Article delete and force-delete artifact cleanup contracts from `ui-endpoints`
- Canonical `/data` artifact contract in `docs/ARTIFACTS.md`
- Runtime topology and release automation contract in `docs/ARCHITECTURE.md`

Impacts:

- `docs/ARCHITECTURE.md`
- `docs/ARTIFACTS.md`
- `docs/DESIGN.md`
- `docs/REBUILD.md`
- CI/CD workflows
- Compose deployment files

## Rebuild Notes

- Rebuild Snapshotter as a Python 3.12 `uv` project under `src/snapshotter`.
- Use `boto3` for S3-compatible uploads.
- Do not infer backup consistency beyond the documented SQLite-online-backup plus best-effort-artifacts guarantee.
- Restore remains manual: download the archive, stop Archivist services, replace the `/data` volume contents from archive `data/`, and start services again.

## Security / Privacy Notes

- S3 access key and secret key are secret material.
- Snapshot archives are plaintext `.tar.gz` files before upload.
- Confidentiality relies on TLS in transit, private bucket access, IAM policy, and provider-side storage controls.
- Logs must not include secret values or article content.

## Observability / Logging Notes

- Snapshotter emits structured JSON logs to stdout.
- Logs include snapshot start, archive creation, upload success, upload failure, cleanup failure, and best-effort copy misses for disappearing artifact files.
- Logs must include enough context to identify archive name and object key without logging credentials.

## Open Questions

- None.

## Related Documents

- `./PLAN.md`
- `../../ARCHITECTURE.md`
- `../../ARTIFACTS.md`
- `../../DESIGN.md`
