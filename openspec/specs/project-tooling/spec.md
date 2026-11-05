# project-tooling Specification

## Purpose
Gives every generated project the developer tooling a production service
needs from day one: a task runner, a minimal container image, continuous
integration, and lint and dependency-update configuration, correct for the
selected components and usable on Linux, macOS, and Windows.

## Requirements

### Requirement: Tooling files in every project
A project generated from any supported selection SHALL contain a `Makefile`,
a `Dockerfile`, a `.dockerignore`, a GitHub Actions workflow
`.github/workflows/ci.yml`, a Dependabot configuration
`.github/dependabot.yml`, and a golangci-lint configuration `.golangci.yml`.
These files SHALL be the same for every architecture apart from the slot
directories they name, and SHALL contain only the parts that the selected
components need.

#### Scenario: Files for every architecture
- **WHEN** the user runs `gobento new shop --arch <any architecture> --framework nethttp --db sqlite --data sqlc --dry-run`
- **THEN** the listed files include `Makefile`, `Dockerfile`, `.dockerignore`, `.github/workflows/ci.yml`, `.github/dependabot.yml`, and `.golangci.yml`

### Requirement: Make targets with plain-command fallbacks
The `Makefile` SHALL offer at least the targets `help` (the default), `build`,
`run`, `test`, `cover`, `vet`, `lint`, `fmt`, `tidy`, `check`, `docker-build`,
`docker-run`, and `clean`, and each target SHALL run plain `go` or `docker`
commands. The generated README SHALL list, for every task, the command to run
without make, so the project is usable on Windows without make. `make build`
SHALL write the binary to `bin/`, with an `.exe` suffix on Windows.

#### Scenario: Build with make
- **WHEN** `make build` runs in a generated project named `shop`
- **THEN** it creates `bin/shop` and exits 0

#### Scenario: Default target
- **WHEN** `make` runs without a target in a generated project
- **THEN** it lists the targets with a one-line description each

#### Scenario: Without make
- **WHEN** a user reads the generated README
- **THEN** its Development section shows, next to each make target, the `go` or `docker` command that does the same

### Requirement: Static, non-root container image
The `Dockerfile` SHALL build the application in a Go build stage with cgo
disabled (`CGO_ENABLED=0`) and SHALL copy only the binary into a minimal
runtime image without a shell or package manager, which runs it as a
non-root user. The Go build image SHALL follow the project's selected Go
version, and the runtime image SHALL be pinned by digest.

#### Scenario: Image builds and serves
- **WHEN** `docker build` runs in a project generated with `--db sqlite` and the image is started with port 8080 published
- **THEN** `GET /health` responds 200

#### Scenario: Image runs as non-root
- **WHEN** the image of any generated project is inspected
- **THEN** its user is the numeric non-root user `65532:65532`

#### Scenario: Go version follows the selection
- **WHEN** a project is generated with `--go 1.26`
- **THEN** the `Dockerfile` builds with the Go 1.26 image

### Requirement: Writable data for file databases in the image
When the selected database stores its data in a file, the image SHALL keep
that file in a volume at `/data` that the non-root user can write, and SHALL
set `DATABASE_URL` to a file in it. The generated README SHALL explain how to
keep the data with a named volume.

#### Scenario: SQLite data in the container
- **WHEN** the image of a project generated with `--db sqlite` is started and receives `POST /items`
- **THEN** the item is stored in a database file under `/data`, and `GET /items` returns it

### Requirement: Server databases in the container
When the selected database needs a server, the README SHALL explain how to
point the container at a server with `DATABASE_URL`, including a server that
runs on the Docker host, and `make docker-run` SHALL pass `DATABASE_URL` to the
container when it is set.

#### Scenario: PostgreSQL image builds
- **WHEN** `docker build` runs in a project generated with `--db postgres`
- **THEN** the build succeeds

### Requirement: Continuous integration workflow
The CI workflow SHALL run on pushes to `main` and on pull requests, with
read-only repository permissions, and SHALL check that `go.mod` and `go.sum`
are tidy, build, vet, lint with golangci-lint, run the tests with the race
detector, and build the Docker image. Every action SHALL be pinned by full
commit SHA. The workflow SHALL pass actionlint.

#### Scenario: Valid workflow
- **WHEN** actionlint checks the generated `.github/workflows/ci.yml` of any supported selection
- **THEN** it reports no problems

#### Scenario: Actions pinned by SHA
- **WHEN** a project is generated
- **THEN** every `uses:` in its workflow names a 40-character commit SHA, followed by a comment with the release version

#### Scenario: Database tests run in CI
- **WHEN** a project is generated with `--db postgres`
- **THEN** its workflow starts a PostgreSQL service container and sets `TEST_DATABASE_URL` for the tests, so the database tests run instead of being skipped

