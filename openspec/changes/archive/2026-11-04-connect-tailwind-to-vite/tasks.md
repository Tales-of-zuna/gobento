# Tasks

## 1. One set of points

- [x] 1.1 Declare `vite-config-imports`, `vite-plugins`, and `css-imports` once, in `components/frontend/kind.yaml`; remove `vite-imports`, `vite-plugins`, and `vite-styles` from `base.yaml` and the point `npm-dev-dependencies`; rename the points in the React, Vue, and Svelte templates and render the stylesheet's import without a leading blank line
- [x] 1.2 In `internal/compose`, let a component's templates and notes include (`contrib`, `contribs`) the base layer's points and its own kind's; fail for another kind's point; keep the check for contributions no template includes; verify with `TestComposeBaseLayerErrors`, `TestComponentTemplatesIncludePoints`, and `TestComposeKindPoints`

## 2. npm packages are rules data

- [x] 2.1 List `tailwindcss` and `@tailwindcss/vite` 4.3.4 in the `npm:` section (role `dev-dependency`, `css/tailwind`); add `TestTailwindReleasedTogether`; check the versions on the npm registry (Vite 8 peer range, MIT)
- [x] 2.2 Give binaries and npm packages optional `traits:` conditions like modules (parse, `CheckTraits`, `Resolve`, catalog document, `gobento list versions` ONLY WHEN); limit the CLI to `frontend-build: go` and the npm packages to `vite`; verify with compat and prepare tests
- [x] 2.3 Name the dependencies npm adds in the dry run; verify with `TestNewViteFrontend`
- [x] 2.4 Work around npm 12 refusing registry tarballs with bundled dependencies in lock-only installs (npm/cli#9800): retry with `--allow-remote=all` and require every lock file tarball on the registry's host; verify with `TestGenerateScaffoldAllowRemote` and a real npm 12.2.0

## 3. Tests and goldens

- [x] 3.1 Remove `internal/compose/testdata/vite-frontend`; test the hook against the shipped React, Vue, and Svelte (`TestViteHook`) and the lost contributions on a modified shipped React (`TestViteHookWithoutPoints`)
- [x] 3.2 Add the golden `testdata/golden-react-tailwind`, and the Vite frontends with Tailwind to the expected candidates and supported selections (324)
- [x] 3.3 Update the Tailwind component's description, comment, and notes for Vite

## 4. Website and docs

- [x] 4.1 Run `make catalog`; in the configurator, show Tailwind's files only for go builds, and binaries and npm packages by their trait conditions; add schema traits and a Vite + Tailwind test helper
- [x] 4.2 Update README, CONTRIBUTING, `docs/maintainers.md`, and the website docs (quick start, installation, CLI, concepts, adding a component) for Tailwind with the Vite frontends, the single set of points, trait conditions, and the npm 12 retry

## 5. Verification

- [x] 5.1 Generate React + Tailwind and htmx + Tailwind projects, `make build`, run them, and check in headless Chromium that Tailwind's stylesheet is applied and items can be added
- [x] 5.2 `go test -short ./...`, `go test ./...` with PostgreSQL, `make lint`, `make catalog-check`, actionlint, website checks, and `openspec validate --strict`
