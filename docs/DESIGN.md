# DESIGN.md

Records durable decisions that must survive rebuilds. This can be ADR-like but does not need heavy ceremony.

A decision discovered during implementation should be promoted here if it affects more than one task or must remain true across rebuilds.

---

## Decision Record Format

Each decision can be as lightweight as a heading plus rationale paragraph, or as structured as a full ADR. Use the level of ceremony that matches the stakes.

Suggested minimal format:

```text
### DSGN-NNN: <Title>

**Date:** YYYY-MM-DD
**Status:** accepted | superseded | under review

**Context:** Why this decision was needed.

**Decision:** What was decided.

**Consequences:** What changes as a result. What becomes easier or harder.
```

Only decisions with `Status: accepted` are binding for rebuild. Decisions marked `superseded` or `under review` are retained as historical context and must not be used as required behavior when they conflict with accepted decisions.

---

## Decisions

### DSGN-001: Reliability and Simplicity Over Feature Breadth

**Date:** 2026-04-30
**Status:** accepted

**Context:** Archivist v0 is a personal article archiving system intended to be reliable and easy to rebuild before it is feature-complete.

**Decision:** v0 prioritizes deterministic processing, clear failure states, minimal dependencies, and a small deployment surface. Features that increase product breadth but are not required for the core archive-review loop are deferred.

**Consequences:** v0 excludes Playwright, full-text search, filtering, advanced tagging, PWA/offline support, multi-user support, and browser extensions. Future feature specs may add deferred capabilities only by updating the relevant canonical docs. OpenTelemetry observability was later added by DSGN-019 and the `otel-observability` feature.

### DSGN-002: Single-User Product Boundary

**Date:** 2026-04-30
**Status:** accepted

**Context:** The initial product serves one person through Telegram and a private web UI.

**Decision:** v0 has one bootstrapped personal user and no user-facing account management. Telegram ingestion and UI/API access are private authenticated surfaces.

**Consequences:** v0 does not include tenant isolation, account management, user roles, per-user storage partitioning, registration, or user administration. Runtime ownership resolution is refined by DSGN-017.

### DSGN-017: Runtime User ID Resolution

**Date:** 2026-06-05
**Status:** accepted

**Context:** Archivist already persists `users`, `articles.user_id`, `jobs.user_id`, and authenticated sessions with a user id. Runtime code still assigned the personal user ULID in several paths, which made ownership fields look user-aware while behavior remained single-user-by-constant.

**Decision:** Authentication bootstrap is the primary production code path allowed to hardcode the initial personal user id and fixed personal Telegram sender id `1559957191`. Bootstrap sets the personal row's `telegram_user_id` to `1559957191` only when it is null and preserves an existing non-null value; it does not read deployment-configured Telegram sender allowlists or personal sender settings for that seed. Worker CLI enqueue is the only runtime exception: it uses `jobs.DefaultUserID = 01ASB2XFCZJY7WHZ2FNRTMQJCT`, checks `users.id` by that id, never infers ownership from user-table cardinality, and never creates the user. Other runtime Gateway and Worker behavior resolves `user_id` from persisted user rows, authenticated session state, claimed jobs, or article ownership. Password login loads all non-empty Argon2id PHC password hashes, verifies every candidate, and issues a session only when exactly one user matches the submitted password. Multiple password-bearing rows are valid; duplicate password matches fail closed. Telegram webhook authorization is based on the existence of a `users.telegram_user_id` mapping; unknown Telegram senders receive no reply and create no rows. Gateway and Worker attach `user_id` to logs and traces when known, using the exact key `user_id`; Snapshotter remains user-agnostic.

**Consequences:** v0 still has no registration, roles, tenant administration, or user-facing multi-user UI. The existing personal account remains bootstrapped. Runtime ownership no longer relies on a hardcoded personal-id fallback except for Worker CLI enqueue's explicit default-user existence check. Future user provisioning can add rows, password hashes, and Telegram mappings without changing Gateway, UI, or Worker processing ownership semantics; password-only login remains ambiguous only when the submitted password matches multiple users, in which case it fails closed. Worker CLI enqueue continues to target the bootstrapped personal account unless a future canonical decision changes that command.

### DSGN-003: SQLite Owns State and Filesystem Owns Artifacts

**Date:** 2026-04-30
**Status:** accepted

**Context:** Article records and job state need transactional updates, while raw HTML, Markdown, and summaries are better stored as files.

