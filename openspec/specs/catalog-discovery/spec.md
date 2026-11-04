# catalog-discovery Specification

## Purpose
Lets users and scripts see, from the CLI alone, what gobento supports: the
components of each kind, the supported selections, and the version rules,
read from the catalog and rules embedded in the binary.

## Requirements

### Requirement: List the components
`gobento list components` SHALL print every component kind in canonical order
with its title, the `gobento new` flag that selects it, and its description,
and under each kind every component in catalog order with its name, title,
description, and the traits it provides and requires. `--kind <kind>` SHALL
restrict the output to one kind; an unknown kind SHALL be an error that names
the known kinds.

#### Scenario: All components
- **WHEN** a user runs `gobento list components`
- **THEN** the output lists each kind with its flag, for example `Database (--db)`, and under it each component of the catalog, for example `postgres` with the title `PostgreSQL` and `provides: conn=database/sql, sql-dialect=postgresql, sql-schema=schema.sql`

#### Scenario: One kind
- **WHEN** a user runs `gobento list components --kind framework`
- **THEN** the output lists only the HTTP framework kind and its components

#### Scenario: Unknown kind
- **WHEN** a user runs `gobento list components --kind db`
- **THEN** gobento exits with status 1 and reports that `db` is not a kind, naming `architecture, framework, database, data, frontend`

### Requirement: List the supported selections
`gobento list selections` SHALL print every supported selection, one per
line with one column per kind, in the order of the catalog's candidates, and
the number of selections. The flags `--arch`, `--framework`, `--db`, `--data`,
and `--frontend` SHALL restrict the output to the selections that contain the
named component; a name that is not a component of the kind SHALL be an error
that names the components of that kind.

#### Scenario: Every supported selection
- **WHEN** a user runs `gobento list selections`
- **THEN** the output has one row per selection that `gobento new` accepts, and no other rows

#### Scenario: Narrowed by component
- **WHEN** a user runs `gobento list selections --data gorm`
- **THEN** every row lists `gorm` as the data-access component, and every supported selection with `gorm` is listed

#### Scenario: Unknown component
- **WHEN** a user runs `gobento list selections --db mysql`
- **THEN** gobento exits with status 1 and reports that `mysql` is not a database, naming the supported ones

### Requirement: Optional kinds in the listing
`gobento list components` SHALL mark an optional kind as optional and SHALL
list `none`, the value that selects no component of the kind, before its
components. `gobento list selections` SHALL show `none` in the column of an
optional kind for a selection without a component of that kind, and
`--frontend none` SHALL restrict the output to the selections without a
frontend.

#### Scenario: Optional kind
- **WHEN** a user runs `gobento list components --kind frontend`
- **THEN** the output starts with `Frontend (--frontend, optional)` and lists `none` before `htmx`

#### Scenario: Without a frontend
- **WHEN** a user runs `gobento list selections --frontend none`
- **THEN** every row shows `none` as the frontend, and every supported selection without a frontend is listed

### Requirement: List the version rules
`gobento list versions` SHALL print the supported Go range and the default Go
version, and every module of the rules file, sorted by path, with its path, role,
range, default, the components that use it, the trait values it is
conditioned on, and its constraints, followed by every vendored asset with its
name, version, the components that use it, and its source.

#### Scenario: Module with a trait condition and a constraint
- **WHEN** a user runs `gobento list versions`
- **THEN** the row of `github.com/libtnb/sqlite` names `data/gorm`, states that it is used only when `sql-dialect=sqlite`, and its constraint that `gorm.io/gorm` must be `>=v1.31.2` when it is `>=v1.2.0` is listed

### Requirement: List the tooling versions
After the vendored assets, `gobento list versions` SHALL print every project
tooling version of the rules file: each container image with its key, name,
tag, and digest if it has one, each GitHub Action with its version and commit
SHA, and each tool with its version.

#### Scenario: Tooling versions
- **WHEN** a user runs `gobento list versions`
- **THEN** the output lists the runtime image with its digest, the Go build image with the tag `{go}`, `actions/checkout` with its version and 40-character commit SHA, and `golangci-lint` with its version

### Requirement: List everything
`gobento list` without a subcommand SHALL print the components, the
supported selections, and the version rules, in that order.

#### Scenario: Overview
- **WHEN** a user runs `gobento list`
- **THEN** the output contains the output of `gobento list components`, `gobento list selections`, and `gobento list versions`

### Requirement: JSON output
With `--json`, every form of `gobento list` SHALL print one JSON document
instead of text. `gobento list --json` SHALL print the same document the
website is built from, with the fields `kinds`, `components`, `go`, `modules`,
`assets`, `tooling`, `supported`, and `rules`. Lists SHALL be empty lists rather than null.

#### Scenario: Same data as the website
- **WHEN** `gobento list --json` runs for a gobento build
- **THEN** its output is byte for byte the catalog document the exporter writes for the website from the same catalog

### Requirement: JSON output of a subcommand
With `--json`, each subcommand SHALL print an object with only the fields it
lists, in the format of the full document: `components` prints `kinds` and
`components`, `selections` prints `supported`, and `versions` prints `go`,
`modules`, `assets`, and `tooling`. Filters SHALL apply to the JSON as to the text.

#### Scenario: Scripts read the components
- **WHEN** a script runs `gobento list components --kind database --json`
- **THEN** it receives an object whose `kinds` holds only the database kind and whose `components` holds the database components with their `id`, `kind`, `name`, `title`, `description`, `provides`, and `requires`

#### Scenario: Scripts read the tooling
- **WHEN** a script runs `gobento list versions --json`
- **THEN** its `tooling` object holds `images`, `actions`, and `tools`, and the `runtime` image's `digest` is the one in the rules file

### Requirement: Listing is derived from the embedded catalog
Everything `gobento list` prints SHALL come from the catalog and version rules
embedded in the binary, with the supported selections computed by the same
check `gobento new` uses. The output SHALL be deterministic: kinds in
canonical order, components and selections in catalog order, modules sorted by
path, and traits sorted by name. `gobento list` SHALL NOT ask
questions, write files, or use the network.

#### Scenario: A component is added
- **WHEN** a component is added to the catalog and gobento is rebuilt
- **THEN** `gobento list` shows it and every new supported selection, without any change to the command's code

#### Scenario: No terminal needed
- **WHEN** `gobento list` runs with stdin and stdout redirected
- **THEN** it prints the same output as in a terminal and asks nothing

### Requirement: Help points at the listing
`gobento --help` and `gobento new --help` SHALL mention `gobento list` as
the way to see the available components, selections, and versions.

#### Scenario: Help for new
- **WHEN** a user runs `gobento new --help`
- **THEN** the help names `gobento list` for the supported components, selections, and versions

### Requirement: List the Node.js rule and npm packages
`gobento list versions` (and `gobento list`) SHALL show the supported
Node.js range and release line, and every npm package with its role,
version, and components; the JSON output SHALL include them.

#### Scenario: npm packages in the listing
- **WHEN** the user runs `gobento list versions`
- **THEN** the output shows the Node.js range and `create-vite` with its version, the role `scaffolder`, and the frontends that use it
