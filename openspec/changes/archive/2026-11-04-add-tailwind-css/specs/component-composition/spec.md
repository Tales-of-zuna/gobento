# Spec Delta

## MODIFIED Requirements

### Requirement: Component kinds
gobento SHALL organize components into six kinds: architecture, framework,
database, data access, frontend, and CSS framework. The frontend and CSS
framework kinds are optional; the others are required. A selection SHALL
contain exactly one component of each required kind and at most one
component of each optional kind. No component of an optional kind SHALL be
named `none`, which stands for selecting no component of that kind.

#### Scenario: One component per kind
- **WHEN** a selection is resolved
- **THEN** it contains exactly one architecture, one framework, one database, and one data-access component, either one frontend or none, and either one CSS framework or none

#### Scenario: Component named none
- **WHEN** the catalog contains a frontend or CSS framework component named `none`
- **THEN** loading the catalog fails with an error naming the component and explaining that `none` is reserved

### Requirement: Components contribute through declared points
The base layer SHALL declare named contribution points, and a component kind
MAY declare points for its components' templates. A component of any kind
MAY contribute to a declared point with one template per point, and a
template of the base layer or of a component of the declaring kind SHALL
include the selected components' contributions to a point, in kind order,
where it names that point. A contribution SHALL depend only on traits and the
template data, never on which other component is selected.

#### Scenario: Database contributes a CI service
- **WHEN** a project is generated with a database that contributes to the point for CI service containers
- **THEN** the generated workflow contains that service, and a project generated with another database that does not contribute contains no service

#### Scenario: New kinds contribute without base layer changes
- **WHEN** a component of a kind added later contributes to an existing point, such as the build stage of the Dockerfile
- **THEN** its contribution appears in the generated file without any change to the base layer's templates

#### Scenario: Frontend point filled by a CSS framework
- **WHEN** the frontend kind declares a point for the stylesheet, the selected frontend includes it, and the selected CSS framework contributes to it
- **THEN** the frontend's stylesheet is the contribution, and without a contribution it is the frontend's own

### Requirement: Undeclared contribution points are rejected
A catalog in which a component contributes to a point that neither the base
layer nor any kind declares SHALL be rejected, as SHALL a kind that declares
a point of the base layer or of another kind. A template that names a point
its base layer or kind does not declare, a non-empty contribution to a point
of a kind without a selected component, and a non-empty contribution that no
rendered template includes SHALL make composition fail.

#### Scenario: Unknown contribution point
- **WHEN** a component in the catalog has a contribution template for a point that neither the base layer nor any kind declares
- **THEN** loading the catalog fails with an error naming the component and the point, and listing the declared points

#### Scenario: Base template names an unknown point
- **WHEN** a base layer template asks for the contributions to a point that `base.yaml` does not declare
- **THEN** composition fails with an error naming the template and the point, and gobento writes nothing

#### Scenario: Point of an unselected kind
- **WHEN** a selected component's contribution renders non-empty for a point of a kind of which no component is selected
- **THEN** composition fails naming the contribution and the point, and a contribution that renders empty for the selection is ignored

#### Scenario: Contribution that no template includes
- **WHEN** a selected component's contribution renders non-empty for a point of the selected frontend's kind, but no template of that frontend includes the point
- **THEN** composition fails naming the contribution, the point, and the frontend, so the selection is not supported

## ADDED Requirements

### Requirement: Templates that render empty produce no file
A file template whose output for a selection is empty or whitespace only
SHALL produce no file, so a component MAY leave out files based on the
selection's traits.

#### Scenario: File for one build system only
- **WHEN** a component's root file template renders only when the trait `frontend-build` is `go` and the selection provides `vite`
- **THEN** the planned files and the generated project contain no such file
