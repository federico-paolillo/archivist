# ARCHITECTURE.md

Describes global system architecture: executables, services, boundaries, data ownership, storage, runtime topology, and integration patterns.

Architecture decisions that constrain all features belong here or in `docs/DESIGN.md`.

---

## System Overview

Archivist is a single-user personal article archiving system.

The v0 system accepts article URLs through Telegram, stores article state in SQLite, processes queued article jobs with a single worker, writes large artifacts to the filesystem, generates structured LLM summaries, and exposes a minimal authenticated web UI for review and administration.

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
  - expose authenticated API endpoints for the UI;
  - support basic admin actions, including retry and delete.

### Worker

- Runtime: Go.
- Deployment: single instance in v0.
- Responsibilities:
  - atomically dequeue jobs from SQLite;
  - fetch article HTML over plain HTTP(S);
  - store the raw HTML snapshot;
  - extract readable content using configured extraction candidates;
  - convert selected extracted content to Markdown;
  - score extraction candidates and select the best candidate;
  - call the configured LLM summarizer;
  - validate summary JSON;
  - persist artifacts and final article/job state.

### Web UI

- Runtime/tooling: Preact with Vite.
- Responsibilities:
  - show the article list;
  - show article detail;
  - display summary, key points, tags, Markdown content, original link, and failure messages;
  - expose retry and delete actions.

## Data Storage

SQLite is the source of truth for article state, job state, processing status, artifact paths, and error state.

Filesystem storage under `/data` stores larger raw and derived artifacts:

```text
/data/
  archive.db
  articles/
    {article_id}/
      snapshot.html
      content.md
      summary.json
      metadata.json
```

Artifact writes must be atomic: write to a temporary path and then rename into place. Optional artifact hashes may be stored for integrity checks and debugging.

Core article state includes:

- `id`
- `original_url`
- `canonical_url`
- `title`
- `domain`
- `status`: `pending`, `processing`, `ready`, or `failed`
- `selected_extractor`
- `extractor_score`
- artifact paths for snapshot, Markdown, summary, and metadata
- `error_message`
- `created_at`
- `processed_at`

Core job state includes:

- `id`
- `article_id`
- `type`
- `status`: `queued`, `running`, `succeeded`, `failed`, `retrying`, or `dead`
- `attempts`
- `run_after`
- `locked_at`
- `locked_by`
- `error_message`
- `created_at`

## Service Boundaries and Communication

The gateway and worker communicate through SQLite, not direct RPC.

The UI communicates only with the gateway API. It must not read SQLite or filesystem artifacts directly.

The worker owns processing jobs and filesystem artifact production. The gateway owns request authentication, article/job creation, UI-facing API behavior, and admin actions.

## External Integrations

### Telegram

Telegram is the only v0 ingestion channel. The gateway accepts Telegram webhook requests and rejects requests that do not match the configured webhook secret or allowed Telegram user ID.

### Article Websites

The worker fetches article HTML using direct HTTP(S) requests. v0 does not use Playwright, headless browser rendering, or browser automation.

### Extraction Providers and Libraries

The worker attempts content extraction with:

- Jina Reader when enabled;
- a Go readability implementation.

### LLM Provider

The worker uses a provider-agnostic summarization interface. Provider, API key, and model are configuration values.

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
- Cookie authentication uses `AUTH_COOKIE_SECRET`, a long random secret supplied through configuration.
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
AUTH_COOKIE_SECRET
LLM_PROVIDER
LLM_API_KEY
LLM_MODEL
JINA_ENABLED
```

## Key Constraints

- Single-user system in v0.
- Single Go worker instance in v0.
- SQLite-backed metadata and job queue.
- Filesystem-backed artifact storage.
- No external queue system in v0.
- No Playwright or headless browser rendering in v0.
- No full-text search, filtering, advanced tagging, browser extension, PWA/offline mode, or observability stack in v0.
