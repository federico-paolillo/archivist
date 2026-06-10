# Archivist

> You send to Archivist via Telegram a link with a web-page you like and Archivist will: snapshot, summarize and quiz you 

## Launch locally

Archivist runs locally as four processes: the ASP.NET Core Gateway API, the Go Worker, the Vite UI, and the Python Snapshotter.

Create a local data directory and export the shared configuration before starting the processes:

```bash
mkdir -p .local/data

export ARCHIVIST_SQLITE_PATH="$PWD/.local/data/archive.db"
export ARCHIVIST_DATA_DIR="$PWD/.local/data"
export ARCHIVIST_AUTH_BOOTSTRAP_PASSWORD="<2048 printable ASCII characters>"
```

The Worker also requires the extraction and summary provider configuration:

```bash
export ARCHIVIST_JINA_API_KEY="<jina-api-key>"
export ARCHIVIST_LLM_PROVIDER="anthropic"
export ARCHIVIST_LLM_API_KEY="<anthropic-api-key>"
export ARCHIVIST_LLM_MODEL="claude-haiku-4-5-20251001"
```

The Snapshotter also requires S3-compatible Object Storage configuration:

```bash
export ARCHIVIST_SNAPSHOTTER_INTERVAL_SECONDS="86400"
export ARCHIVIST_SNAPSHOTTER_WORK_DIR="/tmp/archivist-snapshotter"
export ARCHIVIST_SNAPSHOTTER_S3_ENDPOINT_URL="<s3-endpoint-url>"
export ARCHIVIST_SNAPSHOTTER_S3_REGION="<s3-region>"
export ARCHIVIST_SNAPSHOTTER_S3_BUCKET="<bucket>"
export ARCHIVIST_SNAPSHOTTER_S3_ACCESS_KEY_ID="<access-key-id>"
export ARCHIVIST_SNAPSHOTTER_S3_SECRET_ACCESS_KEY="<secret-access-key>"
export ARCHIVIST_SNAPSHOTTER_OBJECT_PREFIX=""
```

Set Telegram configuration only when testing webhook ingestion:

```bash
export ARCHIVIST_Telegram__BotToken="<telegram-bot-token>"
export ARCHIVIST_Telegram__WebhookSecret="<telegram-webhook-secret>"
```

Start the Gateway API in one terminal:

```bash
cd src/gateway
dotnet run --project Archivist.Gateway.Api
```

Smoke-check the Gateway:

```bash
curl http://localhost:5178/ping/
```

Start the Worker in a second terminal:

```bash
cd src/worker
go run ./cmd/app process
```

Start the UI in a third terminal:

```bash
cd src/ui
npm install
npm run dev
```

Start the Snapshotter in a fourth terminal after the Python project has been bootstrapped:

```bash
cd src/snapshotter
uv sync --locked --all-extras --dev
uv run archivist-snapshotter
```

The standalone Vite dev server does not proxy `/api`. The full authenticated browser flow requires a same-origin HTTPS reverse proxy that strips `/api` before forwarding to Gateway and sends `X-Forwarded-Proto: https`.

Generate a local-only TLS certificate for `localhost`:

```bash
mkdir -p .local/tls

openssl req -x509 -newkey rsa:2048 -sha256 -days 30 -nodes \
  -keyout .local/tls/localhost.key \
  -out .local/tls/localhost.crt \
  -subj "/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1,IP:0:0:0:0:0:0:0:1"
```

The repository includes `Caddyfile.local` for the local HTTPS ingress:

```caddyfile
https://localhost:8443 {
	tls .local/tls/localhost.crt .local/tls/localhost.key

	handle_path /api/* {
		reverse_proxy localhost:5178 {
			header_up Host localhost:8443
			header_up X-Forwarded-Proto https
			header_up X-Forwarded-Host localhost:8443
			header_up X-Forwarded-For {remote_host}
		}
	}

	handle {
		reverse_proxy 127.0.0.1:5173
	}
}
```

Start the local ingress in another terminal:

```bash
caddy run --config Caddyfile.local
```

Open `https://localhost:8443` and temporarily trust the local certificate when the browser asks. Do not commit `.local/tls/`.

## Docker Compose Deployment

The Compose stack runs Gateway, Worker, UI, Snapshotter, a private OpenTelemetry Collector, and, in development only, Grafana LGTM for manual OTEL validation.

Compose uses a shared base file plus small overlays. `docker-compose.yaml` contains the common service topology. `docker-compose.local.yaml` adds local builds, static local defaults, env-file-backed secrets and external target selectors, and the development-only Grafana LGTM service.

Create `.env.local` from `.env.local.example`, fill the required secrets and external target values, then start the local stack. Static local defaults such as `/data/archive.db`, `/data`, local ports, model defaults, and the local Collector endpoint live in `docker-compose.local.yaml`.

```bash
cp .env.local.example .env.local
docker compose --env-file .env.local -f docker-compose.yaml -f docker-compose.local.yaml up --build
```

