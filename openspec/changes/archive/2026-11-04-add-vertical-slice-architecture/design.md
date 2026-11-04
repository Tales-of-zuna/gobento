# Design

## Context

An architecture component declares slots (`entry`, `models`, `handlers`,
`database`, `store`); each slot is one directory and one Go package. Framework,
database, and data-access templates refer to slots only by name
(`.Slots.handlers.Package`, `.Slots.models.Import`), and the architecture owns
`cmd/<name>/main.go` and the contract package (`Item`, `ItemStore`). Only `mvc`
ships today, so the shipped components have never been placed in another
layout. See proposal.md for the motivation.

## Goals / Non-Goals

**Goals:**

- A vertical slice layout that a Go developer would recognize and keep, built
  only from the existing slot names, so `nethttp`, `sqlite`, and `sqlc` are
  reused byte-for-byte as templates.
- No import cycles and sensible package names in every generated project.

**Non-Goals:**

- Several features per project, or a feature name chosen by the user. The
  example feature stays `items`, as in `mvc`.
- Moving `/health` or the router out of the framework's handlers package:
  frameworks provide `NewRouter`, and changing that contract would touch every
  framework (gin and chi are being added in parallel).
- Sharing template files between architectures.

## Decisions

### D1. Component name `vertical-slice`

The `--arch` value is what users type and what the configurator shows.
`vertical-slice` is the common name of the pattern and reads unambiguously in
`gobento new shop --arch vertical-slice`. Alternatives: `slice` is ambiguous in
Go (slices are a language type); `vslice` is short but not a term anyone
searches for. Component names already allow `-`.

### D2. Layout

| Slot       | Directory                    | Package       |
| ---------- | ---------------------------- | ------------- |
| `entry`    | `cmd/{{.Project.Name}}`      | `main`        |
| `models`   | `internal/items`             | `items`       |
| `handlers` | `internal/items/itemshttp`   | `itemshttp`   |
| `store`    | `internal/items/itemsstore`  | `itemsstore`  |
| `database` | `internal/platform/database` | `database`    |

- The feature directory `internal/items` is the slice: everything about items
  is below it. Its root package holds the types and the `ItemStore` contract,
  the inner ring that both subpackages import. The subpackages never import
  each other, and the root imports neither, so no cycle is possible; `main`
  wires them, exactly as in `mvc`.
- Subpackage names follow the standard library's prefix convention
  (`net/http/httptest`, `net/http/httputil`): `itemshttp` and `itemsstore`
  stay distinct when imported side by side in `main.go`
  (`itemshttp.NewRouter(itemsstore.New(db))`) and when a second feature adds
  `ordershttp`. Plain `http` and `store` were rejected: `http` collides with
  `net/http` in every handler file, and a second feature would add a second
  package `store`, forcing import aliases.
- Directory names equal package names, so the slot `package:` override is not
  needed (Go convention: the last path element is the package name).
- The database connection is shared by all features, so it does not belong to
  the slice: `internal/platform/database` is the common Go name for such
  cross-cutting infrastructure. The schema file stays with the database
  component (`schema.sql`), unchanged.
- Alternative considered: handlers in the feature root (`internal/items`)
  with the types in `internal/items/itemsmodel`. Rejected: the types would be
  a dependency of their own parent package, which reads inside-out, and the
  store would have to import a sibling of the handlers.

### D3. Nested slot directories are allowed, intrusions are rejected

Slots `handlers` and `store` lie inside the directory of `models`. Compose
already allows this: slot directories must only be distinct. But an owner may
add subpackages below its slot (as sqlc does with `sqlcgen`), so the owner of
an outer slot could write into an inner slot's directory, breaking "one owner
per slot" without a path conflict. Compose now rejects any slot file whose path
lies inside a different slot directory nested within its own (mirroring the
existing check for root files inside slot directories). This is the only core
change, and no shipped component is affected by it.

### D4. Architecture-owned files are separate copies

`vertical-slice` has its own `main.go`, `.gitignore`, and README templates. The
entry point is currently identical in wiring to `mvc`'s, but it belongs to the
architecture by contract, and the README must explain a different layout. A
shared-partials mechanism would be a core change for about 90 lines and is left
for when a third architecture shows what is truly common.

### D5. Tests and goldens

- `TestGolden` becomes table-driven: the existing `testdata/golden` keeps
  `mvc/nethttp/sqlite/sqlc`, and `testdata/golden-vertical-slice` holds
  `vertical-slice/nethttp/sqlite/sqlc`. Keeping the old directory unchanged
  limits conflicts with the parallel per-component golden restructuring.
- A shipped-catalog test checks the reuse claim directly: for each framework
  file, the `vertical-slice` output equals the `mvc` output once the slot
  package names and import paths are swapped.
- Tests that relied on the single architecture being chosen automatically pass
  `--arch mvc` (or `Components[architecture]`); a new CLI test covers the
  missing `--arch` error with the real catalog.

## Risks / Trade-offs

- [Omitting `--arch` now fails] → Intended by the existing rule; the error
  names `--arch` and both values, every documented example is updated, and the
  interactive prompts being added in parallel will ask instead.
- [`itemshttp` also serves `/health`, which is not about items] → The README
  says so; moving the router needs a framework contract change (non-goal).
- [Two copies of `main.go` can drift] → Both are covered by goldens and the
  integration test; D4 records when to factor them out.

## Migration Plan

Users of `gobento new <name>` without flags add `--arch mvc` to keep the
previous output. Nothing else changes for `mvc` projects; its golden files are
unchanged.
