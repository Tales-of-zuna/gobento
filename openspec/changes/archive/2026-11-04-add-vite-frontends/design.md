# Design

## Context

See proposal.md for the motivation. The htmx change (archived as
`2026-11-04-add-frontend-htmx-templ`, D10) fixed the frame: the optional
`frontend` kind fills the architecture's optional `web` slot with
`func NewHandler(items models.ItemStore) http.Handler`, the entry point
mounts the API under `/api/` and the web handler at `/`, and templates see
`.APIPrefix` and `.Slots.<x>.Filled`. The base layer (`components/base`)
writes the Makefile, Dockerfile, CI workflow, Dependabot, and
`.dockerignore`; components extend them through contribution points
(`contrib/<point>.tmpl`), which only base templates could include so far.

Generation (`internal/generate`) runs preflight, writes the plan and
`go.mod` into a temporary directory, runs `go get` and `go mod tidy`,
rechecks `go.mod`, and renames the result into place; any failure or
cancellation removes everything. It runs no tool but `go`.

create-vite 9.3.3 is npm's `latest`. Its templates `react-ts`, `vue-ts`, and
`svelte-ts` scaffold Vite 8 projects, and both packages declare
`engines.node: ^20.19.0 || >=22.12.0`. Non-interactive use is
`create-vite <dir> --no-interactive --no-immediate --template <name>`.
Node.js 26 is the Active LTS (26.13.0), 24 is in maintenance, and 20 is end
of life.

## Goals / Non-Goals

**Goals:**

- Scaffold with the official tool at generation time, own only what must be
  gobento's, and keep the three frontends alike.
- Keep generation's guarantees: all input checked before writing, nothing
  left behind, Ctrl-C cleans up.
- `go build` and `go test` keep working without Node.js in every generated
  project.
- Versions as data, including the scaffolder and Node.js.
- Leave a hook for the Tailwind CSS component developed in parallel.

**Non-Goals:**

- Other create-vite templates (JavaScript variants, Preact, Solid, Lit),
  meta-frameworks (Next.js, Nuxt, SvelteKit), server-side rendering, or
  client-side routers.
- Choosing the Node.js version per project (no `--node` flag) or another
  package manager than npm.
- Running `npm install` at generation time.

## Decisions

### D1. Scaffold at generation time, after writing, before placing

A scaffold step runs between writing the plan (and `go.mod`) and the `go`
steps:

1. In a staging directory inside the temporary directory, run
   `npm create --yes vite@<version> <base> -- --no-interactive --no-immediate --template <template>`
   with no standard input, so neither npm (`--yes`) nor create-vite asks.
2. Remove the files the component lists in `scaffold.remove` (the demo
   files the overlays make unused); a missing one is an error, so a
   create-vite bump that renames files is noticed by the tests.
3. Move every staged file into `<dir>` unless gobento wrote that path:
   gobento's overlays win.
4. Run `npm install --package-lock-only --ignore-scripts --no-audit
   --no-fund` (with the npm `dependency` and `dev-dependency` packages of
   the selection, if any), which resolves the tree into
   `package-lock.json` without `node_modules`.

The npm processes run in their own process group on Unix, and the context
cancels the whole group; on Windows `taskkill /T` ends the process tree.
Without this, Ctrl-C would stop npm but not the `node` process it started,
which could keep writing into the directory cleanup is removing.

Alternatives considered:

- *Scaffold first, then write the plan over it.* The same result, but the
  plan's files would have to be written twice or the scaffolder run in a
  non-empty directory, which create-vite refuses without `--overwrite`
  (which deletes everything).
- *`npm install` at generation time.* Hundreds of megabytes, minutes on
  slow networks, and `node_modules` moved by the final rename. The lock
  file pins the tree just as well; `npm ci` installs it.
- *Vendoring the scaffolded files as gobento templates.* Contradicts the
  project principle and goes stale with every create-vite release.
- *`npx create-vite`.* `npm create` is create-vite's documented command;
  both resolve the same package.

### D2. Preflight checks Node.js only when needed

`Request.Scaffolds` lists the scaffold steps; when it is non-empty,
preflight looks up `node` and `npm`, runs `node --version`, and checks it
against the rules' Node.js range (`>=v22.12.0`: create-vite's and Vite's
engines without the end-of-life 20 line). Problems are reported with the go
and target problems, before anything is written, and name the component,
the scaffolder, and the range. Without a scaffolded component nothing
Node-related runs. The node and npm commands are overridable in the request,
so unit tests use the test binary as a fake `node` and `npm`, as they
already do for `go`.

