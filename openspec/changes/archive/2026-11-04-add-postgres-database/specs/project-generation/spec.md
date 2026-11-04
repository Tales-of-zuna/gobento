## ADDED Requirements

### Requirement: PostgreSQL database
gobento SHALL offer PostgreSQL as a database (`--db postgres`), using a
pure-Go driver so that the generated project builds without a C compiler on
Linux, macOS, and Windows. The generated README SHALL explain how to start a
local PostgreSQL server, how to point the application and its tests at a
server, and how the schema is applied.

#### Scenario: PostgreSQL project
- **WHEN** the user runs `gobento new shop --module example.com/shop --db postgres --data sqlc`
- **THEN** a project is created whose `go.mod` requires the PostgreSQL driver at its default version and no SQLite driver
- **AND** its README contains a command that starts a PostgreSQL server matching the default connection string

#### Scenario: Database is a choice
- **WHEN** the user runs `gobento new shop` without `--db`
- **THEN** gobento exits non-zero naming `--db` and listing `postgres` and `sqlite`, and writes nothing

## MODIFIED Requirements

### Requirement: Generated project builds and passes its tests
A project generated from any supported selection SHALL compile, pass `go vet`
and `gofmt -l`, and pass its own `go test ./...` without edits by the user.

#### Scenario: Build check after generation
- **WHEN** a project is generated with the selection mvc/nethttp/sqlite/sqlc
- **THEN** `go build ./...`, `go vet ./...` and `go test ./...` succeed inside it, and `gofmt -l .` lists no files

#### Scenario: Build check without a database server
- **WHEN** a project is generated with the selection mvc/nethttp/postgres/sqlc on a machine without a PostgreSQL server and `TEST_DATABASE_URL` is unset
- **THEN** `go build ./...`, `go vet ./...` and `go test ./...` succeed inside it, and `gofmt -l .` lists no files

#### Scenario: Database tests run against a server
- **WHEN** a project is generated with the selection mvc/nethttp/postgres/sqlc and `TEST_DATABASE_URL` names a reachable PostgreSQL server
- **THEN** `go test ./...` succeeds and its database and store tests run against that server instead of being skipped

#### Scenario: Example functionality is present
- **WHEN** the generated project is running
- **THEN** `GET /health` responds 200
- **AND** an item created with `POST /items` is returned by `GET /items`, persisted in the selected database
