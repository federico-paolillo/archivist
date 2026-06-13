# DESIGN.md

Records accepted durable decisions that must survive rebuilds.

A decision discovered during implementation should be promoted here if it affects more than one task or must remain true across rebuilds.

---

## Decision Record Format

Each decision can be as lightweight as a heading plus rationale paragraph, or as structured as a full ADR. Use the level of ceremony that matches the stakes.

Suggested minimal format:

```text
### DSGN-NNN: <Title>

**Date:** YYYY-MM-DD
**Status:** accepted

**Context:** Why this decision was needed.

**Decision:** What was decided.

**Consequences:** What changes as a result. What becomes easier or harder.
```

All decisions in this file are accepted and binding for rebuild.

---

## Decisions

### DSGN-001: Reliability and Simplicity Over Feature Breadth

**Date:** 2026-04-30
**Status:** accepted

**Context:** Archivist is a personal article archiving system intended to be reliable and easy to rebuild.

**Decision:** Archivist prioritizes deterministic processing, clear failure states, minimal dependencies, and a small deployment surface. The runtime surface is the core archive-review loop: Telegram ingestion, SQLite persistence, Worker processing, filesystem artifacts, summary generation, authenticated browser review, deletion/recovery actions, backups, and OpenTelemetry observability.

**Consequences:** The system stays small enough to rebuild from canonical documents. Capabilities outside the current runtime surface require canonical specs and design updates before implementation.

### DSGN-002: Bootstrapped Account and User-Aware Runtime Ownership

**Date:** 2026-04-30
**Status:** accepted

**Context:** Archivist serves one bootstrapped personal account while persisting user ownership on articles, jobs, sessions, and Telegram sender mappings.

**Decision:** Authentication bootstrap is the production code path allowed to hardcode the initial personal user id and fixed personal Telegram sender id `1559957191`. Bootstrap sets the personal row's `telegram_user_id` to `1559957191` only when it is null and preserves an existing non-null value; it does not read deployment-configured Telegram sender allowlists or personal sender settings for that seed. Worker CLI enqueue is the only runtime exception: it uses `jobs.DefaultUserID = 01ASB2XFCZJY7WHZ2FNRTMQJCT`, checks `users.id` by that id, never infers ownership from user-table cardinality, and never creates the user. Other runtime Gateway and Worker behavior resolves `user_id` from persisted user rows, authenticated session state, claimed jobs, or article ownership. Password login loads all non-empty Argon2id PHC password hashes, verifies every candidate, and issues a session only when exactly one user matches the submitted password. Multiple password-bearing rows are valid; duplicate password matches fail closed. Telegram webhook authorization is based on the existence of a `users.telegram_user_id` mapping; unknown Telegram senders receive no reply and create no rows. Gateway and Worker attach `user_id` to logs and traces when the Archivist user is known, using the exact key `user_id`; Snapshotter remains user-agnostic.

**Consequences:** The existing personal account remains bootstrapped. Runtime ownership does not rely on a hardcoded personal-id fallback except for Worker CLI enqueue's explicit default-user existence check. Additional user rows, password hashes, and Telegram mappings can exist without changing Gateway, UI, or Worker processing ownership semantics; password-only login remains ambiguous only when the submitted password matches multiple users, in which case it fails closed. Worker CLI enqueue targets the bootstrapped personal account unless a new canonical decision changes that command.

### DSGN-003: SQLite Owns State and Filesystem Owns Artifacts

**Date:** 2026-04-30
**Status:** accepted

**Context:** Article records and job state need transactional updates, while raw HTML, Markdown, and summaries are better stored as files.

**Decision:** SQLite is the authoritative store for user, article, job, notification, and error state. The filesystem stores raw and derived artifacts under `/data/articles/{article_id}/`, with filenames defined in `docs/ARTIFACTS.md`.

**Consequences:** Rebuild-critical state must be recoverable from SQLite plus deterministic artifact paths derived from `DATA_DIR` and `article_id`. Artifact writes use temporary files followed by rename. SQLite does not store artifact path columns.

