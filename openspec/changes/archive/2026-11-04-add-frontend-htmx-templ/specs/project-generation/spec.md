## MODIFIED Requirements

### Requirement: Generate a project from a component selection
`gobento new <name>` SHALL create a Go project from one selected architecture,
HTTP framework, database, and data-access component, and optionally one
frontend (`--frontend`), using the given Go module path, or the project name
when `--module` is omitted.

#### Scenario: Fully specified by flags
- **WHEN** the user runs `gobento new shop --module example.com/shop --arch mvc --framework nethttp --db sqlite --data sqlc`
- **THEN** a directory `shop` is created containing a Go module named `example.com/shop`
- **AND** the command exits with status 0 and prints the next commands to run

#### Scenario: GORM data access
- **WHEN** the user runs `gobento new shop --arch mvc --framework nethttp --db sqlite --data gorm`
- **THEN** a directory `shop` is created whose store is implemented with GORM on the SQLite database
- **AND** the command exits with status 0

#### Scenario: htmx frontend
- **WHEN** the user runs `gobento new shop --arch mvc --framework nethttp --db sqlite --data sqlc --frontend htmx`
- **THEN** a directory `shop` is created that serves HTML pages at `/` and the API under `/api`
- **AND** the command exits with status 0

#### Scenario: Module path defaults to the project name
- **WHEN** the user runs `gobento new shop --arch mvc --framework nethttp --db sqlite --data sqlc` without `--module`
- **THEN** the generated module is named `shop`

#### Scenario: Architecture must be chosen
- **WHEN** the user runs `gobento new shop` without `--arch` while `mvc` and `vertical-slice` are both supported architectures
- **THEN** gobento exits non-zero naming `--arch` and listing `mvc` and `vertical-slice`, and writes nothing

