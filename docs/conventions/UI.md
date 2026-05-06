# UI.md

Describes conventions and best-practices when working on the UI module.

## In general

Frontend code lives under `src/ui/` and targets Preact with Tailwind CSS.

- Tailwind CSS v4 is configured through the Vite plugin and `src/ui/src/app.css`.
- Put shared component utility classes in `@layer components` in `app.css`.
- Reuse existing classes such as `btn-primary`, `btn-secondary`, `btn-outline`, and `input-field` instead of repeating large class strings.
- Apply Poor Man's DI following the Composition Root pattern in frontend. Use a `deps.ts` file to collect all dependencies and initialize them once in `main.tsx` using a `makeDeps(): Deps` function.
- Use factory functions for transient dependencies in `deps.ts`
- `preact-iso` offers a simple router for Preact with conventional and hooks-based APIs. We use that for routing

## Configuration

- The UI reads `VITE_API_BASE_PATH` at build time to select the same-origin public API base path.
- `VITE_API_BASE_PATH` defaults to `/api`.
- Normalize the API base path before use so requests do not contain double slashes.
- Public deployment reverse-proxies `/api/*` to the Gateway's unprefixed route contracts by stripping `/api`.
- Do not add `/api` prefixes to Gateway route definitions from UI code.

## Authentication

- The UI checks `GET ${apiBasePath}/auth/session` during startup or protected-route entry and shows `/login` on `401`.
- The UI submits credentials only to `POST ${apiBasePath}/login` and does not store the password in local storage, session storage, IndexedDB, or URL state.
- Invalid login navigates to `/login/failed`, which renders a blank black page with no visible information.
- Authenticated requests must use same-origin `fetch` with credentials included.
- The logout control calls `POST ${apiBasePath}/logout` and returns the UI to `/login` on success or `401`.

## Articles

- The UI article API client calls `GET ${apiBasePath}/articles`, `GET ${apiBasePath}/articles/{id}`, and `DELETE ${apiBasePath}/articles/{id}`.
- Browser page routes are `/articles` and `/articles/<article_id>`.
- Markdown article content is untrusted. Render with raw HTML disabled or sanitized; do not execute raw HTML, scripts, inline event handlers, or `javascript:` links.
- Retry/requeue controls are out of scope until a backend retry contract exists.

## Testing

- Frontend tests are colocated as `src/ui/src/**/*.test.{ts,tsx}`.
- Use Vitest, jsdom, React Testing Library, and `@testing-library/user-event`.
- Prefer assertions on user-visible behavior over component internals.
- Leverage Vitest `setupFiles` for common, repeated initialization during tests.
- Run frontend verification from `src/ui/`:

```bash
npm run format
npm run lint
npm run build
npm run test
```
