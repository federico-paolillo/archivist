# Implementation Diary: UI/API Authentication

This file is an append-only historical log. It records implementation outcomes, validation, decisions, and follow-ups.

It is not the canonical source of requirements. Durable decisions must be promoted to canonical documents.

---

## Log

## 2026-05-04 — AUTHN-001: Authn ALM Artifacts

Status:
- Completed.

Summary:
- Added canonical ALM artifacts for the v0 single-user UI/API authentication surface.

Changes:
- Created authn canonical artifacts, tasks, and ExecPlans.
- Promoted durable auth decisions to architecture, design, and conventions.
- Updated Telegram ingestion's shared `users` table contract so auth can own `password_hash` and Telegram ingestion can own `telegram_user_id`.

Decisions:
- `users.telegram_user_id` is nullable at rest so auth bootstrap can create the personal user before Telegram ingestion maps the Telegram identity.
- Gateway restart invalidation was expected in the prior cookie-ticket design; this was superseded by the 2026-05-05 opaque-session amendment.
- The 2048-character login secret is retained as requested but bounded by request-size validation and throttling.

Validation:
- ALM consistency was reviewed after reverting accidental code changes.

Follow-ups:
- Implement `TELING-001` before `AUTHN-002`, or explicitly update the shared users persistence contract as part of `AUTHN-002`.
- Future multi-replica auth requirements were superseded by the 2026-05-05 opaque-session amendment.

Canonical Updates:
- `docs/specs/authn/SPEC.md`
- `docs/specs/authn/PLAN.md`
- `docs/specs/authn/tasks/*.md`
- `docs/specs/authn/plans/*.execplan.md`
- `docs/specs/INDEX.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/GATEWAY.md`
- `docs/conventions/UI.md`
- `docs/specs/telegram-ingestion/SPEC.md`

## 2026-05-05 — AUTHN-DOC: DSGN-015 Opaque Session Amendment

Status:
- Completed.

Summary:
- Amended the authn ALM documentation to use opaque server-issued session cookies behind `ISessionStore`.

Changes:
- Replaced the prior cookie-ticket language with `__Host-app-auth`, `AppCookieAuthenticationHandler`, `AddAppCookie()`, `ISessionStore`, and `SessionEntry`.
- Documented `/login`, `/logout`, handler behavior, cookie lifecycle ownership, and multi-replica Redis guidance.
- Kept authn implementation tasks blocked/proposed; this was documentation-only.

Decisions:
- The cookie value is a pure random capability and carries no embedded metadata.
- Session validity is determined by server-side store presence and absolute expiry.
- The custom auth handler participates in the standard ASP.NET Core pipeline but does not issue, clear, rotate, or refresh cookies.

Validation:
- Documentation text checks were run to find stale active references to the old cookie-ticket design.

Follow-ups:
- Future implementation of `AUTHN-003` must follow the amended `DSGN-015` and `AUTHN-003` ExecPlan.

Canonical Updates:
- `docs/DESIGN.md`
- `docs/ARCHITECTURE.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/GATEWAY.md`
- `docs/specs/authn/SPEC.md`
- `docs/specs/authn/PLAN.md`
- `docs/specs/authn/tasks/AUTHN-001-authn-canonical-docs-and-design-decisions.md`
- `docs/specs/authn/tasks/AUTHN-003-gateway-cookie-authentication-endpoints.md`
- `docs/specs/authn/plans/AUTHN-002-password-persistence-and-bootstrap.execplan.md`
- `docs/specs/authn/plans/AUTHN-003-gateway-cookie-authentication.execplan.md`

## 2026-05-09 — AUTHN-002: Password persistence and bootstrap

Status:
- Completed.

Summary:
- Implemented gateway auth bootstrap: `users` table schema initialization, personal user row seeding, Argon2id PHC password hashing, and bootstrap logic.
- All acceptance criteria satisfied and 27 tests pass.

