## ADDED Requirements

### Requirement: Scaffold step
When the selection includes a component with a scaffolder, generation SHALL
run it after writing gobento's files to the temporary directory and before
moving the project into place, merge its files under the component's
directory with gobento's files taking precedence, and record the npm
dependency tree in a lock file without installing the dependencies. A
failure or interruption SHALL leave nothing behind, including files written
by processes the scaffolder started.

#### Scenario: Scaffolder fails
- **WHEN** the scaffolder or npm fails during generation (for example, without network access)
- **THEN** gobento exits non-zero with npm's output, and no generated files remain on disk

#### Scenario: Interrupted while scaffolding
- **WHEN** the user presses Ctrl-C while the scaffolder runs
- **THEN** gobento stops the scaffolder and every process it started, removes everything it created, and exits with status 130

### Requirement: Node.js preflight
Before writing, when the selection includes a component with a scaffolder,
gobento SHALL check that the `node` and `npm` commands are available and
that the Node.js version lies in the rules file's Node.js range, and report
every problem together with the other preflight problems. Without such a
component, gobento SHALL NOT look for Node.js.

#### Scenario: Node.js is missing
- **WHEN** the user runs `gobento new shop ... --frontend react` and no `node` command is found on PATH
- **THEN** gobento exits non-zero naming the frontend, the scaffolder, and the required Node.js range, with instructions to install Node.js, and writes nothing

#### Scenario: Node.js is too old
- **WHEN** the local Node.js version is below the floor of the Node.js range
- **THEN** gobento exits non-zero naming both versions, and writes nothing

#### Scenario: API-only projects need no Node.js
- **WHEN** the user generates a project without a frontend, or with `--frontend htmx`, on a machine without Node.js
- **THEN** generation succeeds

### Requirement: Dry run with a scaffolder
With `--dry-run`, gobento SHALL list the files it writes itself and SHALL
say which scaffolder adds more files to which directory, and that
generation needs Node.js and network access to the npm registry. It SHALL
NOT run the scaffolder or require Node.js.

#### Scenario: Dry run of a React project
- **WHEN** the user runs `gobento new shop ... --frontend react --dry-run`
- **THEN** the output lists `frontend/vite.config.ts` and the other gobento files, says that `npm create vite@<version>` with the template `react-ts` adds more files to `frontend/` and writes `frontend/package-lock.json`, and nothing is written

### Requirement: Next steps build the frontend
After generating a project with a Vite frontend, the next steps SHALL
include the commands that install and build the frontend before the
command that runs the server.

#### Scenario: Next steps
- **WHEN** a project with `--frontend vue` is generated
- **THEN** the next steps list `npm --prefix frontend ci` and `npm --prefix frontend run build` before `go run ./cmd/shop`

### Requirement: Vite frontends are verified in CI
gobento's CI SHALL set up Node.js for the integration test on Linux, macOS,
and Windows. For each generated project with a Vite frontend, it SHALL
install the frontend's dependencies with `npm ci`, run its checks and its
build, and then build and test the Go project with the built app embedded.

#### Scenario: Frontend build in the integration test
- **WHEN** CI generates a covered selection with a Vite frontend with the default versions
- **THEN** `npm ci` and `npm run build` succeed in `frontend/`, the build is in the web package's `dist/app/`, and `go build ./...` and `go test ./...` succeed