**Decision:** SQLite is the authoritative store for state. The filesystem stores large raw and derived artifacts under `/data/articles/{article_id}/`, with filenames defined in `docs/ARTIFACTS.md`.

**Consequences:** Rebuild-critical state must be recoverable from SQLite plus deterministic artifact paths derived from `DATA_DIR` and `article_id`. Artifact writes use temporary files followed by rename. Optional hashes may be added for integrity or debugging, but they are not required for v0.

### DSGN-004: SQLite Job Queue With One Worker

**Date:** 2026-04-30
**Status:** superseded by DSGN-011

**Context:** v0 needs retryable background processing without introducing an external queue.

**Decision:** Jobs are stored in SQLite and processed by a single Go worker. Dequeue must be atomic using `UPDATE ... RETURNING`.

**Consequences:** v0 avoids external queue infrastructure. Jobs support queued, running, succeeded, failed, retrying, and dead states. Failed jobs retry up to three times with backoff before becoming dead and surfacing an error in the UI.

### DSGN-005: Preserve Raw and Derived Article Data

**Date:** 2026-04-30
**Status:** accepted

**Context:** Extraction and summarization may fail or improve over time. Raw input is needed for debugging and reprocessing.

**Decision:** The worker stores the raw HTML snapshot and derived Markdown and summary artifacts for each processed article.

**Consequences:** The system can expose failure state clearly and can support future reprocessing. Storage usage grows with retained artifacts and is managed through delete behavior and filesystem backups.

### DSGN-006: Candidate-Based Extraction With Scoring

**Date:** 2026-04-30
**Status:** superseded by DSGN-012

**Context:** Article extraction quality varies by source and extraction method.

**Decision:** The worker attempts multiple extraction candidates, scores them, selects the highest-scoring candidate, and fails extraction when all candidates score below `0.6`.

**Consequences:** v0 extraction is best-effort and explicit about failure. Candidate scoring may consider title presence, content length, paragraph count, link density, boilerplate ratio, sentence density, error-page detection, canonical URL detection, language detection, and heading structure.

### DSGN-012: Local-First Markdown Extraction With Jina Fallback

**Date:** 2026-05-04
**Status:** accepted

**Context:** The v0 pipeline needs Markdown before summary generation, but cost control matters. Local extraction is cheaper than provider calls, while some pages still need an external fallback.

**Decision:** Markdown extraction uses a deterministic local-first sequence behind a Worker-owned `MarkdownExtractor` abstraction. The Worker first uses `codeberg.org/readeck/go-readability/v2` against the saved HTML snapshot and calls `CheckDocument()` before accepting local output. If `CheckDocument()` returns false, local extraction fails, or local Markdown conversion fails, the Worker logs the fallback decision and calls Jina Reader. Successful Markdown is persisted as `content.md`, and Markdown completion is an intermediate stage once summary generation exists.

**Consequences:** v0 does not use extraction candidate scoring or a quality score threshold. The system minimizes paid Jina usage, but still has a provider fallback for unreadable local results. Provider decisions must be logged, and Jina insufficient-balance failures must be distinguishable from generic provider failures. Pipeline orchestration depends on Archivist interfaces, not Jina SDK or adapter types.

### DSGN-007: Provider-Agnostic Structured Summaries

**Date:** 2026-04-30
**Status:** superseded by DSGN-013

**Context:** Summaries are generated by an LLM, but the product should not be tied to a single provider.

**Decision:** Summarization uses a provider-agnostic interface and produces strict JSON with `summary`, `key_points`, `tags`, and `template_version`.

**Consequences:** Summary output must be schema-validated. In v0, malformed LLM output fails the job instead of entering an automatic retry loop.

### DSGN-013: Provider-Agnostic Text Summaries

**Date:** 2026-05-04
**Status:** accepted

**Context:** The final v0 pipeline needs a summary that Gateway can send directly to Telegram. Structured summary fields are not required for v0 and would force a schema contract before the product needs one.

**Decision:** Summary generation uses a provider-agnostic `SummarizerService` interface and persists text-only output to `summary.md`. Claude through Anthropic is the first provider. Official provider SDKs must be used when suitable SDKs exist, while provider-specific SDK types remain contained inside provider adapters. `summary.json` and SQLite summary columns are out of scope for v0.

**Consequences:** Gateway and UI read human-readable summary text from `summary.md`. Future structured summaries require a new canonical decision and artifact/schema contract.

### DSGN-014: Summary Generation Defines Final V0 Processing Success

