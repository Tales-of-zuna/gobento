## ADDED Requirements

### Requirement: Binaries and npm packages take trait conditions
A platform binary or an npm package in the rules file MAY list trait
conditions like a module. It SHALL then be used only when one of its
components is selected and the selection provides one of the listed values
for every listed trait; otherwise templates, the scaffold step, and the
configurator do not see it. `gobento list versions` and the catalog
document SHALL show the conditions.

#### Scenario: Tailwind CLI only for htmx
- **WHEN** a project is prepared with `--frontend react --css tailwind`
- **THEN** it uses the npm packages `tailwindcss` and `@tailwindcss/vite` and not the `tailwindcss` binary, and with `--frontend htmx --css tailwind` it uses the binary and no npm package of Tailwind

#### Scenario: Condition on an unknown trait
- **WHEN** a binary or an npm package lists a trait condition on a trait no component provides
- **THEN** loading the catalog fails naming the entry and the trait
