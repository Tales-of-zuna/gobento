# Design

## Context

`gobento new` is built in `internal/cli/new.go`: cobra parses flags, `runNew`
builds a `generate.Spec`, `generate.Prepare` validates everything at once
(name, module path, selection, versions, composition), and either the dry-run
listing or `generate.Generate` follows. `<name>` is a required argument. The
selection is checked against `compose.Supported(cat)`, the list of selections
that resolve and compose; a kind with several supported names and no flag
becomes a `generate.SelectionError`, which the CLI prints as
`--framework: choose a framework; supported: ...`.

Interrupts: `cmd/gobento` cancels the context on SIGINT/SIGTERM, and
`app.exitStatus` maps a cancelled context plus a `context.Canceled` error to
status 130 with `gobento: interrupted; nothing was left behind`.

The agreed prompt library is charmbracelet huh. Its current release is
`charm.land/huh/v2` v2.0.3 (the v1 line, `github.com/charmbracelet/huh`, is
the previous major). huh renders a full-screen form with bubbletea, or, in
accessible mode, plain line prompts read from an `io.Reader`.

## Goals / Non-Goals

**Goals:**
- Ask only for what gobento cannot decide, offer only supported options, and
  keep every non-interactive run byte-for-byte as today.
- Keep the decision logic (which questions, which options, the resulting
  input) free of huh, and unit-test it with a scripted asker.
- Same validation rules as `Prepare`, same abort semantics as an interrupt.

**Non-Goals:**
- Prompts for `--go`, `--pin`, or `--dir` (see Decisions).
- Going back to an earlier question; a run is short, and Ctrl-C plus the
  printed command is the way to start over.
- A configuration file or remembered answers.

## Decisions

### When gobento asks
Questions are possible only when stdin and stdout are both terminals
(`golang.org/x/term.IsTerminal`, which works on Linux, macOS, and Windows
consoles) and `--no-input` is not given. Anything else (pipes, CI, redirected
output, Git Bash's mintty, which is not a Windows console) keeps today's
behavior. `--no-input` follows the common CLI convention (clig.dev) for
terminals where questions are unwanted. No environment variable disables
questions: CI has no terminal, so one is not needed.

What is asked:
- **Name**, only when no `<name>` argument is given. Then the **module path**,
  pre-filled with the name; an empty answer keeps the name. The module path is
  asked only together with the name: `gobento new shop` keeps its default
  module without a question, so a command copied from the website or the docs
  runs without questions whenever it names every kind with several options
  (the configurator always does).
- **Each kind** whose flag is omitted, in kind order (architecture,
  framework, database, data access), when it has several options among the
  supported selections that match the flags and the earlier answers. One
  remaining option is chosen silently.

Alternative considered: always asking for the module path. Rejected because
it would add a question to the most common command (`gobento new shop`) and
to every copied website command.

### Filtering instead of validating
The interview keeps the list of supported selections (`compose.Supported`)
that match everything decided so far. A question's options are the distinct
names of its kind in that list, sorted like the error message lists them.
Every answer narrows the list, so an unsupported combination can never be
chosen and no "incompatible" error needs to be designed. The order follows
`catalog.Kinds()`, so the architecture (which defines slots) is decided first.

### Invalid command-line input means no questions
Before asking, gobento checks what was given: the name and module path (with
the same rules `Prepare` uses), the `--pin` syntax, and whether the component
flags match any supported selection. If anything is wrong, it asks nothing
and `Prepare` reports every problem as today. Asking first and then failing on
a typo in a flag would waste the answers; the report stays "every problem at
once". Version range problems (`--go`, `--pin` values) are still found by
`Prepare` after the answers, before anything is written.

### Same validation rules
`internal/generate` gets `CheckName(name)` and `CheckModule(path, name)`,
extracted from `Prepare` without changing their messages; `Prepare` calls
them, and the name and module questions use them as validators, so an answer
that the question accepts is one `Prepare` accepts.

### Code layout: decision logic separate from rendering
- `internal/cli/prompt.go`: a `question` value (field, title, help, options
  or text with a default and validator), an `asker` interface
  (`ask(ctx, question) (string, error)`), `canAsk` (the pre-check) and
  `interview`, which walks the decisions above and fills the input. It also
  builds the equivalent command. No huh import.
- `internal/cli/prompt_huh.go`: `huhAsker`, one single-field huh form per
  question (an `Input` or a `Select`). One form per question, rather than one
  form with dynamic options, keeps the dependency between answers in plain Go
  where it is tested, at the cost of back-navigation (see Non-Goals).
- `app` gains `stdin`, `interactive func() bool`, `getenv`, and an optional
  `asker`, so tests inject a scripted asker and a fake terminal check.
  Existing tests build `app` without them and stay non-interactive.

### Aborting
- Full-screen mode: Ctrl-C and Esc are the form's quit keys (Esc added to
  huh's default Ctrl-C; the select filter, which also uses Esc, is disabled,
  as the lists are short). huh returns `ErrUserAborted`, which the CLI turns
  into `errAborted`; `exitStatus` maps it to 130 with the existing
  interrupted message. SIGTERM cancels the context, which stops the form, and
  the cancelled context gives the existing 130 path.