**Date:** 2026-05-04
**Status:** accepted

**Context:** Snapshotting and Markdown extraction were planned as interim terminal success points while later pipeline stages were unspecified.

**Decision:** Final v0 success means `summary.md` has been atomically written and the article/job/notification terminal success transaction has committed. Snapshot and Markdown completion are intermediate stages once summary generation is implemented.

**Consequences:** Rebuilds must not mark articles `ready`, jobs `succeeded`, or success notifications pending at snapshot or Markdown boundaries in the final v0 pipeline. Failures in any pipeline stage remain terminal in v0 and use ARC-coded public errors.

### DSGN-008: Gateway Owns Telegram Replies

**Date:** 2026-05-01
**Status:** superseded by DSGN-011

**Context:** Telegram ingestion needs both immediate acknowledgement replies and later terminal success/failure replies. The architecture already keeps gateway and worker communication through SQLite rather than direct RPC.

**Decision:** The gateway owns all Telegram Bot API calls. The gateway sends immediate replies for accepted and invalid authorized webhook messages. The worker records terminal notification intent by writing SQLite outbox rows when Telegram-originated jobs complete or fail. The gateway dispatches those terminal outbox rows as replies to the original Telegram message.

**Consequences:** The worker never depends on Telegram Bot API clients or gateway callback endpoints. Terminal notification delivery is retryable through SQLite outbox state. Telegram delivery failures do not mutate terminal article/job state.

### DSGN-009: Persist External Identity Correlation Early

**Date:** 2026-05-02
**Status:** superseded by DSGN-010

**Context:** v0 is single-user, but Telegram-originated work should be correlatable to future Archivist user records without reprocessing historical Telegram messages. Telegram sender user ID is distinct from chat ID and reply message ID.

**Decision:** Accepted Telegram ingestions persist `telegram_user_id` as sender identity metadata and upsert an external identity correlation row keyed by `(provider, external_user_id)`, initially using provider `telegram` and the Telegram sender user ID. The correlation row includes nullable `archivist_user_id` because no canonical Archivist users table exists yet.

**Consequences:** The system gains a durable future join point for multi-tenancy and per-user processing. v0 authorization remains `TELEGRAM_ALLOWED_USER_ID`; external identity correlation does not drive authorization, routing, ownership, or job behavior until a future feature changes the canonical docs.

### DSGN-010: Seed Telegram Identity Correlation With Personal Account

**Date:** 2026-05-03
**Status:** superseded by DSGN-011

**Context:** v0 is single-user and the first Archivist user account is known ahead of the broader user model. Keeping `archivist_user_id` non-null simplifies future correlation and avoids a later backfill for the personal Telegram account.

**Decision:** Accepted Telegram ingestions upsert external identity correlation rows with provider `telegram`, the Telegram sender user ID as `external_user_id`, and `archivist_user_id = 01ASB2XFCZJY7WHZ2FNRTMQJCT`. This ULID is the personal Archivist account ID.

**Consequences:** Historical Telegram-originated jobs can be correlated to the personal Archivist account once user-aware features exist. v0 authorization remains `TELEGRAM_ALLOWED_USER_ID`; the personal account ULID must not be used as a catch-all for future additional Telegram users.

### DSGN-011: Simplified V0 Persistence Without Automatic Retries

**Date:** 2026-05-03
**Status:** accepted

**Context:** v0 should minimize persistence shape and operational obligations. Automatic retries require observability and operational policy that are out of scope. Telegram delivery failures and worker failures must still be visible enough for manual requeue by sending the URL again.

**Decision:** v0 persistence uses `users`, `articles`, `jobs`, and `notifications`. The personal `users.id` is `01ASB2XFCZJY7WHZ2FNRTMQJCT`, with `telegram_user_id` stored directly on that row. Jobs have only `queued`, `running`, `succeeded`, and `failed` states and are claimed atomically with `UPDATE ... RETURNING`; there are no retry, backoff, or lock-owner fields. Notifications have only `pending`, `sent`, and `failed` states and are not retried. Notification dispatch derives reply targets from jobs and success content from article artifacts. Article artifact paths are computed from `DATA_DIR` and `article_id`; artifact path columns and extraction telemetry columns are not stored.

**Consequences:** The database remains small and rebuildable. Failures are terminal in v0 and surface through persisted error fields. Users can manually re-send URLs to create new processing jobs. Future retry support, multi-worker locking, richer notification metadata, or extraction observability must be introduced by new canonical specs and design updates.

