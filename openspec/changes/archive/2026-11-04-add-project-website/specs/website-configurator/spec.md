# Spec Delta

## Purpose

Lets visitors assemble the exact `gobento new` command for the stack they want
by choosing from what gobento actually supports, with compatibility explained,
so nobody has to learn flags or guess versions.

## ADDED Requirements

### Requirement: Options come from the component catalog
The configurator SHALL offer exactly the components, Go versions, and module
version ranges defined in the repository's component catalog and version rules
at build time, with the rules-file defaults preselected.

#### Scenario: New component appears
- **WHEN** a component is added to the catalog and the site is rebuilt
- **THEN** the configurator offers it without any change to website code

#### Scenario: Defaults
- **WHEN** a visitor opens the configurator
- **THEN** each kind with a single option, plus the default Go and module versions, is preselected

### Requirement: Live command output
The configurator SHALL show the `gobento new` command for the current choices,
update it immediately on every change, and omit flags whose value equals what
gobento would choose by default.

#### Scenario: Changing a choice
- **WHEN** the visitor changes the project name to `shop` and the module path to `example.com/shop`
- **THEN** the displayed command becomes `gobento new shop --module example.com/shop` followed by any non-default flags

#### Scenario: Copy command
- **WHEN** the visitor activates the copy control
- **THEN** the exact displayed command is placed on the clipboard and a visible confirmation appears

### Requirement: Input validation matches the CLI
The configurator SHALL apply the same rules as `gobento new` to the project name
and module path, and SHALL show the reason for any invalid value next to the
field instead of producing a command.

#### Scenario: Invalid project name
- **WHEN** the visitor enters `My App` as the project name
- **THEN** the field shows the allowed characters and no copyable command is offered

### Requirement: Incompatible choices are disabled and explained
The configurator SHALL disable options that cannot be combined with the current
selection (unmet traits or violated version constraints) and SHALL show why
each is disabled.

#### Scenario: Version too new for Go
- **WHEN** the selected Go version is older than a module version requires
- **THEN** that module version is disabled and its explanation names the required Go version

### Requirement: Shareable configuration
The configurator's state SHALL be encoded in the page URL, so opening a shared
URL restores the same choices.

#### Scenario: Share a link
- **WHEN** a visitor copies the URL after choosing options and another visitor opens it
- **THEN** the second visitor sees the same choices and the same command

#### Scenario: Unknown value in a shared link
- **WHEN** a shared URL names a component or version the catalog no longer contains
- **THEN** the configurator falls back to the default for that choice and shows a notice

### Requirement: Generated commands are accepted by the CLI
Every command the configurator can produce SHALL pass the validation of
`gobento new` for the same catalog.

#### Scenario: Cross-check in CI
- **WHEN** the website test suite runs against the real catalog
- **THEN** a command is produced for every supported selection, and each one passes `gobento new`'s validation without generating a project
