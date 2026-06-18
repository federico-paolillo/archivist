# INDEX.md

Global feature index and navigation map. It lists all features, dependencies, and links to their specs and plans.

This file prevents scattered feature folders from becoming unmanageable.

---

## Feature Table

| Slug | Title | Depends On | Spec | Plan |
|---|---|---|---|---|
| `telegram-ingestion` | Telegram Ingestion | — | [`SPEC.md`](./telegram-ingestion/SPEC.md) | [`PLAN.md`](./telegram-ingestion/PLAN.md) |
| `authn` | UI/API Authentication | `telegram-ingestion` | [`SPEC.md`](./authn/SPEC.md) | [`PLAN.md`](./authn/PLAN.md) |
| `article-processing` | URL-To-Article Processing Pipeline | `telegram-ingestion` | [`SPEC.md`](./article-processing/SPEC.md) | [`PLAN.md`](./article-processing/PLAN.md) |
| `markdown-extraction` | Markdown Extraction With Fallbacks | `article-processing` | [`SPEC.md`](./markdown-extraction/SPEC.md) | [`PLAN.md`](./markdown-extraction/PLAN.md) |
| `summary-generation` | Summary Generation | `markdown-extraction` | [`SPEC.md`](./summary-generation/SPEC.md) | [`PLAN.md`](./summary-generation/PLAN.md) |
| `ui-endpoints` | UI Article Endpoints | `authn`, `telegram-ingestion`, `summary-generation` | [`SPEC.md`](./ui-endpoints/SPEC.md) | [`PLAN.md`](./ui-endpoints/PLAN.md) |
| `ui` | Final Browser UI | `authn`, `ui-endpoints` | [`SPEC.md`](./ui/SPEC.md) | [`PLAN.md`](./ui/PLAN.md) |
| `snapshotter` | Snapshotter | `summary-generation`, `ui-endpoints` | [`SPEC.md`](./snapshotter/SPEC.md) | [`PLAN.md`](./snapshotter/PLAN.md) |
| `otel-observability` | OpenTelemetry Observability | `snapshotter` | [`SPEC.md`](./otel-observability/SPEC.md) | [`PLAN.md`](./otel-observability/PLAN.md) |

---
