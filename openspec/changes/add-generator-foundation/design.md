# Design

## Context

The repository only has OpenSpec scaffolding and the repository standards from
change `establish-repo-standards` (license, CI baseline, Makefile,
`.gitattributes`). That change lands first. See proposal.md for motivation and
`specs/` for required behavior. The local toolchain is Go 1.27. Go upstream
supports 1.26 and 1.27, and every dependency below needs at least 1.26.
gobento must run on Linux, macOS, and Windows without cgo.

## Goals / Non-Goals

**Goals:**
- A composition engine and data formats that later frameworks, databases, and
  architectures plug into without engine changes.
- Every failure (input, composition, versions, toolchain) is detected in memory,
  before the first byte is written.
- One combination proven end to end on three operating systems.

**Non-Goals:**
- Interactive prompts. They are deferred to the change that adds a second
  option for some kind, because with one option per kind there is nothing to
  ask.
- Trait-variant template directories (`files@<value>/`). They are deferred until
  a second SQL dialect exists. The trait *check* is in scope.
- Third-party component plugins, updating existing projects, and offline
  generation.

## Decisions

### D1. Repository layout

```
cmd/gobento/          main: wires the CLI only
internal/cli/         cobra commands and flag parsing
internal/catalog/     loads manifests + versions.yaml from an fs.FS; lists trait-compatible selections
internal/compat/      version ranges, constraint evaluation, resolution
internal/compose/     slot resolution, conflict detection, rendering into an in-memory Plan; supported selections
internal/generate/    preflight, writing a Plan, Go toolchain steps, atomic placement, cleanup
components/           data only: component dirs, versions.yaml, embed.go
```

Packages depend only downward: cli → generate → compose → catalog, compat.

*Alternative:* a single package. Rejected because compat and compose are what
contributors change most, and they need isolated tests.

### D2. Slots are Go packages, contracts are Go code

An architecture declares slots. Each slot is one directory, which means one Go
package, filled by exactly one component kind:

```yaml
# components/architecture/mvc/component.yaml
name: mvc
title: Model-View-Controller
slots:
  entry:    { dir: "cmd/{{.Project.Name}}", package: main, filled_by: architecture }
  models:   { dir: internal/models,         filled_by: architecture }
  handlers: { dir: internal/controllers,    filled_by: framework, required: true }
  database: { dir: internal/database,       filled_by: database,  required: true }
  store:    { dir: internal/store,          filled_by: data,      required: true }
```

The architecture owns the entry point and the contract. Its `main.go` calls
fixed functions:

| Slot | Contract the filling component must provide |
|---|---|
| `models` (architecture) | `type Item struct{ID int64; Name string; CreatedAt time.Time}`, `type ItemStore interface{ List(ctx) ([]Item, error); Create(ctx, name string) (Item, error) }` |
| `database` | `func Open(ctx context.Context, dsn string) (T, error)`, where `T` has `Close()` and is named by the `conn` trait |
| `store` | `func New(db T) models.ItemStore` |
| `handlers` | `func NewRouter(items models.ItemStore) http.Handler` |

`main.go` uses `db, err := database.Open(...)`, so it never names `T`. The
compiler checks every contract in the generated project. Each component's
package is its own, so identifier clashes are impossible (for example, sqlc's
generated `store.Item` and `models.Item` coexist).

Traits are plain key/value data. sqlite provides
`{sql-dialect: sqlite, conn: database/sql}`, and sqlc requires `sql-dialect` in
`[sqlite]` and `conn` in `[database/sql]`. The pgx or GORM components later
declare their own `conn` values, so incompatible pairs are rejected by data, not
by code.

Components may also add files at the project root (for example `sqlc.yaml`).
These are checked for path conflicts like any other file.

**Manifest schema.** `component.yaml` has exactly these keys. Unknown keys are
rejected, so a typo fails loading instead of being ignored:

