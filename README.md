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
export ARCHIVIST_Telegram__AllowedUserId="<telegram-user-id>"
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