### DSGN-005: Preserve Raw and Derived Article Data

**Date:** 2026-04-30
**Status:** accepted

**Context:** Extraction and summarization may fail or require diagnosis during operation. Raw input is needed for debugging and reprocessing by explicit operator action.

**Decision:** The Worker stores the raw HTML snapshot and derived Markdown and summary artifacts for each processed article.

**Consequences:** The system can expose failure state clearly. Storage usage grows with retained artifacts and is managed through delete behavior and filesystem backups.

### DSGN-008: Gateway Owns Telegram Replies

**Date:** 2026-05-01
**Status:** accepted

**Context:** Telegram ingestion needs both immediate acknowledgement replies and terminal success/failure replies. Gateway and Worker communicate through SQLite rather than direct RPC.

**Decision:** Gateway owns all Telegram Bot API calls. Gateway sends immediate replies for accepted and invalid authorized webhook messages. Worker records terminal notification intent by writing SQLite notification rows when Telegram-originated jobs complete or fail. Gateway dispatches those terminal notification rows as replies to the original Telegram message.

**Consequences:** Worker never depends on Telegram Bot API clients or Gateway callback endpoints. Telegram delivery failures do not mutate terminal article/job state. Notification persistence remains the Gateway/Worker handoff for terminal Telegram replies.

### DSGN-011: Simplified Persistence Without Automatic Retries

**Date:** 2026-05-03
**Status:** accepted

**Context:** Archivist minimizes persistence shape and operational obligations. Automatic retries require retry policy, state-transition, idempotency, and operator-control decisions that are not part of the current queue contract. Telegram delivery failures and worker failures must still be visible enough for manual requeue by sending the URL again.

**Decision:** Persistence uses `users`, `articles`, `jobs`, and `notifications`. The personal `users.id` is `01ASB2XFCZJY7WHZ2FNRTMQJCT`, with `telegram_user_id` stored directly on that row. Jobs have only `queued`, `running`, `succeeded`, and `failed` states and are claimed atomically with `UPDATE ... RETURNING`; retry, backoff, and lock-owner fields are absent. Notifications have only `pending`, `sent`, and `failed` states and are not retried. Notification dispatch derives reply targets from jobs and success content from article artifacts. Article artifact paths are computed from `DATA_DIR` and `article_id`; artifact path columns and extraction telemetry columns are not stored.

**Consequences:** The database remains small and rebuildable. Failures are terminal and surface through persisted error fields. Users can manually re-send URLs to create new processing jobs. Queue retry, multi-worker locking, richer notification metadata, or additional persisted extraction diagnostics require canonical specs and design updates.

### DSGN-012: Local-First Markdown Extraction With Jina Fallback

**Date:** 2026-05-04
**Status:** accepted

**Context:** The pipeline needs Markdown before summary generation, but cost control matters. Local extraction is cheaper than provider calls, while some pages still need an external fallback.

**Decision:** Markdown extraction uses a deterministic local-first sequence behind a Worker-owned `MarkdownExtractor` abstraction. The Worker first uses `codeberg.org/readeck/go-readability/v2` against the saved HTML snapshot and calls `CheckDocument()` before accepting local output. If `CheckDocument()` returns false, local extraction fails, or local Markdown conversion fails, the Worker logs the fallback decision and calls Jina Reader. Successful Markdown is persisted as `content.md`, and Markdown completion is an intermediate stage before summary generation.

**Consequences:** The system does not use extraction candidate scoring or a quality score threshold. The system minimizes paid Jina usage, but still has a provider fallback for unreadable local results. Provider decisions must be logged, and Jina insufficient-balance failures must be distinguishable from generic provider failures. Pipeline orchestration depends on Archivist interfaces, not Jina SDK or adapter types.

### DSGN-013: Provider-Agnostic Text Summaries

