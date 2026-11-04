## MODIFIED Requirements

### Requirement: Component reference is generated from the catalog
The Component reference SHALL list every component in the catalog with its
kind, description, the traits it provides or requires, the supported and
default versions of its modules, including the trait values a module is
conditioned on, and the version of every asset it vendors, generated at build
time. It SHALL mark optional kinds as optional.

#### Scenario: Version bump
- **WHEN** a module's default version changes in the rules file and the site is rebuilt
- **THEN** the component reference shows the new default without any edit to docs content

#### Scenario: Trait-conditioned module
- **WHEN** a module in the rules file is conditioned on `sql-dialect: postgresql`
- **THEN** the component reference lists it under its components and states that it is used only when `sql-dialect` is `postgresql`

#### Scenario: Vendored asset
- **WHEN** the rules file records htmx as an asset of the htmx frontend
- **THEN** the component reference lists htmx and its version under that frontend, marked as vendored
