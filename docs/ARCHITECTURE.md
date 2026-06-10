# ARCHITECTURE.md

Describes global system architecture: executables, services, boundaries, data ownership, storage, runtime topology, and integration patterns.

Architecture decisions that constrain all features belong here or in `docs/DESIGN.md`.

---

## System Overview

Archivist is a personal article archiving system with one bootstrapped user in v0 and a user-aware ownership model.

The v0 system accepts article URLs through Telegram, stores article state in SQLite, processes queued article jobs with a single worker, writes large artifacts to the filesystem, generates text LLM summaries, and exposes a minimal authenticated web UI for review and administration.

High-level flow:

```text
Telegram Bot
  -> ASP.NET Core Minimal API gateway
  -> SQLite articles and jobs
  -> Go worker
  -> SQLite plus filesystem artifacts under /data
  -> Python snapshotter uploads /data backups to S3-compatible Object Storage
  -> ASP.NET Core API
  -> Preact/Vite UI
```

The system favors a small, rebuildable deployment over horizontal scale. SQLite owns authoritative state. Filesystem artifacts are derived or retained content associated with article records. Runtime Gateway and Worker code must resolve article and job ownership from SQLite, authenticated session state, or claimed job state. Authentication bootstrap may hardcode the initial personal user id and fixed personal Telegram sender id `1559957191`. Worker CLI enqueue is the only runtime exception: it uses `jobs.DefaultUserID = 01ASB2XFCZJY7WHZ2FNRTMQJCT`, checks that `users.id` exists by that id, never infers ownership from user-table cardinality, and never creates the user.

## Executables and Services

### Gateway API

- Runtime: ASP.NET Core Minimal API.
- Responsibilities:
  - receive Telegram webhook requests;
  - authenticate Telegram webhook requests using the configured secret;
  - accept URLs only from Telegram senders mapped by `users.telegram_user_id`;
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
- Production command: `archivist-worker process`.
- Responsibilities:
  - atomically dequeue jobs from SQLite;
  - fetch article HTML over plain HTTP(S);
  - store the raw HTML snapshot;
  - extract readable content with go-readability v2 first;
  - fall back to Jina Reader when local readability cannot produce Markdown;
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

### Snapshotter

- Runtime/tooling: Python 3.12 with `uv`, `ruff`, `ty`, and `pytest`.
- Deployment: one background service in the Docker application stack.
- Production command: `archivist-snapshotter`.
- Responsibilities:
  - sleep for the configured interval before taking the first snapshot;
  - periodically stage `/data` backups;
  - copy the configured SQLite database through the SQLite online backup API;
  - copy non-database `/data` files best-effort while Gateway and Worker may continue writing;
  - create a single `.tar.gz` archive named `archivist-<yyyy-mm-dd>-<unix-timestamp>.tar.gz`;
  - include a root `manifest.json` describing the snapshot and its consistency limits;
  - upload the archive to S3-compatible Object Storage with explicit endpoint, region, bucket, object key, access key, and secret key configuration;
  - emit structured stdout logs without secrets or article content.

Snapshotter does not own restore, remote retention, pruning, encryption, Gateway/UI status, or writer coordination in v0. Failed snapshot attempts are logged, temporary files are cleaned up, and the service continues to the next interval.

### OpenTelemetry Collector

- Runtime: official OpenTelemetry Collector Contrib container.
- Deployment: one private service on the Docker internal network.
- Responsibilities:
  - receive OTLP HTTP traces and logs from Gateway, Worker, and Snapshotter;
  - tail-sample traces, keeping all traces with at least one span status of `ERROR` and 10% of non-error traces;
  - export telemetry to the configured Grafana-compatible OTLP backend;
  - keep OTLP receiver ports private to the Docker network.

The development Compose stack also includes a minimal Grafana LGTM container for manual local validation. Grafana LGTM is development-only and is not part of the production runtime contract.

### Public Browser/API Routing

The browser UI owns page routes such as `/login` and `/articles`. To avoid collisions with Gateway's intentionally unprefixed API routes, the UI calls Gateway through a configured same-origin API base path.

- Default public UI API base: `/api`.
- Vite build-time configuration: `VITE_API_BASE_PATH`, default `/api`.
- Reverse proxy behavior: public `/api/*` requests are forwarded to Gateway with the `/api` prefix stripped.
- Gateway route contracts remain unprefixed, for example `POST /login`, `GET /articles`, and `DELETE /articles/{id}`.
- Public `/api/login`, `/api/logout`, and `/api/auth/session` reach Gateway as `POST /login`, `POST /logout`, and `GET /auth/session`.
- Public root-level `/login`, `/articles`, and `/articles/{article_id}` remain UI routes.

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

