# Design

## Context

See proposal.md for the motivation. Today a generated project is the union
of the four selected components' templates. Root files (`README.md`,
`.gitignore`, `sqlc.yaml`) are written by whichever component has them under
`files/root/`, and the composer rejects two components producing the same
root file. The architecture owns `README.md`; other components extend it only
through `notes.md.tmpl`, which is appended at the end. Template data is
`.Project`, `.Slots`, `.Traits`, and `.Versions` (Go modules only), and
`components/versions.yaml` knows nothing but the Go version and Go modules.

The new files (Makefile, Dockerfile, CI workflow, ...) are the same for every
architecture, but parts of them depend on the database (a PostgreSQL service
in CI, a writable data volume for SQLite) and on the data access (sqlc code
generation). Another change is designing an optional `frontend` kind
(htmx/templ) that will also need to add steps to the Dockerfile and CI.

## Goals / Non-Goals

**Goals:**

- One owner per tooling file, no copies per architecture.
- Database- or data-specific parts come from the component that knows them,
  through a mechanism any present or future kind can use.
- Every version in the new files (images, actions, golangci-lint) is data in
  `components/versions.yaml`, checked when the catalog loads.
- Every new file is exercised for every supported selection by gobento's
  tests where the tool is available (make, actionlint, golangci-lint,
  Docker).

**Non-Goals:**

- No new component kind and no optional kinds (the frontend change owns
  that); no user flag to leave tooling out.
- No deployment manifests (Kubernetes, Compose files) and no image
  publishing in the generated CI.
- `gobento list` and the website's catalog data do not show tooling versions
  yet; the generated files show them.

## Decisions

### 1. A gobento-owned base layer owns the tooling files

