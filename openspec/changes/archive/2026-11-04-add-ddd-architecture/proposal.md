# Proposal

## Why

gobento ships a layered (`mvc`) and a feature-oriented (`vertical-slice`)
layout. Many Go services are built in a domain-driven, ports-and-adapters
style instead, and the README roadmap lists "More architectures: DDD" as the
next layout. A third layout with a real application layer also tests the slot
model again: the existing framework, database, and data-access components must
drop in unchanged.

## What Changes

- Add the architecture component `ddd` (title "Domain-driven design"):
  - `internal/domain` (package `domain`): the `Item` entity, the rule every
    item name follows, and the `ItemStore` port.
  - `internal/app` (package `app`): the application service `ItemService`,
    which runs the item use cases, checks the domain rules, and stores items
    through an `ItemStore`. It implements `ItemStore` itself, so the HTTP
    adapter uses it through the same port.
  - `internal/adapters/httpapi`, `internal/adapters/persistence`, and
    `internal/adapters/database`: the HTTP, data-access, and database
    adapters, filled by the framework, data, and database components.
  - `cmd/<name>/main.go` wires the adapters, the service, and the domain.
- It declares the same slot names as `mvc` and `vertical-slice` (`entry`,
  `models`, `handlers`, `database`, `store`), plus one architecture-owned slot
  `app` for the application service. Every shipped framework, database, and
  data-access component is used unchanged.
- The generated README of a `ddd` project explains the layers, the dependency
  direction (adapters → app → domain), and how to add a second aggregate.
- The generated project includes tests for the domain rule and the
  application service.
- With three architectures, the `--arch` error and the interactive question
  list `ddd`, `mvc`, and `vertical-slice`.
- Tests: a full golden directory for `ddd/nethttp/sqlite/sqlc`, the
  supported-selection lists, the layout-independence test, and an integration
  check that the domain imports no other project package. Per-component
  goldens keep using the reference selection's other components, so adding an
  architecture does not rewrite them.
- Website: the configurator previews the `ddd` directories, the exported
  catalog and the supported commands include it, and the home belt shows it.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `architecture-layouts`: adds the DDD architecture, its dependency direction,
  and its README; the README requirement covers `ddd`.
- `project-generation`: the "Architecture must be chosen" scenario lists all
  three architectures, and the build check covers a `ddd` selection.

## Impact

- Code: `components/architecture/ddd/` (new); `internal/compose`
  (`golden_test.go`: goldens table, per-component selection choice,
  layout-independence test over every architecture; `testdata/golden-ddd`);
  `internal/catalog/shipped_test.go`; `internal/cli` (help text, tests);
  `internal/generate/integration_test.go` (domain import check).
- Website code: `website/components/configurator/view-model.ts` (file-tree
  preview) and its test, `website/src/generated/catalog.json` and
  `website/lib/command/supported-commands.txt` (`make catalog`),
  `website/components/components-reference.test.tsx`,
  `website/tests/home.spec.ts` (belt plates), configurator e2e messages.
- Documentation updated in the same change: `README.md` (usage, tree example,
  roadmap tick "DDD"), `website/content/docs/index.mdx`,
  `website/content/docs/concepts/components.mdx`,
  `website/content/docs/concepts/slots-and-contracts.mdx`,
  `website/content/docs/cli/new.mdx` (flag table),
  `website/content/docs/getting-started/quick-start.mdx`, the
  `gobento new --help` text, and the new architecture's generated
  `README.md` template.
