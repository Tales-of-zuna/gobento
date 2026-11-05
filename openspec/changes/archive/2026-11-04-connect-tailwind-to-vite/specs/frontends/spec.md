## ADDED Requirements

### Requirement: Vite frontends take a CSS framework
The Vite frontends SHALL include the frontend kind's points
`vite-config-imports` and `vite-plugins` in `frontend/vite.config.ts` and
`css-imports` at the top of the stylesheet the app imports, so that every
selection with a Vite frontend is also supported with each CSS framework
that accepts `frontend-build: vite`.

#### Scenario: React with Tailwind
- **WHEN** the user runs `gobento new shop --arch mvc --framework nethttp --db sqlite --data sqlc --frontend react --css tailwind` with Node.js and network access, then `make build`
- **THEN** the build succeeds, `frontend/package.json` lists `tailwindcss` and `@tailwindcss/vite` as development dependencies, and the stylesheet the binary serves is compiled by Tailwind