#### Scenario: No server for file databases
- **WHEN** a project is generated with `--db sqlite`
- **THEN** its workflow starts no service container and sets no `TEST_DATABASE_URL`

#### Scenario: Generated code is checked in CI
- **WHEN** a project is generated with `--data sqlc`
- **THEN** its workflow fails when the committed sqlc code differs from what the pinned sqlc generates from the queries and the schema

### Requirement: Lint configuration that the project passes
The generated `.golangci.yml` SHALL configure golangci-lint (configuration
version 2) with formatting checks, and every generated project SHALL pass
`golangci-lint run` with it, at the version the Makefile and the workflow use.

#### Scenario: Fresh project lints clean
- **WHEN** `make lint` runs in a project generated from any supported selection
- **THEN** it reports no issues and exits 0

### Requirement: Dependency updates
The generated `.github/dependabot.yml` SHALL ask for updates of the Go
modules, the GitHub Actions, and the Docker base images, so the pinned
versions do not go stale.

#### Scenario: Ecosystems
- **WHEN** a project is generated
- **THEN** its `.github/dependabot.yml` lists the `gomod`, `github-actions`, and `docker` ecosystems

### Requirement: Frontend code generation in the tooling
When the selected frontend ships generated code, such as the Go code templ
generates from `pages.templ`, `make generate` SHALL regenerate it with the
version from the rules file, the CI workflow SHALL fail when the committed
generated code is not what that version generates, and the README SHALL name
the command without make. The `.dockerignore` and `.gitignore` SHALL exclude
the files templ writes in watch mode. A project without a frontend SHALL get
none of this.

#### Scenario: Regenerate the pages with make
- **WHEN** `make generate` runs in a project generated with `--frontend htmx`
- **THEN** it runs `go generate` in the web package, whose `//go:generate` line runs the pinned templ, and the committed files stay unchanged

#### Scenario: Fresh templ code passes CI
- **WHEN** the CI step for templ runs in a freshly generated project with `--frontend htmx`
- **THEN** regenerating changes no file and the step succeeds

#### Scenario: Stale templ code fails CI
- **WHEN** `pages.templ` is changed and committed without regenerating `pages_templ.go`
- **THEN** the CI step for templ shows the difference and fails

#### Scenario: Watch-mode files stay out of the image
- **WHEN** a project is generated with `--frontend htmx`
- **THEN** its `.dockerignore` and `.gitignore` list `*_templ.txt`

### Requirement: Images serve the frontend
The image of a project with a frontend SHALL serve the pages and their static
files from the binary, with the API under `/api/`, and the README's Docker
section SHALL say so.

#### Scenario: Image serves the pages
- **WHEN** the image of a project generated with `--db sqlite --frontend htmx` is started with port 8080 published
- **THEN** `GET /api/health` responds 200, `GET /` responds 200 with the page, and `GET /static/htmx.min.js` responds 200

### Requirement: Frontend build in the tooling
With a Vite frontend, `make build` SHALL install and build the frontend
before the binary, and `make dev` SHALL run the Vite dev server, which
forwards `/api` to the Go server on port 8080. The CI workflow SHALL set up
Node.js at the rules file's release line with a pinned action, run `npm
ci`, the checks the scaffolded project defines, and the build before the Go
checks. Without a Vite frontend, none of this is generated.

#### Scenario: Build with make
- **WHEN** `make build` runs in a project generated with `--frontend react`
- **THEN** it runs `npm ci` and `npm run build` in `frontend/` and then builds the binary, which serves the built app at `/`

### Requirement: Images and dependency updates cover the frontend
In a project with a Vite frontend, the Dockerfile SHALL build the frontend
in a stage of its own from the rules file's `node` image and copy only the
build into the Go build stage, leaving the runtime image unchanged.
Dependabot SHALL update the npm dependencies of `frontend/`, and
`.dockerignore` SHALL exclude `frontend/node_modules` and a local build.

#### Scenario: Image serves the app
- **WHEN** the image of a project generated with `--db sqlite --frontend vue` is built and started with port 8080 published
- **THEN** `GET /` responds 200 with the built app's page, a client route responds with the same page, and `GET /api/health` responds 200

#### Scenario: Dependency updates for the frontend
- **WHEN** a project is generated with a Vite frontend
- **THEN** its `.github/dependabot.yml` also lists the `npm` ecosystem for `/frontend`

### Requirement: Ignore files from components
The generated `.gitignore` SHALL include the patterns that the selected
components contribute, and only those beyond the architecture's own, so a
frontend's generated or watch-mode files are ignored only with that
frontend.

#### Scenario: templ watch files only with htmx
- **WHEN** a project is generated with `--frontend react`
- **THEN** its `.gitignore` does not list `*_templ.txt`, while one generated with `--frontend htmx` does
