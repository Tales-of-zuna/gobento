# Spec Delta

## ADDED Requirements

### Requirement: Base layer in every selection
gobento SHALL compose a base layer that it owns into every selection, next to
the selected components. The base layer SHALL produce only files at the
project root (and below it, outside every slot directory), SHALL NOT be a
component kind or be selectable, and SHALL be the same for every selection
apart from the template data. A component that produces a root file the base
layer also produces SHALL make the selection fail like any other path
conflict.

#### Scenario: Every selection gets the base layer
- **WHEN** a project is generated from any supported selection
- **THEN** the files of the base layer are part of the project, rendered with the selection's slots, traits, and versions

#### Scenario: Component writes a base layer file
- **WHEN** a selected component has a template for a root file that the base layer produces, such as `Makefile`
- **THEN** gobento exits non-zero naming the file, the base layer, and the component, and writes nothing

### Requirement: Components contribute through declared points
The base layer SHALL declare named contribution points. A component of any
kind MAY contribute to a declared point with one template per point, and a
base layer file SHALL include the selected components' contributions to a
point, in kind order, where it names that point. A contribution SHALL depend
only on traits and the template data, never on which other component is
selected.

#### Scenario: Database contributes a CI service
- **WHEN** a project is generated with a database that contributes to the point for CI service containers
- **THEN** the generated workflow contains that service, and a project generated with another database that does not contribute contains no service

#### Scenario: New kinds contribute without base layer changes
- **WHEN** a component of a kind added later contributes to an existing point, such as the build stage of the Dockerfile
- **THEN** its contribution appears in the generated file without any change to the base layer's templates

### Requirement: Undeclared contribution points are rejected
A catalog in which a component contributes to a point the base layer does not
declare SHALL be rejected, and a base layer template that names an undeclared
point SHALL make composition fail.

#### Scenario: Unknown contribution point
- **WHEN** a component in the catalog has a contribution template for a point the base layer does not declare
- **THEN** loading the catalog fails with an error naming the component and the point, and listing the declared points

#### Scenario: Base template names an unknown point
- **WHEN** a base layer template asks for the contributions to a point that `base.yaml` does not declare
- **THEN** composition fails with an error naming the template and the point, and gobento writes nothing