The canonical artifact path contract is defined in `docs/ARTIFACTS.md`.

Core user state includes:

- `id`: seeded as `01ASB2XFCZJY7WHZ2FNRTMQJCT` for the personal account in v0
- `telegram_user_id`: nullable until bootstrap or another canonical user-provisioning path maps a Telegram sender id, unique when present. Auth bootstrap sets the personal row to `1559957191` only when this value is null and preserves an existing non-null value.
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
- `traceparent`
- `tracestate`

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

`jobs.traceparent` and `jobs.tracestate` are nullable W3C Trace Context carrier fields. Gateway writes them when creating a queued job inside a traced request. Worker extracts them when claiming the job to continue Gateway-originated traces. Worker-created CLI enqueue jobs may leave them null; processing those jobs starts a valid root trace.

Worker CLI enqueue uses `jobs.DefaultUserID = 01ASB2XFCZJY7WHZ2FNRTMQJCT` as the default owner for operator-created jobs. Before inserting an article or job, the Worker must query `users.id` for that exact id. Missing default user fails enqueue. Additional user rows do not change the selected owner or cause failure. Worker CLI enqueue must not create, upsert, or repair `users`.

In final v0 processing, `articles.status = ready`, `jobs.status = succeeded`, and the success notification row are committed only after `summary.md` has been atomically written. HTML snapshotting and Markdown extraction are intermediate stages once summary generation exists. Any terminal failure in fetch, snapshot, Markdown extraction, summarization, or artifact writing marks the article and job failed with an ARC-coded public error.

## Service Boundaries and Communication

The gateway and worker communicate through SQLite, not direct RPC.

The UI communicates only with the gateway API through the configured API base path. It must not read SQLite or filesystem artifacts directly.

The worker owns processing jobs, filesystem artifact production, final article/job state, and creation of terminal notification rows. The gateway owns request authentication, article/job creation, Telegram API calls, terminal notification dispatch, UI-facing API behavior, and admin actions.

## External Integrations

### Telegram

Telegram is the only v0 ingestion channel. The gateway accepts Telegram webhook requests and rejects requests that do not match the configured webhook secret. Sender authorization is based on `users.telegram_user_id`: a sender is authorized only when the sender id maps to an existing Archivist user row.

Authorized Telegram messages must contain exactly one trimmed absolute `http` or `https` URL. Invalid authorized messages receive `Nope, you must send only an URL`. Valid queued URL messages receive `Ok, I will have a look` after the article/job enqueue transaction commits. Completion replies are sent later by the gateway from SQLite notification rows, as replies to the original Telegram message.

The gateway persists the Telegram sender user ID separately from `telegram_chat_id` and `telegram_message_id`. `telegram_user_id` is sender identity metadata. `telegram_chat_id` and `telegram_message_id` are reply-target metadata. Accepted Telegram ingestion stores the resolved Archivist `user_id` on both the article and job rows.

The v0 personal account ULID must not be treated as a catch-all for runtime ingestion. Auth bootstrap seeds the personal row's Telegram sender mapping with `1559957191` only when `users.telegram_user_id` is null. Deployment-configured Telegram sender allowlists or personal sender settings are not bootstrap inputs and are not runtime webhook authorization gates.

The worker must not call Telegram APIs directly.

### Article Websites

The worker fetches article HTML using direct HTTP(S) requests. v0 does not use Playwright, headless browser rendering, or browser automation.

### Extraction Providers and Libraries

The worker attempts Markdown extraction in this order:

1. go-readability v2 from the saved HTML snapshot.
2. Jina Reader fallback when go-readability `CheckDocument()` returns false, local extraction fails, or local Markdown conversion fails.

The Worker logs critical extraction decisions, including fallback from go-readability to Jina. If both local extraction and Jina fallback fail, the job becomes terminally failed.

### LLM Provider

The Worker uses a provider-agnostic summarization interface. Claude through Anthropic is the first v0 provider, but provider, API key, and model are configuration values. The Anthropic implementation must use official Anthropic SDKs when suitable SDKs exist for the implementation language. Provider-specific SDK types must not leak outside the summarizer adapter.

## Runtime Topology

v0 deploys all components together on a single VPS.

Deployment requirements:

