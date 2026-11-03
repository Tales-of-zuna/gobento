# Proposal

## Why

Existing Go project generators (notably go-blueprint) hard-code tool versions in
templates and grow one template per combination, so they fall behind upstream
releases and are hard to extend. gobento needs a foundation that avoids both
problems before any breadth (more frameworks, databases, frontends) is added.
This change sets that foundation and proves it with one complete, buildable
stack.

## What Changes

- New `gobento` CLI with a `new` command that generates a Go project from a
  selection of components (architecture, HTTP framework, database, data access),
  driven by flags. Every input problem is reported at once, before anything is
  written.
- A composition model. An architecture declares slots, and each slot is one Go
  package filled by exactly one component. The architecture owns the Go
  contract between slots (types and function signatures), so the compiler
  checks that components fit together. Traits (such as SQL dialect and
  connection type) express which components are compatible. No template exists
  per combination.
- A version-compatibility rules file covering the Go version and every module:
  supported ranges, defaults, and constraints. Users may choose versions inside
  the ranges. Incompatible choices, and any version that dependency
  resolution would silently raise, are rejected.
- Safe output: refuses non-empty targets, fills empty ones (including `.`),
  checks the local Go toolchain first, and leaves nothing behind on failure or
  Ctrl-C.
- The first supported combination: architecture **MVC**, framework
  **net/http**, database **SQLite** (pure Go, no cgo), data access **sqlc**. The
  generated project has a health endpoint and one example resource, and it
  builds, vets, and passes its own tests.
- CI for the generator: every supported combination is generated, built, and
  tested on Linux, macOS, and Windows, plus a check that the shipped sqlc
  output is current.

## Capabilities

### New Capabilities
- `project-generation`: the `gobento new` command, covering inputs, validation,
  output location rules, toolchain preflight, cleanup guarantees, and the
  promise that a generated project builds and passes its tests.
- `component-composition`: slots, one owner per slot, architecture-owned Go
  contracts, traits, composition errors, and what counts as a supported
  combination.
- `version-compatibility`: the rules data for versions, defaults, overrides,
  constraint checks, and exact pinning in the generated project.

### Modified Capabilities
<!-- None: this is the first behavior change; no specs exist yet. -->

## Impact

- New Go module `github.com/Tales-of-zuna/gobento` (binary `gobento`), Go 1.26+.
- New dependencies: `github.com/spf13/cobra`, `golang.org/x/mod`,
  `go.yaml.in/yaml/v3` (see design.md D8).
- Generation needs network access to the Go module proxy, because selected
  versions are fetched with the Go toolchain.
- Depends on change `establish-repo-standards` (license, CI baseline, Makefile,
  `.gitattributes`), which lands first. This change extends that CI.
- Out of scope: interactive prompts (they arrive with the first kind that has
  a second option), other frameworks, databases, ORMs, frontends, Tailwind,
  architectures other than MVC, and releases.