**Date:** 2026-05-04
**Status:** accepted

**Context:** The processing pipeline needs a summary that Gateway can send directly to Telegram and the UI can render without a schema-specific presentation contract.

**Decision:** Summary generation uses a provider-agnostic `SummarizerService` interface and persists text-only output to `summary.md`. Claude through Anthropic is the configured provider path. Official provider SDKs must be used when suitable SDKs exist, while provider-specific SDK types remain contained inside provider adapters.

**Consequences:** Gateway and UI read human-readable summary text from `summary.md`. Structured summary data requires a canonical decision and artifact/schema contract before implementation.

### DSGN-014: Summary Generation Defines Processing Success

**Date:** 2026-05-04
**Status:** accepted

**Context:** Snapshotting and Markdown extraction are intermediate processing stages. Article success must represent the completed archive-review artifact set.

**Decision:** Processing success means `summary.md` has been atomically written and the article/job/notification terminal success transaction has committed. Snapshot and Markdown completion are intermediate stages.

**Consequences:** Rebuilds must not mark articles `ready`, jobs `succeeded`, or success notifications pending at snapshot or Markdown boundaries. Failures in any pipeline stage are terminal and use ARC-coded public errors.

### DSGN-015: Opaque Session Cookie Authentication For UI/API

**Date:** 2026-05-05
**Status:** accepted

**Context:** The web UI needs a private browser authentication surface for the bootstrapped account. Custom cookie payloads, random cookie names, startup-generated signing secrets, and cookie-ticket key-ring management add concerns that do not improve the current threat model.

**Decision:** UI/API authentication uses password-only login against password-bearing user rows and an opaque server-issued session id cookie. Passwords are generated 2048-character printable ASCII bearer secrets stored only as Argon2id PHC hashes on `users.password_hash`.

The auth cookie is named `__Host-app-auth` and carries only an opaque session identifier: 32 bytes from `RandomNumberGenerator.GetBytes`, base64url-encoded without padding. The cookie is a pure capability. It contains no user id, role, expiry timestamp, or session metadata.

The gateway stores authoritative session state server-side as `sessionId -> { userId, createdAt, absoluteExpiresAt }`. The implementation is an in-memory `ConcurrentDictionary<string, SessionEntry>` behind an `ISessionStore` interface. Expiry is absolute and server-side enforced at 24 hours from issue. The cookie itself is browser-session scoped and must not set `Expires` or `Max-Age`.

Cookie attributes are fixed: `HttpOnly`, `Secure`, `SameSite=Strict`, `Path=/`, no `Domain`, and the `__Host-` prefix. Cookie values are never logged and must be redacted from request and response logging middleware.

Authentication integrates with the normal ASP.NET Core authentication pipeline through a custom `IAuthenticationHandler`, or `AuthenticationHandler<AuthenticationSchemeOptions>`, registered by `AddAppCookie()` on `AuthenticationBuilder`. The default scheme and authentication type are `"app-cookie"`. On success, the handler sets `HttpContext.User` to a minimal `ClaimsPrincipal` containing only `ClaimTypes.NameIdentifier` with the authenticated session's user id.

`POST /login` accepts the password in the request body and rejects non-`POST`, non-same-origin, or requests whose effective public scheme is not `https`. The effective public scheme is `HttpRequest.Scheme` after trusted forwarded-header processing. Login throttling is applied per IP and globally before Argon2id verification so throttling cannot become a CPU amplification vector. Gateway loads every user row with a non-empty Argon2id PHC `password_hash`, verifies the submitted password against every candidate, and treats the login as successful only when exactly one candidate verifies. Multiple password-bearing rows are valid. Zero matches and duplicate matches fail closed with `401`; the response does not disclose whether users exist, hashes exist, no hash matched, or multiple hashes matched. Successful verification always rotates the session: if the request carries an existing valid cookie, the old session-store entry is removed; a fresh 32-byte session id is generated; `{ userId, createdAt = now, absoluteExpiresAt = now + 24h }` is inserted for the matching user id; `__Host-app-auth` is set with the fixed cookie attributes; and the endpoint returns `204 No Content`. The endpoint must not log the password, session id, cookie value, or `Set-Cookie` header.