| Key | Type | Kinds | Meaning |
|---|---|---|---|
| `name` | string, `^[a-z][a-z0-9-]*$` | all | Must equal the directory name. |
| `title` | string, required | all | Short display name, such as `SQLite`. |
| `description` | string, optional | all | One-sentence summary. |
| `slots` | map of slot name → slot | architecture only, at least one | Slot names match `^[a-z][a-z0-9_]*$` (they are template fields, `.Slots.<name>`), and `root` is reserved. |
| `slots.<name>.dir` | string, required | | Directory relative to the project root. It is a template over `.Project`, for example `cmd/{{.Project.Name}}`. |
| `slots.<name>.package` | string, optional | | Go package name. It defaults to the last element of the rendered `dir`, and is needed when that element is not a valid package name (`entry` sets `package: main`). |
| `slots.<name>.filled_by` | kind, required | | `architecture`, `framework`, `database`, or `data`. |
| `slots.<name>.required` | bool, default false | | The selection must fill the slot. |
| `provides` | map of trait → value | all | Traits this component provides, such as `sql-dialect: sqlite`. |
| `requires` | map of trait → list of values | all | For each trait, the values this component accepts, such as `sql-dialect: [sqlite]`. |

A component does not list the slots it fills. They follow from its layout: a
template under `files/<slot>/` fills `<slot>`, and `files/root/` adds files at
the project root (D3).

*Alternatives considered:*
- Code hooks spliced into a shared `main.go` (the first draft). Rejected: hook
  bodies would have to agree on local variable names and imports that nothing
  checks, and it becomes a second template language.
- Components described in Go code (a struct registry). Rejected: every new
  option would become an engine code change. YAML keeps components as data,
  and the loader validates them.

### D3. Component files and rendering