### DSGN-015: Opaque Session Cookie Authentication For V0 UI/API

**Date:** 2026-05-05
**Status:** accepted

**Context:** The v0 web UI needs a private browser authentication surface for one personal user. The original custom-cookie idea included random cookie names, user-id hashes, startup-generated signing secrets, and an in-memory user-id map. The later cookie-ticket design made auth cookie key-ring management part of the deployment surface. Both approaches add concerns that do not improve the v0 threat model.

**Decision:** UI/API authentication uses password-only login against password-bearing user rows and an opaque server-issued session id cookie. Passwords are generated 2048-character printable ASCII bearer secrets stored only as Argon2id PHC hashes on `users.password_hash`.

The auth cookie is named `__Host-app-auth` and carries only an opaque session identifier: 32 bytes from `RandomNumberGenerator.GetBytes`, base64url-encoded without padding. The cookie is a pure capability. It contains no user id, role, expiry timestamp, or session metadata.

The gateway stores authoritative session state server-side as `sessionId -> { userId, createdAt, absoluteExpiresAt }`. The v0 implementation is an in-memory `ConcurrentDictionary<string, SessionEntry>` behind an `ISessionStore` interface. Expiry is absolute and server-side enforced at 24 hours from issue. The cookie itself is browser-session scoped and must not set `Expires` or `Max-Age`.

Cookie attributes are fixed: `HttpOnly`, `Secure`, `SameSite=Strict`, `Path=/`, no `Domain`, and the `__Host-` prefix. Cookie values are never logged and must be redacted from request and response logging middleware.

Authentication integrates with the normal ASP.NET Core authentication pipeline through a custom `IAuthenticationHandler`, or `AuthenticationHandler<AuthenticationSchemeOptions>`, registered by `AddAppCookie()` on `AuthenticationBuilder`. The default scheme and authentication type are `"app-cookie"`. On success, the handler sets `HttpContext.User` to a minimal `ClaimsPrincipal` containing only `ClaimTypes.NameIdentifier` with the authenticated session's user id.

`POST /login` accepts the password in the request body and rejects non-`POST`, non-same-origin, or requests whose effective public scheme is not `https`. The effective public scheme is `HttpRequest.Scheme` after trusted forwarded-header processing. Login throttling is applied per IP and globally before Argon2id verification so throttling cannot become a CPU amplification vector. Gateway loads every user row with a non-empty Argon2id PHC `password_hash`, verifies the submitted password against every candidate, and treats the login as successful only when exactly one candidate verifies. Multiple password-bearing rows are valid. Zero matches and duplicate matches fail closed with `401`; the response does not disclose whether users exist, hashes exist, no hash matched, or multiple hashes matched. Successful verification always rotates the session: if the request carries an existing valid cookie, the old session-store entry is removed; a fresh 32-byte session id is generated; `{ userId, createdAt = now, absoluteExpiresAt = now + 24h }` is inserted for the matching user id; `__Host-app-auth` is set with the fixed cookie attributes; and the endpoint returns `204 No Content`. The endpoint must not log the password, session id, cookie value, or `Set-Cookie` header.

Gateway runs privately behind the trusted Docker reverse proxy in the primary deployment topology. Public TLS termination occurs upstream of Caddy, Caddy forwards plaintext HTTP to Gateway on the Docker internal network, and Gateway must process forwarded headers before authentication and authorization. Gateway must not be exposed directly to the public Internet while trusting forwarded headers. The production VPS is operator-controlled and may sit behind a load balancer with dynamic source IPs, so v0 deliberately documents the private-network invariant instead of requiring static trusted-proxy IP configuration.

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

`RemoveAsync` on a missing key is a no-op. The v0 in-memory implementation removes expired entries on lookup and may also perform a periodic sweep. Multi-replica deployment swaps `ISessionStore` to Redis, preferred over memcached because Redis can provide predictable TTL behavior and avoids authentication semantics depending on memory-pressure eviction.

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

**Reasoning:** The v0 threat model is a single personal user behind effective public HTTPS. The realistic risks are credential theft, XSS, CSRF, and replay. They are not cookie tampering or inspection by an attacker who has not already stolen the cookie. Because the cookie value is a random capability with no embedded meaning, there is no value to encrypt. There is also no value to sign: a forged random string will not exist in `ISessionStore`, so lookup fails and the request is unauthorized. Constant-time comparison is not required because the session id is used as a dictionary key, not compared byte-by-byte against a stored secret. Theft mitigations stay the same as the previous design: `__Host-` prefix, `HttpOnly`, `Secure`, `SameSite=Strict`, host-only scope, browser-session cookie lifetime, effective public HTTPS, server-side absolute expiry, and login throttling. This design also removes the ASP.NET Core key-ring deployment concern because there is no protected cookie payload to share across replicas. Multiple gateway replicas need only a shared session store and shared throttling state.

