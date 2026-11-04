# Spec Delta

## MODIFIED Requirements

### Requirement: Compatibility rules are data
gobento SHALL read supported versions from a rules file that lists the project's
Go version range and, for each module, the components using it, optionally the
trait values a selection must provide for the module to be used, its role
(library or tool), supported range, default, and constraints on other versions,
every vendored asset (a file shipped unmodified from an upstream release)
with its components, version, source, and checksum, and every platform
binary a generated project downloads on demand. Templates SHALL NOT contain
version numbers.

#### Scenario: Updating a supported version
- **WHEN** a maintainer raises a library module's default version in the rules file
- **THEN** newly generated projects use that version without any template change

#### Scenario: Asset version visible to templates
- **WHEN** a component's README notes refer to the version of an asset it vendors
- **THEN** the notes take the version from the rules file, and a generated project's README names the recorded version

## ADDED Requirements

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