- one shared `/data` volume for SQLite and article artifacts;
- gateway, worker, UI, and snapshotter deployed as one application stack;
- only ingress Caddy publishes a host port from the Docker stack;
- Gateway is private on the Docker internal network and has no host-published port;
- Snapshotter backs up `/data` to S3-compatible Object Storage;
- Gateway, Worker, and Snapshotter send OTLP traces and logs to the private Collector;
- the Collector exports telemetry to the configured Grafana-compatible backend and must not publish OTLP ports to the host;
- stdout logging collected by the host or deployment environment.

The v0 topology does not target high scalability, multi-region deployment, or real-time processing guarantees.

### Repository Automation

GitHub Actions is the canonical repository automation surface. CI runs on pushes to `main` and must fail when Gateway, Worker, UI, or Snapshotter build, lint, formatting, type-checking, or test validation fails. Coverage upload is deferred until the component test commands emit coverage reports.

Gateway validation uses the exact .NET SDK pinned by `src/gateway/global.json`. The Gateway container build stage must use the matching `mcr.microsoft.com/dotnet/sdk` image tag so local, CI, and release builds use the same SDK feature band.

CD is a manually dispatched release workflow. It checks out the requested release ref, validates the same component gates as CI, builds and pushes multi-architecture `linux/amd64` and `linux/arm64` images for Gateway, Worker, UI, and Snapshotter to GitHub Container Registry, emits GitHub artifact attestations for each pushed image, tags the resolved release commit, and opens a draft GitHub release. The UI image build receives the resolved release commit SHA through the `VERSION_LABEL` Docker build argument.

Compose uses a shared base plus environment overlays. `docker-compose.yaml` owns the common service topology, networks, volumes, dependencies, and shared runtime constants. `docker-compose.local.yaml` adds local `build:` entries, local image tags, static local defaults, env-file-backed secrets and external target selectors, and the development-only Grafana LGTM service. `docker-compose.prod.yaml` has no `build:` entries and receives Gateway, Worker, UI, and Snapshotter images through `ARCHIVIST_GATEWAY_IMAGE`, `ARCHIVIST_WORKER_IMAGE`, `ARCHIVIST_UI_IMAGE`, and `ARCHIVIST_SNAPSHOTTER_IMAGE`.

The CD workflow generates a release package under `release/compose/`, validates the packaged production Compose model with Docker Compose, and publishes it as a compressed deployment artifact attached to both the workflow run and the draft GitHub release. The package contents are `docker-compose.yaml`, `docker-compose.prod.yaml`, `.env`, digest-pinned `.env.images`, `rp.Caddyfile`, and `otelcol-config.yaml`. Operators deploy with the packaged runtime `.env` first and the release-provided `.env.images` file second so release image pins override accidental local image values, and they must pass both packaged Compose files with `-f docker-compose.yaml -f docker-compose.prod.yaml`.

### Reverse Proxy And TLS Termination

The primary v0 public topology is:

```text
Internet -> Scaleway Load Balancer TLS termination -> Docker host port 65000 -> ingress Caddy plaintext HTTP
  -> /api/* -> Gateway on Docker internal network
  -> other routes -> UI Caddy on Docker internal network
```

DNS binding, public IP binding, certificate provisioning, and public Internet exposure are external to the application stack. The VPS/cloud-provider layer terminates TLS before traffic reaches Caddy. Caddy does not own or present the public certificate in this topology.

Caddy receives plaintext HTTP on host port `65000` after upstream TLS termination. It must listen with `http://:65000` for the primary Scaleway load-balancer topology. A bare `:65000` Caddy site is an HTTPS listener and requires a certificate strategy, so it is prohibited for this primary upstream-terminated deployment.

Caddy must overwrite forwarded headers before proxying to Gateway:

```caddyfile
http://:65000 {
    encode zstd gzip

    handle_path /api/* {
        reverse_proxy gateway:8080 {
            header_up Host {host}
            header_up X-Forwarded-Proto https
            header_up X-Forwarded-For {remote_host}
            header_up X-Forwarded-Host {host}
        }
    }

    handle {
        reverse_proxy ui:8080
    }
}
```

`X-Forwarded-Proto` must be the literal value `https`. Using Caddy's `{scheme}` would forward `http` in this topology and make Gateway reject HTTPS-required auth flows. `GATEWAY_PUBLIC_HOSTS` must match the public hostname supplied through the forwarded host context. Gateway trusts forwarded headers because only the reverse proxy is supposed to reach it on the Docker internal network; this is a documented deployment invariant rather than a static source-IP allowlist because the upstream load balancer may use dynamic source IPs.

If a future deployment sends TLS all the way to Caddy, then Caddy becomes the TLS endpoint and must be configured with a real certificate strategy. That alternate topology must be documented before use.

