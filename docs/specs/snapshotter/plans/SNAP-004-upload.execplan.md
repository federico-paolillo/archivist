---
id: SNAP-004-PLAN
task: ../tasks/SNAP-004-upload.md
status: completed
canonical: true
---

# ExecPlan: SNAP-004 Upload

## Objective

Implement S3-compatible upload through `boto3` with explicit Archivist-prefixed configuration and no remote deletion behavior.

## Linked Task

- `../tasks/SNAP-004-upload.md`

## Implementation Sequence

1. Parse and validate required S3 configuration from environment.
2. Normalize the optional object prefix without leading or trailing duplicate separators.
3. Create a `boto3` S3 client with endpoint URL, region, access key, and secret key.
4. Upload the archive file with `upload_file`.
5. Keep logs secret-safe.

## Validation Plan

```bash
cd src/snapshotter && uv run pytest
```

## Risks

- S3-compatible providers differ in optional features; Snapshotter should use only basic S3 upload behavior.

## Completion Criteria

- Stubbed tests prove the client receives explicit endpoint, region, bucket, key, and credentials.
