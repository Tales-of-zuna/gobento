# Spec Delta

## MODIFIED Requirements

### Requirement: Shipped architectures support frontends
Every shipped architecture SHALL declare the optional `web` slot: `mvc` in
`internal/views`, `vertical-slice` in `internal/web`, and `ddd` in
`internal/adapters/web`, outside every feature directory because pages and
static files are shared by all features.

#### Scenario: MVC with htmx
- **WHEN** the user runs `gobento new shop --module example.com/shop --arch mvc --framework nethttp --db sqlite --data sqlc --frontend htmx`
- **THEN** the project contains the package `views` in `internal/views`, which the entry point in `cmd/shop` mounts at `/`

#### Scenario: Vertical slice with htmx
- **WHEN** the user runs `gobento new shop --module example.com/shop --arch vertical-slice --framework nethttp --db sqlite --data sqlc --frontend htmx`
- **THEN** the project contains the package `web` in `internal/web`, which imports the `items` package and none of its subpackages

#### Scenario: DDD with htmx
- **WHEN** the user runs `gobento new shop --module example.com/shop --arch ddd --framework nethttp --db sqlite --data sqlc --frontend htmx`
- **THEN** the project contains the package `web` in `internal/adapters/web`, which the entry point in `cmd/shop` mounts at `/` with the application service
- **AND** a `POST /items` from the web page reaches the store only through the application service

### Requirement: Domain layers depend inward
In a `ddd` project, the domain package SHALL import no other package of the
project, the application package SHALL import only the domain package, and the
adapter packages, the web adapter included, SHALL not import the application
package; only the entry point SHALL connect the adapters to the application
service. The handlers package SHALL import neither the store nor the database
package.

#### Scenario: Domain stays independent
- **WHEN** a `ddd` project is generated for any supported selection
- **THEN** the domain package imports no other package of the project, directly or transitively, and the project builds

#### Scenario: Handlers stay independent of storage in ddd
- **WHEN** a `ddd` project is generated for any supported selection
- **THEN** the handlers package imports neither the store package nor the database package, directly or transitively

#### Scenario: Web adapter depends on the domain port
- **WHEN** a `ddd` project is generated for any supported selection with a frontend
- **THEN** the web package imports neither the application package nor the store or database package, directly or transitively