**Consequences:** v0 no longer depends on key-ring management for UI/API cookie auth. v0 explicitly requires server-side auth session state, reversing the previous no-session-store line. Multi-replica deployment requires a shared `ISessionStore` implementation, with Redis recommended, plus shared login throttling state. Gateway restart still invalidates all sessions in v0 because the in-memory store is wiped; this is intentional and keeps the prior restart-invalidation behavior. `[Authorize]`, `HttpContext.User`, and `User.Identity.IsAuthenticated` continue to work through the standard ASP.NET Core pipeline, so downstream endpoint code and filters remain normal.

**Amendment History:** 2026-05-05 amendment supersedes the original ASP.NET Core cookie-ticket design under this same decision id. 2026-05-13 amendment defines HTTPS auth checks in terms of the effective public request context after trusted forwarded-header processing.

### DSGN-016: Worker ARC Errors Use Idiomatic Go Error Flow

**Date:** 2026-05-16
**Status:** accepted

**Context:** Worker article-processing failures need stable ARC public messages for SQLite persistence, while provider adapters and orchestration also need diagnostic details for logs. The previous code mixed public `errors.New("[ARC-NNN] ...")` sentinels with result DTOs carrying code-only fields, creating two competing failure patterns.

**Decision:** Worker ARC classification is implemented by `src/worker/internal/arc`, which owns ARC code constants, public messages, and typed sentinel errors. Worker functions return `error` for failures. Provider adapters return `(output, error)`, wrap ARC sentinels with `%w` or typed diagnostic errors when needed, and do not put ARC codes in result DTO fields. Package diagnostic errors preserve operation metadata and unwrap to ARC sentinels or lower-level causes where callers need `errors.Is` or `errors.As`; packages must not add low-value ARC alias surfaces. Pipeline orchestration uses `errors.Is`, `errors.As`, and `arc.CodeOf` for classification, logs diagnostic errors separately, and persists public article/job errors by rendering the ARC code through `arc.PublicMessage`.

**Consequences:** Wrapped provider, HTTP, SDK, and filesystem details remain available to logs without leaking into `articles.error_message` or `jobs.error_message`. Terminal public text is rendered only when the pipeline persists terminal article/job failure state; diagnostic `err.Error()` strings must not be persisted as public ARC text. Rebuilds must not reintroduce adapter result fields such as `ErrorCode` or `ResultStatus` for failure classification. `docs/ERRORS.md` remains the canonical human ARC catalog; `internal/arc` is the Go implementation of that catalog.

### DSGN-021: Stale Running Jobs Are Operator-Recoverable By Force Delete

**Date:** 2026-06-02
**Status:** accepted

**Context:** The Worker claims queued jobs by updating SQLite state to `running` before article processing. If infrastructure fails after claim but before terminal persistence, the job can remain `running` indefinitely. Normal article deletion rejects running jobs to avoid racing an active worker, which leaves abandoned running jobs undeletable.

**Decision:** Archivist treats running jobs as force-delete eligible only after they are stale. A running job is stale when `started_at <= now - 2 hours`; `started_at IS NULL` is stale for recovery because a claimed job without a start timestamp cannot represent a healthy active claim. Force delete is an explicit authenticated same-origin Gateway action for user-owned articles. It deletes the article, associated jobs, associated notifications, and deterministic artifact directory. Normal delete continues to reject running jobs. This decision does not add automatic retry, requeue, rollback-to-queued behavior, lock owners, heartbeats, new job states, or schema changes.

**Consequences:** The UI can expose a deliberate cleanup path for zombie jobs without weakening normal delete safety for active work. The stale threshold is conservative for v0 single-worker operation. Future automatic retry, heartbeat, or multi-worker locking behavior requires a new canonical feature and design update.

### DSGN-018: Snapshotter Provides Simple Object Storage Backups

**Date:** 2026-06-02
**Status:** accepted

**Context:** Archivist stores authoritative state in SQLite and retained article artifacts under `/data`. The v0 single-VPS deployment needs an off-host backup mechanism without adding another database, queue, scheduler, or coordination protocol.

