## ADDED Requirements

### Requirement: Components include contribution points
Besides the base layer's templates, the templates of a selected component
SHALL be able to include the contributions to a declared point, so one
component can extend a file another component owns, such as a CSS component
adding a Vite plugin to a frontend's Vite configuration. Contribution
templates themselves SHALL NOT include points.

#### Scenario: CSS contribution to a Vite frontend
- **WHEN** a selected component contributes to the point `vite-plugins` and a Vite frontend is selected
- **THEN** the contribution appears in the plugins of `frontend/vite.config.ts`, and without a Vite frontend it appears nowhere

### Requirement: Scaffolded components
A component MAY declare in its manifest a scaffolder from the rules file,
the template to pass it, the project directory it creates, and scaffolded
files to remove. Its own root templates under that directory SHALL replace
the scaffolded files of the same path. The catalog check SHALL reject a
scaffold declaration that names an unknown scaffolder or one not listed for
the component, or that has an unsafe directory or path.

#### Scenario: Unknown scaffolder
- **WHEN** a component's manifest names a scaffolder the rules file does not list for it
- **THEN** loading the catalog fails naming the component and the scaffolder
