## MODIFIED Requirements

### Requirement: Compatibility rules are data
gobento SHALL read supported versions from a rules file that lists the project's
Go version range and, for each module, the components using it, optionally the
trait values a selection must provide for the module to be used, its role
(library or tool), supported range, default, and constraints on other versions.
Templates SHALL NOT contain version numbers.

#### Scenario: Updating a supported version
- **WHEN** a maintainer raises a library module's default version in the rules file
- **THEN** newly generated projects use that version without any template change

## ADDED Requirements

### Requirement: Modules conditioned on traits
A rules-file module MAY list trait conditions. Such a module SHALL be used by a
selection only when one of its components is selected and, for every listed
trait, the selection provides one of the listed values. A module that is not
used SHALL NOT appear in the generated `go.mod`, and pinning it SHALL be
rejected like any module the selection does not use. Trait names in conditions
SHALL be checked against the traits the catalog's components provide.

#### Scenario: Dialect-specific driver left out
- **WHEN** a project is generated with `--db sqlite --data gorm`
- **THEN** its `go.mod` requires the SQLite dialect for GORM and does not require `gorm.io/driver/postgres`

#### Scenario: Pin for a module whose trait condition is not met
- **WHEN** the user runs `gobento new shop --db sqlite --data gorm --pin gorm.io/driver/postgres=<a version inside its range>`
- **THEN** gobento exits non-zero naming the module as not used by the selection, and writes nothing

#### Scenario: Unknown trait in a condition
- **WHEN** the rules file conditions a module on a trait that no component in the catalog provides
- **THEN** loading the catalog fails with an error naming the module and the trait
