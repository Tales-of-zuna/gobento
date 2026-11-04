# version-compatibility Specification

## Purpose
Keeps generated projects on current, mutually compatible dependency versions by
describing supported versions as data, letting users choose within those bounds,
and rejecting combinations known not to work.

## Requirements

### Requirement: Compatibility rules are data
gobento SHALL read supported versions from a rules file: the project's Go
range and, for each module, its components, optional trait conditions, role
(library or tool), range, default, and constraints, plus every vendored asset
and platform binary with its components, version, source, and checksums.
Templates SHALL NOT contain version numbers.

#### Scenario: Updating a supported version
- **WHEN** a maintainer raises a library module's default version in the rules file
- **THEN** newly generated projects use that version without any template change

#### Scenario: Asset version visible to templates
- **WHEN** a component's README notes refer to the version of an asset it vendors
- **THEN** the notes take the version from the rules file, and a generated project's README names the recorded version

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
range. Tool module versions and vendored asset versions SHALL be fixed and not
overridable.

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

#### Scenario: Override of a vendored asset
- **WHEN** the user passes `--pin htmx.org=<any version>` with the htmx frontend selected
- **THEN** gobento exits non-zero explaining that vendored asset versions are fixed per gobento release and naming the version this release ships

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

### Requirement: Modules conditioned on traits
A rules-file module MAY list trait conditions. Such a module SHALL be used by a
selection only when one of its components is selected and, for every listed
trait, the selection provides one of the listed values. A module that is not
used SHALL NOT appear in the generated `go.mod`, and pinning it SHALL be
rejected, naming the unmet trait condition when one of its components is
selected. Trait names in conditions SHALL be checked against the traits the
catalog's components provide.

#### Scenario: Dialect-specific driver left out
- **WHEN** a project is generated with `--db sqlite --data gorm`
- **THEN** its `go.mod` requires the SQLite dialect for GORM and does not require `gorm.io/driver/postgres`

#### Scenario: Pin for a module whose trait condition is not met
- **WHEN** the user runs `gobento new shop --db sqlite --data gorm --pin gorm.io/driver/postgres=<a version inside its range>`
- **THEN** gobento exits non-zero saying the module is only used with `sql-dialect` `postgresql`, and writes nothing

#### Scenario: Unknown trait in a condition
- **WHEN** the rules file conditions a module on a trait that no component in the catalog provides
- **THEN** loading the catalog fails with an error naming the module and the trait

### Requirement: Vendored assets match their record
gobento's test suite SHALL check that every asset in the rules file is
shipped, byte-for-byte with the recorded checksum, by a component listed for
it, and, with network access, that the recorded upstream source holds exactly
that file. Loading the rules SHALL reject an asset without components,
version, source, or a well-formed checksum, and an asset that names an
unknown component.

#### Scenario: Vendored file edited
- **WHEN** the vendored htmx file differs from the checksum recorded in the rules file
- **THEN** gobento's tests fail naming the asset and the component

#### Scenario: Malformed checksum
- **WHEN** an asset's checksum is not 64 hexadecimal digits
- **THEN** loading the rules fails with an error naming the asset

### Requirement: Tooling versions are fixed data
The rules file SHALL list the project tooling versions that are not Go
modules: each container image as a name, a tag (which MAY follow the selected
Go version), and optionally a digest; each GitHub Action as a release and its
full commit SHA; and each tool as an exact version. They SHALL be fixed per
gobento release, not overridable with `--pin`, and a malformed entry SHALL be
rejected, naming it.

#### Scenario: Image tag follows the Go version
- **WHEN** the Go build image's tag follows the Go version and the user passes `--go 1.26`
- **THEN** templates see the build image with the tag `1.26`

#### Scenario: Malformed action SHA
- **WHEN** the rules file lists an action whose SHA is not 40 hexadecimal characters
- **THEN** loading the rules fails with an error naming the action

#### Scenario: Pin for a tooling entry
- **WHEN** the user passes `--pin golangci-lint=<any version>`
- **THEN** gobento exits non-zero because no selected component uses such a module, and writes nothing

#### Scenario: Updating a pinned action
- **WHEN** a maintainer changes the version and commit SHA of an action in the rules file
- **THEN** newly generated workflows use that SHA and version without any template change

### Requirement: Platform binaries are fixed data
The rules file SHALL list each platform binary, such as the standalone
Tailwind CLI, with its components, an exact version, optionally the URL of
the release's checksum file, and for each supported `GOOS/GOARCH` platform an
https download URL that contains the version and a SHA-256. Binaries SHALL be
fixed per gobento release and not be overridable with `--pin`.

#### Scenario: Binary visible to templates
- **WHEN** a selected component's template refers to a binary of the rules file
- **THEN** it sees the version and every platform's URL and checksum from the rules file

#### Scenario: URL of another version
- **WHEN** a platform's URL does not contain the binary's version
- **THEN** loading the rules fails with an error naming the binary and the platform

#### Scenario: Malformed platform
- **WHEN** a platform key is not of the form `GOOS/GOARCH`, or its checksum is not 64 lowercase hexadecimal digits
- **THEN** loading the rules fails with an error naming the binary and the platform

#### Scenario: Pin for a binary
- **WHEN** the user passes `--pin tailwindcss=4.0.0`
- **THEN** gobento exits non-zero explaining that binary versions are fixed per gobento release and naming the shipped version, and writes nothing

### Requirement: Platform binaries match their release
With network access, gobento's tests SHALL check that every platform
checksum of a binary that names a checksum file equals the checksum the
release publishes for that file.

#### Scenario: Checksum copied wrongly
- **WHEN** a platform's SHA-256 in the rules file differs from the release's checksum file
- **THEN** gobento's tests fail naming the binary and the platform
