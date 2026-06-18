---
id: AUTHN
slug: authn
title: UI/API Authentication
owner: null
depends_on: [telegram-ingestion]
impacts: [gateway, ui, sqlite]
canonical: true
---
# Feature: UI/API Authentication

## Intent

Protect the Archivist web UI and UI-facing gateway API with cookie authentication for the bootstrapped user set.

## Motivation

Archivist exposes a private browser UI for article review and administration. The UI must be usable from a normal browser without exposing credentials to JavaScript, while keeping the implementation small enough for a single VPS deployment.

## Scope

In scope:

- Password-only login across all password-bearing Archivist user rows, with session issuance only when exactly one stored hash matches the submitted password.
- Argon2id password hash storage in SQLite.
- One-time bootstrap of the personal user's password hash from `AUTH_BOOTSTRAP_PASSWORD`.
- Bootstrap of the personal user's `telegram_user_id` to the fixed Telegram sender id `1559957191` when the row has no Telegram sender mapping.
- Persisted Telegram sender identity mapping on `users.telegram_user_id`; auth bootstrap seeds only the personal mapping and runtime Gateway flows must treat mappings as data read from SQLite.
- Opaque server-issued session id cookies integrated through a custom ASP.NET Core authentication handler.
- Browser-session auth cookies with `HttpOnly`, `Secure`, `SameSite=Strict`, `Path=/`, no `Domain`, and the `__Host-app-auth` cookie name.
- In-memory `ISessionStore` with 24-hour server-side absolute expiry.
- `POST /login`, `POST /logout`, and `GET /auth/session`.
- Protection of UI-facing API endpoints.
- In-memory login throttling for the single gateway instance.
- Basic same-origin defenses for unsafe methods.
- Forwarded-header handling for the trusted Docker reverse proxy deployment topology.
- Auth endpoint/session contracts consumed by the browser UI.
- Auth does not provide account registration, password reset, password rotation UI, user self-service, roles, tenant isolation, user-management UI, multi-replica session sharing, custom HMAC cookies, random cookie names, user-id hash cookie payloads, sliding expiry, refresh tokens, multiple concurrent sessions per user, revocation lists beyond session entry removal, or cryptographic transforms over the cookie value.
- Browser login rendering, login-failure route, article shell, and logout menu are owned by the `ui` feature.

## Users / Actors

- Personal Archivist user.
- Gateway API.
- Preact/Vite UI.
- SQLite database.

## Requirements

