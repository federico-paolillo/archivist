---
id: AUTHN
slug: authn
title: UI/API Authentication
status: done
owner: null
depends_on: [telegram-ingestion]
impacts: [gateway, ui, sqlite]
canonical: true
---

# Feature: UI/API Authentication

## Intent

Protect the Archivist web UI and UI-facing gateway API with single-user cookie authentication.

## Motivation

Archivist v0 exposes a private browser UI for article review and administration. The UI must be usable from a normal browser without exposing credentials to JavaScript, while keeping the implementation small enough for a single VPS deployment.

## Scope

In scope:

- Password-only login for the fixed personal Archivist user.
- Argon2id password hash storage in SQLite.
- One-time bootstrap of the personal user's password hash from `AUTH_BOOTSTRAP_PASSWORD`.
- Opaque server-issued session id cookies integrated through a custom ASP.NET Core authentication handler.
- Browser-session auth cookies with `HttpOnly`, `Secure`, `SameSite=Strict`, `Path=/`, no `Domain`, and the `__Host-app-auth` cookie name.
- In-memory v0 `ISessionStore` with 24-hour server-side absolute expiry.
- `POST /login`, `POST /logout`, and `GET /auth/session`.
- Protection of UI-facing API endpoints.
- In-memory login throttling for the single gateway instance.
- Basic same-origin defenses for unsafe methods.
- Forwarded-header handling for the trusted Docker reverse proxy deployment topology.
- Auth endpoint/session contracts consumed by the final browser UI.

## Out of Scope

Not included:

- Account registration, password reset, password rotation UI, user self-service, roles, or tenant isolation.
- Multiple authorized UI users.
- Multi-replica auth operation in v0.
- Custom HMAC cookies, random cookie names, or user-id hash cookie payloads.
- Sliding expiry, refresh tokens, multiple concurrent sessions per user, or revocation lists beyond plain session entry removal.
- Encryption, signing, MACs, or other cryptographic transforms over the cookie value.
- Final Preact browser login rendering, login-failure route, article shell, and logout menu, which are owned by the `ui` feature.

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
- REQ-006: The personal user row must use `id = 01ASB2XFCZJY7WHZ2FNRTMQJCT`.
- REQ-007: `AUTH_BOOTSTRAP_PASSWORD` must be accepted only when the personal user has no stored password hash.
- REQ-008: The bootstrap password must not be logged, returned, or persisted in plaintext.
- REQ-009: Successful login must issue an opaque session cookie named `__Host-app-auth`.
- REQ-010: The session id must be 32 bytes from `RandomNumberGenerator.GetBytes`, base64url-encoded without padding.
- REQ-011: The browser cookie must not set `Expires` or `Max-Age`.
- REQ-012: The auth cookie must be `HttpOnly`, `Secure`, `SameSite=Strict`, use `Path=/`, and omit `Domain`.
- REQ-013: `ISessionStore` must map `sessionId` to `{ userId, createdAt, absoluteExpiresAt }`.
- REQ-014: Session expiry must be absolute and server-side enforced at 24 hours from issue.
- REQ-015: The v0 `ISessionStore` implementation must be in-memory and may use `ConcurrentDictionary<string, SessionEntry>`.
- REQ-016: Gateway restart invalidates existing v0 sessions by wiping the in-memory store.
- REQ-017: Gateway auth must integrate with ASP.NET Core through a custom `IAuthenticationHandler`, or `AuthenticationHandler<AppCookieOptions>`, registered by `AddAppCookie()`.
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

## Acceptance Criteria

```gherkin
Feature: UI/API authentication

Scenario: Bootstrap stores the personal password hash
  Given the personal user has no password_hash
  And AUTH_BOOTSTRAP_PASSWORD is a valid 2048-character printable ASCII secret
  When the gateway starts
  Then the users table contains the personal user
  And password_hash is an Argon2id PHC string
  And the plaintext bootstrap secret is not logged or stored

Scenario: Successful login
  Given the personal user has a stored password_hash for the submitted password
  When the browser posts the password to /login
  Then the response status is 204
  And Set-Cookie contains "__Host-app-auth"
  And the cookie is HttpOnly, Secure, SameSite=Strict, and Path=/
  And the cookie has no Domain, Expires, or Max-Age
  And the session store contains a fresh session entry expiring 24 hours after issue

Scenario: Invalid login
  Given the personal user has a stored password_hash
  When the browser posts the wrong password to /login
  Then the response status is 401
  And no auth cookie is issued
  And throttle counters are incremented

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
  Given the personal user has a stored password_hash for the submitted password
  And Gateway receives internal HTTP from Caddy
  And trusted forwarded headers set the effective public scheme to https
  And the Origin matches the effective public origin
  When the browser posts the password to /login
  Then the response status is 204
  And the auth cookie attributes remain unchanged

Scenario: Login rejects ineffective HTTPS context
  Given the personal user has a stored password_hash for the submitted password
  And forwarded-header processing leaves the effective public scheme as http
  When the browser posts the password to /login
  Then the response status is 403
  And no auth cookie is issued

Scenario: Unsafe request rejects public origin mismatch
  Given a valid auth cookie is present
  When an unsafe request supplies an Origin or Referer with a mismatched scheme, host, or effective port
  Then the response status is 403
  And the endpoint handler is not reached
```

