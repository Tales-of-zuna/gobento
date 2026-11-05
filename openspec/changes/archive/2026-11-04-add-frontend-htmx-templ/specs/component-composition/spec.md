## MODIFIED Requirements

### Requirement: Component kinds
gobento SHALL organize components into five kinds: architecture, framework,
database, data access, and frontend. The frontend kind is optional; the
others are required. A selection SHALL contain exactly one component of each
required kind and at most one component of each optional kind. No component
of an optional kind SHALL be named `none`, which stands for selecting no
component of that kind.

#### Scenario: One component per kind
- **WHEN** a selection is resolved
- **THEN** it contains exactly one architecture, one framework, one database, and one data-access component, and either one frontend or none

#### Scenario: Component named none
- **WHEN** the catalog contains a frontend component named `none`
- **THEN** loading the catalog fails with an error naming the component and explaining that `none` is reserved

### Requirement: Supported combinations
A selection SHALL be supported exactly when it has one component per required
kind and at most one per optional kind, all required traits are met, and it
passes composition. Leaving out an optional kind is a separate selection,
supported under the same conditions. The set of supported selections SHALL be
derivable from the component catalog alone.

#### Scenario: Trait mismatch is not supported
- **WHEN** the catalog contains a data-access component whose required dialect no database provides
- **THEN** no selection containing it is listed as supported

#### Scenario: Optional kind left out
- **WHEN** the catalog contains a frontend component
- **THEN** the supported selections include, for every supported combination of the required kinds, the same combination without a frontend

## ADDED Requirements

### Requirement: Templates see slot state and the API base path
Templates SHALL be able to test whether a slot declared by the architecture is
filled by a selected component, and SHALL see the path prefix under which the
API router is served: empty without a frontend and `/api` with one. A
template of an optional slot's neighbors SHALL use this instead of naming a
particular component.

#### Scenario: Entry point adapts to a filled slot
- **WHEN** a selection fills the architecture's optional `web` slot
- **THEN** the architecture's templates see that slot as filled and the API base path as `/api`, and render the entry point that mounts both handlers

#### Scenario: Optional slot left empty
- **WHEN** a selection leaves the optional `web` slot unfilled
- **THEN** the architecture's templates see the slot as not filled and the API base path as empty, and no file is written into the slot's directory
