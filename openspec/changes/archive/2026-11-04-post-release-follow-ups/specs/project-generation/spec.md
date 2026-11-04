# Spec Delta

## ADDED Requirements

### Requirement: Report the gobento version
`gobento --version` SHALL print the gobento version. A version set at build
time (release builds) SHALL take precedence; otherwise gobento SHALL report the
main module version recorded by the Go toolchain without its leading `v`, and
`dev` when no version was recorded.

#### Scenario: Release binary
- **WHEN** a release binary built with `-X main.version=0.1.0` runs `gobento --version`
- **THEN** it prints `gobento version 0.1.0`

#### Scenario: Installed with go install
- **WHEN** gobento installed with `go install github.com/Tales-of-zuna/gobento/cmd/gobento@vX.Y.Z` runs `gobento --version`
- **THEN** it prints `gobento version X.Y.Z`

#### Scenario: No recorded version
- **WHEN** gobento runs without a build-time version and Go recorded `(devel)` or nothing
- **THEN** it prints `gobento version dev`
