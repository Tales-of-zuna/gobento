# Tasks

## 1. Catalog and composition

- [x] 1.1 Add the optional kind `css` (`KindCSS`, `Selection.CSS`, `Kinds`, `With`, `String`) and update `internal/catalog` tests; verify `go test ./internal/catalog`
- [x] 1.2 Let a kind declare `contributions:` in `components/<kind>/kind.yaml` for its components' templates; check in the loader that every contribution names a point of the base layer or of a kind, and that no kind redeclares a point of the base layer or of another kind; verify with new invalid-catalog cases
- [x] 1.3 In `internal/compose`, give each component's templates a `contrib` for its kind's points, render contributions before checking their point against the selection, fail for a contribution no rendered template includes, and write no file for a template that renders to whitespace; verify with compose tests for a kind point, a point of an unselected kind, a lost contribution, and an empty template
- [x] 1.4 Update `adding-a-component.mdx`, `concepts/components.mdx`, `concepts/slots-and-contracts.mdx`, and `concepts/traits.mdx` for the `css` kind, kind points, empty templates, and `frontend-build`; verify the documented manifest keys against the loader

## 2. Rules file: platform binaries

- [x] 2.1 Add the `binaries` section to `internal/compat` (parse and validate, `CheckComponents`, resolve for the selected components, `ErrNotPinnable` for pins) and `.Binaries` to the template data; verify with compat tests for each validation error and the pin
- [x] 2.2 Record Tailwind CSS v4.3.4 in `components/versions.yaml` (URLs and SHA-256 from the release, checked against its `sha256sums.txt`; no advisories) and add a network test comparing every checksum with the release's checksum file; verify `go test ./internal/compat` with and without `-short`
- [x] 2.3 Document binaries in `concepts/version-rules.mdx`, the header comment of `versions.yaml`, and the update procedure in `docs/maintainers.md`; verify against the parser's field names

## 3. Components

- [x] 3.1 `frontend/htmx`: provide `frontend-build: go`, declare the frontend kind's points (`stylesheet` and the Vite points) in `components/frontend/kind.yaml`, render `static/styles.css` from it or the plain CSS, move the page title to a Go constant, regenerate `pages_templ.go` with the pinned templ, make the static-file test check the classes the pages use, and describe the stylesheet in the notes according to the point
- [x] 3.2 `css/tailwind`: manifest, `tailwind.css`, the runner `internal/tools/tailwindcss` (download, checksum, cache, `TAILWINDCSS` override) with a test, the shipped minified output as the `stylesheet` contribution, `make-generate` and `ci-steps` contributions, notes, and the Vite contributions, each guarded by `frontend-build`; verify in a generated project that `make generate` changes nothing, `make lint` passes, and the CI step fails after a stale commit
- [x] 3.3 Add a fixture Vite frontend under `internal/compose/testdata` and a compose test for the Vite hook (and for a Vite frontend without the points); regenerate the goldens (`go test ./internal/compose -run TestGolden -update`) with an htmx + Tailwind golden, and review the diff
- [x] 3.4 Document Tailwind in `getting-started/quick-start.mdx`, `concepts/project-tooling.mdx`, and `contributing/adding-a-component.mdx` (updating the Tailwind output, the Vite hook); verify the commands against a generated project

## 4. CLI and listing

- [x] 4.1 Add `--css` to `gobento new` and `gobento list selections`, the "CSS framework" question with "None (plain CSS)", the equivalent command, and binaries in `gobento list versions` (text and JSON); update the CLI tests and snapshots (`internal/cli/testdata/list`)
- [x] 4.2 Add the kind and binaries to `internal/catalogjson`, run `make catalog`, and update `TestShippedCatalogSupported` and the coverage counts
- [x] 4.3 Update `README.md` (usage, components, roadmap), `CONTRIBUTING.md`, `cli/new.mdx`, `cli/list.mdx`, and `docs/index.mdx`; verify the documented flags and output against `gobento new --help` and `gobento list`

## 5. Website

- [x] 5.1 Schema, kinds, and frozen fixture (`lib/test/fixture-catalog.json`) with the `css` kind and binaries; configurator with a sixth tray compartment and the packing stage derived from the number of kinds (keeping the memoized `packed`), "None (plain CSS)", and disabled CSS frameworks without a frontend; command builder and URL state with `--css`
- [x] 5.2 Home page kinds and stack, roadmap lists without Tailwind CSS, and the component reference with binaries
- [x] 5.3 Verify `pnpm test`, `pnpm lint`, `pnpm typecheck`, `npx prettier --check .`, and `pnpm e2e`

## 6. CI and verification

- [x] 6.1 Make `templ-fresh` in `.github/workflows/ci.yml` a matrix over `--css none` and `--css tailwind`, running the Tailwind command in the Tailwind leg; verify the job's commands locally
- [x] 6.2 Run the Tailwind command in `TestIntegration` for covered Tailwind selections and in a project named `container`; verify with `go test -count=1 -run TestIntegration ./internal/generate/` and a local PostgreSQL
- [x] 6.3 Generate an htmx + Tailwind project, run `make generate`, `go build`, run it and check in a browser that the stylesheet is served and applied, and build its image with Docker
- [x] 6.4 `go test -short ./...`, `go test ./...`, `make lint`, `make catalog-check`, and `openspec validate add-tailwind-css --strict` pass