Gateway runs privately behind the trusted Docker reverse proxy in the primary deployment topology. Public TLS termination occurs upstream of Caddy, Caddy forwards plaintext HTTP to Gateway on the Docker internal network, and Gateway must process forwarded headers before authentication and authorization. Gateway must not be exposed directly to the public Internet while trusting forwarded headers. The production VPS is operator-controlled and may sit behind a load balancer with dynamic source IPs, so the documented invariant is private Gateway reachability rather than static trusted-proxy IP configuration.

`POST /logout` reads the cookie and removes the matching session-store entry when present. It always emits `Set-Cookie: __Host-app-auth=; HttpOnly; Secure; SameSite=Strict; Path=/; Max-Age=0` and returns `204 No Content`, regardless of whether the cookie or store entry existed. Clearing only the cookie leaves server-side state alive until expiry; clearing only the store entry leaves the browser sending a dead cookie until replacement. Both actions are required.

The authentication handler only authenticates. On `HandleAuthenticateAsync`, a missing cookie returns `AuthenticateResult.NoResult()`. An unknown session returns failure without exposing a useful distinction to clients. An expired session is removed from the store and fails. A valid session returns `AuthenticateResult.Success(new AuthenticationTicket(principal, Scheme.Name))`. The handler must not issue, clear, rotate, or refresh cookies; `/login` and `/logout` own cookie lifecycle. Sliding expiry is not implemented.

The session-store contract is:

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

`RemoveAsync` on a missing key is a no-op. The in-memory implementation removes expired entries on lookup and may also perform a periodic sweep. Multi-replica deployment swaps `ISessionStore` to Redis, preferred over memcached because Redis can provide predictable TTL behavior and avoids authentication semantics depending on memory-pressure eviction.

The registration surface is:

```csharp
public static AuthenticationBuilder AddAppCookie(this AuthenticationBuilder builder);
```

There is no app-cookie settings object for the canonical cookie name or session lifetime. Gateway code consumes `AppCookieDefaults.CookieName` and `AppCookieDefaults.SessionLifetime` directly in the authentication handler and login/logout endpoints. Configuration values such as `AppCookie:CookieName` or `AppCookie:SessionLifetime` must not alter auth behavior. The session store is resolved from dependency injection.

```csharp
services.AddSingleton<ISessionStore, InMemorySessionStore>();
services.AddAuthentication("app-cookie").AddAppCookie();
services.AddAuthorization();
// app.UseAuthentication(); app.UseAuthorization();
```

Do not add sliding expiry, refresh tokens, multiple concurrent sessions per user, server-side revocation lists beyond entry removal, or encryption, signing, MACs, or other cryptographic transforms over the cookie value without a new design decision.

**Reasoning:** The threat model is a personal deployment behind effective public HTTPS. The realistic risks are credential theft, XSS, CSRF, and replay. They are not cookie tampering or inspection by an attacker who has not already stolen the cookie. Because the cookie value is a random capability with no embedded meaning, there is no value to encrypt. There is also no value to sign: a forged random string will not exist in `ISessionStore`, so lookup fails and the request is unauthorized. Constant-time comparison is not required because the session id is used as a dictionary key, not compared byte-by-byte against a stored secret. Theft mitigations are `__Host-` prefix, `HttpOnly`, `Secure`, `SameSite=Strict`, host-only scope, browser-session cookie lifetime, effective public HTTPS, server-side absolute expiry, and login throttling. This design also removes the ASP.NET Core key-ring deployment concern because there is no protected cookie payload to share across replicas. Multiple gateway replicas need only a shared session store and shared throttling state.

