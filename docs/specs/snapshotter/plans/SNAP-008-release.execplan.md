---
id: SNAP-008-PLAN
task: ../tasks/SNAP-008-release.md
status: completed
canonical: true
---

# ExecPlan: SNAP-008 Release

## Objective

Extend CD release automation so Snapshotter is built, attested, included in image env output, and listed in draft release notes.

## Linked Task

- `../tasks/SNAP-008-release.md`

## Implementation Sequence

1. Add Snapshotter image env name to CD workflow.
2. Run Snapshotter Python validation in CD.
3. Build and push Snapshotter image for `linux/amd64` and `linux/arm64`.
4. Emit Snapshotter image attestation.
5. Pass Snapshotter digest-pinned image to `package-compose-release.sh`.
6. Pass Snapshotter image, digest, and attestation to `create-draft-release.sh`.
7. Update both scripts' argument validation and output.

## Validation Plan

```bash
scripts/package-compose-release.sh test-version gateway worker ui snapshotter
docker compose --env-file release/compose/.env --env-file release/compose/.env.images -f release/compose/docker-compose.yml config --quiet
```

## Risks

- Release scripts are positional; all call sites and usage strings must change together.

## Completion Criteria

- Release package includes `ARCHIVIST_SNAPSHOTTER_IMAGE` and draft release notes include Snapshotter image and attestation.
