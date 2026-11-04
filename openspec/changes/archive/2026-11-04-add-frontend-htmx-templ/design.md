# Design

## Context

See proposal.md for the motivation. The relevant current state:

- `catalog.Kinds()` lists four kinds, and `catalog.Selection` has one field
  per kind. `Candidates` is the product of all kinds; `compose.Supported`
  keeps the candidates that resolve and compose. The CLI (`--arch`,
  `--framework`, `--db`, `--data`), the questions, the website export, and
  the website schema all iterate the kinds, and every kind is required: a
  kind with several options needs its flag without a terminal.
- An architecture's entry point (`cmd/<name>/main.go`) passes
  `handlers.NewRouter(store.New(db))` to `http.Server`. The three framework
  components serve `GET/HEAD /health`, `GET/HEAD/POST /items`, and ship a
  byte-identical API test against `NewRouter`.
- Templates see `.Project`, `.Slots` (every declared slot's `Dir`, `Package`,
  `Import`), `.Traits`, and `.Versions`, with `missingkey=error`. A slot may
  be optional (`required: false`); nothing tells a template whether it is
  filled.
- sqlc's generated code is shipped as templates (`sqlcgen/`), and the
  `sqlc-fresh` CI job regenerates it in a reference project and fails on a
  diff.
- CI's integration job generates every supported selection twice (defaults
  and floors) on three operating systems: 24 selections today, and a third
  architecture is being added in parallel.

## Goals / Non-Goals

**Goals:**

- An optional `frontend` kind whose absence yields today's API at today's
  paths, so API-only remains the default and existing commands keep
  working.
- One serving design that works for server-rendered frontends now and for
  Vite-built single-page apps later, without editing framework components or
  writing per-combination templates.
- A small, generic architecture-side change that a third architecture can
  copy, documented as a checklist.

**Non-Goals:**

- React, Vue, Svelte (Vite) and Tailwind CSS; their path is described in D10.
- More than one frontend per project, frontend-specific API changes, or
  authentication and sessions.
- Changing the API router, its routes, or its shared test.

## Decisions

### D1. An optional kind, where "no component" is the nil selection

`catalog.KindFrontend` joins `Kinds()` after `data`, and `Kind.Optional()`
reports it. `Selection.Frontend` is nil for "none". `Candidates` iterates
nil first and then each frontend, so every required-kind combination is a
candidate both without and with each frontend. Composition accepts a nil
optional kind and still requires the others. `none` is reserved: the loader
rejects a component of an optional kind named `none`, and the CLI maps
`--frontend none` (and an omitted flag) to nil. `Selection.String()` leaves
out a nil optional kind, so API-only selections keep their names
(`mvc/nethttp/sqlite/sqlc`) and a frontend appends one
(`mvc/nethttp/sqlite/sqlc/htmx`).

The rule "a kind with several options needs its flag" applies to required
kinds only. For an optional kind, the omitted flag already has a meaning
(none), so it never fails for being omitted; in a terminal the question is
still asked when there is more than one option, with "None (API only)" first
and preselected. The equivalent command omits `--frontend` for none unless
the flag was given.

Alternatives considered:

- *A real `frontend/none` component with no files.* It would appear in the
  catalog, the component reference, and the website as a component, needs a
  manifest, and makes "none" a name every later optional kind must copy.
- *A boolean `--web` flag plus a separate frontend flag.* Two flags for one
  choice, with invalid combinations to report.
- *Making `--frontend` required like the other kinds.* Every existing
  command, script, and doc example would break, and API-only projects would
  need a flag to say "nothing".

### D2. Routing: the API moves under /api only when a frontend is selected

When the `web` slot is filled, the entry point serves a root
`http.ServeMux`:

```go
mux.Handle("/api/", http.StripPrefix("/api", handlers.NewRouter(items)))
mux.Handle("/", web.NewHandler(items))
```

The API router is unchanged and still tested by the framework's shared test
at its own root; below `/api` it answers exactly as an API-only project.
The frontend owns every other path, including `/` and `/static/`. Without a
frontend, `newHandler` returns `NewRouter(items)` directly, so API-only
projects keep `/health` and `/items` at the root.

Alternatives considered:

- *Frontend at `/` and the API paths unchanged.* `/items` would be both a
  page and an API route. Content negotiation (`Accept: text/html`) makes the
  routing depend on headers, needs the frontend inside each framework's
  router, and breaks the framework test's 404/405 guarantees.
- *Frontend under `/ui/`.* No conflict, but the application's pages would
  never be at `/`, and SPAs dislike being served below a prefix.
- *Frontend routes registered into the framework's router.* One adapter per
  framework and frontend pair: per-combination code.
- *The frontend builds the root mux itself.* Each frontend would re-implement
  the prefix policy; the architecture is the component that already wires
  everything.
- *Always serving the API under `/api`.* Changes every existing project and
  documented command without a reason when there is no frontend.

### D3. What templates see: `.Slots.<name>.Filled` and `.APIPrefix`

