---
id: OTEL-008
feature: otel-observability
title: Add Snapshotter backup and upload spans
depends_on:
  - OTEL-007
blocks:
  - OTEL-009
  - OTEL-011
parallel: false
requires_exec_plan: false
canonical: true
---
# OTEL-008: Add Snapshotter backup and upload spans

## Objective

Add Snapshotter spans around scheduled attempts, archive capture, SQLite backup, artifact copy, tarball creation, upload, and cleanup.

## Done When

- Snapshot attempts produce independent root traces.
- S3 upload uses botocore instrumentation and a manual upload span.
- Snapshotter telemetry does not log S3 credentials.

## Validation

Required checks:

```bash
cd src/snapshotter && uv run pytest
```
