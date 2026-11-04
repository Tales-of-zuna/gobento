## MODIFIED Requirements

### Requirement: Compatibility rules are data
gobento SHALL read supported versions from a rules file that lists the project's
Go version range and, for each module, the components using it, optionally the
trait values a selection must provide for the module to be used, its role
(library or tool), supported range, default, and constraints on other versions.
The rules file SHALL also list every vendored asset, a file that a component
ships unmodified from an upstream release (such as a JavaScript library),
with the components shipping it, its version, its upstream source, and its
checksum. Templates SHALL NOT contain version numbers.

#### Scenario: Updating a supported version
- **WHEN** a maintainer raises a library module's default version in the rules file
- **THEN** newly generated projects use that version without any template change

#### Scenario: Asset version visible to templates
- **WHEN** a component's README notes refer to the version of an asset it vendors
- **THEN** the notes take the version from the rules file, and a generated project's README names the recorded version

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

## ADDED Requirements

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
