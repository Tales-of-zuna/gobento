# Spec Delta

## Purpose

Defines the project layouts gobento ships as architecture components: where each
slot lives, how the generated packages may depend on each other, and how the
generated README explains the layout.

## ADDED Requirements

### Requirement: Vertical slice architecture
gobento SHALL ship an architecture `vertical-slice` that organizes a project by
feature. The example `items` feature SHALL own one directory holding its types
and store contract, with its HTTP handlers and its data access in subpackages
of that directory. Code shared by all features, such as the database
connection, SHALL live outside the feature directory.

#### Scenario: Feature layout
- **WHEN** the user runs `gobento new shop --module example.com/shop --arch vertical-slice --framework nethttp --db sqlite --data sqlc`
- **THEN** the project contains `cmd/shop/main.go`, the package `items` in `internal/items`, the handlers package `itemshttp` in `internal/items/itemshttp`, the store package `itemsstore` in `internal/items/itemsstore`, and the database package in `internal/platform/database`
- **AND** it contains no `internal/models`, `internal/controllers`, or `internal/store` directory

#### Scenario: Shipped components are reused unchanged
- **WHEN** a framework, database, or data-access component is selected with `vertical-slice`
- **THEN** gobento uses the same component templates it uses with `mvc`, placed in the `vertical-slice` directories with the matching package names and import paths

### Requirement: Feature packages depend inward
In a `vertical-slice` project, the feature's handlers and store packages SHALL
depend on the feature's contract package and never on each other, and the
feature's contract package SHALL import none of its subpackages, so the
packages form no import cycle.

#### Scenario: Handlers stay independent of storage
- **WHEN** a `vertical-slice` project is generated for any supported selection
- **THEN** the handlers package imports neither the store package nor the database package, directly or transitively, and the project builds

### Requirement: Generated README explains the layout
The README of a project generated with an architecture SHALL describe the
directories of that architecture and what each one contains. For
`vertical-slice` it SHALL also explain how to add another feature.

#### Scenario: Vertical slice README
- **WHEN** a project is generated with `--arch vertical-slice`
- **THEN** its `README.md` lists `internal/items`, `internal/items/itemshttp`, `internal/items/itemsstore`, and `internal/platform/database` with their contents
- **AND** it describes the steps to add a second feature next to `internal/items`
