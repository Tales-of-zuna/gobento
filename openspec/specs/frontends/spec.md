# frontends Specification

## Purpose
Defines the optional frontend kind: how a project serves web pages next to its
HTTP API without per-combination templates, the web slot contract every
frontend implements, and the htmx + templ frontend.

## Requirements

### Requirement: Frontends are optional
gobento SHALL treat the frontend as an optional component kind: a selection
SHALL contain at most one frontend, and a selection without one SHALL be
valid. Leaving out the frontend SHALL produce an API-only project that
serves the same API at the same paths as before frontends existed.

#### Scenario: No frontend by default
- **WHEN** the user runs `gobento new shop --arch mvc --framework nethttp --db sqlite --data sqlc` without `--frontend`
- **THEN** the project has no web UI, and its HTTP API is served at `/` exactly as without the frontend kind

#### Scenario: Explicitly none
- **WHEN** the user runs the same command with `--frontend none`
- **THEN** gobento generates the same files as without `--frontend`

### Requirement: The web slot contract
An architecture that supports frontends SHALL declare an optional slot `web`
filled by the frontend kind. The frontend component SHALL provide, in that
slot, a function that takes the architecture's store contract
(`models.ItemStore`) and returns an `http.Handler` serving every path outside
the API. The frontend SHALL reach the data only through that contract and
SHALL NOT import the database or store packages, directly or transitively.
A frontend SHALL depend on no particular HTTP framework.

#### Scenario: Web UI depends only on the contract
- **WHEN** a project is generated with any supported selection that includes a frontend
- **THEN** the web package imports neither the database nor the store package, directly or transitively, and imports no HTTP framework module

#### Scenario: Architecture without a web slot
- **WHEN** an architecture does not declare the slot `web` and a frontend is selected with it
- **THEN** the selection is not supported, and gobento exits non-zero naming the frontend and the slot `web`, and writes nothing

### Requirement: The API and the frontend share one server
When a frontend is selected, the generated application SHALL serve the API
router unchanged under the path prefix `/api` (so `GET /api/items` reaches
the API's `GET /items`) and the frontend's handler at every other path. The
API's routes, status codes, and bodies below the prefix SHALL be those of
an API-only project, and the framework component's files SHALL be identical
with and without a frontend. Without a frontend, the API SHALL be served at
`/`.

#### Scenario: API under /api
- **WHEN** a project generated with a frontend is running and receives `GET /api/health`
- **THEN** it responds `200` with `{"status":"ok"}`

#### Scenario: Pages at the root
- **WHEN** a project generated with a frontend is running and receives `GET /`
- **THEN** it responds `200` with an HTML page

#### Scenario: Frontend leaves the framework untouched
- **WHEN** two selections differ only in whether a frontend is selected
- **THEN** the Go files contributed by the framework component are byte-for-byte identical

### Requirement: Entry point tests the routing
The entry point of an architecture SHALL ship a test that checks, without a
database, that the API is served at its base path (`/` without a frontend,
`/api` with one) and, when a frontend is selected, that the frontend answers
`GET /` with HTML.

#### Scenario: Routing test
- **WHEN** `go test ./...` runs in a project generated with a frontend
- **THEN** the entry point's test requests `/api/health` and `/`, and passes

### Requirement: htmx and templ frontend
gobento SHALL offer a frontend `htmx` ("htmx + templ") that renders HTML on
the server with templ components and uses htmx for interaction. Its home page
SHALL list every item, HTML-escaped, and offer a form that adds an item: with
htmx, without a page reload and clearing the form; without JavaScript, by a
form post followed by a redirect to the updated page.

#### Scenario: Add an item with htmx
- **WHEN** the page posts the form to `/items` with the header `HX-Request: true` and the name `first`
- **THEN** the response is `200` with an HTML fragment holding the new item and a fresh form, and `GET /` lists `first`

#### Scenario: Add an item without JavaScript
- **WHEN** a browser posts the form to `/items` without htmx
- **THEN** the response is `303` redirecting to `/`

### Requirement: htmx frontend rejects invalid input
The htmx frontend SHALL reject a blank item name with status `422` and a
message shown next to the field, and SHALL reject unsafe cross-origin
requests with `403`, storing nothing in either case.

#### Scenario: Empty name
- **WHEN** the form is posted with a blank name
- **THEN** the response is `422` and shows a message asking for a name, and no item is stored

#### Scenario: Cross-origin form post
- **WHEN** a form on another site posts to `/items` (the browser sends `Sec-Fetch-Site: cross-site`)
- **THEN** the response is `403` and no item is stored

### Requirement: Frontend assets are served from the binary
The htmx frontend SHALL serve htmx and its stylesheet from files embedded in
the Go binary, with no request to a CDN or other network location, so the
application works offline. The htmx file SHALL be a pinned upstream release,
vendored unmodified. The stylesheet SHALL be the frontend's plain CSS, or
the one the selected CSS framework contributes.

#### Scenario: Offline page
- **WHEN** a project generated with `--frontend htmx` runs on a machine without network access and a browser opens `/`
- **THEN** every script and stylesheet of the page is loaded from the application itself, under `/static/`

#### Scenario: Vendored file matches upstream
- **WHEN** gobento's tests run with network access
- **THEN** the shipped htmx file is byte-for-byte the file of the recorded upstream release

#### Scenario: Stylesheet styles the markup
- **WHEN** a project generated with `--frontend htmx`, with or without a CSS framework, serves `/static/styles.css`
- **THEN** the response is CSS that styles the classes the pages use

### Requirement: Generated templ code is shipped and current
The htmx frontend SHALL ship both the templ source and the Go code templ
generates from it, so a generated project builds and tests without the templ
command. The shipped generated code SHALL be exactly what the pinned templ
version generates for the project, and the generated project SHALL document
the command that regenerates it.

#### Scenario: Build without templ installed
- **WHEN** a project generated with `--frontend htmx` runs `go build ./...` on a machine without the templ command
- **THEN** the build succeeds

#### Scenario: Regeneration changes nothing
- **WHEN** the regeneration command from the generated README runs in a freshly generated project
- **THEN** no file changes

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
Node.js sources. Components that work with only some frontends, such as CSS
frameworks, SHALL require this trait instead of naming frontends, and use it
to decide what to contribute.

#### Scenario: Trait values
- **WHEN** the catalog is listed
- **THEN** `htmx` provides `frontend-build: go`, and `react`, `vue`, and `svelte` provide `frontend-build: vite`

#### Scenario: htmx build system
- **WHEN** a user runs `gobento list components --kind frontend`
- **THEN** `htmx` is listed with `provides: frontend-build=go`
