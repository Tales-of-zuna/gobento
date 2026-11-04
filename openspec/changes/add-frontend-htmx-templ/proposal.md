# Proposal

## Why

gobento generates API-only services today, while the roadmap promises
frontends ("htmx + templ", then React, Vue, and Svelte via Vite, and Tailwind
CSS). A frontend is the first component kind that a project may also leave
out, and the first that adds pages next to an API whose routes and test are
shared by every framework. Its design decides how every later frontend plugs
in, so it comes first, with htmx + templ as the first option because it needs
no Node toolchain and gobento must own its templates (no official scaffolder
exists).

## What Changes

- A fifth component kind, `frontend`, which is **optional**: a selection has
  at most one frontend, and none is a valid choice. API-only projects stay
  the default. `gobento new` gains `--frontend <name>`; omitting it, or
  `--frontend none`, selects no frontend, so the flag is never required even
  though the kind has several options. Terminal questions offer
  "None (API only)" first.
- The architectures `mvc` and `vertical-slice` declare an optional slot
  `web` (`internal/views` and `internal/web`) filled by the frontend. When it
  is filled, the entry point serves a root mux: the unchanged API router
  under `/api/` and the frontend's handler at every other path. Without a
  frontend the API stays at `/`, byte-for-byte as today.
- Templates can see whether a slot is filled (`.Slots.<name>.Filled`) and the
  API's base path (`.APIPrefix`: `""` or `"/api"`).
- A frontend component `frontend/htmx` ("htmx + templ"): server-rendered
  pages written as templ components, an htmx form that adds an item without
  reloading the page (and still works without JavaScript), and htmx
  vendored as a static file embedded in the binary, so the app works
  offline. The templ output is shipped pre-generated, like sqlc's, so
  projects build without the templ CLI; CI checks that it is current.
- The rules file gains `assets`: files gobento vendors, with a fixed version,
  their upstream source, and a checksum that tests compare with the shipped
  file and with the upstream file.
- **BREAKING (CI coverage rule)**: CI keeps generating every supported
  selection without a frontend, and covers selections with a frontend
  pairwise: every pair of components that can appear together with a
  frontend is built together at least once. The full product would double
  the integration matrix with each frontend.
- The website's catalog, configurator, home page, and component reference
  learn the optional kind ("None (API only)") and vendored assets; React,
  Vue, Svelte, and Tailwind CSS stay on the roadmap as planned.

## Capabilities

### New Capabilities

- `frontends`: the optional frontend kind, the `web` slot contract, how the
  API and the frontend share one server, and the htmx + templ frontend.

### Modified Capabilities

- `component-composition`: five kinds with one optional; supported
  combinations include "no component" for optional kinds; slot fill state
  and the API base path are visible to templates.
- `project-generation`: the `--frontend` flag and its `none` value, optional
  kinds never require a flag, the frontend question, the example
  functionality under `/api` when a frontend is selected, and the CI coverage
  rule for supported selections.
- `version-compatibility`: vendored assets as rules-file data, fixed per
  release and checked against the shipped and upstream files.
- `architecture-layouts`: both architectures declare the `web` slot and
  document it in the generated README.
- `website-configurator`: optional kinds offer "None (API only)", and the
  command omits the flag for it.
- `website-docs`: the component reference lists vendored assets with their
  versions.

## Impact

- Code: `internal/catalog` (kind, selection, candidates, `none` name),
  `internal/compat` (assets), `internal/compose` (optional kind, `Filled`,
  `APIPrefix`, assets, pairwise coverage), `internal/generate` (selection,
  integration test), `internal/cli` (flag, questions, equivalent command),
  `internal/tools/catalogexport`, `components/` (both architectures, framework
  notes, new `frontend/htmx`, `versions.yaml`), `.github/workflows/ci.yml`
  (`templ-fresh` job).
- Dependencies of generated projects: `github.com/a-h/templ` (library, exact
  version) and the vendored htmx 2.0.11 file (0BSD license).
- Website: catalog schema and fixture, configurator, URL state, command
  builder, home page kinds and stack, roadmap lists, component reference.
- Documentation updated in this change: `README.md`, `CONTRIBUTING.md`,
  `website/content/docs/index.mdx`, `getting-started/quick-start.mdx`
  (a "With htmx" section), `getting-started/installation.mdx`, `cli/new.mdx`,
  `concepts/components.mdx`, `concepts/slots-and-contracts.mdx` (frontend
  contract and serving), `concepts/version-rules.mdx` (assets),
  `contributing/adding-a-component.mdx` (adding a frontend, what an
  architecture must do to support frontends, regenerating templ output), the
  generated README templates and component notes, and the header comment of
  `components/versions.yaml`.