`components/base/` holds `base.yaml` (title, description, contribution
points), `notes.md.tmpl` (the README's Development section), and
`files/root/` with `Makefile`, `Dockerfile`, `.dockerignore`,
`.github/workflows/ci.yml`, `.github/dependabot.yml`, and `.golangci.yml`.
The loader keeps it apart from the components (`Catalog.Base()`), it is
composed into every selection (`compose.Input.Base`), and its files carry
the component ID `base`. Its notes are appended after every component's
notes, so the README reads Run → Test → Layout → Database → (Changing
queries) → Development. Only `files/root/` is allowed in the base layer:
it has no slots, and its files are architecture-independent by
construction; it reaches into the layout only through `.Slots`, like
every other template (the Dockerfile builds `./{{.Slots.entry.Dir}}`).

Alternatives considered:

- *Copy the files into each architecture* (they already own the README and
  `.gitignore`): three copies to keep identical, and every new architecture
  adds a fourth. Rejected.
- *Make one architecture-independent component own them*, e.g. a "tooling"
  kind: a kind must be selected, appears in flags, questions, the website,
  and `gobento list`, and multiplies the supported selections. An optional
  kind is out of scope and being designed elsewhere. Rejected.
- *Generate the files in Go code* (like `go.mod`): versions would still be
  data, but contributors could not see or golden-test the files as
  templates, and components could not extend them. Rejected.

### 2. Contribution points instead of traits or per-database files

The base layer declares points in `base.yaml`; a component contributes with
`contrib/<point>.tmpl` in its directory. Contributions are rendered with the
normal template data, trimmed, and joined with newlines in kind order; base
templates insert them with `{{contrib "point"}}`, usually with
`{{indent N ...}}` (YAML) or `{{prefix "\t" ...}}` (Make recipes). The loader
rejects a contribution to an undeclared point and lists the declared ones;
`contrib` is available only in the base layer's templates (files and notes),
and fails for an undeclared point.

Points: `make` (Makefile variables or targets), `make-generate` (recipe of
`make generate`, which exists only when something contributes),
`ci-services`, `ci-env`, `ci-steps` (steps before the build in the CI job),
`docker-build` (instructions in the build stage after the sources are
copied), `docker-runtime` (instructions in the runtime stage before
`USER`), `dockerignore`, and `readme-docker` (README text in the Docker
subsection).

Ownership: the base layer owns every file; SQLite contributes `make`
(volume flag for `docker-run`), `docker-build`/`docker-runtime` (writable
`/data`, `DATABASE_URL`), `dockerignore` (`*.db*`), and `readme-docker`;
PostgreSQL contributes `ci-services`, `ci-env` (`TEST_DATABASE_URL`), `make`
(`host.docker.internal` mapping), and `readme-docker`; sqlc contributes
`make-generate` and `ci-steps` (`sqlc diff`).

Alternatives considered:

- *A trait such as `ci-service: postgres` tested in the base templates*:
  the base layer would name database values and change with every new
  database, which is exactly the per-combination coupling the trait design
  avoids. Rejected.
- *Several components writing the same root file, merged by gobento*: the
  merge semantics differ per file format (YAML, Make, Dockerfile), and the
  one-owner rule for files would no longer hold. Rejected.
- *Structured data contributions* (e.g. `ci: {services: {...}}` in
  `component.yaml`): typed, but every new need (templ steps, extra Docker
  instructions) requires a schema change in Go; snippets need only a new
  point name. Rejected for now; a point can become structured later.

**Extension point for the frontend kind.** A component of any kind,
including a future optional one, contributes by adding `contrib/` files;
`Selection.Components()` is the only list the composer walks, so a new
kind is covered once it is in that list. A templ frontend would contribute
`docker-build` (`RUN go tool templ generate`), `ci-steps` (generate and check
for a diff), `make-generate`, and possibly `dockerignore`. If it needs a
place that no point covers (for example a step in a separate job), it adds
a point to `base.yaml` and one `{{contrib ...}}` call to the base template;
no other component changes.

### 3. Tooling versions are a separate section of the rules file

```yaml
tooling:
  images:   { go: {image: golang, tag: "{go}"}, runtime: {image: ..., tag: nonroot, digest: sha256:...}, postgres: {...} }
  actions:  { actions/checkout: {version: v7.0.1, sha: <40 hex>}, ... }
  tools:    { golangci-lint: v2.15.0 }
```

`compat.ParseRules` validates every entry (image name and tag syntax, the
only placeholder `{go}`, `sha256:` digests, semver action versions and
40-hex SHAs, semver tool versions) and `Resolve` returns the rendered
values for the chosen Go version. Templates see `.Images.<name>` (a full
reference, `name:tag[@digest]`), `index .Actions "<owner/repo>"` (a
complete `uses:` value, `owner/repo@<sha> # <version>`, the format
Dependabot keeps up to date), and `index .Tools "<name>"`. The build image's
tag is `{go}`, the selected Go version, so `--go 1.26` builds with Go 1.26
and patch releases are picked up; the runtime image, which is what ships,
is pinned by digest. The PostgreSQL README's `docker run` command and the CI
service use `.Images.postgres`, removing the last version number from the
templates.

Alternatives considered:

- *Go modules with `role: tool`* for golangci-lint: modules need a
  component, would show up in `gobento list versions` as used by `base`,
  and would become pin targets; actions and images are not modules at all.
  Rejected.
- *Digests for the build image, one per Go minor version*: reproducible,
  but the tag must then be a fixed map that drifts from the Go range and
  blocks Go patch updates. Rejected; documented as a trade-off.
- *Hard-coding the versions in templates and letting Dependabot update the
  generated project*: the first generated version would still be a template
  constant that gobento has to bump by hand in several files. Rejected.

### 4. Generated CI is one check job plus a Docker job

`check` (checkout, setup-go from `go.mod`, `ci-steps`, `go mod tidy -diff`,
build, vet, golangci-lint-action, `go test -race`) carries the
`ci-services` and `ci-env` contributions, so every contribution reaches the
job that compiles code; `docker` runs `docker build`. Actions use
`persist-credentials: false` and the workflow has `contents: read`.
Formatting is checked by golangci-lint's `gofmt` and `goimports` formatters
(with the module as local prefix), so there is no separate gofmt step.

### 5. Makefile, with plain commands documented

Make is the most widely installed task runner (also in Git for Windows);
`just` would add an install step. Every recipe is a plain `go` or `docker`
command, golangci-lint runs as `go run …@<version>` like sqlc, so make is
the only prerequisite, and the README's Development table lists the
equivalent command for each target. `BINARY` gets `.exe` when `OS` is
`Windows_NT`.

### 6. Tests

- Unit: catalog (base layer and `contrib/` validation), compat (tooling
  parsing, resolution, errors), compose (contributions order and placement,
  unknown points, base path conflicts, the postgres and sqlite CI and
  Dockerfile differences), goldens (full goldens include the base files;
  component goldens get `contrib/<point>` pseudo-files).
- Integration (`TestIntegration`, network): for the default variant of every
  selection, run actionlint on the workflow, and where `make` exists run
  `make build` and `make lint` (golangci-lint is built once and cached).
- `TestIntegrationDocker`: build and run the images of
  `mvc/nethttp/sqlite/sqlc` (POST and GET an item, inspect the user) and
  build `ddd/gin/postgres/gorm`; skipped with `-short`, without Docker, or
  when the daemon does not run Linux containers. gobento's Linux CI
  integration job already has Docker, so it runs there.

## Risks / Trade-offs

- [golangci-lint flags a future template change] → `make lint` runs for every
  selection in the integration test, so a template that does not lint
  fails gobento's CI, not the user's.
- [The build image floats on Go patch releases] → deliberate for security;
  the runtime image is digest-pinned and Dependabot updates it.
- [Snippet indentation errors in YAML contributions] → contributions are
  written at column zero and indented by the base template; actionlint runs
  on every generated workflow.
- [Integration test time grows] → linting and actionlint run only for the
  default variant; the Docker test builds two images.
- [`make docker-run` for PostgreSQL cannot reach `localhost`] → documented;
  the Makefile maps `host.docker.internal` to the host gateway.

## Migration Plan

Additive for users: existing projects are not touched. For contributors,
`compose.Input` gains `Base` and `Tooling`; callers inside gobento are
updated in this change. Rolling back is reverting the change.
