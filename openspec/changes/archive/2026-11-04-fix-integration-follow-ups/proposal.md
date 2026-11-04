# Proposal

## Why

A review of the integrated wave-1 work (gin/chi, prompts, vertical slice,
PostgreSQL, GORM) found text and messages that no longer match behavior: the
GORM notes in generated PostgreSQL READMEs still describe the store test as
using `DATABASE_URL`, `gobento new --help` claims every kind flag defaults to
"the only supported one" although every kind now has several, pinning a
trait-conditioned module names the wrong cause, and a terminal run without a
name hides flag problems behind the name error.

## What Changes

- The GORM notes drop the stale PostgreSQL paragraph; the database notes'
  "Tests" section (`TEST_DATABASE_URL`, `dbtest`) is the single description.
- Kind flag help says the flag is asked for in a terminal and required
  otherwise.
- Pinning a module of a selected component whose trait condition the
  selection does not meet reports the unmet condition, for example
  `cannot pin gorm.io/driver/postgres, it is only used with sql-dialect
  postgresql`.
- When `<name>` is omitted in a terminal and the flags are invalid (so
  nothing is asked), gobento reports the missing name and every flag problem
  together.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `version-compatibility`: *Modules conditioned on traits* names the unmet
  trait condition when such a module is pinned.
- `project-generation`: *No questions for invalid command-line input* reports
  flag problems together with a missing name.

## Impact

- Code: `components/data/gorm/notes.md.tmpl`, `internal/cli/new.go`,
  `internal/compat/resolve.go` and their tests; goldens refreshed.
- Documentation updated in the same change: `website/content/docs/cli/new.mdx`
  (pin error example) and `website/content/docs/concepts/version-rules.mdx`
  where the message is shown.