Changes:
- `Archivist.Gateway.Application/Auth/Options/AuthOptions.cs`: Typed options for `SQLITE_PATH` and `AUTH_BOOTSTRAP_PASSWORD`.
- `Archivist.Gateway.Application/Auth/Services/IPasswordValidator.cs`: Interface for 2048-char printable ASCII validation.
- `Archivist.Gateway.Application/Auth/Services/IPasswordHasher.cs`: Interface for Argon2id PHC hashing and verification.
- `Archivist.Gateway.Application/Auth/Services/IAuthBootstrapService.cs`: Interface for auth storage initialization.
- `Archivist.Gateway.Application/Auth/Services/Defaults/PasswordValidator.cs`: Validates exactly 2048 printable ASCII characters (0x20–0x7E).
- `Archivist.Gateway.Application/Auth/Services/Defaults/Argon2idPasswordHasher.cs`: Argon2id PHC hasher with m=19456,t=2,p=1 and 16-byte salt. Constant-time verification via `CryptographicOperations.FixedTimeEquals`.
- `Archivist.Gateway.Application/Auth/Services/Defaults/AuthBootstrapService.cs`: Creates `users` table if absent, inserts personal user row if absent, hashes and stores bootstrap password only when `password_hash` is NULL. Existing hashes are preserved; bootstrap password is never logged.
- `Archivist.Gateway.Application/Auth/Extensions/ServiceCollectionExtensions.cs`: `AddAuth()` registers all auth services and options.
- `Archivist.Gateway.Api/Program.cs`: Calls `AddAuth()` and awaits `IAuthBootstrapService.InitializeAsync()` before accepting requests.
- `Archivist.Gateway.Tests/Auth/PasswordValidatorTest.cs`: Unit tests covering all validator boundary conditions.
- `Archivist.Gateway.Tests/Auth/Argon2idPasswordHasherTest.cs`: Unit tests covering hash format, salt uniqueness, and verification.
- `Archivist.Gateway.Tests/Auth/AuthBootstrapServiceTest.cs`: Integration tests covering bootstrap of missing hash, skip-when-present, validation failure, idempotent second call, and verification.
- `Archivist.Gateway.Tests/IntegrationTest.cs`: Updated integration test base to register a no-op `IAuthBootstrapService` stub by default, so API integration tests do not need a database.
- Added NuGet packages: `Konscious.Security.Cryptography.Argon2 1.3.1`, `Microsoft.Data.Sqlite 10.0.7`, `Microsoft.Extensions.Logging.Abstractions 10.0.0`.

Decisions:
- Used `Microsoft.Data.Sqlite` directly for bootstrap rather than EF Core: bootstrap runs before the full application stack is initialized and only needs minimal SQL. EF Core entity configuration belongs to future persistence tasks.
- `CREATE TABLE IF NOT EXISTS` used for schema initialization: this is safe and idempotent when TELING-001 has already created the table with the same schema. The column set matches the TELING-001 contract exactly (`id`, nullable `telegram_user_id`, nullable `password_hash`).
- PHC string format implemented manually: Konscious library provides raw Argon2id computation; the PHC string encoding (`$argon2id$v=19$m=...$salt$hash`) is constructed in the hasher.
- Integration test base class now stubs `IAuthBootstrapService` by default to isolate non-auth integration tests from database requirements. Auth-specific tests call the real service directly.
- `app.Run()` changed to `await app.RunAsync()` to satisfy CA1849 (analyzer enforces non-blocking async host startup).

Validation:
- `dotnet format`: no changes required.
- `dotnet build`: 0 warnings, 0 errors.
- `dotnet test`: 27 passed, 0 failed, 0 skipped.
- Test coverage: bootstrap of missing hash, existing hash preservation, skip without bootstrap password when hash exists, invalid password rejection (too short), missing SQLITE_PATH rejection, stored hash verifies against bootstrap password, idempotent second call.

Follow-ups:
- A 2048-character password is a potential CPU-amplification vector if requests are not throttled at the gateway layer. This risk must be addressed in AUTHN-003 (login throttling before Argon2id verification). This was noted as a risk in the ExecPlan and is mitigated by AUTHN-003 throttling.
- AUTHN-003 now unblocks.

