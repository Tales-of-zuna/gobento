# Spec Delta

## MODIFIED Requirements

### Requirement: Options come from the component catalog
The configurator SHALL offer exactly the components, Go versions, and module
version ranges defined in the repository's component catalog and version rules
at build time, with the rules-file defaults preselected. For an optional kind
it SHALL also offer "None", preselected, which selects no component of that
kind.

#### Scenario: New component appears
- **WHEN** a component is added to the catalog and the site is rebuilt
- **THEN** the configurator offers it without any change to website code

#### Scenario: Defaults
- **WHEN** a visitor opens the configurator
- **THEN** each kind with a single option, plus the default Go and module versions, is preselected

#### Scenario: Frontend is optional
- **WHEN** a visitor opens the configurator
- **THEN** the frontend choice offers "None (API only)", preselected, and every frontend in the catalog

#### Scenario: CSS framework needs a frontend
- **WHEN** no frontend is selected
- **THEN** the CSS framework choice keeps "None (plain CSS)" selected and disables every CSS framework, explaining that it needs a frontend that provides `frontend-build`, and the command has no `--css` flag

#### Scenario: CSS framework with a frontend
- **WHEN** the visitor selects `htmx` and then Tailwind CSS
- **THEN** the command contains `--frontend htmx --css tailwind`