### Requirement: Automatic selection of the only option
When a selection flag of a required kind is omitted and the kind has exactly
one supported value, gobento SHALL select that value. When the kind has
several values and gobento does not ask questions (see "Ask for missing input
in a terminal"), gobento SHALL exit non-zero naming the flag and listing the
values. When the flag of an optional kind is omitted, gobento SHALL select no
component of that kind without asking for the flag; the value `none` SHALL
select no component explicitly. An unknown value of an optional kind SHALL be
reported with the supported values, including `none`.

#### Scenario: Single option is chosen automatically
- **WHEN** the user omits `--db` and `sqlite` is the only supported database
- **THEN** the project is generated with `sqlite`

#### Scenario: Several options require a choice
- **WHEN** the user omits a selection flag of a required kind that has several supported values, and stdin is not a terminal or `--no-input` is given
- **THEN** gobento exits non-zero naming the flag and its supported values, and writes nothing

#### Scenario: Optional kind needs no flag
- **WHEN** the user omits `--frontend` without a terminal while `htmx` is a supported frontend
- **THEN** the project is generated without a frontend

#### Scenario: Unknown frontend
- **WHEN** the user passes `--frontend react` while only `htmx` is a supported frontend
- **THEN** gobento exits non-zero naming `--frontend` and listing `none` and `htmx`, and writes nothing

### Requirement: Generated project builds and passes its tests
A project generated from any supported selection SHALL compile, pass `go vet`
and `gofmt -l`, and pass its own `go test ./...` without edits by the user. It
SHALL also compile with cgo disabled (`CGO_ENABLED=0`), so it needs no C
compiler on Linux, macOS, or Windows.

#### Scenario: Build check after generation
- **WHEN** a project is generated with the selection mvc/nethttp/sqlite/sqlc, vertical-slice/nethttp/sqlite/sqlc, mvc/nethttp/sqlite/gorm, or mvc/nethttp/sqlite/sqlc with the htmx frontend
- **THEN** `go build ./...`, `go vet ./...` and `go test ./...` succeed inside it, and `gofmt -l .` lists no files

#### Scenario: Build check without a database server
- **WHEN** a project is generated with the selection mvc/nethttp/postgres/sqlc on a machine without a PostgreSQL server and `TEST_DATABASE_URL` is unset
- **THEN** `go build ./...`, `go vet ./...` and `go test ./...` succeed inside it, and `gofmt -l .` lists no files

#### Scenario: Database tests run against a server
- **WHEN** a project is generated with the selection mvc/nethttp/postgres/sqlc and `TEST_DATABASE_URL` names a reachable PostgreSQL server
- **THEN** `go test ./...` succeeds and its database and store tests run against that server instead of being skipped

#### Scenario: Build without cgo
- **WHEN** `CGO_ENABLED=0 go build ./...` runs in a project generated from any supported selection
- **THEN** it succeeds

#### Scenario: Example functionality is present
- **WHEN** the generated project is running
- **THEN** `GET /health` responds 200 at the API's base path (`/` without a frontend, `/api` with one)
- **AND** an item created with `POST /items` at that base path is returned by `GET /items` there, persisted in the selected database

### Requirement: Ask for missing input in a terminal
In a terminal (stdin and stdout) without `--no-input`, `gobento new` SHALL
ask, before validating, for the project name when no `<name>` is given, then
for the module path (empty keeps the name), and for each kind whose flag is
omitted and that has several compatible options, in kind order. For an
optional kind, "none" counts as an option and SHALL be offered first, as the
preselected answer. Answers SHALL follow the name and module path rules.
Nothing given by a flag SHALL be asked, and `--dry-run` SHALL ask the same
questions.

#### Scenario: Several frameworks, framework flag omitted
- **WHEN** the user runs `gobento new shop` in a terminal and the catalog has several supported frameworks
- **THEN** gobento asks for the framework, and not for the name, the module path, or a kind with a single option
- **AND** it generates the project with the chosen framework

#### Scenario: No name given
- **WHEN** the user runs `gobento new` in a terminal
- **THEN** gobento asks for the project name, rejecting names that break the project name rules, and then for the module path, using the name when the answer is empty

#### Scenario: Flags skip their questions
- **WHEN** the user runs `gobento new shop --framework nethttp` in a terminal
- **THEN** gobento does not ask for the framework

#### Scenario: Frontend question
- **WHEN** the user runs `gobento new shop --arch mvc --framework nethttp --db sqlite --data sqlc` in a terminal
- **THEN** gobento asks for the frontend, offering "None (API only)" first and then `htmx`

### Requirement: Equivalent command after questions
After asking at least one question, gobento SHALL print on stderr the
equivalent `gobento new` command, with every answer as an argument or flag
and the other flags of the run, so the run can be repeated without
questions. An optional kind answered with "none" SHALL be left out unless its
flag was given, because leaving out the flag means the same.

#### Scenario: Framework chosen by a question
- **WHEN** the user answers `gin` to the framework question of `gobento new shop --dry-run`
- **THEN** stderr shows a command that starts with `gobento new shop` and contains `--framework gin` and `--dry-run`

#### Scenario: No frontend chosen by a question
- **WHEN** the user answers "None (API only)" to the frontend question
- **THEN** the equivalent command contains no `--frontend` flag

## ADDED Requirements

### Requirement: Supported selections are verified in CI
gobento's CI SHALL generate and check projects on Linux, macOS, and Windows,
each with the default versions and with the floor of every range. It SHALL
cover every supported selection without an optional component, and a subset
of the selections with one in which every pair of components that appears
together in such a selection appears together at least once, chosen
deterministically from the catalog.

#### Scenario: Backend combinations are exhaustive
- **WHEN** CI runs the integration test
- **THEN** every supported selection without a frontend is generated and checked

#### Scenario: Frontend combinations are covered pairwise
- **WHEN** CI runs the integration test and the catalog has a frontend
- **THEN** for every architecture, framework, database, and data-access component that can be combined with that frontend, a project with the frontend and that component is checked, and for every two such components that can appear together with the frontend, a project with the frontend and both of them is checked