Canonical Updates:
- `docs/specs/authn/tasks/AUTHN-002-password-persistence-and-bootstrap.md` (status: done)
- `docs/specs/authn/PLAN.md` (AUTHN-002 row: done)
- `docs/specs/authn/plans/AUTHN-002-password-persistence-and-bootstrap.execplan.md` (status: completed)

## 2026-05-06 — DOCS-SANITY: UI Auth Dependency Correction

Status:
- completed

Summary:
- Corrected authn dependency docs so UI routing/auth shell work waits for the validated browser auth contract.

Changes:
- Updated `AUTHN-003`, `AUTHN-004`, and `PLAN.md` blockers.
- Replaced stale global cookie-key wording with in-memory sessions and login throttling.

Decisions:
- `UI-002` depends on `AUTHN-004`, not only on earlier auth endpoint planning.

Validation:
- Structural docs check passed for task/PLAN drift, dependency drift, ExecPlan links, required context, Markdown links, and canonical TODOs.
- Targeted repository searches found no stale ephemeral-cookie-key wording.
- Production build/test validation was not required because this was a docs-only correction.

Follow-ups:
- Implement UI routing only after the browser auth contract is validated.

Canonical Updates:
- `docs/specs/authn/PLAN.md`
- `docs/specs/authn/tasks/AUTHN-003-gateway-cookie-authentication-endpoints.md`
- `docs/specs/authn/tasks/AUTHN-004-protect-ui-api-and-validate-auth-client-contract.md`
- `docs/ARCHITECTURE.md`
- `docs/conventions/GENERAL.md`

---

## 2026-05-10 — AUTHN-003: Gateway Opaque Session Cookie Authentication

Task ID: AUTHN-003
Status: done
Branch: agent/authn-003

### Summary of Changes

Implemented the full gateway session cookie authentication surface:

**Application layer (`Archivist.Gateway.Application/Auth/`)**:
- `AppCookieDefaults` — scheme name `"app-cookie"` and cookie name `"__Host-app-auth"` constants.
- `AppCookieOptions` — extends `AuthenticationSchemeOptions`, defaults: `CookieName = "__Host-app-auth"`, `SessionLifetime = 24h`.
- `AuthOptions` — bootstrap/SQLite options.
- `IPasswordValidator`, `IPasswordHasher`, `IPasswordStore`, `IAuthBootstrapService` interfaces.
- `ISessionStore` + `SessionEntry` record — per SPEC.md contract exactly.
- `ILoginThrottle` — per-IP and global failed-attempt rate limiting interface.
- `PasswordValidator`, `Argon2idPasswordHasher`, `AuthBootstrapService`, `SqlitePasswordStore` — implementations.
- `InMemorySessionStore` — `ConcurrentDictionary`-backed, expired entries removed eagerly on lookup, `TimeProvider`-injected.
- `InMemoryLoginThrottle` — per-IP limit: 10, global limit: 50.
- `AppCookieAuthenticationHandler` — custom auth handler.
- `AuthenticationBuilderExtensions.AddAppCookie()` — registers the scheme.
- `ServiceCollectionExtensions.AddAuth()` — registers all auth services.

**API layer (`Archivist.Gateway.Api/Auth/`)**:
- `LoginRequest` DTO, `Handlers.PostLogin`, `Handlers.PostLogout`, `Handlers.GetSession`.
- `SameOriginFilter` — rejects cross-origin unsafe methods with `403`.
- `Endpoints.MapAuth()`.

**Program.cs** updated: `AddAuth`, auth bootstrap, `UseAuthentication`, `UseAuthorization`, `MapAuth`.

**Tests**: 44 tests covering session store, login throttle, auth handler, endpoint success/failure, cookie attributes, session rotation, logout, throttling, same-origin rejection.

### Decisions

