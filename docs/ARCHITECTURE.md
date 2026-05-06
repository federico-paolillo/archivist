# ARCHITECTURE.md

Describes global system architecture: executables, services, boundaries, data ownership, storage, runtime topology, and integration patterns.

Architecture decisions that constrain all features belong here or in `docs/DESIGN.md`.

---

## System Overview

Archivist is a single-user personal article archiving system.

The v0 system accepts article URLs through Telegram, stores article state in SQLite, processes queued article jobs with a single worker, writes large artifacts to the filesystem, generates text LLM summaries, and exposes a minimal authenticated web UI for review and administration.

High-level flow:

```text
Telegram Bot
  -> ASP.NET Core Minimal API gateway
  -> SQLite articles and jobs
  -> Go worker
  -> SQLite plus filesystem artifacts under /data
  -> ASP.NET Core API
  -> Preact/Vite UI
```

The system favors a small, rebuildable deployment over horizontal scale. SQLite owns authoritative state. Filesystem artifacts are derived or retained content associated with article records.

## Executables and Services

### Gateway API

- Runtime: ASP.NET Core Minimal API.
- Responsibilities:
  - receive Telegram webhook requests;
  - authenticate Telegram webhook requests using the configured secret;
  - accept URLs only from the configured Telegram user;
  - create article records;
  - enqueue article processing jobs;
  - send immediate Telegram replies for accepted or invalid authorized ingestion messages;
  - dispatch terminal Telegram completion replies from SQLite notifications;
  - read article artifacts from `/data` through read-only abstractions;
  - expose authenticated API endpoints for the UI;
  - support basic admin delete actions.

### Worker

- Runtime: Go.
- Deployment: single instance in v0.
- Responsibilities:
  - atomically dequeue jobs from SQLite;
  - fetch article HTML over plain HTTP(S);
  - store the raw HTML snapshot;
  - extract readable content with go-readability v2 first;
  - fall back to Jina Reader when local readability cannot produce Markdown and Jina is enabled;
  - convert selected local extracted content to Markdown;
  - call the configured LLM summarizer;
  - persist text-only summaries;
  - persist artifacts and final article/job state;
  - create terminal notification rows for Telegram-originated jobs.

### Web UI

- Runtime/tooling: Preact with Vite.
- Responsibilities:
  - expose browser routes `/login`, `/login/failed`, `/articles`, and `/articles/{article_id}`;
  - authenticate through the Gateway auth endpoints;
  - show the article list;
  - show article detail;
  - display title, summary Markdown, content Markdown, original link, progress/failure states, and failure messages;
  - expose delete actions.

### Public Browser/API Routing

The browser UI owns page routes such as `/login` and `/articles`. To avoid collisions with Gateway's intentionally unprefixed API routes, the UI calls Gateway through a configured same-origin API base path.

- Default public UI API base: `/api`.
- Vite build-time configuration: `VITE_API_BASE_PATH`, default `/api`.
- Reverse proxy behavior: public `/api/*` requests are forwarded to Gateway with the `/api` prefix stripped.
- Gateway route contracts remain unprefixed, for example `POST /login`, `GET /articles`, and `DELETE /articles/{id}`.

## Data Storage

SQLite is the source of truth for user state, article state, job state, notification state, and error state.

Filesystem storage under `/data` stores larger raw and derived artifacts:

```text
/data/
  archive.db
  articles/
    {article_id}/
      snapshot.html
      content.md
      summary.md
      summary.json  # future structured summary artifact
      metadata.json
```

Artifact writes must be atomic: write to a temporary path and then rename into place. Artifact paths are deterministic from `DATA_DIR` and `article_id`; v0 does not store artifact path columns in SQLite. Optional artifact hashes may be stored for integrity checks and debugging if a future spec requires them.

The canonical artifact path convention is defined in `docs/conventions/ARTIFACTS.md`.

Core user state includes:

- `id`: seeded as `01ASB2XFCZJY7WHZ2FNRTMQJCT` for the personal account in v0
- `telegram_user_id`: nullable until Telegram ingestion maps the configured Telegram user, unique when present
- `password_hash`: nullable only before UI/API authentication bootstrap completes, then an Argon2id PHC string

Core article state includes:

- `id`
- `user_id`
- `original_url`
- `canonical_url`
- `title`
- `status`: `queued`, `ready`, or `failed`
- `error_message`
- `created_at`

Core job state includes:

- `id`
- `user_id`
- `article_id`
- `type`
- `status`: `queued`, `running`, `succeeded`, or `failed`
- `telegram_update_id`
- `telegram_chat_id`
- `telegram_message_id`
- `telegram_user_id`
- `error_message`
- `created_at`
- `started_at`
- `completed_at`
- `expires_at`

Telegram update idempotency is keyed by `jobs.telegram_update_id`.

Notification state includes:

- `id`
- `job_id`
- `status`: `pending`, `sent`, or `failed`
- `error_message`
- `created_at`
- `sent_at`
- `expires_at`

Jobs do not retry automatically in v0. Notifications do not retry automatically in v0. Failed jobs and failed notifications persist error text so the user can manually re-send the URL or the operator can diagnose the issue.

In final v0 processing, `articles.status = ready`, `jobs.status = succeeded`, and the success notification row are committed only after `summary.md` has been atomically written. HTML snapshotting and Markdown extraction are intermediate stages once summary generation exists. Any terminal failure in fetch, snapshot, Markdown extraction, summarization, or artifact writing marks the article and job failed with an ARC-coded public error.

