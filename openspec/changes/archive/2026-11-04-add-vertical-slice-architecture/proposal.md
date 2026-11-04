# Proposal

## Why

gobento ships one architecture, `mvc`, so the claim that frameworks, databases,
and data-access layers fill slots by name and drop into any layout has only
been shown with test fixtures. A second, genuinely different layout proves it
with the shipped components and gives users the feature-oriented layout many Go
services prefer (README roadmap: "More architectures: vertical slice").

## What Changes

- Add the architecture component `vertical-slice` (title "Vertical slice"): the
  example `items` feature lives in `internal/items` (its types and the
  `ItemStore` contract) with its HTTP handlers in `internal/items/itemshttp`
  and its data access in `internal/items/itemsstore`; the shared database
  connection lives in `internal/platform/database`. It declares the same slot
  names as `mvc` (`entry`, `models`, `handlers`, `database`, `store`), so
  `nethttp`, `sqlite`, and `sqlc` are used unchanged.
- The generated README of a `vertical-slice` project explains the layout and how
  to add a second feature.
- Composition rejects a file that a component writes into a slot directory nested
  inside its own slot directory (vertical slice nests slots inside the feature
  directory, so this keeps "one owner per slot" true for nested slots).
- With two architectures, `--arch` has two supported values, so omitting it now
  fails under the existing rule "Automatic selection of the only option".
  **BREAKING** for invocations that relied on the default: `gobento new shop`
  must become `gobento new shop --arch mvc` (the release is pre-1.0).
- Golden-file coverage for the reference selection of the new architecture;
  tests that assumed a single architecture are updated.
- The CI job that regenerates the sqlc output passes `--arch mvc`.
- The website configurator previews the new architecture's directories, and the
  exported catalog lists it and its supported selection.

## Capabilities

### New Capabilities

- `architecture-layouts`: the project layouts gobento ships, starting with the
  vertical slice layout, its package dependency direction, and its README.

### Modified Capabilities

- `component-composition`: "One owner per slot" also covers files that would
  land in a nested slot directory.
- `project-generation`: the scenarios of "Generate a project from a component
  selection" and "Generated project builds and passes its tests" name the
  architecture explicitly and cover every shipped architecture.

## Impact

- Code: `components/architecture/vertical-slice/` (new), `internal/compose`
  (nested-slot check, goldens, supported list), tests in `internal/catalog`,
  `internal/cli`, `internal/generate`; `.github/workflows/ci.yml` (sqlc-fresh).
- Website code: `website/components/configurator/view-model.ts` (file-tree
  preview), `website/src/generated/catalog.json` (`make catalog`),
  `website/lib/command/supported-commands.txt`, website unit tests that list
  architectures.
- Documentation updated in the same change: `README.md` (usage, roadmap),
  `website/content/docs/index.mdx`, `website/content/docs/concepts/components.mdx`,
  `website/content/docs/concepts/slots-and-contracts.mdx`,
  `website/content/docs/concepts/traits.mdx` (if it lists selections),
  `website/content/docs/cli/new.mdx`,
  `website/content/docs/getting-started/quick-start.mdx`,
  `website/content/docs/contributing/adding-a-component.mdx`, `CONTRIBUTING.md`
  where it shows `gobento new` without `--arch`, the `gobento new --help` text,
  and the new architecture's generated `README.md` template.
