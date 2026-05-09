---
feature: authn
status: draft
canonical: true
---

# Feature Plan: UI/API Authentication

## Purpose

This file is the feature-level implementation control board for the v0 UI/API authentication surface.

---

## Task DAG

```text
AUTHN-001 -> AUTHN-002 -> AUTHN-003 -> AUTHN-004 -> AUTHN-005
```

---

## Execution Phases

### Phase 1: Canonical Contracts

- `AUTHN-001` creates the authn feature artifacts and promotes durable auth decisions to architecture, design, and conventions.

### Phase 2: Gateway Foundations

- `AUTHN-002` defines password persistence, bootstrap behavior, and Argon2id verification.
- `AUTHN-003` implements opaque session cookie endpoints, `ISessionStore`, the custom `"app-cookie"` authentication handler, throttling, and same-origin checks.

### Phase 3: UI API Protection And Validation

- `AUTHN-004` validates protected UI-facing gateway endpoint behavior and preserves the auth endpoint contract consumed by the final UI.
- `AUTHN-005` performs the security validation pass and records completion.

---

## Tasks

| ID | Task | Status | Depends On | Blocks | Parallel | ExecPlan |
|---|---|---|---|---|---|---|
| `AUTHN-001` | Authn canonical docs and design decisions | done | - | `AUTHN-002` | no | - |
| `AUTHN-002` | Password persistence and bootstrap | done | `AUTHN-001`, `TELING-001` | `AUTHN-003` | no | `plans/AUTHN-002-password-persistence-and-bootstrap.execplan.md` |
| `AUTHN-003` | Gateway opaque session cookie authentication | blocked | `AUTHN-002` | `AUTHN-004`, `UIEND-002`, `UIEND-003` | no | `plans/AUTHN-003-gateway-cookie-authentication.execplan.md` |
| `AUTHN-004` | Protect UI API and validate auth client contract | blocked | `AUTHN-003` | `AUTHN-005`, `UI-002` | no | - |
| `AUTHN-005` | Security validation pass | blocked | `AUTHN-004` | - | no | - |

---

## Concurrency Rules

- Auth implementation tasks are sequenced because they modify the same schema, auth middleware, and API routes.
- `AUTHN-002` must wait for `TELING-001` unless the implementer explicitly updates the shared persistence foundation first.
- Future UI API feature tasks must treat `AUTHN-003` as a blocking gateway contract.
- The browser auth shell must wait for `AUTHN-004` because that task validates the UI auth client contract consumed by `UI-002`.
- Future Telegram persistence work must preserve `users.password_hash` and nullable-at-rest `users.telegram_user_id`.

---

## Blocking Interfaces or Schemas

- `users.password_hash` Argon2id PHC storage.
- `ISessionStore` and `SessionEntry`.
- `AppCookieAuthenticationHandler`, `AppCookieOptions`, and `AddAppCookie()`.
- `POST /login`, `POST /logout`, and `GET /auth/session`.
- Cookie name and attributes for `__Host-app-auth`.
- 32-byte base64url opaque session id generation and 24-hour server-side absolute expiry.
- Same-origin rejection behavior for unsafe methods.
- `AUTH_BOOTSTRAP_PASSWORD` and `SQLITE_PATH`.

---

## Validation Sequence

1. Run gateway auth bootstrap and endpoint tests.
2. Run gateway build and full test suite.
3. Verify cookie attributes, session rotation, server-side expiry, logout store removal, oversized login rejection, throttling, protected endpoint `401`, and same-origin rejection.

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

- all required tasks are `done` or explicitly `skipped`;
- acceptance criteria in `SPEC.md` are satisfied;
- validation sequence passes or failures are recorded;
- durable implementation decisions have been promoted to canonical documents;
- `DIARY.md` contains final implementation notes;
- `docs/specs/INDEX.md` reflects the final feature status.
