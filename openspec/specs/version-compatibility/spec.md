# version-compatibility Specification

## Purpose
Keeps generated projects on current, mutually compatible dependency versions by
describing supported versions as data, letting users choose within those bounds,
and rejecting combinations known not to work.

## Requirements

### Requirement: Compatibility rules are data
gobento SHALL read supported versions from a rules file that lists the project's
Go version range and, for each module, the components using it, its role
(library or tool), supported range, default, and constraints on other versions.
Templates SHALL NOT contain version numbers.

#### Scenario: Updating a supported version
- **WHEN** a maintainer raises a library module's default version in the rules file
- **THEN** newly generated projects use that version without any template change

### Requirement: Rules file is checked for consistency
gobento SHALL reject a rules file in which a default lies outside its own range,
a constraint or component reference is unknown, or a range is malformed. Its
test suite SHALL check the shipped rules file.

#### Scenario: Default outside range
- **WHEN** the rules file declares default `v1.5.0` for a range `>=v2.0.0`
- **THEN** loading the rules fails with an error naming the module

### Requirement: Default versions
When the user does not choose a version, gobento SHALL use the rules-file default
for the Go version and for every library module the selection uses.

#### Scenario: No version flags
- **WHEN** a project is generated without version overrides
- **THEN** its `go.mod` requires each library module at exactly its default, and its `go` directive is the default Go version

### Requirement: User version overrides
gobento SHALL let the user choose the Go version (`--go`) and library module
versions (`--pin module=version`), accepting only versions inside the supported
range. Tool module versions SHALL be fixed and not overridable.

#### Scenario: Override within range
- **WHEN** the user passes `--go 1.26` and `--pin modernc.org/sqlite=<a version inside its range>`
- **THEN** the generated `go.mod` uses Go 1.26 and that exact sqlite version

#### Scenario: Override outside range
- **WHEN** the user pins a module version outside its supported range
- **THEN** gobento exits non-zero showing the supported range, and writes nothing

#### Scenario: Override for a module not in the selection
- **WHEN** the user pins a module that no selected component uses
- **THEN** gobento exits non-zero naming the module, and writes nothing

#### Scenario: Override of a tool version
- **WHEN** the user pins a tool module such as sqlc
- **THEN** gobento exits non-zero explaining that tool versions are fixed per gobento release

### Requirement: Reject incompatible combinations
gobento SHALL evaluate every constraint across the final set of chosen versions
and SHALL refuse the generation if any constraint is violated, naming the
modules and versions in conflict.

#### Scenario: Go version too old for a module
- **WHEN** the chosen module version requires a newer Go version than the one selected
- **THEN** gobento exits non-zero naming the module, its required Go version, and the selected Go version, and writes nothing

### Requirement: Chosen versions are exactly what the project gets
The generated `go.mod` SHALL contain exactly the chosen Go version and library
versions. If dependency resolution would raise any of them, gobento SHALL fail
and report it as a missing constraint in the rules file.

#### Scenario: Resolution raises a version
- **WHEN** resolving dependencies raises a chosen module version or the `go` directive
- **THEN** gobento exits non-zero naming the module, the chosen version, and the raised version, and leaves no files behind

#### Scenario: Reproducible dependencies
- **WHEN** a project is generated
- **THEN** every direct dependency in its `go.mod` is an exact version and `go mod verify` succeeds
