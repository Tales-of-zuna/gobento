# Spec Delta

## MODIFIED Requirements

### Requirement: Compatibility rules are data
gobento SHALL read supported versions from a rules file that lists the project's
Go version range and, for each module, the components using it, optionally the
trait values a selection must provide for the module to be used, its role
(library or tool), supported range, default, and constraints on other versions.
The rules file SHALL also list the versions of the project tooling that are not
Go module requirements: container images, GitHub Actions, and tools.
Templates SHALL NOT contain version numbers.

#### Scenario: Updating a supported version
- **WHEN** a maintainer raises a library module's default version in the rules file
- **THEN** newly generated projects use that version without any template change

#### Scenario: Updating a pinned action
- **WHEN** a maintainer changes the version and commit SHA of an action in the rules file
- **THEN** newly generated workflows use that SHA and version without any template change

## ADDED Requirements

### Requirement: Tooling versions are fixed data
The rules file's tooling section SHALL record each container image as a
name, a tag, and optionally a digest; each GitHub Action as a release
version and its full 40-character commit SHA; and each tool as an exact
version. An image tag MAY follow the selected Go version. Tooling versions
SHALL be fixed per gobento release and SHALL NOT be overridable with `--pin`.
gobento SHALL reject a tooling section with a malformed SHA, digest,
version, tag, or name, naming the entry.

#### Scenario: Image tag follows the Go version
- **WHEN** the Go build image's tag follows the Go version and the user passes `--go 1.26`
- **THEN** templates see the build image with the tag `1.26`

#### Scenario: Malformed action SHA
- **WHEN** the rules file lists an action whose SHA is not 40 hexadecimal characters
- **THEN** loading the rules fails with an error naming the action

#### Scenario: Pin for a tooling entry
- **WHEN** the user passes `--pin golangci-lint=<any version>`
- **THEN** gobento exits non-zero because no selected component uses such a module, and writes nothing
