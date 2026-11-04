# Spec Delta

## MODIFIED Requirements

### Requirement: Optional kinds need no flag
When the flag of an optional kind is omitted, gobento SHALL select no
component of that kind without requiring the flag; the value `none` SHALL
select no component explicitly. An unknown value of an optional kind SHALL be
reported with the supported values, including `none`.

#### Scenario: No flag for an optional kind
- **WHEN** the user omits `--frontend` and stdin is not a terminal
- **THEN** the project is generated without a frontend

#### Scenario: No CSS flag with a frontend
- **WHEN** the user passes `--frontend htmx`, omits `--css`, and stdin is not a terminal
- **THEN** the project is generated with htmx and without a CSS framework

#### Scenario: Unknown CSS framework
- **WHEN** the user passes `--css bootstrap`
- **THEN** gobento exits non-zero naming `--css` and listing `none` and `tailwind`, and writes nothing

### Requirement: Ask for missing input in a terminal
In a terminal (stdin and stdout) without `--no-input`, `gobento new` SHALL
ask, before validating, for the project name when no `<name>` is given, then
for the module path (empty keeps the name), and for each kind whose flag is
omitted and that has several compatible options, in kind order; for an
optional kind, "none" is an option, offered first. Answers SHALL follow the
name and module path rules.
Nothing given by a flag SHALL be asked, and `--dry-run` SHALL ask the same
questions.

#### Scenario: Several frameworks, framework flag omitted
- **WHEN** the user runs `gobento new shop` in a terminal and the catalog has several supported frameworks
- **THEN** gobento asks for the framework, and not for the name, the module path, or a kind with a single option
- **AND** it generates the project with the chosen framework

#### Scenario: No name given
- **WHEN** the user runs `gobento new` in a terminal
- **THEN** gobento asks for the project name, rejecting names that break the project name rules, and then for the module path, using the name when the answer is empty

#### Scenario: Flags skip their questions
- **WHEN** the user runs `gobento new shop --framework nethttp` in a terminal
- **THEN** gobento does not ask for the framework

#### Scenario: Frontend question
- **WHEN** the user runs `gobento new shop --arch mvc --framework nethttp --db sqlite --data sqlc` in a terminal
- **THEN** gobento asks for the frontend, offering "None (API only)" first and then `htmx`

#### Scenario: CSS framework question
- **WHEN** the user answers `htmx` to the frontend question
- **THEN** gobento then asks for the CSS framework, offering "None (plain CSS)" first and then `tailwind`

### Requirement: Supported selections are verified in CI
gobento's CI SHALL generate and check projects on Linux, macOS, and Windows,
each with the default versions and with the floor of every range. It SHALL
cover every supported selection without an optional component, and a subset
of the selections with one in which every pair of components that appears
together in such a selection appears together at least once, chosen
deterministically from the catalog.

#### Scenario: Backend combinations are exhaustive
- **WHEN** CI runs the integration test
- **THEN** every supported selection without a frontend is generated and checked

#### Scenario: Frontend combinations are covered pairwise
- **WHEN** CI runs the integration test and the catalog has a frontend
- **THEN** for every architecture, framework, database, and data-access component that can be combined with that frontend, a project with the frontend and that component is checked, and for every two such components that can appear together with the frontend, a project with the frontend and both of them is checked

#### Scenario: CSS frameworks are covered pairwise
- **WHEN** CI runs the integration test and the catalog has a CSS framework
- **THEN** for every component that can be combined with that CSS framework, including each frontend, a project with the CSS framework and that component is checked