- Accessible mode reads lines in cooked mode, so Ctrl-C is a SIGINT that
  cancels the context. huh's accessible prompts cannot be cancelled, so the
  form runs in a goroutine and the asker returns `ctx.Err()` as soon as the
  context is done; the process then exits and the blocked read goes with it.
- End of input (Ctrl-D): huh's accessible prompts silently use the default
  on EOF. The asker reads through a wrapper that records EOF and returns
  "input ended before every question was answered", exit status 1. The
  wrapper also reads one byte at a time, because huh starts a new
  `bufio.Scanner` for every field and would otherwise swallow the next
  answers when input arrives in one chunk.
Nothing is written in any of these cases, because questions end before
`Prepare` and generation start.

### Accessibility
Accessible mode is on when `ACCESSIBLE` is non-empty (the convention of
charm's tools, such as huh's examples and gum) or `TERM=dumb` (a terminal
that cannot draw the full-screen form). No flag: it is a property of the
user's setup, set once in the environment, not of one run. Accessible output
goes through `colorprofile.NewWriter`, so colors and styles are dropped where
the terminal cannot show them (and in tests). huh's accessible fields print
only their title, so the asker prints the question's hint line first.

### Full-screen details
- A text question's default (the module path) is a placeholder, not editable
  text, so typing replaces it and Enter keeps it.
- huh's text input panics when it has no width, which happens on a terminal
  that reports a 0x0 size (some pseudo-terminals). When stdout reports no
  usable size (fewer than 20 columns, or an error), the form gets a fixed
  width of 80 columns; otherwise it follows the terminal.
- Tests drive the full-screen form with key bytes on a reader (Enter, `j`,
  Esc, Ctrl-C) and a fixed width, in addition to the accessible mode.

### Equivalent command
After at least one question, stderr shows
`Equivalent command: gobento new <name> [flags]`, with the module path when it
differs from the name, every kind that was asked or given or has several
supported options globally, and `--go`, `--pin`, `--dir`, `--dry-run` as
given, quoted for a POSIX shell. stdout keeps exactly today's output, so
`--dry-run` listings and next steps are unchanged.

### Dry run
`--dry-run` asks the same questions: the point of a dry run is to check the
choices before generating, and the printed equivalent command (with
`--dry-run`) can then be rerun without it.

### Versions stay flag-only
The Go version and module pins default to the rules, which is what almost
every user wants; ranges are narrow (two Go minors today), and pins are
per-module expert choices whose valid set depends on the selection. A
question for them would lengthen every run to serve few users. The help, the
docs, and the equivalent command point to `--go` and `--pin`.

### Dependencies
`charm.land/huh/v2` v2.0.3, the latest release on proxy.golang.org, and
`golang.org/x/term`; their transitive modules come with them. huh is used
only from `internal/cli`, so it stays out of the generated projects. The
`go` directive stays `1.26.0`.

## Risks / Trade-offs

- [Dependency weight: huh pulls bubbletea, lipgloss, and terminal modules]
  → Accepted for the agreed library; only the CLI binary grows, and
  `govulncheck` covers the new modules in CI.
- [Full-screen mode depends on terminal behavior that tests only
  approximate] → The decision logic, the accessible rendering, and key
  handling of the full-screen form are tested; real terminals were checked
  with a manual run in a pseudo-terminal (Linux). Windows consoles are
  covered by huh and bubbletea, not by a gobento test.
- [Interactive runs are possible only with a real terminal on Windows] →
  Documented: in mintty, use `winpty`, Windows Terminal, or flags.
- [Parallel catalog changes (gin, chi, postgres, gorm, vertical slice)] → The
  interview reads everything from the catalog; new components need no CLI
  change.