## Security Boundaries

The v0 deployment has one bootstrapped user, but runtime authorization is user-aware.

Security boundaries:

- Telegram ingestion is limited to senders mapped by `users.telegram_user_id`.
- Telegram webhook requests require `Telegram:WebhookSecret`.
- The entire UI and UI-facing API are protected by cookie authentication.
- UI/API authentication uses password-only login against all password-bearing user rows. Login loads every non-empty Argon2id PHC hash, verifies the submitted password against every candidate, and issues a session only when exactly one user matches. Multiple password-bearing rows are valid; duplicate password matches fail closed.
- Password hashes are Argon2id PHC strings stored on password-bearing `users` rows.
- `AUTH_BOOTSTRAP_PASSWORD` is a one-time secret used only to initialize `users.password_hash` when missing.
- Browser auth uses an opaque server-issued session id in `__Host-app-auth`, with `HttpOnly`, `Secure`, `SameSite=Strict`, `Path=/`, no `Domain`, browser-session cookie lifetime, and 24-hour server-side absolute expiry.
- UI/API auth HTTPS decisions use the effective public request context after trusted forwarded-header processing.
- Gateway must not be exposed directly to the public Internet while trusting forwarded headers. If it is exposed directly, clients may spoof `X-Forwarded-Proto` or `X-Forwarded-Host`, which can make Gateway evaluate auth HTTPS and host decisions using attacker-supplied public context.
- Gateway auth integrates with ASP.NET Core through a custom `"app-cookie"` authentication handler registered by `AddAppCookie()`.
- v0 stores auth sessions in memory behind `ISessionStore`; gateway restart invalidates existing auth sessions by design.
- Multi-replica UI/API auth requires a shared `ISessionStore`, with Redis preferred, and shared login throttling state.
- Secrets must be supplied through environment variables or equivalent deployment secret mechanisms.
- Secrets must not be committed to the repository.

## Configuration

The following logical configuration keys define the v0 runtime surface:

```text
DATA_DIR
SQLITE_PATH
ARCHIVIST_PUBLIC_PORT
Telegram:BotToken
Telegram:WebhookSecret
AUTH_BOOTSTRAP_PASSWORD
LLM_PROVIDER
LLM_API_KEY
LLM_MODEL
JINA_API_KEY
GATEWAY_PUBLIC_HOSTS
VITE_API_BASE_PATH
SNAPSHOTTER_INTERVAL_SECONDS
SNAPSHOTTER_WORK_DIR
SNAPSHOTTER_S3_ENDPOINT_URL
SNAPSHOTTER_S3_REGION
SNAPSHOTTER_S3_BUCKET
SNAPSHOTTER_S3_ACCESS_KEY_ID
SNAPSHOTTER_S3_SECRET_ACCESS_KEY
SNAPSHOTTER_OBJECT_PREFIX
OTEL_SERVICE_NAME
OTEL_RESOURCE_ATTRIBUTES
OTEL_EXPORTER_OTLP_ENDPOINT
OTEL_TRACES_SAMPLER
ARCHIVIST_OTEL_COLLECTOR_IMAGE
ARCHIVIST_OTEL_EXPORTER_OTLP_ENDPOINT
ARCHIVIST_OTEL_EXPORTER_OTLP_AUTHORIZATION
ARCHIVIST_OTEL_TAIL_SAMPLING_PERCENTAGE
ARCHIVIST_OTEL_TAIL_SAMPLING_DECISION_WAIT
```

Gateway uses the default ASP.NET Core application configuration sources, appends `ARCHIVIST_`-prefixed environment variables, and creates the builder without command-line arguments. Standalone Gateway keys remain flat, for example `ARCHIVIST_SQLITE_PATH`. Option-bound groups use hierarchy with double underscores in environment variables, for example `ARCHIVIST_Telegram__BotToken` and `ARCHIVIST_Telegram__WebhookSecret`.

Worker uses configuro from `src/worker/pkg/app/config`, loads `ARCHIVIST_`-prefixed environment variables, and exposes runtime values through the config structs. Canonical Worker environment variables are `ARCHIVIST_SQLITE_PATH`, `ARCHIVIST_DATA_DIR`, `ARCHIVIST_JINA_API_KEY`, `ARCHIVIST_LLM_PROVIDER`, `ARCHIVIST_LLM_API_KEY`, and `ARCHIVIST_LLM_MODEL`. `SQLITE_PATH`, `DATA_DIR`, `JINA_API_KEY`, and `LLM_API_KEY` for the Anthropic provider are required when `config.Load()` runs; missing required values fail startup before the Worker composition root is built.