- Layout: `components/<kind>/<name>/component.yaml` plus
  `files/<slot>/<file>.tmpl` (into that slot's directory) and
  `files/root/<file>.tmpl` (project root).
- Every file under `components/` except `component.yaml`, `versions.yaml`, and
  `embed.go` must end in `.tmpl`, and the catalog test enforces this. As a
  result, no template is ever compiled, vetted, or linted as part of gobento.
  A leading `dot-` maps to `.` (`dot-gitignore.tmpl` → `.gitignore`), so no
  dotfiles are needed in `components/` and plain `//go:embed` works.
- Templates are loaded with `\r\n` normalized to `\n`. The repository's
  `.gitattributes` forces LF anyway.
- `text/template`, with fixed and documented data: `.Project` (Name, Module,
  GoVersion), `.Slots.<name>` (Dir, Package, Import), `.Traits`, and `.Versions`
  (module → version). This shape is the contributor API, documented in
  `docs/components.md`.
- Each component may add `notes.md.tmpl`. Composition renders every
  component's notes and appends them, in kind order, to the root `README.md`
  (for example, sqlc's "how to regenerate"). The template data stays as listed
  above. If any component has notes but no component produces `README.md`,
  that is a composition error.
- `.go` outputs go through `go/format.Source`. A failure there is reported as a
  composition error naming the component and the file.
- Output is a `Plan`: an ordered list of {slash path, mode, content}.
  Composition collects every error (unknown slot, unfilled required slot, two
  owners, root-path conflict, unmet trait, render or format error) and returns
  them together with `errors.Join`. It also rejects:
  - a slot filled by a component of a kind other than its `filled_by`;
  - two components providing different values for the same trait;
  - a component producing `go.mod` or `go.sum`, which generation writes (D5);
  - a root file that lies inside a slot directory;
  - two slots whose rendered directories are the same.
- A selection is *supported* when its traits are met and it composes. The
  catalog lists trait-compatible candidates; `compose.Supported` keeps those
  that compose without error, so the supported list is derived from the
  catalog alone (spec *Supported combinations*).

### D4. Version rules file

`components/versions.yaml` holds all version data, so a bump is a one-line,
reviewable, automatable change:

```yaml
go:
  range: ">=1.26 <1.28"
  default: "1.27"
modules:
  modernc.org/sqlite:
    components: [database/sqlite]
    role: library
    range: ">=v1.61.0 <v2.0.0"
    default: v1.61.1
    constraints:
      - when: ">=v1.61.0"
        require: { go: ">=1.26" }
  github.com/sqlc-dev/sqlc:
    components: [data/sqlc]
    role: tool                 # not in go.mod; fixed per gobento release
    range: "=v1.31.1"
    default: v1.31.1
```

(The values shown come from the version research on 2026-11-03. Range floors are
confirmed during apply.)

- A range is a space-separated list of comparators (`>=`, `>`, `<=`, `<`, `=`),
  all ANDed, compared with `golang.org/x/mod/semver`. Go versions are
  normalized internally (`1.27` → `v1.27.0`).
- Resolution: defaults, then `--go` and `--pin` overrides, then range checks,
  then every matching constraint. All errors are collected. Constraints are
  evaluated only when every chosen version is valid and in range, so a rejected
  pin does not also produce misleading constraint errors.
- Tool modules have exact ranges, so `--pin` on them is rejected.
- Only versions inside a range are offered, and CI tests both the defaults and
  the range floors (D9), so the range claim is tested.

*Alternative:* versions inside each component manifest. Rejected because
cross-component constraints would be spread across files, and automated bumps
would touch many places.

### D5. Generation pipeline

1. **Validate in memory:** name, module path (`module.CheckImportPath`; unlike
   `CheckPath`, it accepts dot-less paths such as `shop`), selection (an
   omitted kind with one value is auto-selected), versions (D4), and
   composition (D3).
2. **Preflight:** find `go`, and compare `go env GOVERSION` with the selected Go
   version. If the selected version is newer and `GOTOOLCHAIN=local`, fail.
   If it is newer otherwise, warn that Go will download it. Check the target
   directory (missing, or empty, is OK).
3. **Write** the Plan into `.gobento-tmp-<random>` in the target's parent (same
   filesystem). `go.mod` is rendered directly from data (module and `go`
   directive), not created by `go mod init`. The `go` directive is written as a
   full release version (`1.27` → `go 1.27.0`, that is `NormalizeGo` without
   the `v`). Go orders `1.26` before `1.26.0`, so a short directive would be
   raised by dependencies that declare `go 1.26.0`, such as modernc.org/sqlite,
   and step 5 would reject it.
4. **Resolve:** run one `go get mod1@v1 mod2@v2 ...`, then `go mod tidy`. All go
   commands get a controlled environment: `GOWORK=off`, `GOFLAGS=-mod=mod`, and
   the user's `GOPROXY`/`GONOSUMDB`/`GOPRIVATE` passed through unchanged.
5. **Re-check:** parse `go.mod` (`golang.org/x/mod/modfile`). The `go` directive
   and every direct requirement must equal the chosen values. A difference is
   reported as a missing rules constraint. A `toolchain` line added by the local
   toolchain is removed (`go mod edit -toolchain=none`), so the project does not
   depend on the generating machine.
6. **Place:**
   - If the target is missing, rename the temp dir to the target.
   - If the target exists and is empty (including `.`), move each top-level
     entry into it, then remove the temp dir.
   - On Windows, a rename or move that fails with an access-denied error is
     retried with short backoff, because antivirus and indexers briefly hold
     new files.
7. **Cleanup and interrupts:** everything runs under `signal.NotifyContext`
   (os.Interrupt), and go subprocesses use `exec.CommandContext`. On any error
   or interrupt, the temp dir is removed, along with any entries already moved
   into a pre-existing target. The target directory itself is never deleted.

### D6. sqlc output is pre-generated and shipped

The research showed sqlc does *not* need cgo, because it falls back to a
WebAssembly Postgres parser on Windows and without cgo. We still ship its output
instead of running it during generation, because running it would compile a
roughly 76 MB tool on the user's machine to regenerate a fixed example, which
is slow and network-heavy.

- The shipped `sqlc.yaml` sets `omit_sqlc_version: true`, so generated files
  contain no version numbers (spec: *Compatibility rules are data*). The
  package name comes from `.Slots.store.Package`.
- The generated README explains how to regenerate:
  `go run github.com/sqlc-dev/sqlc/cmd/sqlc@<.Versions value> generate`.
- **CI freshness check:** generate the reference project, run the pinned sqlc
  inside it, and fail if any file changed. This compares real output with real
  output, with no placeholder rewriting.
- The driver is `modernc.org/sqlite`, registered as `"sqlite"` (to be confirmed
  in source during apply). It builds without cgo on linux, darwin, and windows,
  for amd64 and arm64.

*Alternative:* a `tool` directive in the generated `go.mod`. Rejected because it
adds sqlc's large dependency tree to every generated project.

### D7. Generated project (first slice)

```
<name>/
  cmd/<name>/main.go           architecture: config from env (PORT, DATABASE_URL), wiring per D2, graceful shutdown
  internal/models/             architecture: Item, ItemStore
  internal/controllers/        framework: NewRouter, health + items handlers, tests with a fake ItemStore
  internal/database/           database: Open, embedded schema.sql applied at startup
  internal/store/              data: queries.sql, sqlc output, New() adapter, test against real SQLite
  sqlc.yaml, README.md, .gitignore, go.mod, go.sum
```

The SQLite test uses `t.TempDir()` and closes the DB in `t.Cleanup`, because
Windows cannot remove open files.

### D8. CLI

- `spf13/cobra` v1.10.2 provides commands, help, and completion. It is the de
  facto standard for Go CLIs.
- Flags on `new`: `--module`, `--arch`, `--framework`, `--db`, `--data`, `--go`,
  `--pin module=version` (repeatable), `--dir`, and `--dry-run`. The name
  `--pin` avoids confusion with cobra's root `--version`. `--dry-run` stops
  after pipeline step 1 of D5 (validation and composition, all in memory) and
  lists the Plan's paths. It skips the toolchain and target checks, so it
  works on machines without Go-proxy access, such as CI test runners. The website's configurator test uses it to prove that
  every command the site produces is accepted (change `add-project-website`).
- Error output lists every problem, one per line, each with a suggested fix
  (for example, the supported values).
- Other dependencies: `golang.org/x/mod` v0.42.0 (semver, module, modfile) and
  `go.yaml.in/yaml/v3` v3.0.6 (the maintained successor of the archived
  `gopkg.in/yaml.v3`; v4 is release-candidate only). huh is not needed in this
  change.

### D9. Testing and CI

- **Unit tests:** compat (ranges, resolution, every constraint error), catalog
  (the shipped catalog, the `.tmpl` rule, invalid fixtures, supported-selection
  listing including a trait mismatch), and compose (each error scenario, plus a
  second fixture data component proving the framework's files are
  byte-identical when only the data component changes).
- **Golden test:** compose the reference selection and compare the Plan to
  `internal/compose/testdata/golden/`. Regenerate it with `-update`.
- **Integration test** (`internal/generate`, skipped with `-short`): for every
  supported selection, generate twice (defaults, then range floors) into
  `t.TempDir()`. Then run `go build`, `go vet`, `go test`, `gofmt -l`, and
  `go mod verify`, assert the `go.mod` pins, and assert with `go list -deps`
  that the handlers package does not import the database or store packages.
- **CI** (extends the baseline workflow from `establish-repo-standards`):
  - Unit tests on Go 1.26 and 1.27.
  - The integration job on the newest Go in the rules range, on
    ubuntu/macos/windows, with module cache enabled.
  - The sqlc freshness job on ubuntu.

## Risks / Trade-offs

- [Generation needs the network] → It fails early with the `go` error, and
  nothing is left behind.
- [The integration matrix grows as the product of components] → Supported
  selections are derived from the catalog. Shard the job when it passes about
  15 minutes, and if needed run full `go test` only on defaults and
  build+vet on the rest.
- [The template data shape becomes a contributor API] → It is documented and
  covered by the golden test.
- [MVS raises a pinned version or the `go` directive] → Detected in D5 step 5
  and reported as a missing constraint.
- [Windows file locking during placement] → Retry with backoff. The target
  directory is never deleted.
- [The contract in D2 is specific to the example `Item` resource] → This is
  intended for the first slice. A later architecture may define a different
  contract, and components declare which architecture contracts they
  implement when a second architecture arrives.

## Open Questions

- Should the component catalog move to its own repository to release
  independently of the CLI? This does not affect this change.
