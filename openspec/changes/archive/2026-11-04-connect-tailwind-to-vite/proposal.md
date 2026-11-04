# Proposal

## Why

Tailwind CSS and the Vite frontends were built in parallel, and each brought
its own hook for a CSS framework to reach a Vite build. Tailwind declared
the points `vite-config-imports`, `vite-plugins`, `css-imports`, and
`npm-dev-dependencies` in `components/frontend/kind.yaml` and tested them
against a fixture frontend; the Vite frontends declared `vite-imports`,
`vite-plugins`, and `vite-styles` in `base.yaml` and an `npm:` section whose
`dev-dependency` packages the scaffold step installs. After the merge, the
two do not meet: React, Vue, and Svelte include points Tailwind does not
contribute to, so no Vite frontend can be combined with `--css tailwind`,
and npm package versions are written in two places (a contribution template
and the rules file).

## What Changes

- One set of Vite points, declared once, by the kind whose templates
  include them: `components/frontend/kind.yaml` declares `stylesheet` (go)
  and `vite-config-imports`, `vite-plugins`, and `css-imports` (vite). The
  `vite-imports`, `vite-plugins`, and `vite-styles` points leave `base.yaml`.
- The point `npm-dev-dependencies` is removed. npm package versions are
  rules-file data only: `tailwindcss` and `@tailwindcss/vite` join the
  `npm:` section with the role `dev-dependency` for `css/tailwind`, and the
  scaffold step installs them into Vite frontends. A test keeps their
  version equal to the standalone CLI's in `binaries:`.
- A component's templates may include the points of the base layer and of
  its own kind (the architectures' `.gitignore` includes the base point
  `gitignore`; the Vite frontends include their kind's points); naming a
  point of another kind fails composition.
- React, Vue, and Svelte with `--css tailwind` become supported: their Vite
  configuration imports and calls `@tailwindcss/vite`, and their stylesheet
  starts with `@import "tailwindcss";`. The fixture Vite frontend in
  `internal/compose/testdata/vite-frontend` is removed; the hook is tested
  against the shipped frontends, and the lost-contribution case stays a
  unit test on a modified shipped frontend.
- Docs, the Tailwind component's description and notes, the catalog
  document, and the website describe Tailwind with the Vite frontends.

## Impact

- `components/base/base.yaml`, `components/frontend/kind.yaml`, the Vite
  frontends' `vite.config.ts` and stylesheet templates,
  `components/css/tailwind` (manifest, contributions, notes),
  `components/versions.yaml`.
- `internal/compose` (who may include which point), its tests and goldens
  (a new React + Tailwind golden), `internal/catalog` shipped tests,
  `internal/generate` scaffold tests.
- The supported selections grow by 108 (every architecture, framework,
  database, and data access with React, Vue, or Svelte and Tailwind); the
  integration test covers them pairwise with the existing rule.
- Website catalog, supported commands, and docs.
