---
feature: authn
status: done
canonical: true
---

# Feature Plan: UI/API Authentication

## Purpose

This file is the feature-level implementation control board for the final UI/API authentication surface.

---

## Task DAG

```text
TELING-001 -> AUTHN-002 -> AUTHN-003 -> AUTHN-004
AUTHN-004 -> UIEND-002
AUTHN-004 -> UIEND-003
AUTHN-004 -> UI-002
```

---

## Execution Phases

### Phase 1: Gateway Persistence

- `AUTHN-002` defines the `users` password fields, personal-user bootstrap, Argon2id password hashing, bootstrap Telegram sender mapping, and password verification inputs.

### Phase 2: Gateway Authentication Runtime

- `AUTHN-003` implements opaque session cookie endpoints, `ISessionStore`, the custom `"app-cookie"` authentication handler, login throttling, trusted forwarded-header processing, effective public HTTPS validation, and same-origin checks for unsafe methods.

### Phase 3: Final Auth Contract Validation

- `AUTHN-004` validates the protected UI-facing Gateway contract consumed by `ui-endpoints` and `ui`, including authenticated API `401/403` behavior, auth endpoint behavior, effective public scheme/host/port handling, cookie attributes, and same-origin rejection.

---

## Tasks

| ID | Task | Status | Depends On | Blocks | Parallel | ExecPlan |
|---|---|---|---|---|---|---|
| `AUTHN-002` | Password persistence and bootstrap | done | `TELING-001` | `AUTHN-003` | no | null |
| `AUTHN-003` | Gateway opaque session cookie authentication | done | `AUTHN-002` | `AUTHN-004` | no | null |
| `AUTHN-004` | Protect UI API and validate auth client contract | done | `AUTHN-003` | `UIEND-002`, `UIEND-003`, `UI-002` | no | null |

---

## Concurrency Rules

- Auth implementation tasks are sequenced because they modify the same schema, middleware, endpoint routing, request-origin interpretation, and tests.
- `AUTHN-002` must wait for `TELING-001` because it extends the shared persistence foundation.
- `AUTHN-004` is the final auth dependency for UI article endpoints and browser UI work.
- UI endpoint and browser UI tasks must not depend on pre-proxy auth behavior; they consume the post-forwarding, effective-HTTPS auth contract validated by `AUTHN-004`.
- Telegram persistence work must preserve `users.password_hash` and nullable-at-rest `users.telegram_user_id`.

---

## Blocking Interfaces or Schemas

- `users.password_hash` Argon2id PHC storage.
- `users.telegram_user_id` as the persisted Telegram sender mapping.
- `ISessionStore` and `SessionEntry`.
- `AppCookieAuthenticationHandler`, `AppCookieDefaults`, and `AddAppCookie()`.
- `POST /login`, `POST /logout`, and `GET /auth/session`.
- Cookie name and attributes for `__Host-app-auth`.
- 32-byte base64url opaque session id generation and 24-hour server-side absolute expiry.
- Password-only exact-one-match login across all password-bearing users.
- Same-origin rejection behavior for unsafe methods.
- Effective public scheme, host, and port after trusted forwarded-header processing.
- `GATEWAY_PUBLIC_HOSTS`, `AUTH_BOOTSTRAP_PASSWORD`, and `SQLITE_PATH` standalone logical keys, supplied to Gateway through configuration providers or `ARCHIVIST_`-prefixed environment variables.

---

## Validation Sequence

1. Run gateway auth bootstrap and endpoint tests.
2. Run gateway build and full test suite.
3. Verify password hash bootstrap, bootstrap Telegram sender seeding, exact-one-match login, duplicate-match failure, forwarded `https` login success, effective `http` login `403`, forwarded public host constraints, cookie attributes, session rotation, server-side expiry, logout store removal, oversized login rejection, throttling, protected endpoint `401`, and same-origin rejection.

Validation commands:

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

---

## Open Planning Questions

- None.

---

## Completion Criteria

The feature is complete when:

- all required tasks are `done`;
- acceptance criteria in `SPEC.md` are satisfied;
- validation sequence passes or failures are recorded;
- durable implementation decisions have been promoted to canonical documents;
- `docs/specs/INDEX.md` reflects the final feature status.
