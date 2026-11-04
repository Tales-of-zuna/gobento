# Spec Delta

## MODIFIED Requirements

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
- **THEN** gobento exits with status 1 and reports that `db` is not a kind, naming `architecture, framework, database, data, frontend, css`

### Requirement: List the supported selections
`gobento list selections` SHALL print every supported selection, one per
line with one column per kind, in the order of the catalog's candidates, and
the number of selections. The flags `--arch`, `--framework`, `--db`, `--data`,
`--frontend`, and `--css` SHALL restrict the output to the selections that
contain the named component; a name that is not a component of the kind SHALL
be an error that names the components of that kind.

#### Scenario: Every supported selection
- **WHEN** a user runs `gobento list selections`
- **THEN** the output has one row per selection that `gobento new` accepts, and no other rows

#### Scenario: Narrowed by component
- **WHEN** a user runs `gobento list selections --data gorm`
- **THEN** every row lists `gorm` as the data-access component, and every supported selection with `gorm` is listed

#### Scenario: Unknown component
- **WHEN** a user runs `gobento list selections --db mysql`
- **THEN** gobento exits with status 1 and reports that `mysql` is not a database, naming the supported ones

#### Scenario: Narrowed by CSS framework
- **WHEN** a user runs `gobento list selections --css tailwind`
- **THEN** every row lists a frontend and `tailwind` as the CSS framework

### Requirement: List the version rules
`gobento list versions` SHALL print the supported Go range and the default Go
version, and every module of the rules file, sorted by path, with its path, role,
range, default, the components that use it, the trait values it is
conditioned on, and its constraints, followed by every vendored asset with its
name, version, the components that use it, and its source, and every platform
binary with its name, version, the components that use it, and its platforms.

#### Scenario: Module with a trait condition and a constraint
- **WHEN** a user runs `gobento list versions`
- **THEN** the row of `github.com/libtnb/sqlite` names `data/gorm`, states that it is used only when `sql-dialect=sqlite`, and its constraint that `gorm.io/gorm` must be `>=v1.31.2` when it is `>=v1.2.0` is listed

#### Scenario: Platform binary
- **WHEN** a user runs `gobento list versions`
- **THEN** the output lists `tailwindcss` with its version, `css/tailwind`, and its platforms, such as `linux/amd64`
