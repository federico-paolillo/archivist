---
id: MUSER-003
feature: user-id-parameterization
title: Worker user-id propagation and telemetry
status: done
depends_on: [MUSER-001]
blocks: [MUSER-005]
parallel: true
exec_plan: ../plans/MUSER-003-worker-user-id-propagation-and-telemetry.execplan.md
canonical: true
---

# MUSER-003: Worker User-ID Propagation And Telemetry

## Objective

Apply the accepted Worker CLI enqueue exception to bootstrap-only hardcoding, enforce job/article user consistency, and attach `user_id` telemetry.

## Scope

This task includes:

- Worker CLI enqueue default-user handling through `jobs.DefaultUserID = 01ASB2XFCZJY7WHZ2FNRTMQJCT`.
- Worker CLI enqueue existence check for `users.id = jobs.DefaultUserID`.
- Worker CLI enqueue failure when the default user row is missing.
- Worker CLI enqueue prohibition on user-table cardinality inference and user creation.
- Job repository ownership-scoped reads and mutations.
- Pipeline propagation of `Job.UserID`.
- Worker logs and spans using `user_id`.
- Worker tests for default-user existence checking, mismatch safety, and telemetry.

## Out of Scope

This task does not include:

- Gateway changes.
- New CLI flags.
- User creation, repair, or provisioning from Worker CLI enqueue.
- Snapshotter telemetry.

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `../plans/MUSER-003-worker-user-id-propagation-and-telemetry.execplan.md`
- `../../../ARCHITECTURE.md`
- `../../../DESIGN.md`
- `../../telegram-ingestion/SPEC.md`
- `../../article-processing/SPEC.md`
- `../../job-recovery/SPEC.md`
- `.agents/skills/archivist-worker/SKILL.md`

## Acceptance Criteria

```gherkin
Scenario: Worker processes only owned article/job pairs
  Given a claimed job has a user_id
  When Worker reads or mutates the associated article
  Then the SQL operation is scoped to that user_id

Scenario: Worker CLI enqueue uses the explicit default user
  Given a user row exists with id "01ASB2XFCZJY7WHZ2FNRTMQJCT"
  When Worker CLI enqueue creates an article and job
  Then the article and job use `jobs.DefaultUserID`

Scenario: Worker CLI enqueue does not infer ownership from user count
  Given a user row exists with id "01ASB2XFCZJY7WHZ2FNRTMQJCT"
  And other user rows also exist
  When Worker CLI enqueue creates an article and job
  Then the article and job use `jobs.DefaultUserID`

Scenario: Worker CLI enqueue fails when the default user is missing
  Given no user row exists with id "01ASB2XFCZJY7WHZ2FNRTMQJCT"
  When Worker CLI enqueue runs
  Then no user, article, or job row is created
  And enqueue fails
```

## Done When

- Worker production code defines the personal user id only as `jobs.DefaultUserID` for CLI enqueue's accepted exception.
- CLI enqueue checks for `users.id = jobs.DefaultUserID` and fails when that row is missing.
- CLI enqueue does not infer ownership from the number of `users` rows and does not create the user.
- Job/article mismatch tests prove processing fails safely.
- Worker job-scoped logs and spans include `user_id`.

## Validation

Required checks:

```bash
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

## Dependencies

Depends on:

- `MUSER-001`

Blocks:

- `MUSER-005`

## ExecPlan

ExecPlan:

```text
../plans/MUSER-003-worker-user-id-propagation-and-telemetry.execplan.md
```

## Open Questions

- None.
