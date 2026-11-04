# Proposal

## Why

A freshly generated project builds and passes its tests, but it has no
developer tooling around the code: no task runner, no container image, no
CI, and no linter configuration. Every user has to write the same Makefile,
Dockerfile, and workflow by hand before the project is production-friendly,
and they are easy to get subtly wrong (cgo in the image, running as root,
unpinned actions, a CI that never runs the PostgreSQL tests).

## What Changes

- Every generated project gets a `Makefile`, a multi-stage `Dockerfile`
  (static `CGO_ENABLED=0` build, distroless non-root runtime image), a
  `.dockerignore`, a GitHub Actions workflow `.github/workflows/ci.yml`
  (tidy check, build, vet, golangci-lint, tests with the race detector, and
  a Docker image build), `.github/dependabot.yml` (Go modules, actions, and
  base images), and a `.golangci.yml`. The README gets a "Development"
  section that lists every task with the plain `go` or `docker` command for
  systems without make (Windows).
- A new gobento-owned **base layer** (`components/base/`) produces these
  files for every selection, so they are written once instead of once per
  architecture. It is not a component kind and is never selected.
- Components **contribute** to the base layer's files through named
  contribution points declared by the base layer (for example `ci-services`
  or `docker-runtime`), with templates in `contrib/<point>.tmpl`. PostgreSQL
  contributes a CI service container and `TEST_DATABASE_URL`; SQLite
  contributes a writable `/data` volume to the image; sqlc contributes
  `make generate` and a CI check that its generated code is current. Future
  kinds (such as the planned frontend) use the same points.
- The rules file gets a `tooling` section for versions that are not Go
  module requirements: container images (with digests where fixed),
  GitHub Actions (version and commit SHA), and tools (golangci-lint).
  Templates read them as `.Images`, `.Actions`, and `.Tools`; they still
  contain no version numbers. The PostgreSQL README's server command moves
  to the same data.
- gobento's integration test also lints every generated workflow with
  actionlint, runs the generated Makefile's targets where make exists, and
  builds and runs the Docker image of one SQLite and one PostgreSQL project
  where a Linux Docker daemon is available.

## Capabilities

### New Capabilities

- `project-tooling`: the developer tooling every generated project contains
  (Makefile, Dockerfile, CI workflow, lint and Dependabot configuration) and
  the guarantees about it.

### Modified Capabilities

- `component-composition`: a base layer composed into every selection, and
  contribution points through which components extend the base layer's files.
- `version-compatibility`: the rules file also records container images,
  GitHub Actions, and tool versions that are not Go modules.

## Impact

- Code: `internal/catalog` (base layer, `contrib/` directories, contribution
  point validation), `internal/compat` (tooling section, resolution),
  `internal/compose` (base layer composition, `contrib`/`indent`/`prefix`
  template functions, new template data), `internal/generate` (passing the
  base layer and tooling; integration and Docker tests), `components/`
  (base layer, contributions of sqlite, postgres, and sqlc, versions.yaml).
- Golden files: every full golden and the component goldens of sqlite,
  postgres, and sqlc (new `contrib/` pseudo-files).
- No new Go dependencies. The generated projects reference the golang and
  distroless images, `actions/checkout`, `actions/setup-go`,
  `golangci/golangci-lint-action`, and golangci-lint.
- Documentation: `README.md` (generated tree, how it works),
  `CONTRIBUTING.md`, `website/content/docs/getting-started/quick-start.mdx`
  (what gets generated), a new page
  `website/content/docs/concepts/project-tooling.mdx`,
  `website/content/docs/concepts/version-rules.mdx` (tooling section),
  `website/content/docs/contributing/adding-a-component.mdx` (contributions
  and the base layer), `website/content/docs/cli/new.mdx` (dry-run file
  list), `website/content/docs/concepts/components.mdx`, and the
  configurator's file-tree preview
  (`website/components/configurator/view-model.ts`).