**Consequences:** UI/API cookie auth does not depend on key-ring management. Auth requires server-side session state. Multi-replica deployment requires a shared `ISessionStore` implementation, with Redis recommended, plus shared login throttling state. Gateway restart invalidates all in-memory sessions. `[Authorize]`, `HttpContext.User`, and `User.Identity.IsAuthenticated` work through the standard ASP.NET Core pipeline, so downstream endpoint code and filters remain normal.

### DSGN-016: Worker ARC Errors Use Idiomatic Go Error Flow

**Date:** 2026-05-16
**Status:** accepted

**Context:** Worker article-processing failures need stable ARC public messages for SQLite persistence, while provider adapters and orchestration also need diagnostic details for logs. Rebuilds need one idiomatic Go error-flow contract for ARC classification.

**Decision:** Worker ARC classification is implemented by `src/worker/internal/arc`, which owns ARC code constants, public messages, and typed sentinel errors. Worker functions return `error` for failures. Provider adapters return `(output, error)`, wrap ARC sentinels with `%w` or typed diagnostic errors when needed, and do not put ARC codes in result DTO fields. Package diagnostic errors preserve operation metadata and unwrap to ARC sentinels or lower-level causes where callers need `errors.Is` or `errors.As`; packages must not add low-value ARC alias surfaces. Pipeline orchestration uses `errors.Is`, `errors.As`, and `arc.CodeOf` for classification, logs diagnostic errors separately, and persists public article/job errors by rendering the ARC code through `arc.PublicMessage`.

**Consequences:** Wrapped provider, HTTP, SDK, and filesystem details remain available to logs without leaking into `articles.error_message` or `jobs.error_message`. Terminal public text is rendered only when the pipeline persists terminal article/job failure state; diagnostic `err.Error()` strings must not be persisted as public ARC text. Rebuilds must not reintroduce adapter result fields such as `ErrorCode` or `ResultStatus` for failure classification. `docs/ERRORS.md` remains the canonical human ARC catalog; `internal/arc` is the Go implementation of that catalog.

### DSGN-018: Snapshotter Provides Simple Object Storage Backups

**Date:** 2026-06-02
**Status:** accepted

**Context:** Archivist stores authoritative state in SQLite and retained article artifacts under `/data`. The single-VPS deployment needs an off-host backup mechanism without adding another database, queue, scheduler, or coordination protocol.

**Decision:** Archivist includes a Python Snapshotter service that sleeps for a configured interval before its first run, then periodically stages `/data`, creates `archivist-<yyyy-mm-dd>-<unix-timestamp>.tar.gz`, and uploads it to S3-compatible Object Storage. The configured SQLite database is copied into the staged tree with the SQLite online backup API. Non-database artifacts are copied best-effort from the live filesystem without coordinating Gateway or Worker writes/deletes. The archive includes a root `manifest.json` that records the archive name, object key, timestamp, source paths, Snapshotter version, and consistency note. Snapshotter logs failed attempts and continues to the next interval. Remote retention is handled by bucket lifecycle policy; Snapshotter never deletes remote objects. Restore is manual by downloading the archive, stopping services, replacing `/data` with archive `data/`, and starting services. Archives are plaintext before upload; confidentiality relies on TLS, private bucket access, IAM, and provider-side storage controls.

**Consequences:** The backup mechanism stays small and rebuildable. SQLite backups are internally consistent, but the full DB-plus-artifacts archive is not a transactional point-in-time snapshot under concurrent article writes or deletes. Operators that require a fully coherent `/data` image must stop Gateway and Worker before the next Snapshotter run or add a canonical writer-coordination feature. App-managed retention, restore tooling, backup status UI, metrics, or client-side encryption require new canonical specs and design updates.

### DSGN-019: OpenTelemetry Observability Uses SDK Propagation And Collector Tail Sampling

**Date:** 2026-06-03
**Status:** accepted

**Context:** Archivist needs operational visibility across Gateway HTTP entrypoints, asynchronous Worker processing, LLM/provider calls, artifact writes, Snapshotter backups, and deployment failures.

