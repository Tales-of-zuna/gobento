# Design

## Context

The CLI loads its catalog from `components.FS()` with `internal/catalog`,
which also parses the version rules (`internal/compat`); `compose.Supported`
computes the supported selections. `internal/tools/catalogexport` converts the
same data into the JSON document the website reads
(`website/src/generated/catalog.json`, schema in
`website/lib/catalog/schema.ts`), and `make catalog-check` keeps the committed
copy current. Its types and conversion live in `package main`, so nothing else
can use them. A third architecture is being added in parallel, so neither the
command nor its tests may assume today's components or counts.

## Goals / Non-Goals

**Goals:**

- One read-only command that shows every component, every supported
  selection, and every version rule, as aligned text and as stable JSON.
- One source for the JSON shape, shared by the CLI and the website exporter.
- Output that is deterministic and derived from the catalog alone.

**Non-Goals:**

- Searching or describing a single component in depth (`gobento show`), the
  generated file layout, or slot contracts: the website covers those.
- A versioned JSON schema separate from the website's; see D3.
- Color or terminal-width-aware wrapping.

## Decisions

### D1: `gobento list` with three subcommands

`gobento list components`, `gobento list selections`, and
`gobento list versions`, plus `gobento list` for all three. Subcommands read
better in help and docs than one command with section flags, each owns only
the flags that make sense for it (`--kind` for components; the `gobento new`
component flags for selections), and new sections can be added later without
changing existing output. `--json` is a persistent flag of `list`, so it works
in every form. Alternatives: separate top-level commands (`gobento
components`, ...) clutter the root help; flags like `gobento list --components`
cannot carry per-section filters cleanly.

The selection filters reuse the names and spelling of `gobento new`'s flags
(`--arch`, `--framework`, `--db`, `--data`), so a user can narrow the list with
the flags they are about to pass to `new`. `--kind` takes the kind IDs used in
the catalog and in `--json` (`architecture`, `framework`, `database`, `data`);
the text output shows each kind's flag next to its title so the mapping is
visible.

### D2: Move the export types into `internal/catalogjson`

The types (`Document`, `Kind`, `Component`, `GoRule`, `Module`, `Constraint`,
`Requirement`, `Rules`, ...), the kind titles and descriptions, the conversion
from a loaded `*catalog.Catalog`, and the encoder move from
`internal/tools/catalogexport` into a new package `internal/catalogjson`.
`catalogexport` keeps only its flag parsing and atomic file write, and the CLI
renders both its JSON and its text from the same `Document`. Because the text
is rendered from the document rather than from the catalog directly, the
website, the JSON, and the text cannot disagree: anything the website shows
about components or versions is exactly what `gobento list` shows.

Alternatives considered: importing `catalogexport` from the CLI is impossible
(`package main`); duplicating the types in the CLI would drift; putting them in
`internal/catalog` would mix the website's presentation schema (kind titles,
form messages, flattened constraints) into the loader. The JSON document is a
presentation of the catalog, so it gets its own package. The exporter's output
is unchanged, which `make catalog-check` proves.

The website's CI path filters (`pages.yml`, `website.yml`) watch
`internal/tools/catalogexport/`; they gain `internal/catalogjson/`, because a
change there can change `catalog.json`.

### D3: The JSON is the website's catalog document

`gobento list --json` prints the full document, byte for byte what the
exporter writes; each subcommand prints an object with the subset of top-level
fields it lists (`kinds` and `components`; `supported`; `go` and `modules`).
Every JSON output is therefore a partial version of one documented schema, and
a script can decode any of them into the same types. The schema already has a
consumer that breaks loudly on change (the website's zod schema and tests), so
it is stable in practice: fields are added, not renamed or removed, and a
change to it is a change to `website/lib/catalog/schema.ts` in the same
commit. No separate version field is added; it would change `catalog.json` for
no current consumer. The filters of `components` and `selections` apply to the
JSON too, so the JSON and the text of a command always describe the same
items.

### D4: Text layout

Text is aligned with `text/tabwriter` (two spaces of padding, no tabs in the
output), with a blank line between sections and kinds:

- components: a heading per kind (`Database (--db)`) and its description,
  then per component a row of name and title, followed by indented lines for
  the description, `provides: k=v, ...` and `requires: k=a|b, ...` (traits
  sorted by name; required values in catalog order). Lines are not wrapped:
  descriptions are one sentence and wrapping would make the output depend on
  the terminal.
- selections: a header row of kind IDs in upper case, one row per selection,
  and a final count line (`24 supported selections.`), or a message naming
  the filter when nothing matches.
- versions: a `Go` line with range and default, a module table (`MODULE`,
  `ROLE`, `RANGE`, `DEFAULT`, `USED BY`, `ONLY WHEN`, with `-` for no trait
  condition), and a `Constraints` list (`<module> <when> needs <target> <range>`),
  omitted when there are none.

`gobento list` prints the three sections under the headings `Components`,
`Selections`, and `Versions`.

### D5: Tests that survive new components

Snapshot tests render a frozen document from `internal/cli/testdata/list/catalog.json`
(a copy of today's export) into `testdata/list/*.txt`, rewritten with
`go test ./internal/cli -run TestListSnapshots -update`. Adding a component to
the real catalog changes neither file. Tests against the embedded catalog
compare with values computed from the catalog (`compose.Supported`, the rules,
the components) and assert a few known entries (for example `postgres`,
`gorm`), never counts. A round-trip test decodes `gobento list --json` into
`catalogjson.Document`, compares it with the document built from the catalog,
and re-encodes it to the same bytes.

## Risks / Trade-offs

- [The JSON shape is tied to the website's needs] → It already contains
  everything a script needs about components and versions; additions for the
  CLI are additive and go through the website schema too.
- [`gobento list` output grows with the product of component counts] → The
  selections section is the only multiplicative one; filters and the
  per-section subcommands keep it usable.
- [Moving files conflicts with parallel changes to the exporter] → The move is
  mechanical and the exporter's behavior is unchanged; conflicts are limited
  to `internal/tools/catalogexport/main.go`.
