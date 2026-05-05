# UI.md

Describes conventions and best-practices when working on the UI module.

## In general

Frontend code lives under `src/ui/` and targets Preact with Tailwind CSS.

- Tailwind CSS v4 is configured through the Vite plugin and `ui/app/app.css`.
- Put shared component utility classes in `@layer components` in `app.css`.
- Reuse existing classes such as `btn-primary`, `btn-secondary`, `btn-outline`, and `input-field` instead of repeating large class strings.
- Apply Poor Man's DI following the Composition Root pattern in frontend. Use a `deps.ts` file to collect all dependencies and initialize them once in `main.tsx` using a `makeDeps(): Deps` function.
- Use factory functions for transient dependencies in `deps.ts`
- `preact-iso` offers a simple router for Preact with conventional and hooks-based APIs. We use that for routing

## Authentication

- The UI checks `GET /auth/session` during startup and shows the password-only login form on `401`.
- The UI submits credentials only to `POST /login` and does not store the password in local storage, session storage, IndexedDB, or URL state.
- Authenticated requests must use same-origin `fetch` with credentials included.
- The logout control calls `POST /logout` and returns the UI to the login state on success or `401`.

## Testing

- Frontend tests are colocated as `ui/src/**/*.test.{ts,tsx}`.
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
