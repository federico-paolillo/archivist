---
name: archivist-gateway
description: Use when implementing, reviewing, or planning Archivist Gateway changes under src/gateway, including ASP.NET Core Minimal API routes, EF Core persistence, authentication, configuration, forwarded headers, artifact reads, and Gateway validation.
---

# Archivist Gateway

Use this skill for Gateway work under `src/gateway/`.

## Required Context

Start with the orientation bundle:

```text
AGENTS.md
docs/REBUILD.md
docs/specs/INDEX.md
```

Load canonical docs by task trigger:

- `docs/ARCHITECTURE.md`: executables, service boundaries, storage, deployment, configuration, routing, forwarded headers, or artifact access.
- `docs/DESIGN.md`: accepted durable decisions.
- `docs/ERRORS.md`: ARC public error transport or display.
- `docs/ARTIFACTS.md`: article artifact read/delete behavior.
- Relevant feature `SPEC.md`, `PLAN.md`, task files, and linked ExecPlans before implementation.

## Stack

- Gateway code lives under `src/gateway/`.
- Target: .NET 10 with C# 14.
- `src/gateway/Directory.Build.props` enables nullable reference types, implicit usings, latest analysis, and warnings as errors.
- Use EF Core for database access.

## Coding Rules

- Use file-scoped namespaces.
- Prefer sealed classes and records.
- Do not add unnecessary `using` directives; implicit usings are enabled.
- Add `/// <summary>` comments on classes and interfaces stating their purpose.
- Use exceptions for invalid programmer input, impossible states, and infrastructure failures outside the expected application contract.
- Use lightweight, test-driven interfaces only when they isolate external dependencies or make meaningful tests possible.
- Default implementations of interfaces go under a `Defaults/` folder of the parent feature folder.
- Register feature services through `ServiceCollectionExtensions` methods. Do not scatter dependency registration.
- Application features should stay organized by domain area.

## Minimal APIs

- Route modules live under `Archivist.Gateway.Api/<Feature>/`.
- Use `Endpoints.cs` for route-group mapping and `Handlers.cs` for static handler methods.
- Put HTTP DTOs under `Models/`.
- Gateway route contracts are unprefixed, for example `/articles`, `/login`, and `/auth/session`. `/api` is a frontend/proxy convention.
- Prefer typed results from `Microsoft.AspNetCore.Http.HttpResults`, including high-level `TypedResults` methods such as `TypedResults.InternalServerError()`.
- Map expected application problems to appropriate HTTP responses at the API boundary.

## Authentication And Forwarding

- Follow the auth and forwarded-header contracts in canonical specs and `docs/ARCHITECTURE.md`.
- UI/API auth routes are `POST /login`, `POST /logout`, and `GET /auth/session`.
- Use the custom `"app-cookie"` authentication handler registered through `AddAppCookie()`.
- Cookie auth must return `401` or `403` for API requests instead of redirecting.
- Unsafe HTTP methods must reject cross-site requests before endpoint handling.
- Startup must enable trusted forwarded-header processing before authentication, authorization, and endpoint mapping.

## Configuration

- Keep ASP.NET Core default application configuration sources.
- Create the builder without command-line arguments.
- Append `builder.Configuration.AddEnvironmentVariables("ARCHIVIST_")`.
- Keep expected Gateway configuration keys and sections in `Settings.cs`.
- Production code must not scatter raw configuration-key literals.
- When adding Gateway configuration, update `docs/ARCHITECTURE.md` and affected specs/tasks.

## Persistence And Artifacts

- EF Core entity classes live under `Archivist.Gateway.Application/Entities`.
- Use `AsNoTracking()` for read-only EF projections.
- Do not add migrations unless the persistence schema changes.
- Gateway artifact access is read-only except for the article hard-delete behavior specified by canonical docs.
- Artifact paths and delete semantics must follow `docs/ARTIFACTS.md`.

## Testing And Validation

- Tests are mandatory for backend behavior changes.
- Prefer integration tests when the real DI graph matters.
- Prefer focused unit tests for pure helpers, value objects, builders, encoders, and renderers.
- API integration tests should use proper request DTOs and typed HTTP helpers where possible.

Run from `src/gateway/`:

```bash
dotnet format
dotnet build
dotnet test
```

## Output

Report:

- task ID when applicable;
- Gateway areas changed;
- API/auth/persistence/artifact impact;
- canonical docs updated or why none were needed;
- validation commands and results;
- blockers or follow-ups.
