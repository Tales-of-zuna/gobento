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
- **WHEN** the user runs `gobento new shop --db sqlite --data gorm`
- **THEN** a directory `shop` is created whose store is implemented with GORM on the SQLite database
- **AND** the command exits with status 0

#### Scenario: Module path defaults to the project name
- **WHEN** the user runs `gobento new shop --data sqlc` without `--module`
- **THEN** the generated module is named `shop`

### Requirement: Dry run
`gobento new --dry-run` SHALL perform every input validation and composition
step, list the files that would be created, and exit without writing anything,
checking the target directory, or running the Go toolchain. It SHALL exit 0
when the input is valid, and otherwise exit non-zero with the same problem
report as a normal run.

#### Scenario: Valid dry run
- **WHEN** the user runs `gobento new shop --data sqlc --dry-run`
- **THEN** gobento lists the planned files, exits 0, and creates no directory

#### Scenario: Invalid dry run
- **WHEN** the user runs `gobento new "My App" --dry-run`
- **THEN** gobento exits non-zero with the same problems a normal run would report, and creates nothing

### Requirement: Generated project builds and passes its tests
A project generated from any supported selection SHALL compile, pass `go vet`
and `gofmt -l`, and pass its own `go test ./...` without edits by the user. It
SHALL also compile with cgo disabled (`CGO_ENABLED=0`), so it needs no C
compiler on Linux, macOS, or Windows.

#### Scenario: Build check after generation
- **WHEN** a project is generated with the selection mvc/nethttp/sqlite/sqlc or mvc/nethttp/sqlite/gorm
- **THEN** `go build ./...`, `go vet ./...` and `go test ./...` succeed inside it, and `gofmt -l .` lists no files

#### Scenario: Build without cgo
- **WHEN** `CGO_ENABLED=0 go build ./...` runs in a project generated from any supported selection
- **THEN** it succeeds

#### Scenario: Example functionality is present
- **WHEN** the generated project is running
- **THEN** `GET /health` responds 200
- **AND** an item created with `POST /items` is returned by `GET /items`, persisted in the selected database
