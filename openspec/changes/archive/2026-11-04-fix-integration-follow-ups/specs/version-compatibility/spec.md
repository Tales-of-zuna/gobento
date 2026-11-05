## MODIFIED Requirements

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
