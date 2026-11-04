# Proposal

## Why

gobento's main promise is that users choose their stack, but v0.1.0 has only
one HTTP framework (`net/http`). gin and chi are the two most used Go routers,
and adding them is the smallest real test that the slot model scales: they
must drop into the existing `handlers` slot without changes to the
architecture, database, or data-access components.

## What Changes

- New framework component `framework/gin` (github.com/gin-gonic/gin) filling
  the `handlers` slot; `NewRouter` returns the `*gin.Engine` as an
  `http.Handler`.
- New framework component `framework/chi` (github.com/go-chi/chi/v5) filling
  the `handlers` slot with a `chi.Router`.
- Both serve exactly the API `nethttp` serves: the same routes, status codes,
  JSON error bodies, 1 MiB body limit, and rejection of unknown fields and
  trailing data. Every framework component ships the same,
  framework-independent API test file, and a gobento test keeps those files
  identical, so the API contract cannot drift between frameworks.
- The shared API test accepts any `application/json` media type (gin adds
  `; charset=utf-8`).
- `components/versions.yaml` gains ranges for gin (`>=v1.11.0 <v2.0.0`,
  default v1.12.0) and chi (`>=v5.3.0 <v6.0.0`, default v5.5.0); the chi floor
  excludes every release with a known Go vulnerability advisory.
- Golden tests cover the files every component contributes, not only one
  reference project, so each new component's output is reviewed.
- **BREAKING (CLI, pre-1.0):** framework now has several supported values, so
  the existing *Automatic selection of the only option* rule makes
  `--framework` required: `gobento new shop` without it exits non-zero listing
  `chi, gin, nethttp`. Interactive prompts, developed in parallel, ask instead
  when a terminal is attached.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `component-composition`: adds the requirement that every framework component
  serves the same API contract, enforced by one shared test.

## Impact

- New: `components/framework/gin/`, `components/framework/chi/`.
- Changed: `components/versions.yaml`,
  `components/framework/nethttp/files/handlers/router_test.go.tmpl` (media-type
  check), `internal/catalog/shipped_test.go`, `internal/compose/golden_test.go`
  and `testdata/golden/`, CLI and prepare tests that expected one framework,
  `.github/workflows/ci.yml` (the sqlc-fresh job passes `--framework nethttp`).
- CI: the integration test grows from 1 to 3 selections (defaults and floors
  each) on Linux, macOS, and Windows.
- Website data: `website/src/generated/catalog.json` and
  `website/lib/command/supported-commands.txt` are regenerated; the
  configurator reads the new options from the catalog.
- Documentation updated in the same change: `README.md` (usage, roadmap),
  `website/content/docs/index.mdx`, `getting-started/quick-start.mdx`,
  `cli/new.mdx` (examples, error output), `concepts/components.mdx`,
  `contributing/adding-a-component.mdx` (the shared API test), and the new
  components' `notes.md.tmpl` for the generated README.
