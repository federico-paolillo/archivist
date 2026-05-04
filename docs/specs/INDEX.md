# INDEX.md

Global feature index and navigation map. It should list all features, status, dependencies, and links to their specs and plans. 

This file prevents scattered feature folders from becoming unmanageable.

---

## Feature Table

| Slug | Title | Status | Depends On | Spec | Plan |
|---|---|---|---|---|---|
| `telegram-ingestion` | Telegram Ingestion | draft | — | [`SPEC.md`](./telegram-ingestion/SPEC.md) | [`PLAN.md`](./telegram-ingestion/PLAN.md) |
| `article-processing` | URL-To-Article Processing Pipeline | draft | `telegram-ingestion` | [`SPEC.md`](./article-processing/SPEC.md) | [`PLAN.md`](./article-processing/PLAN.md) |
| `markdown-extraction` | Markdown Extraction With Fallbacks | draft | `article-processing` | [`SPEC.md`](./markdown-extraction/SPEC.md) | [`PLAN.md`](./markdown-extraction/PLAN.md) |

Status values: `draft` \| `in_progress` \| `done` \| `blocked` \| `skipped`

---
