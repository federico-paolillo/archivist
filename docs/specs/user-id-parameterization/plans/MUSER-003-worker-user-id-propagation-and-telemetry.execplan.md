---
id: MUSER-003-PLAN
task: ../tasks/MUSER-003-worker-user-id-propagation-and-telemetry.md
status: completed
canonical: true
---

# ExecPlan: MUSER-003 Worker User-ID Propagation And Telemetry

## Objective

Make Worker CLI enqueue use the accepted `jobs.DefaultUserID` exception, keep processing ownership derived from job state, and attach `user_id` telemetry.

## Linked Task

- `../tasks/MUSER-003-worker-user-id-propagation-and-telemetry.md`

## Required Context

- `../SPEC.md`
- `../PLAN.md`
- `../tasks/MUSER-003-worker-user-id-propagation-and-telemetry.md`
- `../../telegram-ingestion/SPEC.md`
- `../../article-processing/SPEC.md`
- `../../job-recovery/SPEC.md`
- `.agents/skills/archivist-worker/SKILL.md`

## Assumptions

- The Worker CLI enqueue command remains for operator self-enqueueing.
- Worker CLI enqueue is an explicit exception to bootstrap-only hardcoding.
- `jobs.DefaultUserID` is `01ASB2XFCZJY7WHZ2FNRTMQJCT`.
- Worker CLI enqueue must verify that `users.id = jobs.DefaultUserID` exists before inserting CLI-created article and job rows.
- Worker CLI enqueue must not infer ownership from user-table cardinality and must not create or repair the default user row.

## Non-Goals

- New CLI flags.
- Multi-user Worker scheduling policy.
- Snapshotter changes.

## Implementation Sequence

1. Define Worker `jobs.DefaultUserID` as `01ASB2XFCZJY7WHZ2FNRTMQJCT` for CLI enqueue's accepted exception.
2. Add repository logic that checks for `users.id = jobs.DefaultUserID` and fails when that row is absent.
3. Ensure repository logic does not infer ownership from the number of `users` rows and does not create or repair the default user row.
4. Insert CLI-created article/job rows with `jobs.DefaultUserID` only after the existence check passes.
5. Scope job claim and article/job mutations to consistent job/article ownership.
6. Propagate `Job.UserID` into pipeline repository calls that read or mutate article state.
7. Add `user_id` to Worker slog attributes and spans for job-scoped operations.
8. Update Worker tests for enqueue default-user existence checking, missing-default-user failure, no cardinality inference, mismatch safety, processing, and telemetry.

## Validation Plan

```bash
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

## Documentation Updates Required

- Update existing Worker-related specs only when implementation discovers a missing rebuild contract.

## Risks

- Updating repository method signatures can touch many pipeline tests.
- Mismatched job/article ownership must fail without leaking another user's article state.

## Rollback / Recovery Notes

Revert Worker repository and pipeline changes if validation fails before integration.

## Completion Criteria

- Worker production code has no literal personal-user id outside `jobs.DefaultUserID` for CLI enqueue.
- CLI enqueue tests prove default-user existence checking, missing-default-user failure, no cardinality inference, and no user creation.
- Processing tests prove persisted job/article ownership behavior.
