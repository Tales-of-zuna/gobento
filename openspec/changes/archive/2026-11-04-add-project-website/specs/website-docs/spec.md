# Spec Delta

## Purpose

Provides gobento's user and contributor documentation on the website, with
reference pages generated from the component catalog so that they always
match what the CLI supports.

## ADDED Requirements

### Requirement: Documentation structure
The documentation SHALL contain at least these sections: Getting started,
Concepts (slots, contracts, traits, version rules), CLI reference, Component reference,
and Contributing (including adding a component), all navigable from a persistent sidebar.

#### Scenario: Navigation
- **WHEN** a visitor opens any docs page
- **THEN** the sidebar lists every section, highlights the current page, and offers previous and next links

### Requirement: Component reference is generated from the catalog
The Component reference SHALL list every component in the catalog with its
kind, description, the traits it provides or requires, and the supported and
default versions of its modules, generated at build time.

#### Scenario: Version bump
- **WHEN** a module's default version changes in the rules file and the site is rebuilt
- **THEN** the component reference shows the new default without any edit to docs content

### Requirement: Static search
The documentation SHALL provide full-text search that works on the static site
without a server.

#### Scenario: Searching
- **WHEN** a visitor searches for "traits"
- **THEN** results include the Concepts page that explains traits, and selecting a result opens it

### Requirement: Edit on GitHub
Every docs page SHALL link to its source file on GitHub for editing.

#### Scenario: Suggesting an edit
- **WHEN** a visitor activates "Edit this page"
- **THEN** GitHub's editor opens for that page's source file on the `main` branch

### Requirement: Docs code blocks are copyable
Every multi-line or command code block in the documentation SHALL have a copy
control.

#### Scenario: Copy a command
- **WHEN** a visitor activates the copy control on a command block
- **THEN** the block's content is placed on the clipboard and a confirmation appears