Local Compose may call production Telegram, LLM, S3-compatible Object Storage, or OTEL backends if you copy production values into `.env.local`. This is intentional for realistic local validation.

For local OTEL validation, the local Collector exports to the Grafana LGTM container. Grafana is available at `http://localhost:40300`. Login with username `admin` and password `admin`.

Relevant local OTEL variables:

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://otelcol:4318

ARCHIVIST_OTEL_COLLECTOR_IMAGE=otel/opentelemetry-collector-contrib:0.153.0
ARCHIVIST_OTEL_EXPORTER_OTLP_ENDPOINT=http://lgtm:4318
ARCHIVIST_OTEL_EXPORTER_OTLP_AUTHORIZATION=
ARCHIVIST_OTEL_TAIL_SAMPLING_PERCENTAGE=10
ARCHIVIST_OTEL_TAIL_SAMPLING_DECISION_WAIT=10s
```

`OTEL_EXPORTER_OTLP_ENDPOINT` is the application SDK endpoint. It points to Archivist's private Collector. `ARCHIVIST_OTEL_EXPORTER_OTLP_ENDPOINT` is the Collector exporter endpoint. In local development it points to Grafana LGTM. Compose does not expose application-side trace/log exporter disable switches; telemetry is always configured.

### Production reverse-proxy security warning

The production Gateway trusts `X-Forwarded-Proto`, `X-Forwarded-Host`, and related forwarded headers because Archivist is designed to run on an operator-controlled VPS where only ingress Caddy publishes the public application port and Gateway is reachable only on the private Docker network. This is intentional for deployments behind load balancers with dynamic source IPs, where static trusted-proxy IP configuration is brittle.

**Do not publish Gateway directly to the Internet.** If Gateway is exposed directly while it trusts forwarded headers, a client can send spoofed forwarded values. What might happen: Gateway could evaluate login HTTPS checks, public host checks, URL generation, or audit context using attacker-supplied scheme/host data instead of the real connection context. Keep only ingress Caddy publicly reachable, keep Gateway private, and make Caddy overwrite forwarded headers before proxying `/api/*`.

## Production OTEL Deployment

Production releases package `docker-compose.yaml`, `docker-compose.prod.yaml`, `.env`, `.env.images`, `rp.Caddyfile`, and `otelcol-config.yaml`. The packaged `.env` is copied from `.env.example` and is the production variable reference. Fill every `<specify>` value before deployment. Production Compose has no default fallbacks: required variables must be set and non-empty, while optional variables listed with empty values may remain empty.

Deploy with the packaged `.env` first and `.env.images` second so release image pins win:

```bash
docker compose --env-file .env --env-file .env.images -f docker-compose.yaml -f docker-compose.prod.yaml up -d
```

The production stack includes the private Archivist Collector but does not include Grafana LGTM. Configure the Collector exporter for your Grafana-compatible OTLP backend:

```bash
ARCHIVIST_OTEL_COLLECTOR_IMAGE=otel/opentelemetry-collector-contrib:0.153.0
ARCHIVIST_OTEL_EXPORTER_OTLP_ENDPOINT=<grafana-compatible-otlp-http-endpoint>
ARCHIVIST_OTEL_EXPORTER_OTLP_AUTHORIZATION=<authorization-header-value>
ARCHIVIST_OTEL_TAIL_SAMPLING_PERCENTAGE=10
ARCHIVIST_OTEL_TAIL_SAMPLING_DECISION_WAIT=10s
```

Keep Collector OTLP receiver ports private on the Docker network. Do not publish `4318` to the host. Only ingress Caddy should publish the public application port.

The Collector tail-samples traces:

- all traces with at least one span status of `ERROR` are retained;
- 10% of non-error traces are retained.

Applications export traces with always-on SDK sampling so the Collector can make the sampling decision after seeing trace outcomes.
Gateway emits selective HTTP failure logs for security-relevant `401`/`403` responses and operational `5xx` responses, while suppressing routine unauthenticated `GET /auth/session` probes and successful request noise.

Collector runtime outages must not stop Gateway, Worker, or Snapshotter core behavior. Standard OTEL exporters may drop telemetry after bounded retries/timeouts while the application continues. Invalid telemetry configuration may still fail startup.

## Manual OTEL Validation

After starting the local Compose stack:

1. Open `http://localhost:40300`.
2. Submit an article through Telegram or enqueue a URL with the Worker CLI.
3. In Grafana, inspect traces for Gateway inbound HTTP spans.
4. Confirm Worker processing continues the Gateway trace for Telegram-created jobs.
5. Confirm Worker CLI-enqueued jobs create traces without requiring a parent.
6. Confirm Snapshotter emits independent snapshot attempt traces.
7. Inspect logs and confirm records inside spans include `trace_id` and `span_id`.
8. Search logs/traces by `article_id`, `job_id`, URL, or provider request ID as attributes.
9. Confirm those high-cardinality values are not Loki labels or metric labels.
10. Stop `otelcol` and confirm core application behavior continues while telemetry export fails non-fatally.