- REQ-001: `POST /login` must accept only the `password` credential in the request body.
- REQ-002: The accepted login secret must be exactly 2048 printable ASCII characters.
- REQ-003: The gateway must reject malformed or oversized login requests before Argon2id verification.
- REQ-004: The `users` table must contain `password_hash` as an Argon2id PHC string containing algorithm, parameters, salt, and hash.
- REQ-005: Argon2id verification must use at least `m=19456,t=2,p=1`; implementations may increase cost only by updating this spec and validation expectations.
- REQ-006: Authentication bootstrap must create the initial personal user row with `id = 01ASB2XFCZJY7WHZ2FNRTMQJCT`.
- REQ-007: `AUTH_BOOTSTRAP_PASSWORD` must be accepted only when the personal user has no stored password hash.
- REQ-008: The bootstrap password must not be logged, returned, or persisted in plaintext.
- REQ-009: Successful login must issue an opaque session cookie named `__Host-app-auth`.
- REQ-010: The session id must be 32 bytes from `RandomNumberGenerator.GetBytes`, base64url-encoded without padding.
- REQ-011: The browser cookie must not set `Expires` or `Max-Age`.
- REQ-012: The auth cookie must be `HttpOnly`, `Secure`, `SameSite=Strict`, use `Path=/`, and omit `Domain`.
- REQ-013: `ISessionStore` must map `sessionId` to `{ userId, createdAt, absoluteExpiresAt }`.
- REQ-014: Session expiry must be absolute and server-side enforced at 24 hours from issue.
- REQ-014A: The auth cookie name `__Host-app-auth` and 24-hour server-side session lifetime are canonical constants. Gateway must consume `AppCookieDefaults` directly; configuration values that attempt to change them must have no effect unless this spec promotes those values to configurable deployment behavior.
- REQ-015: The `ISessionStore` implementation must be in-memory and may use `ConcurrentDictionary<string, SessionEntry>`.
- REQ-016: Gateway restart invalidates existing sessions by wiping the in-memory store.
- REQ-017: Gateway auth must integrate with ASP.NET Core through a custom `IAuthenticationHandler`, or `AuthenticationHandler<AuthenticationSchemeOptions>`, registered by `AddAppCookie()`.
- REQ-018: The auth scheme and authentication type must be `"app-cookie"`.
- REQ-019: Authenticated requests must receive a minimal `ClaimsPrincipal` containing only `ClaimTypes.NameIdentifier` with the user id.
- REQ-020: The handler must not issue, clear, rotate, or refresh cookies.
- REQ-021: `POST /login` must accept JSON `{ "password": string }`, return `204` and `Set-Cookie` on success, and return a generic `401` on invalid credentials.
- REQ-022: `/login` must reject non-`POST`, non-same-origin, and requests whose effective public scheme is not `https`.
- REQ-023: Login throttling must run per IP and globally before Argon2id verification.
- REQ-024: Successful login must always rotate the session id and remove any existing valid session from the request cookie.
- REQ-025: `POST /logout` must remove the store entry when present, always clear `__Host-app-auth`, and return `204`.
- REQ-026: `GET /auth/session` must return `204` when authenticated and `401` otherwise.
- REQ-027: UI-facing API endpoints must require the auth cookie and must return `401` or `403`, not HTML login redirects.
- REQ-028: Unsafe cross-site requests must be rejected using same-origin request checks. State-changing `GET` endpoints are prohibited.
- REQ-029: Cookie values, passwords, and `Set-Cookie` headers must not be logged.
- REQ-030: Final browser rendering for login, login failure, protected routes, and logout belongs to the `ui` feature and must consume these auth endpoints without changing their contracts.
- REQ-031: The effective public scheme is `HttpRequest.Scheme` after trusted forwarded-header processing.
- REQ-032: Gateway must run privately on the Docker internal network behind Caddy and must not be exposed directly to the public Internet when forwarded headers are trusted.
- REQ-033: Gateway startup must process `X-Forwarded-Proto` and `X-Forwarded-For` before authentication middleware, authorization middleware, and endpoint mapping.
- REQ-034: Forwarded host values must be constrained by `GATEWAY_PUBLIC_HOSTS`.
- REQ-035: Same-origin checks for unsafe methods must compare post-forwarding scheme, host, and effective port.
- REQ-036: Authentication bootstrap must set the personal row's `telegram_user_id` to the fixed Telegram sender id `1559957191` when the row has no Telegram sender mapping, and must preserve an existing non-null `telegram_user_id`.
- REQ-037: Authentication bootstrap must not read deployment-configured Telegram sender allowlists or personal sender settings to seed the personal Telegram sender mapping.
- REQ-038: Successful login must load every user row with a non-empty Argon2id PHC `password_hash`, verify the submitted password against every candidate, and issue the session for the matched row's `id` only when exactly one candidate matches.
- REQ-039: Multiple password-bearing user rows are valid.
- REQ-040: Login must fail closed when no password-bearing row exists, when no candidate hash matches, or when more than one candidate hash matches the submitted password.
- REQ-041: Auth must not introduce registration, user-management endpoints, or user self-service.
- REQ-042: Gateway runtime code must derive authenticated user identity from password verification, session state, or persisted Telegram sender mapping, not from the personal user ULID constant outside bootstrap.
- REQ-043: Telegram sender authorization must resolve a user only through `users.telegram_user_id`; unknown Telegram senders must not create or mutate users, articles, jobs, or notifications.
- REQ-044: Gateway logs and spans must attach `user_id` after password, session, or Telegram sender resolution when the Archivist user is known.

## Acceptance Criteria

