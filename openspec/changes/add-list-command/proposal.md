# Proposal

## Why

To learn what gobento can generate (the components of each kind, which
combinations are supported, and which versions are allowed) a user has to open
the website or read `components/`. The CLI already holds all of this in its
embedded catalog and rules, but only reveals it piecemeal, through error
messages and questions. Scripts and tools have no stable way to read it at all.

## What Changes

- Add `gobento list`, a read-only discovery command that never asks a question
  and never touches the file system or network:
  - `gobento list components [--kind <kind>]`: each kind with its title, flag,
    and description, then each component's name, title, description, and the
    traits it provides and requires.
  - `gobento list selections [--arch ...] [--framework ...] [--db ...] [--data ...]`:
    the supported selections, optionally narrowed by component.
  - `gobento list versions`: the Go range and default, and each module's role,
    range, default, components, trait conditions, and constraints.
  - `gobento list`: all three sections.
- Every form prints aligned, human-readable text by default and a stable JSON
  document with `--json`. The JSON of `gobento list --json` is exactly the
  document the website is built from (`website/src/generated/catalog.json`);
  each subcommand prints the subset of its top-level fields that it lists.
- Move the export types and conversion out of `internal/tools/catalogexport`
  into a new internal package, `internal/catalogjson`, used by both the CLI
  and the exporter, so the two cannot drift. The exporter's output does not
  change.
- `gobento --help` and `gobento new --help` point at `gobento list`.

## Capabilities

### New Capabilities

- `catalog-discovery`: listing the components, supported selections, and
  version rules from the CLI, as text and as JSON.

### Modified Capabilities

- `website-docs`: the CLI reference has one page per command, so it gains a
  page for `gobento list`.

## Impact

- New: `internal/catalogjson/` (types, conversion, encoder, tests, moved from
  `internal/tools/catalogexport`), `internal/cli/list.go` and its tests and
  `testdata/list/` snapshots.
- Changed: `internal/tools/catalogexport` (thin wrapper around
  `internal/catalogjson`), `internal/cli/cli.go` and `new.go` (help text, the
  new command), CI path filters in `.github/workflows/pages.yml` and
  `website.yml` (the exporter's code now also lives in `internal/catalogjson`).
- No change to `website/src/generated/catalog.json` or any generated project.
- Documentation updated in this change: `README.md` (usage), new
  `website/content/docs/cli/list.mdx` wired into
  `website/content/docs/cli/meta.json`, `website/content/docs/cli/new.mdx` and
  `website/content/docs/components.mdx` (pointers to `gobento list`),
  `website/lib/catalog/schema.ts` and `website/lib/test/catalog.ts` (comments
  naming the exporter), `CONTRIBUTING.md` and `docs/maintainers.md` (where the
  exporter code lives, CI path filters), and the website's docs e2e page list.