## Data and State

SQLite remains the source of truth for the personal user.

### `users`

- `id`: ULID, seeded as `01ASB2XFCZJY7WHZ2FNRTMQJCT`.
- `telegram_user_id`: nullable until Telegram ingestion maps the configured Telegram user; unique when present.
- `password_hash`: nullable only before auth bootstrap completes; after bootstrap it stores an Argon2id PHC string.

The v0 system has one user row. Auth does not introduce provisioning state, roles, tenants, external identity tables, password history, refresh tokens, or roles.

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

The v0 implementation is in-memory and may use `ConcurrentDictionary<string, SessionEntry>`. `RemoveAsync` on a missing key is a no-op. Expired entries are removed on lookup and may also be removed by periodic sweep.

## Interfaces

- `POST /login`
  - Request body: `{ "password": string }`.
  - The request must be `POST`, same-origin, and have effective public scheme `https`.
  - The effective public scheme is `HttpRequest.Scheme` after forwarded-header processing.
  - Per-IP and global login throttling run before Argon2id verification.
  - Success: always remove any existing valid session from the request cookie, create a fresh session id, store a `SessionEntry`, and return `204 No Content` with `Set-Cookie`.
  - Failure: `401 Unauthorized`; the response must not disclose whether bootstrap is missing, the password is wrong, or throttling contributed.
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
  - Default cookie name: `"__Host-app-auth"`.
  - Default session lifetime: `TimeSpan.FromHours(24)`.

```csharp
public static AuthenticationBuilder AddAppCookie(
    this AuthenticationBuilder builder,
    Action<AppCookieOptions>? configure = null);
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

## Dependencies

Depends on:

- `telegram-ingestion` persistence foundation for the shared `users` table contract.
- `docs/ARCHITECTURE.md` v0 single-user gateway/UI boundary.
- `docs/ARCHITECTURE.md` reverse-proxy deployment topology.
- `docs/conventions/GATEWAY.md` ASP.NET Core Minimal API conventions.
- `docs/conventions/UI.md` Preact/Vite conventions.

Impacts:

- Gateway API authentication middleware and endpoint mapping.
- Gateway persistence initialization for the `users` table.
- UI auth API client contract.
- Telegram ingestion user-table expectations.

## Rebuild Notes

- Existing code is not authoritative; rebuilds must follow this spec and linked tasks.
- Do not replace opaque session cookies with encrypted cookie payloads or custom signed cookie formats without a new design decision.
- Do not add sliding expiry, refresh tokens, multiple concurrent sessions per user, or revocation lists beyond session entry removal.
- Do not encrypt, sign, MAC, or otherwise transform the cookie value.
- Gateway restart invalidating cookies is intentional in v0.
- Do not implement final Preact login page rendering in this feature; `docs/specs/ui/SPEC.md` owns browser UI routes and rendering behavior.
- The trusted reverse proxy must overwrite forwarded headers. Do not expose Gateway directly to the public Internet while trusting forwarded headers.

## Security / Privacy Notes

- The 2048-character password behaves like a bearer token. It must be generated outside the app with a CSPRNG and a printable ASCII alphabet.
- `AUTH_BOOTSTRAP_PASSWORD`, stored password hashes, session ids, auth cookies, and `Set-Cookie` headers are secret material and must not be logged.
- The cookie value contains no user id, role, expiry, or metadata. Store presence and server-side expiry determine validity.
- SameSite cookies are not the only CSRF control. Unsafe requests must also enforce same-origin request checks.
- Same-origin request checks must compare the post-forwarding effective public scheme, host, and port.
- Login throttling is in-memory and is acceptable only because v0 has one gateway instance.

## Multiple-Replica Notes

To support multiple gateway replicas:

- Replace the in-memory `ISessionStore` with a shared implementation. Redis is preferred over memcached because eviction semantics under memory pressure are safer for auth state.
- Move login throttling counters to a shared store.
- Keep the fixed `__Host-app-auth` cookie name and opaque session id format.
- No shared cookie key ring is required because the cookie carries no protected payload.

## Observability / Logging Notes

- Log login success/failure as aggregate security events without the submitted password, hash, cookie, or bootstrap value.
- Login failure logs may include remote address and throttle state when available.

## Open Questions

- None.

## Related Documents

- `./PLAN.md`
- `./DIARY.md`
- `./plans/AUTHN-002-password-persistence-and-bootstrap.execplan.md`
- `./plans/AUTHN-003-gateway-cookie-authentication.execplan.md`
- `./tasks/AUTHN-006-reverse-proxy-forwarded-headers-and-effective-https-auth.md`
- `./plans/AUTHN-006-reverse-proxy-forwarded-headers.execplan.md`