```gherkin
Feature: UI/API authentication

Scenario: Bootstrap stores the personal password hash
  Given the personal user has no password_hash
  And AUTH_BOOTSTRAP_PASSWORD is a valid 2048-character printable ASCII secret
  When the gateway starts
  Then the users table contains the personal user
  And password_hash is an Argon2id PHC string
  And telegram_user_id is 1559957191 when it was previously null
  And the plaintext bootstrap secret is not logged or stored

Scenario: Bootstrap preserves an existing Telegram sender mapping
  Given the personal user has a non-null telegram_user_id
  When the gateway starts
  Then telegram_user_id is unchanged

Scenario: Successful login
  Given exactly one password-bearing user has a stored password_hash for the submitted password
  And other password-bearing users do not match the submitted password
  When the browser posts the password to /login
  Then the response status is 204
  And Set-Cookie contains "__Host-app-auth"
  And the cookie is HttpOnly, Secure, SameSite=Strict, and Path=/
  And the cookie has no Domain, Expires, or Max-Age
  And the session store contains a fresh session entry expiring 24 hours after issue
  And the session entry user_id is the matching user's id

Scenario: Invalid login
  Given password-bearing users exist
  When the browser posts the wrong password to /login
  Then the response status is 401
  And no auth cookie is issued
  And throttle counters are incremented

Scenario: Duplicate password match fails closed
  Given two password-bearing users have password_hash values that verify the submitted password
  When the browser posts the password to /login
  Then the response status is 401
  And no auth cookie is issued

Scenario: Protected API request
  Given no valid auth cookie is present
  When the browser requests a protected UI API endpoint
  Then the response status is 401

Scenario: Logout
  Given a valid auth cookie is present
  When the browser posts to /logout
  Then the response status is 204
  And the session store entry is removed
  And the auth cookie is cleared with Max-Age=0

Scenario: Authentication handler accepts a valid session
  Given the request contains a valid "__Host-app-auth" cookie
  And the session store contains a non-expired session entry
  When the authentication handler runs
  Then HttpContext.User is authenticated with authentication type "app-cookie"
  And the principal contains only a NameIdentifier claim with the user id

Scenario: Authentication handler rejects an expired session
  Given the request contains a "__Host-app-auth" cookie for an expired session
  When the authentication handler runs
  Then the session store entry is removed
  And authentication fails

Scenario: Login succeeds behind trusted reverse proxy
  Given exactly one password-bearing user has a stored password_hash for the submitted password
  And Gateway receives internal HTTP from Caddy
  And trusted forwarded headers set the effective public scheme to https
  And the Origin matches the effective public origin
  When the browser posts the password to /login
  Then the response status is 204
  And the auth cookie attributes remain unchanged

Scenario: Login rejects ineffective HTTPS context
  Given exactly one password-bearing user has a stored password_hash for the submitted password
  And forwarded-header processing leaves the effective public scheme as http
  When the browser posts the password to /login
  Then the response status is 403
  And no auth cookie is issued

Scenario: Unsafe request rejects public origin mismatch
  Given a valid auth cookie is present
  When an unsafe request supplies an Origin or Referer with a mismatched scheme, host, or effective port
  Then the response status is 403
  And the endpoint handler is not reached

Scenario: Telegram sender mapping resolves a user
  Given users.telegram_user_id maps a Telegram sender to user "U1"
  When Gateway accepts a Telegram webhook from that sender
  Then downstream article and job ownership use user "U1"

Scenario: Unknown Telegram sender is ignored
  Given no users.telegram_user_id row matches a Telegram sender
  When Gateway receives a Telegram webhook from that sender
  Then no user, article, job, or notification state is created or mutated
```

## Data and State

SQLite remains the source of truth for authenticated users.

### `users`

- `id`: ULID, seeded as `01ASB2XFCZJY7WHZ2FNRTMQJCT`.
- `telegram_user_id`: nullable until bootstrap or another canonical user-provisioning path maps a Telegram sender; unique when present. Auth bootstrap sets the personal row to `1559957191` only when this value is null.
- `password_hash`: nullable only before auth bootstrap completes; after bootstrap it stores an Argon2id PHC string.

The system starts with one bootstrapped user row. Additional password-bearing rows may exist through canonical provisioning outside auth, and password-only login supports them by matching the submitted password against all non-empty Argon2id PHC hashes. Gateway runtime identity is data-driven after bootstrap: password login uses the exactly matched password owner, session authentication uses the stored session user id, and Telegram sender authorization uses `users.telegram_user_id`. Auth does not introduce registration, provisioning state, roles, tenants, external identity tables, password history, refresh tokens, or user self-service.

### `ISessionStore`

Gateway auth sessions are server-side state keyed by the opaque cookie value.

```csharp
public interface ISessionStore
{
    Task<SessionEntry?> GetAsync(string sessionId, CancellationToken ct);
    Task SetAsync(string sessionId, SessionEntry entry, CancellationToken ct);
    Task RemoveAsync(string sessionId, CancellationToken ct);
}

public sealed record SessionEntry(
    string UserId,
    DateTimeOffset CreatedAt,
    DateTimeOffset AbsoluteExpiresAt);
```

The implementation is in-memory and may use `ConcurrentDictionary<string, SessionEntry>`. `RemoveAsync` on a missing key is a no-op. Expired entries are removed on lookup and may also be removed by periodic sweep.

## Interfaces

- `POST /login`
  - Request body: `{ "password": string }`.
  - The request must be `POST`, same-origin, and have effective public scheme `https`.
  - The effective public scheme is `HttpRequest.Scheme` after forwarded-header processing.
  - Per-IP and global login throttling run before Argon2id verification.
  - Success: always remove any existing valid session from the request cookie, load all non-empty Argon2id PHC password hashes, verify every candidate, create a fresh session id for the matched user's `id` only when exactly one candidate matches, store a `SessionEntry`, and return `204 No Content` with `Set-Cookie`.
  - Failure: `401 Unauthorized`; the response must not disclose whether bootstrap is missing, no hash matched, duplicate hashes matched, the password is wrong, or throttling contributed.
