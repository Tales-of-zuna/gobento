# Proposal

## Why

Tailwind CSS is the most requested way to style the pages of a Go web
project, and it is the next item on gobento's roadmap. The htmx frontend
ships hand-written CSS only. Tailwind also needs Node.js unless its
standalone CLI is used, and the CLI is a binary for each platform, which the
rules file cannot describe today. The Vite frontends being added in
parallel need the same option, so the way a CSS framework reaches a
frontend has to work for server-rendered pages and for Vite builds alike.

## What Changes

- A new optional component kind `css` (flag `--css`, values `none` and
  `tailwind`), after `frontend` in kind order. A CSS framework requires the
  trait `frontend-build`, which every frontend provides (`go` for htmx,
  `vite` for the Vite frontends). Without a frontend, `none` is its only
  supported value, so the CLI and the website never ask for it.
- Component kinds, not only the base layer, may declare contribution points
  (`components/<kind>/kind.yaml`) for their components' templates. The
  frontend kind declares `stylesheet` for frontends built by Go and four
  points for Vite builds. A contribution that no template includes is an
  error. A template that renders to nothing but whitespace produces no
  file, so a component can leave out files based on traits.
- `frontend/htmx` provides `frontend-build: go`. Its markup keeps the same
  semantic classes for every CSS choice; its stylesheet is the `stylesheet`
  contribution, or its plain CSS when nothing contributes.
- `css/tailwind` for `frontend-build: go` ships `tailwind.css`, the
  Tailwind source for the htmx markup, and the pre-built, minified
  stylesheet it compiles to. It also ships a small Go program,
  `internal/tools/tailwindcss`, that downloads the pinned standalone
  Tailwind CLI for the current platform, checks its SHA-256, caches it, and
  runs it, so projects build with plain `go build` and need neither Node.js
  nor the CLI until the styles change. Its contributions add a `make
  generate` line and a CI step that fails when the committed stylesheet is
  stale.
- For `frontend-build: vite`, `css/tailwind` contributes to the frontend
  kind's Vite points: `vite-config-imports`, `vite-plugins`, `css-imports`,
  and `npm-dev-dependencies` (`@tailwindcss/vite` and `tailwindcss` of the
  same release). This change implements and tests the hook with a fixture
  Vite frontend; the Vite components include the points when they are
  merged.
- The rules file gets a `binaries` section: an exact version and, for each
  supported platform (`GOOS/GOARCH`), an https download URL and a SHA-256.
  It is fixed per gobento release, cannot be pinned, and is checked when the
  catalog loads and, with network access, against the release's checksum
  file. Tailwind CSS v4.3.4 is the first entry.
- CLI: `--css` for `gobento new` and `gobento list selections`, the
  question "CSS framework" asked after the frontend (with "None (plain
  CSS)" first), the equivalent command, and binaries in `gobento list
  versions`.
- Website: the catalog export, schema, and fixture gain the `css` kind and
  binaries; the configurator gets a sixth tray compartment and offers the
  CSS framework only with a frontend that supports one; the home page's
  kinds and stack show Tailwind CSS; Tailwind CSS leaves the roadmap lists.
- gobento's CI checks the shipped Tailwind output in a reference project,
  and the integration test covers `css` pairwise like the frontend.

## Capabilities

### New Capabilities

- `css-frameworks`: the optional `css` kind, how a CSS framework reaches a
  frontend through the `frontend-build` trait and the frontend's
  contribution points, and Tailwind CSS for htmx (standalone CLI, pre-built
  output, tooling) and for Vite frontends (the hook).

### Modified Capabilities

- `component-composition`: six kinds with two optional; kinds may declare
  contribution points; lost contributions fail; templates that render empty
  produce no file.
- `frontends`: frontends provide `frontend-build`; the htmx stylesheet can
  come from a CSS framework through the frontend kind's `stylesheet` point.
- `project-generation`: the `--css` flag, the CSS question, and CI coverage
  of every optional kind.
- `version-compatibility`: platform binaries as rules-file data.
- `catalog-discovery`: the `css` kind, `--css` for selections, and binaries
  in the version listing.
- `website-configurator`: the CSS framework is offered only with a frontend
  that supports one.
- `website-docs`: the component reference lists a component's binaries.

## Impact

- Code: `internal/catalog` (kind, selection, kind contribution points),
  `internal/compat` (binaries), `internal/compose` (points, empty files,
  binaries data, coverage), `internal/generate` (integration test),
  `internal/cli` (flags, questions, listing), `internal/catalogjson`,
  `components/` (new `css/tailwind`, `frontend/htmx`, `versions.yaml`),
  `.github/workflows/ci.yml` (Tailwind freshness).
- Generated projects with `--css tailwind`: the standalone Tailwind CLI
  v4.3.4 (MIT), downloaded on demand by `make generate`, never needed to
  build; no Go dependency is added.
- Website: catalog schema and frozen fixture, configurator and packing
  stage (six compartments), command builder and URL state, home page kinds
  and stack, roadmap lists, component reference.
- Documentation updated in this change: `README.md`, `CONTRIBUTING.md`,
  `website/content/docs/index.mdx`, `getting-started/quick-start.mdx`,
  `cli/new.mdx`, `cli/list.mdx`, `concepts/components.mdx`,
  `concepts/slots-and-contracts.mdx`, `concepts/project-tooling.mdx`,
  `concepts/version-rules.mdx`, `contributing/adding-a-component.mdx`
  (component contribution points, empty templates, the Vite hook, updating
  the Tailwind output), the htmx and Tailwind notes of generated READMEs,
  and the header comment of `components/versions.yaml`.
