# Proposal

## Why

The optional frontend kind (`--frontend htmx`) and the `ddd` architecture
were built on separate branches. After both were merged, `mvc` and
`vertical-slice` support frontends but `ddd` does not: it declares no `web`
slot, so every `ddd` selection with a frontend is unsupported and the
configurator and `gobento list` offer htmx only for the other two layouts.
Every shipped architecture should support every shipped frontend, as the
contributor checklist "Supporting frontends in an architecture" describes.

## What Changes

- `architecture/ddd` declares the optional slot
  `web: { dir: internal/adapters/web, filled_by: frontend }`. The web pages are
  another adapter around the domain, next to `httpapi`, `persistence`, and
  `database`.
- The `ddd` entry point builds its handler in `newHandler`, which takes the
  application service through the `domain.ItemStore` port: with a frontend, an
  `http.ServeMux` serves the API router under `{{.APIPrefix}}/` and the web
  handler at `/`, both given the application service; without one, the API
  router alone, as before.
- `ddd` ships `cmd/<name>/main_test.go`, the routing test of the other
  architectures, run through the application service on an in-memory store.
- The generated `ddd` README lists the web adapter when a frontend is
  selected, tells the reader to open `http://localhost:8080/`, and writes the
  `curl` paths with `{{.APIPrefix}}`.
- The integration test's layer check treats the web package as an adapter: it
  must not import the application package (it reaches the service only
  through the domain port) and, like every web package, neither the database
  nor the store package.
- Tests: a full golden directory `golden-ddd-htmx`, the supported-selection
  list and `TestShippedFrontendMount` for every architecture; `compose.Coverage`
  then includes `ddd` selections with htmx. `make catalog` refreshes the
  website catalog (72 supported selections) and the CLI cross-check commands.
- Docs: the slot table, the `gobento new` dry-run description, and the
  contributor checklist name the `ddd` web directory.

## Capabilities

### Modified Capabilities

- `architecture-layouts`: every shipped architecture, `ddd` included, supports
  frontends; the `ddd` dependency rule covers the web adapter.

## Impact

- `components/architecture/ddd/` (manifest, entry point and its new test,
  README).
- `internal/compose` and `internal/generate` tests and goldens.
- `website/src/generated/catalog.json`, `website/lib/command/supported-commands.txt`,
  and the docs under `website/content/docs/`.
- No change to the CLI, the frontend component, or the framework components.