1. Application project uses `<FrameworkReference Include="Microsoft.AspNetCore.App" />`.
2. `IOptions<AppCookieOptions>` explicitly configured via `services.Configure<AppCookieOptions>(_ => { })`.
3. `SameOriginFilter` normalizes default ports (80/HTTP, 443/HTTPS).
4. `InMemoryLoginThrottle` global counter does not reset on `RecordSuccess` — intentional.

### Validation

```
dotnet format   — clean
dotnet build    — succeeded, 0 warnings, 0 errors
dotnet test     — Passed: 44, Failed: 0, Skipped: 0
```

### Follow-ups

- `AUTHN-004`: protect UI-facing API endpoints.
- `AUTHN-005`: security validation pass.

### Canonical Updates

- `docs/specs/authn/tasks/AUTHN-003-gateway-cookie-authentication-endpoints.md` — status: done
- `docs/specs/authn/PLAN.md` — AUTHN-003 row: done
- `docs/specs/authn/plans/AUTHN-003-gateway-cookie-authentication.execplan.md` — status: completed
- `docs/specs/authn/DIARY.md` — this entry

## 2026-06-05 — AUTHN-DOC: Multi-Password Login And Bootstrap Telegram Sender Correction

Status:
- completed correction

Summary:
- Corrected the canonical auth contract so password-only login supports multiple password-bearing users and auth bootstrap seeds the personal Telegram sender mapping from the fixed accepted id.

Changes:
- Replaced the stale "single password-bearing user row" login rule with candidate-based verification across all non-empty Argon2id PHC password hashes.
- Specified that login issues a session only when exactly one candidate matches; duplicate password matches fail closed.
- Documented that multiple password-bearing rows are valid.
- Documented that auth bootstrap sets the personal row's `telegram_user_id` to `1559957191` only when null and preserves an existing non-null value.
- Removed `Telegram:AllowedUserId` as an auth bootstrap input.

Decisions:
- Password-only login ambiguity is resolved by the submitted password, not by requiring a single password-bearing row.
- No registration, user-management endpoints, or user self-service are introduced.

Validation:
- Documentation-only correction. `git diff --check` was run by the docs worker.

Follow-ups:
- None.

Canonical Updates:
- `docs/specs/authn/SPEC.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`

## 2026-05-12 — AUTHN-004: Protect UI API and validate auth client contract

Task ID: AUTHN-004
Status: done

Summary:
- Added AUTHN-004 regression coverage for unauthenticated protected Gateway access through the authenticated session probe.
- Confirmed `GET /auth/session` returns `204` with a valid app-cookie session and `401` without one.
- Confirmed the existing `/login` and `/logout` endpoint tests still pass.

Decisions:
- Did not add a placeholder `/articles` route in AUTHN-004. `GET /articles`, `GET /articles/{id}`, and `DELETE /articles/{id}` are owned by the `ui-endpoints` feature, and adding a stub here would create an incomplete public API contract.
- Used `GET /auth/session` as the protected Gateway contract probe because it already participates in the real ASP.NET Core authentication and authorization pipeline and is the auth check consumed by the final UI.

Validation:
- `cd src/gateway && dotnet test`
- Result: Passed: 101, Failed: 0, Skipped: 0.

Follow-ups:
- `ui-endpoints` must apply `RequireAuthorization()` to concrete article endpoints when those routes are implemented.
- `AUTHN-005` remains the next authn security validation pass.

Canonical Updates:
- `docs/specs/authn/tasks/AUTHN-004-protect-ui-api-and-validate-auth-client-contract.md` — status: done, validation recorded.
- `docs/specs/authn/PLAN.md` — AUTHN-004 row: done.

## 2026-05-15 — AUTHN-006 and AUTHN-005: Forwarded-header auth and security validation

Task IDs: AUTHN-006, AUTHN-005
Status: done