**Decision:** Gateway, Worker, and Snapshotter emit traces and logs through OpenTelemetry SDKs to a private official OpenTelemetry Collector Contrib service. Application telemetry is always configured in Compose; no application-side trace/log exporter disable switches are part of the deployment contract. Application SDKs export traces always-on; sampling is performed by the Collector with tail sampling that keeps all error traces and 10% of non-error traces. Application OTLP log export includes only `Info`/`Information` and higher records; debug logs may remain local diagnostics but must not be exported through OTLP. Gateway-to-Worker asynchronous propagation uses W3C Trace Context stored on queued jobs as `traceparent` and `tracestate`. Applications use .NET `Activity`, Go/Python OpenTelemetry SDKs, and standard W3C propagators before custom code. Collector runtime outages must not stop core application behavior.

**Consequences:** The SQLite job schema has nullable carrier fields. The deployment topology includes a private Collector and a development-only Grafana LGTM service. Logs include `trace_id` and `span_id` when emitted inside a span. High-cardinality values such as `article_id`, `job_id`, URLs, and provider request IDs remain searchable trace/log attributes and must not be promoted to Loki labels or metric labels.

### DSGN-020: Gateway Delete Accepts A Documented SQLite/Filesystem Atomicity Limit

**Date:** 2026-06-06
**Status:** accepted

**Context:** Gateway hard delete removes SQLite rows and the deterministic article artifact directory under `DATA_DIR`. SQLite transactions cannot include filesystem deletion. The ordering rechecks ownership and job state inside a SQLite write transaction, deletes associated notification/job/article rows, deletes the artifact directory, then commits the SQLite transaction. This preserves database state when artifact cleanup fails, but if artifact cleanup succeeds and the subsequent SQLite commit fails, rollback cannot restore the deleted artifacts.

**Decision:** Archivist accepts this rare cross-resource atomicity limitation instead of adding repair queues, tombstones, or a cleanup service. Normal delete and force delete keep the transaction-before-artifact-cleanup ordering so artifact cleanup failures leave SQLite state intact and the user can retry. If artifact cleanup succeeds but the commit fails, the operator repairs the inconsistent article by deleting the now-artifactless article state through an operational SQLite fix or by restoring `{DATA_DIR}/articles/{article_id}` from a Snapshotter/object-storage backup before retrying. This known limitation must remain visible in canonical docs and deployment guidance.

**Consequences:** The implementation stays small and does not add schema or queue complexity. Successful API responses still mean both SQLite state and artifact directory were removed. Strict cross-resource atomicity requires a canonical feature that introduces durable cleanup state, a repair command, or a storage design where article state and artifacts can be committed atomically.

### DSGN-021: Stale Running Jobs Are Operator-Recoverable By Force Delete

**Date:** 2026-06-02
**Status:** accepted

**Context:** The Worker claims queued jobs by updating SQLite state to `running` before article processing. If infrastructure fails after claim but before terminal persistence, the job can remain `running` indefinitely. Normal article deletion rejects running jobs to avoid racing an active worker, which leaves abandoned running jobs undeletable.

**Decision:** Archivist treats running jobs as force-delete eligible only after they are stale. A running job is stale when `started_at <= now - 2 hours`; `started_at IS NULL` is stale for recovery because a claimed job without a start timestamp cannot represent a healthy active claim. Force delete is an explicit authenticated same-origin Gateway action for user-owned articles. It deletes the article, associated jobs, associated notifications, and deterministic artifact directory. Normal delete continues to reject running jobs. This decision does not add automatic retry, requeue, rollback-to-queued behavior, lock owners, heartbeats, new job states, or schema changes.

**Consequences:** The UI can expose a deliberate cleanup path for zombie jobs without weakening normal delete safety for active work. The stale threshold is conservative for single-worker operation. Automatic retry, heartbeat, or multi-worker locking behavior requires a new canonical feature and design update.