**Decision:** Archivist adds a Python Snapshotter service that sleeps for a configured interval before its first run, then periodically stages `/data`, creates `archivist-<yyyy-mm-dd>-<unix-timestamp>.tar.gz`, and uploads it to S3-compatible Object Storage. The configured SQLite database is copied into the staged tree with the SQLite online backup API. Non-database artifacts are copied best-effort from the live filesystem without coordinating Gateway or Worker writes/deletes. The archive includes a root `manifest.json` that records the archive name, object key, timestamp, source paths, Snapshotter version, and consistency note. Snapshotter logs failed attempts and continues to the next interval. Remote retention is handled by bucket lifecycle policy; Snapshotter never deletes remote objects. Restore is manual by downloading the archive, stopping services, replacing `/data` with archive `data/`, and starting services. Archives are plaintext before upload; confidentiality relies on TLS, private bucket access, IAM, and provider-side storage controls.

**Consequences:** The backup mechanism stays small and rebuildable. SQLite backups are internally consistent, but the full DB-plus-artifacts archive is not a transactional point-in-time snapshot under concurrent article writes or deletes. Operators that require a fully coherent `/data` image must stop Gateway and Worker before the next Snapshotter run or introduce a future writer-coordination feature. Future app-managed retention, restore tooling, backup status UI, metrics, or client-side encryption require new canonical specs and design updates.

### DSGN-019: OpenTelemetry Observability Uses SDK Propagation And Collector Tail Sampling

**Date:** 2026-06-03
**Status:** accepted

**Context:** Archivist needs operational visibility across Gateway HTTP entrypoints, asynchronous Worker processing, LLM/provider calls, artifact writes, Snapshotter backups, and deployment failures. Earlier v0 scope excluded a dedicated observability stack, but this feature explicitly adds post-v0 observability.

**Decision:** Gateway, Worker, and Snapshotter emit traces and logs through OpenTelemetry SDKs to a private official OpenTelemetry Collector Contrib service. Application telemetry is always configured in Compose; no application-side trace/log exporter disable switches are part of the deployment contract. Application SDKs export traces always-on; sampling is performed by the Collector with tail sampling that keeps all error traces and 10% of non-error traces. Application OTLP log export includes only `Info`/`Information` and higher records; debug logs may remain local diagnostics but must not be exported through OTLP. Gateway-to-Worker asynchronous propagation uses W3C Trace Context stored on queued jobs as `traceparent` and `tracestate`. Applications use .NET `Activity`, Go/Python OpenTelemetry SDKs, and standard W3C propagators before custom code. Collector runtime outages must not stop core application behavior.

**Consequences:** The SQLite job schema gains nullable carrier fields. The deployment topology gains a private Collector and a development-only Grafana LGTM service. Logs include `trace_id` and `span_id` when emitted inside a span. High-cardinality values such as `article_id`, `job_id`, URLs, and provider request IDs remain searchable trace/log attributes and must not be promoted to Loki labels or metric labels.


### DSGN-020: Gateway Delete Accepts A Documented SQLite/Filesystem Atomicity Limit

**Date:** 2026-06-06
**Status:** accepted

**Context:** Gateway hard delete removes SQLite rows and the deterministic article artifact directory under `DATA_DIR`. SQLite transactions cannot include filesystem deletion. The existing v0 ordering rechecks ownership and job state inside a SQLite write transaction, deletes associated notification/job/article rows, deletes the artifact directory, then commits the SQLite transaction. This preserves database state when artifact cleanup fails, but if artifact cleanup succeeds and the later SQLite commit fails, rollback cannot restore the deleted artifacts.

**Decision:** Archivist accepts this rare cross-resource atomicity limitation for v0 instead of adding repair queues, tombstones, or a new cleanup service. Normal delete and force delete keep the transaction-before-artifact-cleanup ordering so artifact cleanup failures leave SQLite state intact and the user can retry. If artifact cleanup succeeds but the commit fails, the operator repairs the inconsistent article by deleting the now-artifactless article state through an operational SQLite fix or by restoring `{DATA_DIR}/articles/{article_id}` from a Snapshotter/object-storage backup before retrying. This known limitation must remain visible in canonical docs and deployment guidance.

**Consequences:** The implementation stays small and does not add schema or queue complexity. Successful API responses still mean both SQLite state and artifact directory were removed. Operators that require strict cross-resource atomicity need a future feature that introduces durable cleanup state, a repair command, or a storage design where article state and artifacts can be committed atomically.
