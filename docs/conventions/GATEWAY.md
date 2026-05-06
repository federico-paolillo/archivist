# GATEWAY.md

Describes conventions and best-practices when working on the Gateway module.

## In general

Backend code lives under `src/gateway/` and targets .NET 10 with C# 14. `src/gateway/Directory.Build.props` enables nullable reference types, implicit usings, latest analysis, and warnings as errors.

Conventions:

- Do not add unnecessary `using` directives; implicit usings are enabled.
- Add `/// <summary>` comments on classes and interfaces stating their purpose.
- Use exceptions for invalid programmer input, impossible states, and infrastructure failures that are not part of the expected application contract.
- Default implementations of interfaces go under a `Defaults/` folder of the parent feature folder.
- Keep interfaces lightweight and test-driven. Do not introduce abstractions for one-off code unless they isolate an external dependency or make meaningful tests possible.
- Keep application features organized by domain area, for example `Articles`, `Orders`, and `Receipts`.
- Register feature services through `ServiceCollectionExtensions` methods instead of scattering registrations in unrelated projects.
- **Always prefer** using `TypeResults` high-level methods from ASP .NET Minimal APIs. **Do** `TypedResults.InternalServerError()` instead of `TypedResults.StatusCode(StatusCodes.Status500InternalServerError)`
- Make ad-hoc `ServiceCollectionExtensions` extensions and provide `AddXxx()` methods to register dependencies. Don't scatter around the codebase dependencies registration code.
- Application code can loosely follow the Transaction Script patter by making an `xxxHandler` class for every use case.
- Always use file scoped namespace declarations.
- Prefer sealing classes and records.

## Minimal APIs

- Route modules live under `Archivist.Gateway.Api/<Feature>/`.
- Use `Endpoints.cs` for route-group mapping and `Handlers.cs` for static handler methods.
- Put HTTP DTOs under `Models/`.
- Keep backend routes unprefixed: e.g. `/articles`, `/orders`, and `/devices`. `/api` is a frontend/proxy convention.
- Prefer typed results from `Microsoft.AspNetCore.Http.HttpResults`.
- Map expected application problems to appropriate HTTP responses at the API boundary.

## Authentication

- UI/API auth routes are unprefixed: `POST /login`, `POST /logout`, and `GET /auth/session`.
- Use the custom `"app-cookie"` authentication handler registered through `AddAppCookie()` for browser auth.
- The auth cookie name is `__Host-app-auth`; set `HttpOnly`, `Secure`, `SameSite=Strict`, `Path=/`, and no `Domain`.
- The cookie value is an opaque 32-byte CSPRNG session id encoded as base64url without padding. Do not encrypt, sign, MAC, or embed user/session metadata in the cookie value.
- Auth sessions are stored behind `ISessionStore`. The v0 implementation is in-memory; multi-replica deployments must replace it with Redis or another explicit shared store before adding gateway replicas.
- Cookie auth must return `401` or `403` for API requests instead of redirecting to login or access-denied pages.
- Unsafe HTTP methods must reject cross-site requests before endpoint handling.
- Login verification must validate shape and request size before Argon2id work and must use in-memory throttling in v0.

## Persistence

- EF Core entity classes live under `Archivist.Gateway.Application/Entities`.
- Use `AsNoTracking()` for read-only EF projections.
- Do not add migrations unless the persistence schema changes.
- Auth persistence owns `users.password_hash` and may ensure the personal user row exists before Telegram ingestion maps `telegram_user_id`.

## Artifact Reads

Gateway may read article artifacts under `DATA_DIR` through a read-only artifact abstraction. This abstraction must not expose write, create, rename, or delete operations. Outside the UI article hard-delete path, Gateway code must not mutate `/data` artifacts; Worker owns artifact production and deletion behavior defined by feature specs.

Terminal success notification dispatch reads `{DATA_DIR}/articles/{article_id}/summary.md` once summary generation is implemented. Missing or unreadable summary artifacts fail notification delivery without changing terminal article or job state.

Article hard delete is the only v0 Gateway exception to read-only artifact access. It must use a separate admin-delete cleanup abstraction scoped to deleting `{DATA_DIR}/articles/{article_id}/` for a validated article ULID after authenticated ownership and running-job checks. Do not add general write, create, rename, or arbitrary delete operations to the read-only artifact abstraction.

## Testing

- Tests are mandatory for backend code changes. Cover at least the happy path and the main failure path when the change adds behavior.
- Use integration tests for application/database/API behavior when the real DI graph matters.
- Use focused unit tests for pure helpers, value objects, builders, encoders, and renderers.
- In API integration tests, prefer proper request DTO types and typed HTTP helpers such as `PostAsJsonAsync`. Use raw JSON or `StringContent` only for malformed payloads, unknown fields, or explicit serialization-boundary tests.
- Run backend verification from `src/gateway/`:

```bash
dotnet format
dotnet build
dotnet test
```