Summary:
- Implemented trusted reverse-proxy forwarded-header handling before authentication, authorization, and route mapping.
- Added `GATEWAY_PUBLIC_HOSTS` startup enforcement outside Development and forwarded-host allowlisting through ASP.NET Core forwarded-header options.
- Changed login to require post-forwarding `Request.Scheme == "https"`.
- Updated same-origin filtering to compare post-forwarding scheme, host, and effective port.
- Added regression coverage for forwarded HTTPS login success, effective HTTP login rejection, scheme/host/port origin mismatches for login/logout/delete, forwarded-host allowlisting, missing production public hosts, cookie attributes, and Gateway's unprefixed auth route contract.

Decisions:
- Gateway processes `X-Forwarded-For`, `X-Forwarded-Proto`, and `X-Forwarded-Host` with `ForwardLimit = 1`.
- Known proxy/network lists are cleared because v0 uses the private Docker network as the trust boundary and does not define `GATEWAY_TRUSTED_PROXY_RANGES`.
- The test harness verifies Gateway behavior and the requirement that `/api/*` is not mapped in Gateway. It does not execute Caddy, so actual public `/api` prefix stripping remains a documented deployment assumption.
- Publication hygiene: `AMEND.md`, `REVIEW.md`, `REFACTOR.md`, and `.claude/worktrees/` remain local temporary review/worktree artifacts and are not part of the staged publication set.
- The `src/gateway/.gitignore` `*.lscache` rule is intentional and included to keep local Gateway tooling cache files out of source publication.

Validation:
- `cd src/gateway && dotnet format` — passed.
- `cd src/gateway && dotnet build` — passed with 0 warnings and 0 errors.
- `cd src/gateway && dotnet test` — passed: 124 tests, 0 failed.
- `cd src/worker && go tool lefthook run build` — passed.
- `cd src/worker && go tool lefthook run format` — passed; first run emitted a non-fatal stale golangci cache warning referencing `/Users/federico.paolillo/src/archivist-worktrees/wave4-mdext-005`.
- `cd src/worker && go tool lefthook run lint` — initial run failed because golangci used stale cache entries for the deleted `/Users/federico.paolillo/src/archivist-worktrees/wave4-mdext-005` path; after `go tool golangci-lint cache clean`, rerunning the required lint command passed.
- `cd src/worker && go tool lefthook run test` — passed.

Follow-ups:
- Deployment validation should verify Caddy uses `http://:443`, overwrites forwarded headers, sets `X-Forwarded-Proto: https`, and strips `/api` before forwarding to Gateway.

Canonical Updates:
- `docs/specs/authn/SPEC.md` — status: done.
- `docs/specs/authn/PLAN.md` — AUTHN-006 and AUTHN-005 rows: done.
- `docs/specs/authn/tasks/AUTHN-006-reverse-proxy-forwarded-headers-and-effective-https-auth.md` — status: done, validation recorded.
- `docs/specs/authn/plans/AUTHN-006-reverse-proxy-forwarded-headers.execplan.md` — status: completed.
- `docs/specs/authn/tasks/AUTHN-005-security-validation-pass.md` — status: done, validation recorded.
- `docs/specs/INDEX.md` — authn status: done.

## 2026-05-31 — AUTHN-REVIEW-P2: AppCookieSettings canonical naming correction

Task ID: AUTHN-REVIEW-P2
Status: done

Summary:
- Resolved the active P2 review finding where `docs/DESIGN.md` still named the auth cookie settings type `AppCookieOptions`.
- Updated DSGN-015 to use `AppCookieSettings`, matching the auth spec and Gateway implementation.

Decisions:
- No runtime auth behavior changed. This was a canonical documentation consistency fix only.
- Historical diary references to the earlier `AppCookieOptions` name remain historical implementation record and are not canonical rebuild guidance.

Validation:
- `cd src/gateway && dotnet format` — passed.
- `cd src/gateway && dotnet build` — passed.
- `cd src/gateway && dotnet test` — passed: 162 tests.
- `git diff --check` — passed.

Follow-ups:
- None.

Canonical Updates:
- `docs/DESIGN.md` — DSGN-015 now uses `AppCookieSettings`.
