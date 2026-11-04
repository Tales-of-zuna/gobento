# Proposal

## Why

The roadmap promises React, Vue, and Svelte frontends next to htmx + templ,
and the htmx change left their path open (its design, D10): the same `web`
slot, a single-page app built with Vite and embedded in the Go binary. These
three frameworks have an official scaffolder, `create-vite`, so by the
project's principle gobento must use it instead of owning their templates.
That needs what gobento has never had: a step that runs a tool other than
`go` during generation, Node.js as a prerequisite, and npm versions as rules
data.

## What Changes

- Three frontend components, `frontend/react`, `frontend/vue`, and
  `frontend/svelte`, all TypeScript. Generation runs
  `npm create vite@<version> frontend -- --template <react|vue|svelte>-ts`
  non-interactively, adds a few gobento-owned files on top (the Vite
  configuration, an items page calling `/api/items`, the API client, the
  stylesheet), and records the dependency tree in `package-lock.json`. The
  three behave alike: the same files, the same page, the same tooling.
- The web slot package of a Vite frontend embeds the built app and serves it
  with an `index.html` fallback for client-side routes, never answering
  under `/api/`. Until the frontend is built it serves a page saying how to
  build it, so `go build` and `go test` work without Node.js.
- A scaffold step in generation, after the files are written to the
  temporary directory and before they are moved into place. It keeps the
  guarantees of generation: nothing is left behind when it fails or is
  interrupted. Preflight checks `node` and `npm` (and the Node.js version)
  only when such a frontend is selected; `--dry-run` lists gobento's files
  and says that the scaffolder adds more.
- The rules file gains `node` (the supported Node.js range and the release
  line of the generated tooling) and `npm` (npm packages with a fixed
  version: the scaffolder, and packages a component adds to a scaffolded
  frontend). Templates see `.Node` and `.NPM`; the tooling gains the
  `node` image and `actions/setup-node`.
- Generated tooling for Vite frontends: `make build` builds the frontend
  first and `make dev` runs the Vite dev server; the CI workflow sets up
  Node.js, installs, checks, and builds the frontend before the Go checks;
  the Dockerfile builds it in a `node` stage; Dependabot updates the npm
  dependencies; `.dockerignore` skips `node_modules`.
- New contribution points: `make-build`, `docker-stages`, `dependabot`,
  `gitignore`, and three points inside the Vite frontends' files
  (`vite-imports`, `vite-plugins`, `vite-styles`) that a CSS component, such
  as the Tailwind CSS component being added in parallel, fills. Component
  templates (not only the base layer's) may include contribution points.
- Every frontend provides the trait `frontend-build`: `go` for htmx (built
  by the Go toolchain alone) and `vite` for the new ones.
- The website lists the three frontends, drops them from the roadmap, and
  shows the scaffolded directory in the configurator's file preview.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `frontends`: the Vite frontends (scaffolding, the example page, serving
  the embedded build with a client-route fallback, the unbuilt placeholder),
  and the `frontend-build` trait.
- `project-generation`: the scaffold step, the Node.js preflight, the dry
  run note, cleanup of scaffolded files, next steps, and CI coverage with
  Node.js.
- `version-compatibility`: Node.js and npm package versions as rules data.
- `project-tooling`: frontend build in `make`, CI, the Dockerfile,
  Dependabot, and `.dockerignore`; the `gitignore` point.
- `component-composition`: new contribution points and contributions in
  component templates.
- `catalog-discovery`: `gobento list` shows the Node.js rule and npm
  packages.
- `website-configurator`: the file preview shows the scaffolded directory
  of a Vite frontend.

## Impact

- Code: `internal/catalog` (manifest `scaffold`), `internal/compat`
  (`node`, `npm`, the `{node}` image placeholder), `internal/compose`
  (`.Node`, `.NPM`, contributions in component templates),
  `internal/generate` (scaffold step, Node.js preflight, process-tree
  cancellation, integration test), `internal/cli` (dry run, next steps,
  list), `internal/catalogjson`, `components/` (three new frontends,
  htmx trait and `gitignore` contribution, base templates, architecture
  `.gitignore`, `versions.yaml`), `.github/workflows/ci.yml` (Node.js in the
  integration job).
- Dependencies: generation needs Node.js `>=22.12.0` and npm, and network
  access to the npm registry, only for a Vite frontend. create-vite 9.3.3
  scaffolds Vite 8 projects.
- Website: catalog schema and fixture (`node`, `npm`), configurator file
  preview, roadmap lists, home page, component reference.
- Documentation updated in this change: `README.md`, `CONTRIBUTING.md`,
  `website/content/docs/index.mdx`, `getting-started/installation.mdx`
  (Node.js), `getting-started/quick-start.mdx` ("With React, Vue, or
  Svelte"), `cli/new.mdx` (frontends, preflight errors, dry run), `cli/list.mdx`,
  `concepts/components.mdx`, `concepts/slots-and-contracts.mdx`,
  `concepts/traits.mdx`, `concepts/version-rules.mdx` (`node`, `npm`),
  `concepts/project-tooling.mdx`, `contributing/adding-a-component.mdx`
  (scaffolded components, new points), the generated README notes of the
  new frontends, and the header of `components/versions.yaml`.
