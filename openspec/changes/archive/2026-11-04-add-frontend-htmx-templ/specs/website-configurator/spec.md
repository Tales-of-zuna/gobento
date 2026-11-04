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

### Requirement: Live command output
The configurator SHALL show the `gobento new` command for the current choices,
update it immediately on every change, and omit flags whose value equals what
gobento would choose by default, including the flag of an optional kind set
to "None".

#### Scenario: Changing a choice
- **WHEN** the visitor changes the project name to `shop` and the module path to `example.com/shop`
- **THEN** the displayed command becomes `gobento new shop --module example.com/shop` followed by any non-default flags

#### Scenario: Copy command
- **WHEN** the visitor activates the copy control
- **THEN** the exact displayed command is placed on the clipboard and a visible confirmation appears

#### Scenario: Choosing a frontend
- **WHEN** the visitor chooses htmx as the frontend
- **THEN** the command gains `--frontend htmx`, and choosing "None (API only)" again removes the flag
