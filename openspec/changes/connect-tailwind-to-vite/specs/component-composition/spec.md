## MODIFIED Requirements

### Requirement: Components contribute through declared points
The base layer, and any component kind, MAY declare named contribution
points, each point once. A component of any kind MAY contribute to a
declared point with one template per point. The templates of the base layer
and of every selected component MAY include the base layer's points; the
templates of a component MAY also include the points of its own kind. Where
a template names a point, it SHALL include the selected components'
contributions to it, in kind order. A contribution SHALL depend only on
traits and the template data, never on which other component is selected.

#### Scenario: Database contributes a CI service
- **WHEN** a project is generated with a database that contributes to the point for CI service containers
- **THEN** the generated workflow contains that service, and a project generated with another database that does not contribute contains no service

#### Scenario: New kinds contribute without base layer changes
- **WHEN** a component of a kind added later contributes to an existing point, such as the build stage of the Dockerfile
- **THEN** its contribution appears in the generated file without any change to the base layer's templates

#### Scenario: Frontend point filled by a CSS framework
- **WHEN** the frontend kind declares a point for the stylesheet, the selected frontend includes it, and the selected CSS framework contributes to it
- **THEN** the frontend's stylesheet is the contribution, and without a contribution it is the frontend's own

#### Scenario: Component template includes a base point
- **WHEN** an architecture's `.gitignore` template includes the base layer's point `gitignore` and a selected component contributes to it
- **THEN** the generated `.gitignore` contains the contribution

### Requirement: Undeclared contribution points are rejected
A catalog in which a component contributes to a point that neither the base
layer nor any kind declares SHALL be rejected, as SHALL a kind that declares
a point of the base layer or of another kind. A template that names a point
that neither the base layer nor its own kind declares, a non-empty
contribution to a point of a kind without a selected component, and a
non-empty contribution that no rendered template includes SHALL make
composition fail.

#### Scenario: Unknown contribution point
- **WHEN** a component in the catalog has a contribution template for a point that neither the base layer nor any kind declares
- **THEN** loading the catalog fails with an error naming the component and the point, and listing the declared points

#### Scenario: Base template names an unknown point
- **WHEN** a base layer template asks for the contributions to a point that `base.yaml` does not declare
- **THEN** composition fails with an error naming the template and the point, and gobento writes nothing

#### Scenario: Component template names a point of another kind
- **WHEN** a database component's template asks for the contributions to a point that only the data kind declares
- **THEN** composition fails with an error naming the template, the component, and the point

#### Scenario: Point of an unselected kind
- **WHEN** a selected component's contribution renders non-empty for a point of a kind of which no component is selected
- **THEN** composition fails naming the contribution and the point, and a contribution that renders empty for the selection is ignored

#### Scenario: Contribution that no template includes
- **WHEN** a selected component's contribution renders non-empty for a point of the selected frontend's kind, but no template of that frontend includes the point
- **THEN** composition fails naming the contribution, the point, and the frontend, so the selection is not supported

### Requirement: Components include contribution points
Besides the base layer's templates, the templates of a selected component
SHALL be able to include the contributions to a point of the base layer or
of the component's own kind, so one component can extend a file another
component owns, such as a CSS component adding a Vite plugin to a
frontend's Vite configuration through a point of the frontend kind.
Contribution templates themselves SHALL NOT include points.

#### Scenario: CSS contribution to a Vite frontend
- **WHEN** a selected component contributes to the frontend kind's point `vite-plugins` and a Vite frontend is selected
- **THEN** the contribution appears in the plugins of `frontend/vite.config.ts`, and without a Vite frontend it appears nowhere
