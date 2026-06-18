---
name: archivist-ui
description: Use when implementing, reviewing, or planning Archivist UI changes under src/ui, including Preact, Vite, Tailwind CSS, routing, dependency composition, Markdown rendering, authentication client behavior, tests, and UI validation.
---

# Archivist UI

Use this skill for UI work under `src/ui/`.

## Required Context

Start with the orientation bundle:

```text
AGENTS.md
docs/REBUILD.md
docs/specs/INDEX.md
```

Load canonical docs by task trigger:

- `docs/ARCHITECTURE.md`: public routing, API base path, deployment, storage, or service boundaries.
- `docs/DESIGN.md`: accepted durable UI/design decisions.
- `docs/ARTIFACTS.md`: artifact display assumptions.
- `docs/ERRORS.md`: ARC public error display.
- Relevant feature `SPEC.md`, `PLAN.md`, task files, and active-run ExecPlans before implementation.

## Stack

- Frontend code lives under `src/ui/`.
- Runtime/build: TypeScript ESM, Vite, Preact, `preact-iso`.
- Styling: Tailwind CSS v4 through the Vite plugin and `src/ui/src/app.css`.
- Tests: Vitest, jsdom, Testing Library, `@testing-library/user-event`.
- Lint/format: Biome.

## Project Structure

- `src/ui/src/app.tsx` owns route composition only: `LocationProvider`, `ErrorBoundary`, `Router`, and route declarations.
- Pages live under `src/ui/src/pages/<pagename>/<pagename>.tsx`.
- Page directory names use route-oriented kebab-case.
- Page-specific components live under `src/ui/src/pages/<pagename>/components/`.
- Globally reusable components live under `src/ui/src/components/`.
- Page code must not import page-specific components from another page.
- Promote a component to `src/ui/src/components/` only when it is genuinely reused.

## Dependency Composition

- Use Poor Man's DI through `src/ui/src/deps.ts`.
- Collect dependencies in `deps.ts` and initialize them once in `main.tsx` with `makeDeps(): Deps`.
- Use factory functions in `deps.ts` for transient dependencies.
- Page modules receive dependencies through props from the route composition root.

## Styling And UX

- Put shared component utility classes in `@layer components` in `app.css`.
- Reuse existing classes such as `btn-primary`, `btn-secondary`, `btn-outline`, and `input-field`.
- Preserve semantic HTML, accessible names, keyboard behavior, focus states, and responsive layout.
- For personal-use UI fixes, prefer minimal standards-compliant accessibility semantics that match existing interactions over heavyweight ARIA patterns. Do not declare ARIA roles such as `menu`, `listbox`, or `grid` unless the full keyboard and focus-management pattern is implemented.

## API, Auth, And Markdown

- `VITE_API_BASE_PATH` selects the same-origin public API base path and defaults to `/api`.
- Normalize API base paths to avoid double slashes.
- Do not add `/api` prefixes to Gateway route definitions from UI code.
- Authenticated requests must use same-origin `fetch` with credentials included.
- The UI must not store passwords in local storage, session storage, IndexedDB, or URL state.
- Markdown article content is untrusted. Render with raw HTML disabled or sanitized; do not execute raw HTML, scripts, inline event handlers, or `javascript:` links.
- Rendered Markdown links opened in a new browsing context must use `rel="noopener noreferrer"`.
- Retry/requeue controls require a backend retry contract before UI implementation.

## Testing And Validation

- Frontend tests are colocated as `src/ui/src/**/*.test.{ts,tsx}`.
- Prefer assertions on user-visible behavior over component internals.
- Use Vitest setup files for common repeated initialization.

Run from `src/ui/`:

```bash
npm run format
npm run lint
npm run build
npm run test
```

## Output

Report:

- task ID when applicable;
- UI areas changed;
- routing/auth/API/Markdown impact;
- canonical docs updated or why none were needed;
- validation commands and results;
- blockers or follow-ups.
