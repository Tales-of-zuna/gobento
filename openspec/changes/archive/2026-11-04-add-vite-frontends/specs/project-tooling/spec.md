## ADDED Requirements

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
