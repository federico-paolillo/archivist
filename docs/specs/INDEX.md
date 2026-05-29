# INDEX.md

Global feature index and navigation map. It should list all features, status, dependencies, and links to their specs and plans. 

This file prevents scattered feature folders from becoming unmanageable.

---

## Feature Table

| Slug | Title | Status | Depends On | Spec | Plan |
|---|---|---|---|---|---|
| `telegram-ingestion` | Telegram Ingestion | done | — | [`SPEC.md`](./telegram-ingestion/SPEC.md) | [`PLAN.md`](./telegram-ingestion/PLAN.md) |
| `authn` | UI/API Authentication | done | `telegram-ingestion` | [`SPEC.md`](./authn/SPEC.md) | [`PLAN.md`](./authn/PLAN.md) |
| `article-processing` | URL-To-Article Processing Pipeline | done | `telegram-ingestion` | [`SPEC.md`](./article-processing/SPEC.md) | [`PLAN.md`](./article-processing/PLAN.md) |
| `markdown-extraction` | Markdown Extraction With Fallbacks | done | `article-processing` | [`SPEC.md`](./markdown-extraction/SPEC.md) | [`PLAN.md`](./markdown-extraction/PLAN.md) |
| `worker-runtime-configuration` | Worker Runtime Configuration | done | `markdown-extraction` | [`SPEC.md`](./worker-runtime-configuration/SPEC.md) | [`PLAN.md`](./worker-runtime-configuration/PLAN.md) |
| `summary-generation` | Summary Generation | done | `markdown-extraction`, `worker-runtime-configuration` | [`SPEC.md`](./summary-generation/SPEC.md) | [`PLAN.md`](./summary-generation/PLAN.md) |
| `ui-endpoints` | UI Article Endpoints | done | `authn`, `telegram-ingestion`, `summary-generation` | [`SPEC.md`](./ui-endpoints/SPEC.md) | [`PLAN.md`](./ui-endpoints/PLAN.md) |
| `ui` | Final Browser UI | done | `authn`, `ui-endpoints` | [`SPEC.md`](./ui/SPEC.md) | [`PLAN.md`](./ui/PLAN.md) |

Status values: `draft` \| `in_progress` \| `done` \| `blocked` \| `skipped`

---
