---
id: SNAP-004
feature: snapshotter
title: Implement S3-compatible upload
status: done
depends_on:
  - SNAP-002
blocks:
  - SNAP-005
parallel: true
exec_plan: null
canonical: true
---

# SNAP-004: Implement S3-Compatible Upload

## Objective

Implement S3-compatible upload using `boto3`, explicit Scaleway-compatible endpoint configuration, object key construction, startup validation, and secret-safe logs.

## Acceptance Criteria

```gherkin
Scenario: Snapshotter uploads to configured object storage
  Given Snapshotter has a tar.gz archive and S3-compatible config
  When upload runs
  Then boto3 receives endpoint URL, region, bucket, object key, access key, and secret key
  And Snapshotter does not delete remote objects
```

## Done When

- Upload tests cover object prefix normalization, required config validation, and stubbed upload calls.
- Logs never include access key secret values.

## Validation

```bash
cd src/snapshotter && uv run pytest
```

## Required Context

- `../SPEC.md`
- `../PLAN.md`

## Open Questions

- None.
