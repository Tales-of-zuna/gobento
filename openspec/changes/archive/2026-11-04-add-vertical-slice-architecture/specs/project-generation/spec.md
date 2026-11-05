## MODIFIED Requirements

### Requirement: Generate a project from a component selection
`gobento new <name>` SHALL create a Go project from one selected architecture,
HTTP framework, database, and data-access component, using the given Go module
path, or the project name when `--module` is omitted.

#### Scenario: Fully specified by flags
- **WHEN** the user runs `gobento new shop --module example.com/shop --arch mvc --framework nethttp --db sqlite --data sqlc`
- **THEN** a directory `shop` is created containing a Go module named `example.com/shop`
- **AND** the command exits with status 0 and prints the next commands to run

#### Scenario: Module path defaults to the project name
- **WHEN** the user runs `gobento new shop --arch mvc` without `--module`
- **THEN** the generated module is named `shop`

#### Scenario: Architecture must be chosen
- **WHEN** the user runs `gobento new shop` without `--arch` while `mvc` and `vertical-slice` are both supported architectures
- **THEN** gobento exits non-zero naming `--arch` and listing `mvc` and `vertical-slice`, and writes nothing

### Requirement: Generated project builds and passes its tests
A project generated from any supported selection SHALL compile, pass `go vet`
and `gofmt -l`, and pass its own `go test ./...` without edits by the user.

#### Scenario: Build check after generation
- **WHEN** a project is generated with the selection mvc/nethttp/sqlite/sqlc or vertical-slice/nethttp/sqlite/sqlc
- **THEN** `go build ./...`, `go vet ./...` and `go test ./...` succeed inside it, and `gofmt -l .` lists no files

#### Scenario: Example functionality is present
- **WHEN** the generated project is running
- **THEN** `GET /health` responds 200
- **AND** an item created with `POST /items` is returned by `GET /items`, persisted in the selected database
