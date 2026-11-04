# website-docs Specification

## Purpose
Provides gobento's user and contributor documentation on the website, with
reference pages generated from the component catalog so that they always
match what the CLI supports.

## Requirements

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
