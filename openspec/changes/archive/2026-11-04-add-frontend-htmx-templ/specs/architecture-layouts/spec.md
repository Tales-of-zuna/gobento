## MODIFIED Requirements

### Requirement: Generated README explains the layout
The README of a project generated with an architecture SHALL describe the
directories of that architecture and what each one contains, including the
`web` directory when a frontend is selected, and SHALL show the requests to
try at the API's base path. For `vertical-slice` it SHALL also explain how to
add another feature.

#### Scenario: Vertical slice README
- **WHEN** a project is generated with `--arch vertical-slice`
- **THEN** its `README.md` lists `internal/items`, `internal/items/itemshttp`, `internal/items/itemsstore`, and `internal/platform/database` with their contents
- **AND** it describes the steps to add a second feature next to `internal/items`

#### Scenario: README with a frontend
- **WHEN** a project is generated with `--frontend htmx`
- **THEN** its `README.md` lists the web directory, tells the reader to open `http://localhost:8080/`, and shows `curl` requests under `/api`

## ADDED Requirements

### Requirement: Shipped architectures support frontends
Both shipped architectures SHALL declare the optional `web` slot: `mvc` in
`internal/views` and `vertical-slice` in `internal/web`, outside every
feature directory because pages and static files are shared by all features.

#### Scenario: MVC with htmx
- **WHEN** the user runs `gobento new shop --module example.com/shop --arch mvc --framework nethttp --db sqlite --data sqlc --frontend htmx`
- **THEN** the project contains the package `views` in `internal/views`, which the entry point in `cmd/shop` mounts at `/`

#### Scenario: Vertical slice with htmx
- **WHEN** the user runs `gobento new shop --module example.com/shop --arch vertical-slice --framework nethttp --db sqlite --data sqlc --frontend htmx`
- **THEN** the project contains the package `web` in `internal/web`, which imports the `items` package and none of its subpackages
