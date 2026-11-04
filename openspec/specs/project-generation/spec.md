# project-generation Specification

## Purpose
Lets a developer create a new, ready-to-build Go project from a chosen set of
components with a single `gobento new` command that is safe to run anywhere.

## Requirements

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

### Requirement: Automatic selection of the only option
When a selection flag is omitted and its component kind has exactly one
supported value, gobento SHALL select that value. When the kind has several
values and gobento does not ask questions (see "Ask for missing input in a
terminal"), gobento SHALL exit non-zero naming the flag and listing the values.

#### Scenario: Single option is chosen automatically
- **WHEN** the user omits `--db` and `sqlite` is the only supported database
- **THEN** the project is generated with `sqlite`

#### Scenario: Several options require a choice
- **WHEN** the user omits a selection flag whose kind has several supported values, and stdin is not a terminal or `--no-input` is given
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
- **WHEN** a project is generated with the selection mvc/nethttp/sqlite/sqlc or vertical-slice/nethttp/sqlite/sqlc
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

### Requirement: Report the gobento version
`gobento --version` SHALL print the gobento version. A version set at build
time (release builds) SHALL take precedence; otherwise gobento SHALL report the
main module version recorded by the Go toolchain without its leading `v`, and
`dev` when no version was recorded.

#### Scenario: Release binary
- **WHEN** a release binary built with `-X main.version=0.1.0` runs `gobento --version`
- **THEN** it prints `gobento version 0.1.0`

#### Scenario: Installed with go install
- **WHEN** gobento installed with `go install github.com/Tales-of-zuna/gobento/cmd/gobento@vX.Y.Z` runs `gobento --version`
- **THEN** it prints `gobento version X.Y.Z`

#### Scenario: No recorded version
- **WHEN** gobento runs without a build-time version and Go recorded `(devel)` or nothing
- **THEN** it prints `gobento version dev`

### Requirement: Ask for missing input in a terminal
In a terminal (stdin and stdout) without `--no-input`, `gobento new` SHALL
ask, before validating, for the project name when no `<name>` is given, then
for the module path (empty keeps the name), and for each kind whose flag is
omitted and that has several compatible options, in kind order. Answers SHALL
follow the name and module path rules. Nothing given by a flag SHALL be asked,
and `--dry-run` SHALL ask the same questions.

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

### Requirement: Questions offer only supported options
A component question SHALL offer only the options that form a supported
selection together with the component flags and the earlier answers. A kind
left with a single compatible option SHALL be selected without a question.

#### Scenario: Later options depend on earlier answers
- **WHEN** an earlier answer or flag rules out some options of a later kind
- **THEN** the later question offers only the options that form a supported selection with it, and is skipped when one option remains

### Requirement: No questions for invalid command-line input
If a name, module path, `--pin` value, or component flag given on the command
line is invalid, or the component flags match no supported selection, gobento
SHALL ask nothing and report the problems as it does without a terminal.

#### Scenario: Invalid flag value
- **WHEN** the user runs `gobento new shop --framework express` in a terminal
- **THEN** gobento asks nothing, exits non-zero listing the supported frameworks, and writes nothing

### Requirement: Equivalent command after questions
After asking at least one question, gobento SHALL print on stderr the
equivalent `gobento new` command, with every answer as an argument or flag
and the other flags of the run, so the run can be repeated without
questions.

#### Scenario: Framework chosen by a question
- **WHEN** the user answers `gin` to the framework question of `gobento new shop --dry-run`
- **THEN** stderr shows a command that starts with `gobento new shop` and contains `--framework gin` and `--dry-run`

### Requirement: Disable questions
gobento SHALL NOT ask any question when `--no-input` is given or when stdin or
stdout is not a terminal; it then behaves as if questions did not exist: a
missing `<name>` is an argument error, and a kind with several options and no
flag is reported as in "Automatic selection of the only option".

#### Scenario: Piped input
- **WHEN** the user runs `echo | gobento new` in a shell
- **THEN** gobento exits non-zero asking for exactly one project name, without asking a question

#### Scenario: No input flag in a terminal
- **WHEN** the user runs `gobento new shop --no-input` in a terminal and a kind has several options
- **THEN** gobento exits non-zero naming the kind's flag and its supported values, without asking a question

### Requirement: Abort while asking
When the user cancels a question with Ctrl-C or Esc, or the process is
interrupted while a question is open, gobento SHALL stop, write nothing, print
`gobento: interrupted; nothing was left behind`, and exit with status 130.
When the input ends before every question is answered, gobento SHALL exit
non-zero and write nothing.

#### Scenario: Ctrl-C at a question
- **WHEN** the user presses Ctrl-C while gobento asks for the framework
- **THEN** gobento exits with status 130, reports that it was interrupted, and creates no directory

#### Scenario: Input ends
- **WHEN** the input ends (for example, Ctrl-D in accessible mode) before the last question is answered
- **THEN** gobento exits non-zero saying the input ended, and creates no directory

### Requirement: Accessible questions
When the `ACCESSIBLE` environment variable is set to a non-empty value, or
`TERM` is `dumb`, gobento SHALL ask its questions as plain, line-based text
prompts, listing the options of a choice as numbered lines, instead of a
full-screen interface. The questions, options, validation, and results SHALL
be the same in both modes.

#### Scenario: Screen reader user
- **WHEN** the user runs `ACCESSIBLE=1 gobento new` in a terminal
- **THEN** each question is printed as a line of text, each option of a choice as a numbered line, and the answer is read as a typed line

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
- **WHEN** the user runs `gobento new shop` without `--db` and without a terminal
- **THEN** gobento exits non-zero naming `--db` and listing `postgres` and `sqlite`, and writes nothing
