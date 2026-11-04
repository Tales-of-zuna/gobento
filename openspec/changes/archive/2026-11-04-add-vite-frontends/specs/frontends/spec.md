## ADDED Requirements

### Requirement: Vite frontends from the official scaffolder
gobento SHALL offer the TypeScript frontends `react`, `vue`, and `svelte`,
built with Vite, whose sources the official scaffolder `create-vite`
creates in `frontend/` at generation time, at the rules file's version and
without questions. gobento SHALL add only its Vite configuration, example
page, API client, and stylesheet, replacing scaffolded files of the same
name, and SHALL record the dependency tree in `frontend/package-lock.json`.

#### Scenario: React project
- **WHEN** the user runs `gobento new shop --arch mvc --framework nethttp --db sqlite --data sqlc --frontend react` with Node.js and network access
- **THEN** `frontend/` holds the files of create-vite's `react-ts` template, gobento's `frontend/vite.config.ts` and example page, and a `frontend/package-lock.json`, and no question was asked

#### Scenario: The three frontends are alike
- **WHEN** the same selection is generated with `--frontend react`, `--frontend vue`, and `--frontend svelte`
- **THEN** each project has the same gobento-owned files apart from the framework's component file names, the same web package, and the same tooling, and its page behaves the same

### Requirement: Vite example page
The example page of a Vite frontend SHALL list the items it gets from
`GET /api/items` and offer a form that adds an item with `POST /api/items`
without reloading the page, then shows it and clears the field. An error
response of the API, such as a blank name, SHALL be shown as a message on
the page.

#### Scenario: Add an item
- **WHEN** a browser opens `/` of a running project generated with a Vite frontend whose frontend is built, enters `first`, and submits the form
- **THEN** the page lists `first` without a reload, the field is empty, and `GET /api/items` returns `first`

#### Scenario: Blank name
- **WHEN** the form is submitted with a blank name
- **THEN** the page shows the API's message and lists no new item

### Requirement: Vite frontends serve the embedded build
The web package of a Vite frontend SHALL serve the built app embedded in the
binary: a file of the build at its path, and the app's `index.html` for
every other `GET` or `HEAD` path without a file extension, so client-side
routes work on reload. It SHALL answer `404` for a missing file with an
extension and for every path under the API prefix, and `405` for other
methods.

#### Scenario: Built file
- **WHEN** a project with a built Vite frontend receives `GET` for a file of the build under `/assets/`
- **THEN** it responds `200` with that file

### Requirement: Vite frontends build without Node.js
Until the frontend is built, the web package of a Vite frontend SHALL answer
with an HTML page that says how to build it, so the project builds and
passes its tests without Node.js.

#### Scenario: Unbuilt frontend
- **WHEN** `go build ./...` and `go test ./...` run in a freshly generated project before `npm run build`
- **THEN** both succeed, and `GET /` responds `200` with a page naming the build command


#### Scenario: Client route
- **WHEN** a project with a built Vite frontend receives `GET /items/42`
- **THEN** it responds `200` with the app's `index.html`

#### Scenario: Missing asset
- **WHEN** it receives `GET /assets/missing.js`
- **THEN** it responds `404`

#### Scenario: API paths are never the app
- **WHEN** the web handler receives a request for a path under `/api/`
- **THEN** it responds `404` and never the app's `index.html`

### Requirement: Frontends declare how they are built
Every frontend SHALL provide the trait `frontend-build`: `go` when the Go
toolchain alone builds it (htmx), and `vite` when it is built with Vite from
Node.js sources. Components that work with only some frontends SHALL
require this trait instead of naming frontends.

#### Scenario: Trait values
- **WHEN** the catalog is listed
- **THEN** `htmx` provides `frontend-build: go`, and `react`, `vue`, and `svelte` provide `frontend-build: vite`
