# architecture-layouts Specification

## Purpose
Defines the project layouts gobento ships as architecture components: where each
slot lives, how the generated packages may depend on each other, and how the
generated README explains the layout.

## Requirements

### Requirement: Vertical slice architecture
gobento SHALL ship an architecture `vertical-slice` that organizes a project by
feature. The example `items` feature SHALL own one directory holding its types
and store contract, with its HTTP handlers and its data access in subpackages
of that directory. Code shared by all features, such as the database
connection, SHALL live outside the feature directory.

#### Scenario: Feature layout
- **WHEN** the user runs `gobento new shop --module example.com/shop --arch vertical-slice --framework nethttp --db sqlite --data sqlc`
- **THEN** the project contains `cmd/shop/main.go`, the package `items` in `internal/items`, the handlers package `itemshttp` in `internal/items/itemshttp`, the store package `itemsstore` in `internal/items/itemsstore`, and the database package in `internal/platform/database`
- **AND** it contains no `internal/models`, `internal/controllers`, or `internal/store` directory

#### Scenario: Shipped components are reused unchanged
- **WHEN** a framework, database, or data-access component is selected with `vertical-slice`
- **THEN** gobento uses the same component templates it uses with `mvc`, placed in the `vertical-slice` directories with the matching package names and import paths

### Requirement: Feature packages depend inward
In a `vertical-slice` project, the feature's handlers and store packages SHALL
depend on the feature's contract package and never on each other, and the
feature's contract package SHALL import none of its subpackages, so the
packages form no import cycle.

#### Scenario: Handlers stay independent of storage
- **WHEN** a `vertical-slice` project is generated for any supported selection
- **THEN** the handlers package imports neither the store package nor the database package, directly or transitively, and the project builds

### Requirement: Generated README explains the layout
The README of a project generated with an architecture SHALL describe the
directories of that architecture and what each one contains, including the
`web` directory when a frontend is selected, and SHALL show the requests to
try at the API's base path. For `vertical-slice` it SHALL also explain how to
add another feature. For `ddd` it SHALL also explain the direction of
dependencies between the layers and how to add another aggregate.

#### Scenario: Vertical slice README
- **WHEN** a project is generated with `--arch vertical-slice`
- **THEN** its `README.md` lists `internal/items`, `internal/items/itemshttp`, `internal/items/itemsstore`, and `internal/platform/database` with their contents
- **AND** it describes the steps to add a second feature next to `internal/items`

#### Scenario: DDD README
- **WHEN** a project is generated with `--arch ddd`
- **THEN** its `README.md` lists `internal/domain`, `internal/app`, `internal/adapters/httpapi`, `internal/adapters/persistence`, and `internal/adapters/database` with their contents
- **AND** it states that the adapters and the application layer depend on the domain, and the domain depends on no other layer
- **AND** it describes the steps to add a second aggregate next to `Item`

#### Scenario: README with a frontend
- **WHEN** a project is generated with `--frontend htmx`
- **THEN** its `README.md` lists the web directory, tells the reader to open `http://localhost:8080/`, and shows `curl` requests under `/api`

### Requirement: Domain-driven design architecture
gobento SHALL ship an architecture `ddd` with a domain layer (the example
`Item` entity, its rules, and the store port), an application layer (a
service that runs the item use cases on the domain), and adapter packages for
the HTTP handlers, the data access, and the database connection. It SHALL use
the same framework, database, and data-access components as every other
architecture.

#### Scenario: Layered layout
- **WHEN** the user runs `gobento new shop --module example.com/shop --arch ddd --framework nethttp --db sqlite --data sqlc`
- **THEN** the project contains `cmd/shop/main.go`, the package `domain` in `internal/domain`, the package `app` in `internal/app`, the handlers package `httpapi` in `internal/adapters/httpapi`, the store package `persistence` in `internal/adapters/persistence`, and the database package in `internal/adapters/database`
- **AND** it contains no `internal/models`, `internal/controllers`, `internal/store`, or `internal/items` directory

#### Scenario: Requests go through the application service
- **WHEN** a `ddd` project is running and receives `POST /items`
- **THEN** the handlers reach the store only through the application service, which checks the domain rule for item names before storing the item

#### Scenario: Shipped components are reused unchanged in ddd
- **WHEN** a framework, database, or data-access component is selected with `ddd`
- **THEN** gobento uses the same component templates it uses with `mvc`, placed in the `ddd` directories with the matching package names and import paths

### Requirement: Domain layers depend inward
In a `ddd` project, the domain package SHALL import no other package of the
project, the application package SHALL import only the domain package, and the
adapter packages SHALL not import the application package; only the entry
point SHALL connect the adapters to the application service. The handlers
package SHALL import neither the store nor the database package.

#### Scenario: Domain stays independent
- **WHEN** a `ddd` project is generated for any supported selection
- **THEN** the domain package imports no other package of the project, directly or transitively, and the project builds

#### Scenario: Handlers stay independent of storage in ddd
- **WHEN** a `ddd` project is generated for any supported selection
- **THEN** the handlers package imports neither the store package nor the database package, directly or transitively

### Requirement: Shipped architectures support frontends
The `mvc` and `vertical-slice` architectures SHALL declare the optional `web`
slot: `mvc` in `internal/views` and `vertical-slice` in `internal/web`,
outside every feature directory because pages and static files are shared by
all features.

#### Scenario: MVC with htmx
- **WHEN** the user runs `gobento new shop --module example.com/shop --arch mvc --framework nethttp --db sqlite --data sqlc --frontend htmx`
- **THEN** the project contains the package `views` in `internal/views`, which the entry point in `cmd/shop` mounts at `/`

#### Scenario: Vertical slice with htmx
- **WHEN** the user runs `gobento new shop --module example.com/shop --arch vertical-slice --framework nethttp --db sqlite --data sqlc --frontend htmx`
- **THEN** the project contains the package `web` in `internal/web`, which imports the `items` package and none of its subpackages