- `POST /logout`
  - Removes the matching session-store entry when present.
  - Always clears the auth cookie with `Set-Cookie: __Host-app-auth=; HttpOnly; Secure; SameSite=Strict; Path=/; Max-Age=0`.
  - Returns `204 No Content`.
- `GET /auth/session`
  - Authenticated: `204 No Content`.
  - Unauthenticated: `401 Unauthorized`.
- `AddAppCookie()`
  - Registers the custom auth handler in the standard ASP.NET Core authentication pipeline.
  - Default scheme name: `"app-cookie"`.
  - Cookie name and session lifetime are not configurable through the auth scheme options; consumers use `AppCookieDefaults.CookieName` and `AppCookieDefaults.SessionLifetime` directly.

```csharp
public static AuthenticationBuilder AddAppCookie(this AuthenticationBuilder builder);
```

Wire-up:

```csharp
services.AddSingleton<ISessionStore, InMemorySessionStore>();
services.AddAuthentication("app-cookie").AddAppCookie();
services.AddAuthorization();
// app.UseAuthentication(); app.UseAuthorization();
```

- Authentication handler behavior:
  - Missing `__Host-app-auth` cookie returns `AuthenticateResult.NoResult()`.
  - Unknown session id returns authentication failure without a client-visible distinction from expiry.
  - Expired session entries are removed and authentication fails.
  - Valid session entries produce an `AuthenticationTicket` whose principal has one `ClaimTypes.NameIdentifier` claim and authentication type `"app-cookie"`.
  - The handler does not issue, clear, rotate, or refresh cookies.
- Configuration:
  - `AUTH_BOOTSTRAP_PASSWORD`: one-time 2048-character bootstrap password. Secret.
  - `SQLITE_PATH`: SQLite database path.
  - `GATEWAY_PUBLIC_HOSTS`: comma-separated public host allowlist for trusted forwarded host values. Not secret.
  - Gateway reads these standalone logical keys from configuration providers or `ARCHIVIST_`-prefixed environment variables, for example `ARCHIVIST_AUTH_BOOTSTRAP_PASSWORD`, `ARCHIVIST_SQLITE_PATH`, and `ARCHIVIST_GATEWAY_PUBLIC_HOSTS`.

## Dependencies

Depends on:

- `telegram-ingestion` persistence foundation for the shared `users` table contract.
- `docs/ARCHITECTURE.md` bootstrapped-user gateway/UI boundary.
- `docs/ARCHITECTURE.md` reverse-proxy deployment topology.

Implementation agents should use `.agents/skills/archivist-gateway/SKILL.md` and `.agents/skills/archivist-ui/SKILL.md` for module coding guidance. These skills are not feature dependencies or rebuild sources of truth.

Impacts:

- Gateway API authentication middleware and endpoint mapping.
- Gateway persistence initialization for the `users` table.
- UI auth API client contract.
- Telegram ingestion user-table expectations.
- Gateway Telegram sender identity resolution.

## Rebuild Notes

- Existing code is not authoritative; rebuilds must follow this spec and linked tasks.
- Do not replace opaque session cookies with encrypted cookie payloads or custom signed cookie formats without a new design decision.
- Do not add sliding expiry, refresh tokens, multiple concurrent sessions per user, or revocation lists beyond session entry removal.
- Do not encrypt, sign, MAC, or otherwise transform the cookie value.
- Gateway restart invalidating cookies is intentional with the in-memory session store.
- Do not implement final Preact login page rendering in this feature; `docs/specs/ui/SPEC.md` owns browser UI routes and rendering behavior.
- The trusted reverse proxy must overwrite forwarded headers. Do not expose Gateway directly to the public Internet while trusting forwarded headers.

## Security / Privacy Notes

- The 2048-character password behaves like a bearer token. It must be generated outside the app with a CSPRNG and a printable ASCII alphabet.
- `AUTH_BOOTSTRAP_PASSWORD`, stored password hashes, session ids, auth cookies, and `Set-Cookie` headers are secret material and must not be logged.
- The cookie value contains no user id, role, expiry, or metadata. Store presence and server-side expiry determine validity.
- SameSite cookies are not the only CSRF control. Unsafe requests must also enforce same-origin request checks.
- Same-origin request checks must compare the post-forwarding effective public scheme, host, and port.
- Login throttling is in-memory and relies on the single gateway instance deployment.

## Observability / Logging Notes

- Log login success/failure as aggregate security events without the submitted password, hash, cookie, or bootstrap value.
- Login failure logs may include remote address and throttle state when available.

## Open Questions

- None.

## Related Documents

- `./PLAN.md`
- `./tasks/AUTHN-002-password-persistence-and-bootstrap.md`
- `./tasks/AUTHN-003-gateway-cookie-authentication-endpoints.md`
- `./tasks/AUTHN-004-protect-ui-api-and-validate-auth-client-contract.md`
