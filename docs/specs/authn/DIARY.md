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
