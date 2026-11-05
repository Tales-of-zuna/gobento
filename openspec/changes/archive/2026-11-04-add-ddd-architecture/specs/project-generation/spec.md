## MODIFIED Requirements

### Requirement: Generate a project from a component selection
`gobento new <name>` SHALL create a Go project from one selected architecture,
HTTP framework, database, and data-access component, using the given Go module
path, or the project name when `--module` is omitted.

#### Scenario: Fully specified by flags
- **WHEN** the user runs `gobento new shop --module example.com/shop --arch mvc --framework nethttp --db sqlite --data sqlc`
- **THEN** a directory `shop` is created containing a Go module named `example.com/shop`
- **AND** the command exits with status 0 and prints the next commands to run

#### Scenario: GORM data access
- **WHEN** the user runs `gobento new shop --arch mvc --framework nethttp --db sqlite --data gorm`
- **THEN** a directory `shop` is created whose store is implemented with GORM on the SQLite database
- **AND** the command exits with status 0

#### Scenario: Module path defaults to the project name
- **WHEN** the user runs `gobento new shop --arch mvc --framework nethttp --db sqlite --data sqlc` without `--module`
- **THEN** the generated module is named `shop`

#### Scenario: Architecture must be chosen
- **WHEN** the user runs `gobento new shop` without `--arch` and without questions while `ddd`, `mvc`, and `vertical-slice` are the supported architectures
- **THEN** gobento exits non-zero naming `--arch` and listing `ddd`, `mvc`, and `vertical-slice`, and writes nothing

### Requirement: Generated project builds and passes its tests
A project generated from any supported selection SHALL compile, pass `go vet`
and `gofmt -l`, and pass its own `go test ./...` without edits by the user. It
SHALL also compile with cgo disabled (`CGO_ENABLED=0`), so it needs no C
compiler on Linux, macOS, or Windows.

#### Scenario: Build check after generation
- **WHEN** a project is generated with the selection mvc/nethttp/sqlite/sqlc, vertical-slice/nethttp/sqlite/sqlc, ddd/nethttp/sqlite/sqlc, or mvc/nethttp/sqlite/gorm
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
- **THEN** `GET /health` responds 200
- **AND** an item created with `POST /items` is returned by `GET /items`, persisted in the selected database