`SlotData.Filled` is true when a selected component has templates for the
slot. `Data.APIPrefix` is `"/api"` when a frontend is selected and `""`
otherwise. The architecture's entry point tests `.Slots.web.Filled` to import
and mount the web package; the framework notes and the README write paths as
`{{.APIPrefix}}/items`. Both values describe slots and the selection's shape,
not component names, so the trait rule ("never depend on which component is
selected") still holds. Only architectures that declare a `web` slot refer
to `.Slots.web`; framework components use only `.APIPrefix`, which every
architecture gets.

Alternatives considered:

- *A map of selected component names (`.Selected.frontend`).* Invites
  templates to branch on names, which the trait rule forbids.
- *Leaving undeclared or unfilled slots out of `.Slots`.* `missingkey=error`
  would turn every reference to an unfilled optional slot into an error, and
  declared slots would behave differently from filled ones.
- *Conditional files (an empty render means "no file").* Not needed: the
  entry test (D7) is useful in both shapes, and implicit file removal would
  be a new, surprising composition rule.

### D4. The web slot contract and the htmx component

The architectures declare `web` as optional, filled by `frontend`:
`internal/views` in `mvc` (the "V" of MVC) and `internal/web` in
`vertical-slice` (pages, layout, and static files are shared by all features,
like `internal/platform`). The contract is
`func NewHandler(items models.ItemStore) http.Handler`, the same shape as
`NewRouter`. It uses only the standard library's `net/http` (method
patterns), so one frontend works with every framework.

`frontend/htmx` serves:

- `GET /{$}`: the page listing items with an add form (`hx-post="/items"`,
  `hx-target="#items"`, `hx-swap="beforeend"`).