## Service Boundaries and Communication

The gateway and worker communicate through SQLite, not direct RPC.

The UI communicates only with the gateway API through the configured API base path. It must not read SQLite or filesystem artifacts directly.

The worker owns processing jobs, filesystem artifact production, final article/job state, and creation of terminal notification rows. The gateway owns request authentication, article/job creation, Telegram API calls, terminal notification dispatch, UI-facing API behavior, and admin actions.

## External Integrations

### Telegram

Telegram is the only v0 ingestion channel. The gateway accepts Telegram webhook requests and rejects requests that do not match the configured webhook secret or allowed Telegram user ID.

Authorized Telegram messages must contain exactly one trimmed absolute `http` or `https` URL. Invalid authorized messages receive `Nope, you must send only an URL`. Valid queued URL messages receive `Ok, I will have a look` after the article/job enqueue transaction commits. Completion replies are sent later by the gateway from SQLite notification rows, as replies to the original Telegram message.

The gateway persists the Telegram sender user ID separately from `telegram_chat_id` and `telegram_message_id`. `telegram_user_id` is sender identity metadata. `telegram_chat_id` and `telegram_message_id` are reply-target metadata.

The v0 personal account ULID must not be treated as a catch-all for future multi-user ingestion. Any feature that accepts additional Telegram users must define explicit identity-linking behavior before changing `TELEGRAM_ALLOWED_USER_ID` authorization semantics.

The worker must not call Telegram APIs directly.

### Article Websites

The worker fetches article HTML using direct HTTP(S) requests. v0 does not use Playwright, headless browser rendering, or browser automation.

### Extraction Providers and Libraries

The worker attempts Markdown extraction in this order:

1. go-readability v2 from the saved HTML snapshot.
2. Jina Reader fallback when go-readability `CheckDocument()` returns false, local extraction fails, or local Markdown conversion fails, and Jina is enabled.

The Worker logs critical extraction decisions, including fallback from go-readability to Jina. If both local extraction and Jina fallback fail, the job becomes terminally failed.

### LLM Provider

The Worker uses a provider-agnostic summarization interface. Claude through Anthropic is the first v0 provider, but provider, API key, and model are configuration values. The Anthropic implementation must use official Anthropic SDKs when suitable SDKs exist for the implementation language. Provider-specific SDK types must not leak outside the summarizer adapter.

## Runtime Topology

v0 deploys all components together on a single VPS.

Deployment requirements:

- one shared `/data` volume for SQLite and article artifacts;
- gateway, worker, and UI deployed as one application stack;
- filesystem snapshot backup for `/data`;
- stdout logging collected by the host or deployment environment.

The v0 topology does not target high scalability, multi-region deployment, or real-time processing guarantees.

## Security Boundaries

The system is single-user.

Security boundaries:

- Telegram ingestion is limited to one configured `TELEGRAM_ALLOWED_USER_ID`.
- Telegram webhook requests require `TELEGRAM_WEBHOOK_SECRET`.
- The entire UI and UI-facing API are protected by cookie authentication.
- UI/API authentication uses password-only login for the fixed personal user.
- Password hashes are Argon2id PHC strings stored on the personal `users` row.
- `AUTH_BOOTSTRAP_PASSWORD` is a one-time secret used only to initialize `users.password_hash` when missing.
- Browser auth uses an opaque server-issued session id in `__Host-app-auth`, with `HttpOnly`, `Secure`, `SameSite=Strict`, `Path=/`, no `Domain`, browser-session cookie lifetime, and 24-hour server-side absolute expiry.
- Gateway auth integrates with ASP.NET Core through a custom `"app-cookie"` authentication handler registered by `AddAppCookie()`.
- v0 stores auth sessions in memory behind `ISessionStore`; gateway restart invalidates existing auth sessions by design.
- Multi-replica UI/API auth requires a shared `ISessionStore`, with Redis preferred, and shared login throttling state.
- Secrets must be supplied through environment variables or equivalent deployment secret mechanisms.
- Secrets must not be committed to the repository.

## Configuration

The following configuration keys define the v0 runtime surface:

```text
DATA_DIR
SQLITE_PATH
TELEGRAM_BOT_TOKEN
TELEGRAM_ALLOWED_USER_ID
TELEGRAM_WEBHOOK_SECRET
AUTH_BOOTSTRAP_PASSWORD
LLM_PROVIDER
LLM_API_KEY
LLM_MODEL
JINA_ENABLED
JINA_API_KEY
```

`JINA_API_KEY` is optional configuration for authenticated Jina Reader requests and must be treated as secret material when supplied.

`AUTH_BOOTSTRAP_PASSWORD` is required only before the personal user's `password_hash` has been initialized. It must be exactly 2048 printable ASCII characters and must be treated as secret material.

The UI build uses `VITE_API_BASE_PATH` to choose the same-origin public API base. It defaults to `/api` and is not secret material.

## Key Constraints

- Single-user system in v0.
- Single Go worker instance in v0.
- Single gateway instance in v0 for in-memory auth throttling and ephemeral cookie keys.
- SQLite-backed metadata and job queue.
- Filesystem-backed artifact storage.
- No automatic worker retries or Telegram notification retries in v0.
- No external queue system in v0.
- No Playwright or headless browser rendering in v0.
- No full-text search, filtering, advanced tagging, browser extension, PWA/offline mode, or observability stack in v0.
