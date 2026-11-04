# Proposal

## Why

`gobento new` only takes flags. When a component kind has several supported
options and its flag is omitted, gobento stops with
`--framework: choose a framework; supported: ...`, and a project name must
always be given as an argument. With gin and chi joining net/http (and more
databases, data-access libraries, and architectures on the roadmap), most
first runs would end in that error. The roadmap item "Interactive prompts"
makes the friendly path the default in a terminal, without changing anything
for scripts and CI.

## What Changes

- In a terminal (stdin and stdout are both terminals), `gobento new` asks for
  what it cannot decide by itself:
  - the project name, when no `<name>` argument is given, and then the module
    path (Enter keeps the name);
  - each component kind (architecture, framework, database, data access) that
    still has several supported options, once the earlier answers and flags
    are taken into account. Options come from the supported selections, so an
    unsupported combination cannot be chosen.
- Flags keep working and skip their question. A kind left with a single
  compatible option is chosen without a question.
- If a value given on the command line is invalid, gobento asks nothing and
  reports every problem, exactly as without a terminal.
- After answering, gobento prints the equivalent non-interactive command on
  stderr, so the run can be repeated in a script.
- New flag `--no-input` disables every question. Without a terminal (pipes,
  CI), prompts are disabled too, and behavior is exactly today's: a kind with
  several options and no flag is an error naming the flag and listing the
  values, and a missing name is an argument error.
- Ctrl-C or Esc during the questions stops gobento with exit status 130 and
  `gobento: interrupted; nothing was left behind`; nothing is written.
- Accessible mode: with `ACCESSIBLE` set to a non-empty value, or
  `TERM=dumb`, questions are plain numbered text prompts that screen readers
  can follow, instead of the full-screen interface.
- `--dry-run` asks the same questions, then lists the files as before.
- Versions (`--go`, `--pin`) stay flag-only; the defaults come from the
  compatibility rules.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `project-generation`: "Automatic selection of the only option" is scoped to
  non-interactive runs, and new requirements cover asking for missing input,
  disabling questions, aborting, and accessible prompts.

## Impact

- Code: `internal/cli` (new prompt logic and its huh rendering, `--no-input`,
  optional `<name>` in a terminal, exit status for an aborted prompt);
  `internal/generate` exports the name and module checks that `Prepare`
  already applies, so prompts validate answers with the same rules.
- Dependencies: `charm.land/huh/v2` (the prompt library) and
  `golang.org/x/term` (terminal detection), with their transitive modules.
- Documentation: `README.md` (usage and roadmap),
  `website/content/docs/cli/new.mdx` (synopsis, `--no-input`, prompts,
  accessibility, exit status), `website/content/docs/getting-started/quick-start.mdx`,
  `website/content/docs/concepts/components.mdx` (what happens when a flag is
  omitted), and the `gobento new --help` text.
- Website configurator: unchanged. Its commands always name the project and
  every kind with several options, so they never trigger a question.
