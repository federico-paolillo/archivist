# UI.md

Describes conventions and best-practices when working on the UI module.

## In general

Frontend code lives under `src/ui/` and targets Preact with Tailwind CSS.

- Tailwind CSS v4 is configured through the Vite plugin and `ui/app/app.css`.
- Put shared component utility classes in `@layer components` in `app.css`.
- Reuse existing classes such as `btn-primary`, `btn-secondary`, `btn-outline`, and `input-field` instead of repeating large class strings.
- Apply Poor Man's DI following the Composition Root pattern in frontend. Use a `deps.ts` file to collect all dependencies and initialize them once in `main.tsx`
- Use factory functions for transient dependencies in `deps.ts`

## Testing

- Frontend tests are colocated as `ui/src/**/*.test.{ts,tsx}`.
- Use Vitest, jsdom, React Testing Library, and `@testing-library/user-event`.
- Prefer assertions on user-visible behavior over component internals.
- Run frontend verification from `src/ui/`:

```bash
npm run format
npm run lint
npm run build
npm run test
```
