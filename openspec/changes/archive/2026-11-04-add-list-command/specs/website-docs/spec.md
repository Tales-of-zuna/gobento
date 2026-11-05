## MODIFIED Requirements

### Requirement: Documentation structure
The documentation SHALL contain at least these sections: Getting started,
Concepts (slots, contracts, traits, version rules), CLI reference (one page
per command: `gobento new` and `gobento list`), Component reference, and
Contributing (including adding a component), all navigable from a persistent
sidebar.

#### Scenario: Navigation
- **WHEN** a visitor opens any docs page
- **THEN** the sidebar lists every section, highlights the current page, and offers previous and next links

#### Scenario: Every command has a page
- **WHEN** a visitor opens the CLI reference
- **THEN** the sidebar lists a page for `gobento new` and a page for `gobento list`
