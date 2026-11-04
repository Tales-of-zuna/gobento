# Tasks

## 1. Fixes

- [ ] 1.1 Remove the stale PostgreSQL paragraph from `components/data/gorm/notes.md.tmpl`; refresh goldens and check a postgres+gorm README
- [ ] 1.2 Reword the kind flag help in `internal/cli/new.go`; verify `gobento new --help`
- [ ] 1.3 Name the unmet trait condition when a trait-conditioned module of a selected component is pinned (`internal/compat/resolve.go`), with a unit test; update the CLI test expectation
- [ ] 1.4 Report the missing name together with flag problems when nothing can be asked (`internal/cli/new.go`), with a test using a scripted terminal

## 2. Documentation

- [ ] 2.1 Update the pin error example in `website/content/docs/cli/new.mdx` (and any other page showing it); verify the documented output against the binary
