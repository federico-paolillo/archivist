# Implementation Diary: User ID Parameterization

This file is an append-only historical log. It records implementation outcomes, validation, decisions, and follow-ups.

It is not the canonical source of requirements. Durable decisions must be promoted to canonical documents.

---

## Log

## 2026-06-05 — MUSER-001: Canonical User-ID Resolution Contract

Status:
- completed

Summary:
- Created the `user-id-parameterization` feature artifacts, task DAG, and ExecPlans.

Changes:
- Added the feature spec, plan, task files, and ExecPlans.
- Updated `docs/specs/INDEX.md`, `docs/REBUILD.md`, `docs/ARCHITECTURE.md`, and `docs/DESIGN.md`.

Decisions:
- Authentication bootstrap is the only production path allowed to hardcode the personal user ULID.
- Runtime ownership resolves from SQLite, session state, claimed jobs, or article ownership.
- `Telegram:AllowedUserId` remains only as the current bootstrap seed for `users.telegram_user_id`.

Validation:
- `git diff --check` passed.

Follow-ups:
- None.

Canonical Updates:
- `docs/specs/user-id-parameterization/**`
- `docs/specs/INDEX.md`
- `docs/REBUILD.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`

## 2026-06-05 — MUSER-002: Gateway User-ID Resolution And Telemetry

Status:
- completed

Summary:
- Implemented Gateway runtime `user_id` resolution for password login, Telegram webhook authorization, Telegram ingestion ownership, and Gateway telemetry.

Changes:
- Password login now reads the single password-bearing user row and issues sessions for that row's `id`.
- Auth bootstrap seeds the personal row's `telegram_user_id` from the existing `Telegram:AllowedUserId` config when the mapping is empty.
- Telegram webhook authorization resolves senders through `users.telegram_user_id`; unknown senders create no rows and receive no reply.
- Telegram ingestion receives resolved `user_id` and no longer creates or upserts `users`.
- Gateway spans/scopes attach `user_id` when known.

Decisions:
- No Gateway production code outside `AuthBootstrapService` retains the personal ULID.

Validation:
- Gateway worker: `dotnet format`, `dotnet build -maxcpucount:1`, `dotnet test --no-build`, and `git diff --check` passed before coordinator bootstrap adjustment.
- Coordinator: `cd src/gateway && dotnet format --verify-no-changes && dotnet build -maxcpucount:1 && dotnet test --no-build` passed with 194 tests.

Follow-ups:
- None.

Canonical Updates:
- `docs/specs/authn/SPEC.md`
- `docs/specs/telegram-ingestion/SPEC.md`
- `docs/specs/ui-endpoints/SPEC.md`
- `docs/specs/otel-observability/SPEC.md`

## 2026-06-05 — MUSER-003: Worker User-ID Propagation And Telemetry

Status:
- completed

Summary:
- Implemented Worker default-user resolution, job/article ownership scoping, and `user_id` telemetry.

Changes:
- Worker CLI enqueue resolves exactly one persisted `users.id`.
- Worker enqueue fails when zero or multiple users exist.
- Worker job claim requires job/article ownership consistency.
- Article URL reads, canonical URL updates, title updates, and terminal completion are scoped by `Job.UserID`.
- Worker job-scoped logs and spans attach `user_id`.

Decisions:
- Worker has no production literal personal-user id.

Validation:
- Worker worker: `go test ./pkg/jobs ./internal/pipeline ./internal/app`, `go test ./...`, `go tool lefthook run build`, `go tool lefthook run format`, `go tool lefthook run lint`, and `go tool lefthook run test` passed.
- Coordinator: `cd src/worker && go test ./...` passed.
- Coordinator: `cd src/worker && go tool lefthook run format` passed with no fixes.

Follow-ups:
- None.

Canonical Updates:
- `docs/specs/user-id-parameterization/SPEC.md`

## 2026-06-05 — MUSER-004: Cross-Feature Docs And Observability Cleanup

Status:
- completed

Summary:
- Aligned existing canonical specs with the new user-id parameterization contract.

Changes:
- Updated auth, Telegram ingestion, UI endpoint, and OpenTelemetry specs to remove runtime personal-user hardcoding requirements and define `user_id` telemetry.

Decisions:
- Snapshotter remains excluded from `user_id` telemetry.

Validation:
- `git diff --check` passed.

Follow-ups:
- None.

Canonical Updates:
- `docs/specs/authn/SPEC.md`
- `docs/specs/telegram-ingestion/SPEC.md`
- `docs/specs/ui-endpoints/SPEC.md`
- `docs/specs/otel-observability/SPEC.md`

## 2026-06-05 — MUSER-005: Integration Validation

Status:
- completed

Summary:
- Integrated Gateway, Worker, and documentation slices and ran cross-module validation.

Changes:
- Fixed bootstrap seeding after integration review identified the fresh-install mapping gap.
- Ran production hardcoded-id scan confirming the personal ULID remains only in `AuthBootstrapService`.

