# Spec Delta

## Purpose

Lets a developer create a new, ready-to-build Go project from a chosen set of
components with a single `gobento new` command that is safe to run anywhere.

## ADDED Requirements

### Requirement: Generate a project from a component selection
`gobento new <name>` SHALL create a Go project from one selected architecture,
HTTP framework, database, and data-access component, using the given Go module
path, or the project name when `--module` is omitted.

#### Scenario: Fully specified by flags
- **WHEN** the user runs `gobento new shop --module example.com/shop --arch mvc --framework nethttp --db sqlite --data sqlc`
- **THEN** a directory `shop` is created containing a Go module named `example.com/shop`
- **AND** the command exits with status 0 and prints the next commands to run

#### Scenario: Module path defaults to the project name
- **WHEN** the user runs `gobento new shop` without `--module`
- **THEN** the generated module is named `shop`

### Requirement: Automatic selection of the only option
When a selection flag is omitted and its component kind has exactly one
supported value, gobento SHALL select that value. When the kind has several
values, gobento SHALL exit non-zero naming the flag and listing the values.

#### Scenario: Single option is chosen automatically
- **WHEN** the user omits `--db` and `sqlite` is the only supported database
- **THEN** the project is generated with `sqlite`

#### Scenario: Several options require a choice
- **WHEN** the user omits a selection flag whose kind has several supported values
- **THEN** gobento exits non-zero naming the flag and its supported values, and writes nothing

### Requirement: Project name and module path rules
The project name SHALL start with a lowercase letter and contain only lowercase
letters, digits, `-`, and `_`, and SHALL NOT be a device name that Windows
reserves (`con`, `prn`, `aux`, `nul`, `com1`–`com9`, `lpt1`–`lpt9`). The module path SHALL be a valid Go import path,
as accepted by `go mod init`.

#### Scenario: Name reserved on Windows
- **WHEN** the user runs `gobento new con`
- **THEN** gobento exits non-zero explaining that the name is reserved on Windows, and writes nothing

#### Scenario: Invalid project name
- **WHEN** the user runs `gobento new "My App"`
- **THEN** gobento exits non-zero explaining the allowed characters, and writes nothing

#### Scenario: Invalid module path
- **WHEN** the user passes `--module "Not A Path"`
- **THEN** gobento exits non-zero explaining the module path is invalid, and writes nothing

### Requirement: Report every input problem at once
gobento SHALL validate the name, module path, selections, and versions before
creating any file, and SHALL report all problems found in one run with a
non-zero exit.

#### Scenario: Several problems
- **WHEN** the user passes both an invalid module path and `--framework express`
- **THEN** the output names both problems, lists the supported frameworks, and nothing is written

### Requirement: Dry run
`gobento new --dry-run` SHALL perform every input validation and composition
step, list the files that would be created, and exit without writing anything,
checking the target directory, or running the Go toolchain. It SHALL exit 0
when the input is valid, and otherwise exit non-zero with the same problem
report as a normal run.

#### Scenario: Valid dry run
- **WHEN** the user runs `gobento new shop --dry-run`
- **THEN** gobento lists the planned files, exits 0, and creates no directory

#### Scenario: Invalid dry run
- **WHEN** the user runs `gobento new "My App" --dry-run`
- **THEN** gobento exits non-zero with the same problems a normal run would report, and creates nothing

### Requirement: Output location
The project SHALL be generated into `--dir` when given, otherwise into `./<name>`.
gobento SHALL refuse a target that exists and is not empty, and SHALL fill a
target that exists and is empty, including the current directory.

#### Scenario: Target directory is not empty
- **WHEN** the target directory exists and contains files
- **THEN** gobento exits non-zero and leaves the directory unchanged

#### Scenario: Generate into the current empty directory
- **WHEN** the user runs `gobento new shop --dir .` inside an empty directory
- **THEN** the project files are created directly in that directory

### Requirement: Go toolchain preflight
Before writing, gobento SHALL check that the `go` command is available. If the
selected Go version is newer than the local toolchain, it SHALL fail when
automatic toolchain download is disabled, and otherwise warn that one will be
downloaded.

#### Scenario: Go is not installed
- **WHEN** no `go` command is found on PATH
- **THEN** gobento exits non-zero with instructions to install Go, and writes nothing

#### Scenario: Selected Go is newer and downloads are disabled
- **WHEN** the selected Go version is newer than the local toolchain and `GOTOOLCHAIN=local`
- **THEN** gobento exits non-zero naming both versions, and writes nothing

### Requirement: Generated project builds and passes its tests
A project generated from any supported selection SHALL compile, pass `go vet`
and `gofmt -l`, and pass its own `go test ./...` without edits by the user.

#### Scenario: Build check after generation
- **WHEN** a project is generated with the selection mvc/nethttp/sqlite/sqlc
- **THEN** `go build ./...`, `go vet ./...` and `go test ./...` succeed inside it, and `gofmt -l .` lists no files

#### Scenario: Example functionality is present
- **WHEN** the generated project is running
- **THEN** `GET /health` responds 200
- **AND** an item created with `POST /items` is returned by `GET /items`, persisted in the selected database

### Requirement: No partial projects
If generation fails or is interrupted after it began writing, gobento SHALL
leave no generated files behind, SHALL leave a pre-existing target directory as
it was, and SHALL report the cause.

#### Scenario: Dependency fetch fails
- **WHEN** fetching a selected module version fails (for example, no network)
- **THEN** gobento exits non-zero with the underlying error
- **AND** no generated files remain on disk

#### Scenario: User interrupts generation
- **WHEN** the user presses Ctrl-C while dependencies are being fetched
- **THEN** gobento stops, removes everything it created, reports that it was interrupted, and exits with status 130