- `POST /items`: form-encoded `name`. With `HX-Request: true` it returns the
  new `<li>` plus a fresh form marked `hx-swap-oob`, which clears the input
  without inline scripts. Without htmx it answers `303 See Other` to `/`
  (Post/Redirect/Get). A blank name is `422`: for htmx with `HX-Retarget:
  #new-item` and `HX-Reswap: outerHTML` and the form with the message (the
  page's `htmx-config` meta lets htmx swap 422 responses); otherwise the full
  page with the message. Bodies over 1 MiB are `413`.
- `GET /static/...`: `htmx.min.js` and `styles.css` from `//go:embed static`.

The handler is wrapped in `http.CrossOriginProtection` (Go 1.25+, inside the
go range), so cross-site form posts are rejected without CSRF tokens.
Rendering goes to a buffer first, so a template error becomes a clean `500`.

### D5. templ: shipped generated code, exact library version, `go generate`

The component ships `pages.templ` and `pages_templ.go`. `github.com/a-h/templ`
is a **library** module (the generated code imports its runtime) with an
exact range, `=v0.3.1100`: the shipped code is what that version generates,
and its header names the version. `handler.go` carries
`//go:generate go run github.com/a-h/templ/cmd/templ@<version> generate`, with
the version from the rules file, so `go generate ./...` regenerates without
installing anything; templ then records file names relative to the package
directory, which the shipped code matches. Templates substitute only names
on lines without templ expressions, so the line and column numbers in the
generated code stay correct. The `templ-fresh` CI job generates a reference
project, runs `go generate ./...`, and fails on any diff, as `sqlc-fresh`
does.

The templ runtime package imports only the standard library, so the
advisories of `golang.org/x/net` v0.59.0 (required by templ's CLI in its
`go.mod`) do not reach generated binaries; `govulncheck` in a generated
project confirms it.

Alternatives considered:

- *Running templ during generation.* Downloads and builds the templ CLI on
  every `gobento new` (seconds and network), and a failure there is a new
  way generation can break.
- *`go get -tool github.com/a-h/templ/cmd/templ`.* Records the CLI's whole
  dependency graph in the generated `go.mod`, including the flagged
  `golang.org/x/net` version.
- *A range of templ versions.* Pinning another runtime than the generator's
  is untested and templ warns about the mismatch.

### D6. htmx vendored and recorded as an asset

htmx 2.0.11 (`dist/htmx.min.js` of the npm package `htmx.org`, 0BSD license)
is shipped verbatim as `files/web/static/htmx.min.js.tmpl`; it contains no
`{{`, so rendering leaves it unchanged, and a test proves the rendered file's
checksum. 2.0.11 is npm's `latest`; htmx 4 is tagged `next` and changes
defaults, so moving to it is a separate change.

The rules file gains a top-level `assets` map, name to
`{components, version, source, path, sha256}`: `source` is the upstream
archive URL and `path` the file inside it. `compat.ParseRules` validates it,
the catalog checks the components exist, `Resolve` returns the assets of the
selected components, and templates see `.Assets` (name to version), used by
the notes. A pin of an asset fails as not pinnable. `TestShippedAssets`
(unit) compares the recorded checksum with the file the component renders;
`TestAssetSources` (network, skipped with `-short`) downloads `source` and
compares `path`. The website export lists assets so the component reference
shows them.

Alternatives considered:

- *A CDN `<script>`.* Needs network at run time, leaks visitors to a third
  party, and fails offline.
- *Downloading htmx during generation.* Adds the npm registry as a
  generation dependency; vendoring keeps generation offline-capable beyond
  the Go proxy and makes the project reproducible.
- *Recording htmx as a module.* It is not a Go module; module rules mean
  `go get`, ranges, and pins, none of which apply.

### D7. The entry point ships a routing test

Both architectures add `cmd/<name>/main_test.go`: it calls `newHandler` with
an in-memory `ItemStore` and checks `{{.APIPrefix}}/health` and
`{{.APIPrefix}}/items` (and `/` returning HTML when the web slot is filled).
This is what proves the mount in every generated project; the framework test
keeps proving the API itself.

### D8. CI coverage: exhaustive without a frontend, pairwise with one

`compose.Coverage(supported)` returns every supported selection without an
optional component, plus a greedy strength-2 covering array of the others:
starting from the set of component pairs that occur together in some
selection with an optional component, it repeatedly takes the first selection
(in `Supported` order) that covers the most uncovered pairs, until none is
left. With two architectures, three frameworks, two databases, two data
layers, and one frontend, that is 24 + 7 projects instead of 48 (the
architecture × framework pairs alone need 6), and the
frontend part grows with the largest pair product (architecture × framework)
rather than with the product of all kinds.

Why pairs are enough here: a frontend touches the rest of the project only
through the architecture (slot directory, package names, entry point, the
`models` contract) and the framework (mounted under `/api`), and through
`go.mod` (module graph interactions are pairwise between the modules
involved). Database and data access never meet the frontend except through
`ItemStore`, which the exhaustive backend matrix already proves for every
combination. The integration test calls `Coverage` and states the counts;
`TestCoverage` proves the pair property for the shipped catalog.

Alternatives considered:

- *The full product.* 48 projects now, 72 with a third architecture, and a
  multiple of that for every later frontend, on three operating systems.
- *Random sampling.* Not reproducible, and a failing combination could
  disappear from the next run.
- *Only one frontend selection per architecture.* Leaves framework mounting
  (gin's and chi's handling of the stripped path) unproven.

### D9. Website

The export gains `optional` on kinds and `assets`; supported selections omit
a kind that has no component. The schema accepts both, the configurator
offers "None (API only)" for optional kinds (preselected, encoded as absent
in the URL), and the command builder omits the flag for none. The home
page's kinds list and stack show the frontend; React, Vue, Svelte, and
Tailwind CSS stay on the roadmap lists.

### D10. The path for Vite frontends and Tailwind CSS (not implemented)

- **Vite frontends** (`frontend/react`, `frontend/vue`, `frontend/svelte`)
  fill the same `web` slot with the same contract. Their `NewHandler` serves
  an embedded `dist/` (`//go:embed all:dist`) with an `index.html` fallback
  for client-side routes, and ignores `items`: the SPA calls `/api/...`
  (`{{.APIPrefix}}`). Their sources live in a root directory such as
  `frontend/`, scaffolded by `npm create vite@<version>` with the template
  named by the component, with `build.outDir` pointing into the web slot and
  a dev-server proxy for `/api` to `localhost:8080`. This needs a new
  generation step (a scaffolder hook after composition, needing Node) and
  rules-file data for npm tools (the `create-vite` version, recorded like an
  asset), proposed in its own change. Until `npm run build` runs, the
  embedded `dist/` holds a placeholder page, so `go build` keeps working
  without Node.
- **Tailwind CSS** becomes a trait-aware option of frontends (for example a
  `css` trait or a second optional kind `styling`), built with the
  standalone Tailwind CLI for htmx and the Vite plugin for SPAs, replacing
  `styles.css`.

## Risks / Trade-offs

- [Pairwise coverage misses a three-way interaction] → The backend matrix
  stays exhaustive; frontend code depends on the backend only through the
  compiler-checked contract; the covering set is deterministic, so any
  failure reproduces.
- [templ output drifts from the shipped code after a templ bump] →
  `templ-fresh` fails, and the contributor guide gives the regeneration
  steps.
- [A vendored asset changes silently] → Checksum tests against the shipped
  file and the upstream archive.
- [Moving the API under `/api` surprises users who add a frontend later] →
  The generated README and the docs show the paths for both shapes, and the
  entry test fails if the mount is changed by accident.
- [`.Slots.web` is referenced by an architecture that does not declare it] →
  `missingkey=error` makes that a composition error caught by the golden
  and supported-selection tests.

## Migration Plan

Additive for users: commands without `--frontend` generate the same API at
the same paths as before; the only difference in API-only projects is the
entry point's `newHandler` function and its routing test (visible in the
golden diff). The parallel DDD architecture
change can merge before or after this one; to support frontends it follows
the checklist in the contributor guide (declare `web`, use `newHandler`, add
the entry test, list the web directory in its README), otherwise frontend
selections with it are simply unsupported.