Decisions:
- Gateway reviewer requested formatting fixes only; `dotnet format` resolved them.
- Worker reviewer approved without findings.

Validation:
- `cd src/gateway && dotnet format --verify-no-changes && dotnet build -maxcpucount:1 && dotnet test --no-build` passed.
- `cd src/worker && go test ./...` passed.
- `cd src/worker && go tool lefthook run build && go tool lefthook run lint && go tool lefthook run test` passed as part of broad validation.
- `cd src/worker && go tool lefthook run format` passed.
- `docker compose --env-file .env.example config --quiet` passed.
- `git diff --check` passed.

Follow-ups:
- None.

Canonical Updates:
- `docs/specs/user-id-parameterization/PLAN.md`
- `docs/specs/user-id-parameterization/tasks/*.md`
- `docs/specs/user-id-parameterization/plans/*.execplan.md`

## 2026-06-05 — MUSER-006: Review And ALM Closure

Status:
- completed

Summary:
- Completed multi-agent review, final validation, and ALM closure.

Changes:
- Marked all MUSER tasks done.
- Marked MUSER ExecPlans completed.
- Marked the feature and index entry done.

Decisions:
- No reviewer findings remain open.

Validation:
- `git diff --check` passed after ALM closure.

Follow-ups:
- None.

Canonical Updates:
- `docs/specs/user-id-parameterization/PLAN.md`
- `docs/specs/user-id-parameterization/DIARY.md`
- `docs/specs/INDEX.md`

## 2026-06-05 — MUSER-003: Worker CLI Default-User Correction

Status:
- completed correction after task completion

Summary:
- Revised completed MUSER-003 ALM artifacts to reflect the accepted Worker CLI enqueue exception to bootstrap-only hardcoding.

Changes:
- Replaced the prior "exactly one `users` row" Worker CLI enqueue rule with explicit `jobs.DefaultUserID = 01ASB2XFCZJY7WHZ2FNRTMQJCT` behavior.
- Specified that Worker CLI enqueue checks `users.id` by that id, fails when the row is missing, never infers ownership from user-table cardinality, and never creates or repairs the user.
- Preserved the feature, task, and ExecPlan completion state.

Decisions:
- Worker CLI enqueue is the only runtime exception to bootstrap-only hardcoding.
- Additional user rows do not affect Worker CLI enqueue ownership.

Validation:
- `git diff --check` passed for tracked worktree changes.
- `git diff --no-index --check` against empty files passed for the currently untracked `docs/specs/user-id-parameterization/` correction documents.
- Targeted `git diff --no-index --check` whitespace checks passed for the currently untracked user-id-parameterization feature documents.

Follow-ups:
- None.

Canonical Updates:
- `docs/specs/user-id-parameterization/SPEC.md`
- `docs/specs/user-id-parameterization/PLAN.md`
- `docs/specs/user-id-parameterization/tasks/MUSER-003-worker-user-id-propagation-and-telemetry.md`
- `docs/specs/user-id-parameterization/plans/MUSER-003-worker-user-id-propagation-and-telemetry.execplan.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`

## 2026-06-05 — MUSER-002: Gateway Password Login And Bootstrap Correction

Status:
- completed correction after task completion

Summary:
- Revised completed MUSER-002 ALM artifacts and canonical docs to reflect the accepted Gateway correction for password-only login and auth bootstrap Telegram sender seeding.

Changes:
- Replaced the prior "single password-bearing row" Gateway login rule with candidate-based verification across all non-empty Argon2id PHC password hashes.
- Specified that multiple password-bearing rows are valid and login succeeds only when exactly one user matches the submitted password.
- Specified that duplicate password matches fail closed without issuing a session.
- Updated auth bootstrap requirements so the personal row's `telegram_user_id` is set to `1559957191` only when null and an existing non-null value is preserved.
- Removed `settings.PersonalTelegramUserId` / `Telegram:AllowedUserId` as bootstrap inputs and removed the obsolete `.env.example` sample key.
- Preserved the `user-id-parameterization` feature, MUSER-002 task, and MUSER-002 ExecPlan statuses as done/completed.

Decisions:
- Password-only login supports future password-bearing rows without introducing usernames, registration, or user-management endpoints.
- Bootstrap owns the fixed personal Telegram sender seed; runtime Telegram authorization remains database mapping based.

Validation:
- `git diff --check` passed for the documentation correction.

Follow-ups:
- None.

Canonical Updates:
- `docs/specs/authn/SPEC.md`
- `docs/specs/user-id-parameterization/SPEC.md`
- `docs/specs/user-id-parameterization/PLAN.md`
- `docs/specs/user-id-parameterization/tasks/MUSER-002-gateway-user-id-resolution-and-telemetry.md`
- `docs/specs/user-id-parameterization/plans/MUSER-002-gateway-user-id-resolution-and-telemetry.execplan.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `.env.example`
