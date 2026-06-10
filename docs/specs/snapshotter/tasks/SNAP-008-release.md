---
id: SNAP-008
feature: snapshotter
title: Extend CD release automation
status: done
depends_on:
  - SNAP-006
  - SNAP-007
blocks:
  - SNAP-009
parallel: false
exec_plan: ../plans/SNAP-008-release.execplan.md
canonical: true
---

# SNAP-008: Extend CD Release Automation

## Objective

Extend CD, release packaging, and release notes so Snapshotter is built, pushed, attested, digest-pinned, and included in the deployment package.

## Acceptance Criteria

```gherkin
Scenario: Release includes Snapshotter
  Given CD runs for a release version
  When images are built and the Compose package is generated
  Then Snapshotter has a pushed multi-arch image, attestation, image env entry, release note entry, and production Compose validation
```

## Done When

- `.github/workflows/cd.yml` builds and attests Snapshotter.
- `scripts/package-compose-release.sh` accepts and writes the Snapshotter image.
- `scripts/create-draft-release.sh` includes Snapshotter image and attestation in release notes.

## Validation

```bash
scripts/package-compose-release.sh test-version gateway worker ui snapshotter
docker compose --env-file release/compose/.env --env-file release/compose/.env.images -f release/compose/docker-compose.yaml -f release/compose/docker-compose.prod.yaml config --quiet
```

## Required Context

- `../SPEC.md`
- `../PLAN.md`
- `.github/workflows/cd.yml`
- `scripts/package-compose-release.sh`
- `scripts/create-draft-release.sh`

## Open Questions

- None.