`--dry-run` stays offline and Node-free: it lists gobento's files and adds
one line per scaffold step naming the command, the directory, and the lock
file.

### D3. Serving: `//go:embed all:dist`, build in `dist/app`

The web package (the same `handler.go` for all three frontends) embeds
`all:dist` (`all:` keeps files such as Vite's `_plugin-vue_export-helper`
chunks) and serves `fs.Sub(dist, "dist/app")`:

- an existing file of the build at its path, with
  `Cache-Control: public, max-age=31536000, immutable` under `/assets/`
  (Vite's content-hashed files);
- `index.html` with `Cache-Control: no-cache` for `/` and every path without
  a file extension (client routes);
- `404` for a missing file with an extension and for every path under
  `{{.APIPrefix}}` (the root mux never sends those here, but the handler
  must not answer them with the app if it is ever mounted differently);
- `405` with `Allow: GET, HEAD` for other methods.

Vite writes to `<web>/dist/app` (`build.outDir`, `emptyOutDir: true`
because it is outside the Vite root). The committed placeholder is
`<web>/dist/.gitignore`, which ignores `/app/` and keeps `dist/` in the
repository, so `go:embed` always matches a file. When `dist/app/index.html`
is missing, the handler answers with a small built-in page that names
`make build` and the npm commands. The handler's test uses `fstest.MapFS`
for a built and an unbuilt app and also checks the embedded one, so it
passes before and after `npm run build`.

Alternatives considered:

- *Vite writes to `dist/` directly, with a committed `dist/.gitkeep` or
  stub `index.html`.* `emptyOutDir` deletes the committed file on every
  build; committing that deletion breaks `go build` for the next clone.
- *Serving `frontend/dist` from disk.* The binary would no longer be
  self-contained and the image would need the files next to it.
- *A fallback for every 404, assets included.* A missing script would load
  HTML as JavaScript and hide the real error.

### D4. Overlays: few files, the same for all three

gobento owns under `frontend/`:

| File | React | Vue | Svelte |
|---|---|---|---|
| Vite configuration | `vite.config.ts` | same | same |
| Example page | `src/App.tsx` | `src/App.vue` | `src/App.svelte` |
| API client | `src/api.ts` | same | same |
| Stylesheet the entry imports | `src/index.css` | `src/style.css` | `src/app.css` |

The scaffolded entry (`main.tsx`/`main.ts`) and `index.html` stay as the
scaffolder wrote them; the stylesheet keeps the entry's own file name, so
no entry needs to change. The Vite configurations differ only in the
framework plugin's import and call; `api.ts`, the stylesheet, the web
package, the notes, and every contribution are the same template text in
the three components, and `TestViteFrontendsAlike` enforces it (as
`TestFrameworksShareTheAPITest` does for the API test). Templates avoid
`{{` outside template actions (Vue uses `v-text`), because overlays are
Go templates.

`scaffold.remove` deletes the demo files the new page leaves unused
(`src/App.css` and `src/assets/` for React, `src/components/` and
`src/assets/` for Vue, `src/lib/` and `src/assets/` for Svelte).

### D5. Versions: `node` and `npm` in the rules file

```yaml
node:
  range: ">=v22.12.0"
  default: "26"          # release line of setup-node and the node image
npm:
  create-vite:
    components: [frontend/react, frontend/svelte, frontend/vue]
    role: scaffolder     # or dependency, dev-dependency
    version: 9.3.3
```

npm packages are fixed per release like tools and assets (pins are
rejected). A scaffolder's name starts with `create-`, and `npm create
<name without create->@<version>` runs it. `dependency` and
`dev-dependency` packages are added to every scaffolded frontend of a
selection that has one of their components: the npm-side hook for a CSS
component. Templates see `.Node` (`26`) and `.NPM` (name to version); the
`node` image is `node:{node}-alpine`, `{node}` being the release line, like
`{go}` for `golang`. It is a build stage only, so like `golang` it is not
pinned by digest; Dependabot's docker ecosystem keeps the tag current.
`actions/setup-node` v7.1.0 is pinned by SHA in `tooling.actions`. The
component manifest names the scaffolder and the template
(`scaffold: {with: create-vite, template: react-ts, dir: frontend}`); the
template is part of the component, the version is data.

Alternatives considered:

- *Recording create-vite as an asset.* Assets are files gobento ships and
  checksums; the scaffolder is a program it runs.
- *A Node.js tool entry (`tooling.tools.node`).* Tools are exact semantic
  versions run by the Makefile; Node.js has a supported range for
  generation and a release line for the tooling, like Go.
- *Pinning the template's dependency versions in the rules.* The scaffolder
  chooses them; the lock file records them.

### D6. Tooling contributions

New contribution points of the base layer:

- `make-build`: commands of `make build` before `go build`
  (`$(NPM) --prefix frontend ci` and `run build`). The Vite frontends also
  contribute `NPM ?= npm` and a `dev` target (`npm run dev` in `frontend/`)
  through `make`.
- `docker-stages`: whole stages before the Go build stage. The `frontend`
  stage installs with `npm ci` (with an npm cache mount) and builds; the Go
  stage copies `<web>/dist/app` from it (`docker-build`). The runtime image
  is unchanged.
- `dependabot`: entries of `updates:` (`npm` for `/frontend`).
- `gitignore`: lines of the root `.gitignore`, which architectures now
  include; htmx's `*_templ.txt` moves there, so Vite projects do not get it.

A new template function, `contribs "<point>"`, returns the contributions as
a list, so the README's Docker section separates the paragraphs of several
components (here the database's and the frontend's) with a blank line;
`contrib` joined them into one paragraph.

CI (`ci-steps`): `actions/setup-node` with `node-version: {{.Node}}` and the
npm cache keyed on `frontend/package-lock.json`, then `npm ci`,
`npm run --if-present lint` (React's oxlint), `npm run --if-present check`
(Svelte's svelte-check), and `npm run build` (which type-checks React and
Vue), all in `frontend/`, before the Go steps, so `go build` and `go test`
see the real build. `.dockerignore` gets `frontend/node_modules` and
`<web>/dist/app`.

### D7. Hook for CSS components: `frontend-build` and three points

Every frontend provides `frontend-build`: `go` (htmx) or `vite`. A CSS
component requires `frontend-build: [go, vite]` and reaches a Vite frontend
through points its files include: `vite-imports` (import lines of
`vite.config.ts`), `vite-plugins` (entries of its `plugins` array), and
`vite-styles` (CSS at the top of the stylesheet, such as
`@import "tailwindcss";`), plus `npm` `dev-dependency` packages. To allow
this, component templates (files and notes, not contribution templates) get
the `contrib` function; every point is still declared in `base.yaml`, which
stays the single registry.

Alternatives considered:

- *The CSS component replaces the stylesheet or the Vite configuration.*
  Two components producing one file is a composition error, and a
  replacement would have to know every frontend's file names.
- *Branching on component names in the overlays.* Forbidden by the trait
  rule.

### D8. CI and coverage

`compose.Coverage` already covers frontends pairwise; with four frontends
the integration test grows accordingly. For a selection with
`frontend-build: vite`, the defaults variant runs `npm ci`, the CI
workflow's checks, and `npm run build` before the Go checks, and checks
that `dist/app/index.html` exists; the floors variant keeps the unbuilt
placeholder, so both states of the web package are built and tested. The
templ freshness check runs only with `frontend-build: go`. gobento's own
integration job sets up Node.js (pinned `actions/setup-node`) on all three
operating systems: the default line on Linux and Windows, and the range
floor `22.12.0` on macOS, so both ends of the Node.js range run.

## Risks / Trade-offs

- [The npm registry or create-vite is unavailable] → Only Vite selections
  need it; the error shows npm's output and nothing is left behind.
- [A create-vite release changes the templates' files] → The version is
  fixed per gobento release; `scaffold.remove` fails loudly on a missing
  file, and the integration test builds every frontend before a bump ships.
- [Scaffolded dependencies get advisories after release] → The lock file
  is per project and Dependabot updates it; `npm audit` was clean for the
  three templates at 9.3.3 on 2026-11-04.
- [Generation takes longer] → About 10 seconds for create-vite and the
  lock-file resolution with a warm npm cache.
- [Windows command lookup] → `npm` is `npm.cmd` on Windows; `exec.LookPath`
  resolves it and the arguments contain no characters cmd.exe treats
  specially. The integration job runs on Windows.

## Migration Plan

Additive: no existing selection changes behavior. API-only and htmx
projects change only in `.gitignore` (unchanged content for htmx, through
the new point) and the Makefile, Dockerfile, and Dependabot templates, whose
output is unchanged without contributions (golden diff). The Tailwind CSS
change uses `frontend-build`, the three Vite points, and npm
`dev-dependency` entries; whichever merges second adapts `versions.yaml`
and `base.yaml` entries in a straightforward merge.
