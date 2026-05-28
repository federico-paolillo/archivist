---
id: UI-003-PLAN
task: ../tasks/UI-003-article-master-detail-and-delete-workflow.md
status: completed
canonical: true
---

# ExecPlan: UI-003 Article Master-Detail View and Delete Workflow

## Objective

Implement the article master-detail UI, detail state machine, safe Markdown rendering, Original action, and Delete confirmation flow.

## Linked Task

- `../tasks/UI-003-article-master-detail-and-delete-workflow.md`

## Required Context

Start from the linked task's `Required Context` and the linked task file:

- `../tasks/UI-003-article-master-detail-and-delete-workflow.md`

Add only ExecPlan-specific context:

- `docs/specs/ui-endpoints/PLAN.md`
- `docs/specs/ui-endpoints/tasks/UIEND-002-gateway-article-read-api.md`
- `docs/specs/ui-endpoints/tasks/UIEND-003-gateway-article-delete-api.md`
- `docs/design/view.png`

## Assumptions

- `UI-002` has implemented the router, API client, authenticated shell, and design primitives.
- `UIEND-002` and `UIEND-003` have implemented the article read/delete APIs.
- Article status values consumed by the UI are `queued`, `ready`, and `failed`; any future non-ready/non-failed value follows the `Come back later.` state.

## Non-Goals

- Do not implement Retry or requeue.
- Do not implement search, filtering, sorting controls, or pagination UI beyond consuming the first fixed page.
- Do not read SQLite or filesystem artifacts directly.
- Do not display raw HTML from article Markdown.

## Implementation Sequence

1. Add article API client methods for list, detail, and delete using the configured API base and credentials.
2. Define TypeScript DTOs matching `ui-endpoints`: lower-camel metadata fields plus `summaryMarkdown` and `contentMarkdown` for detail.
3. Extend the existing `src/ui/src/pages/articles/articles.tsx` page and page-local components under `src/ui/src/pages/articles/components/`; do not move article page implementation back into `app.tsx`.
4. Render the master-detail grid under the authenticated shell: fixed top bar, bordered master pane, bordered detail pane, dense ledger rows, black/gray surfaces, no rounded corners.
5. Load the first article page from `GET /articles` when the authenticated articles shell mounts.
6. Render article rows with id, title fallback, URL, status, and selected inversion matching the design system.
7. On row click, navigate to `/articles/{id}` immediately and start detail loading.
8. For direct `/articles/{id}` routes, load the article detail for the route id after the session gate succeeds.
9. Show a design-compatible spinner during detail loading.
10. Render `/articles` with no selected id as a blank black detail pane.
11. For ready details, render title, summary markdown, content markdown, Original action, and Delete action.
12. Render Markdown with raw HTML disabled. If adding a Markdown dependency, prefer a renderer that escapes raw HTML by configuration; document the dependency reason in the task diary.
13. Validate Markdown links so `javascript:` URLs cannot execute. External links opened from rendered content must use `rel="noopener noreferrer"`.
14. For queued and other non-ready/non-failed details, render centered white text exactly `Come back later.` while retaining available article actions.
15. For failed details, render the persisted `errorMessage` centered in red while retaining available article actions.
16. For detail fetch failures, render the failure text centered in red.
17. Implement `Original` as a link to `canonicalUrl` when present, otherwise `originalUrl`, opening in a new tab/window with `noopener` and `noreferrer`.
18. Implement `Delete` with a modal containing exactly `Are you sure?`, `Yes`, and `Nevermind`.
19. Ensure `Nevermind` closes the modal without sending a request.
20. Ensure `Yes` sends `DELETE /articles/{id}` through the configured API base.
21. On successful delete, remove the article from local list state, navigate to `/articles`, and clear the detail pane.
22. On delete failure, keep the current URL selected and show the failure text centered in red.
23. Add tests for API calls, route updates, no-id detail, loading, ready, queued, failed, fetch-error, Original link, delete cancel, delete success, and delete failure.
24. Update task status, feature plan, and diary after implementation and validation if this task is completed.

## Validation Plan

```bash
cd src/ui && npm run format
cd src/ui && npm run lint
cd src/ui && npm run build
cd src/ui && npm run test
```

Manual checks:

- `/articles` matches the blank-detail master-detail frame.
- `/articles/<article_id>` ready state matches the screenshot constraints.
- Queued, failed, loading, fetch-error, and delete modal states preserve the design system.

## Documentation Updates Required

- Update `../tasks/UI-003-article-master-detail-and-delete-workflow.md` status when complete.
- Update `../PLAN.md` task table when status changes.
- Append a `DIARY.md` entry with validation results.
- Promote any Markdown renderer dependency or sanitizer policy to `docs/conventions/UI.md` if it becomes durable.

## Risks

- Rendering Markdown with unsafe HTML handling can create XSS from archived article content.
- The public `/articles` page route and article API route collision must stay separated through the configured API base.
- The screenshot includes a Retry control, but v0 explicitly excludes Retry; do not copy that button into implementation.

## Rollback / Recovery Notes

- If Markdown rendering cannot be made safe with the selected dependency, render Markdown as escaped preformatted text until a safe renderer is selected and documented.
- If delete behavior fails, keep the article selected and surface the API error without mutating local list state.

## Completion Criteria

- Article UI acceptance criteria pass.
- Markdown rendering safety is covered by tests or documented manual validation.
- Validation commands pass or failures are recorded.
- Manual article route checks are recorded in `DIARY.md`.