`JINA_API_KEY` is required configuration for Jina Reader fallback and must be treated as secret material.

`AUTH_BOOTSTRAP_PASSWORD` is required only before the personal user's `password_hash` has been initialized. It must be exactly 2048 printable ASCII characters and must be treated as secret material.

`GATEWAY_PUBLIC_HOSTS` is a comma-separated public host allowlist for trusted forwarded host values. It is not secret material.

The UI build uses `VITE_API_BASE_PATH` to choose the same-origin public API base. It defaults to `/api` and is not secret material.

Snapshotter reads `ARCHIVIST_`-prefixed application variables and standard `OTEL_*` SDK variables set by Compose. Canonical Snapshotter application variables are `ARCHIVIST_DATA_DIR`, `ARCHIVIST_SQLITE_PATH`, `ARCHIVIST_SNAPSHOTTER_INTERVAL_SECONDS`, `ARCHIVIST_SNAPSHOTTER_WORK_DIR`, `ARCHIVIST_SNAPSHOTTER_S3_ENDPOINT_URL`, `ARCHIVIST_SNAPSHOTTER_S3_REGION`, `ARCHIVIST_SNAPSHOTTER_S3_BUCKET`, `ARCHIVIST_SNAPSHOTTER_S3_ACCESS_KEY_ID`, `ARCHIVIST_SNAPSHOTTER_S3_SECRET_ACCESS_KEY`, and `ARCHIVIST_SNAPSHOTTER_OBJECT_PREFIX`. `ARCHIVIST_SNAPSHOTTER_INTERVAL_SECONDS` defaults to `86400`. The Snapshotter executable defaults `ARCHIVIST_SNAPSHOTTER_WORK_DIR` to `/tmp/archivist-snapshotter`; Compose deployments set it to `/work/archivist-snapshotter` on a disk-backed `snapshotter-work` volume so snapshot staging is not constrained by tmpfs memory. `ARCHIVIST_SNAPSHOTTER_OBJECT_PREFIX` defaults to empty. The S3 endpoint URL, region, bucket, access key id, and secret access key are required at startup. The access key id and secret access key are secret material.

Local Compose is `docker-compose.yaml` plus `docker-compose.local.yaml`; static local defaults are written directly in `docker-compose.local.yaml`, while `.env.local` supplies secrets and external target selectors such as S3 destination values and an optional OTEL backend override. Local runs may intentionally target production Telegram, LLM, S3-compatible Object Storage, or OTEL systems when those values are copied into `.env.local`. Production Compose is `docker-compose.yaml` plus `docker-compose.prod.yaml`; it reads the packaged `.env` and `.env.images` files and has no default fallbacks for configurable values: required variables must be set and non-empty, while documented optional variables may be explicitly empty. Compose configures application SDKs with `OTEL_SERVICE_NAME`, `OTEL_RESOURCE_ATTRIBUTES`, `OTEL_EXPORTER_OTLP_ENDPOINT`, and always-on trace sampling. Application-side trace/log exporter disable switches are not part of the deployment contract. Collector-specific deployment variables use `ARCHIVIST_OTEL_*`. Applications must keep core behavior working during Collector runtime outages. Invalid telemetry configuration may fail startup. Gateway selectively logs security-relevant HTTP `401`/`403` responses and operational `5xx` responses without enabling broad request logging; routine unauthenticated `GET /auth/session` probes are not logged. Gateway marks `5xx` request activities and caught operational failures that return `5xx` as `ERROR` so Collector tail sampling retains those traces. Telemetry must not expose secrets, cookies, Telegram bot tokens, authentication headers, full article HTML, full Markdown, full summaries, provider payloads, or S3 credentials. High-cardinality values such as `user_id`, `article_id`, `job_id`, URLs, and provider request IDs remain trace/log attributes and must not be promoted to Loki labels or metric labels. Gateway and Worker attach `user_id` when the Archivist user is resolved; Snapshotter does not attach `user_id`.

## Key Constraints

- One bootstrapped user in v0; runtime ownership and authorization paths remain user-aware.
- Single Go worker instance in v0.
- Single gateway instance in v0 for in-memory auth sessions and login throttling.
- Single snapshotter instance in v0.
- SQLite-backed metadata and job queue.
- Filesystem-backed artifact storage.
- No automatic worker retries or Telegram notification retries in v0.
- No external queue system in v0.
- No Playwright or headless browser rendering in v0.
- No full-text search, filtering, advanced tagging, browser extension, or PWA/offline mode in v0.
