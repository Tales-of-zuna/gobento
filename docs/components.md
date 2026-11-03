# Writing components

gobento builds a project from four components, one of each kind:

| Kind | Fills | First option |
|---|---|---|
| `architecture` | the entry point, the shared types, and the layout | `mvc` |
| `framework` | the HTTP handlers | `nethttp` |
| `database` | the database connection and schema | `sqlite` |
| `data` | the data access that implements the shared interfaces | `sqlc` |

There is no template per combination. An architecture declares *slots*, each
of the other components fills a slot, and *traits* say which components fit
together. This page explains how that works and how to add a component.

## Layout

Everything lives under [`components/`](../components), which holds data only:

```
components/
  versions.yaml                     version-compatibility rules
  embed.go                          embeds the catalog into the gobento binary
  <kind>/<name>/
    component.yaml                  the manifest
    notes.md.tmpl                   optional notes appended to the README
    files/
      <slot>/...                    templates for the slot's directory
      root/...                      templates for the project root
```

For example, `components/framework/nethttp/files/handlers/router.go.tmpl`
becomes `internal/controllers/router.go` when the architecture puts the
`handlers` slot in `internal/controllers`.

### File rules

- Every file except `component.yaml`, `versions.yaml`, and `embed.go` ends in
  `.tmpl`. The suffix is removed in the output. Because no template is a `.go`
  file, gobento never compiles, vets, or lints them by accident.
- No file or directory name starts with `.`. Write `dot-` instead: a leading
  `dot-` in any path element becomes `.`, so `files/root/dot-gitignore.tmpl`
  becomes `.gitignore` and `files/root/dot-github/workflows/ci.yml.tmpl` becomes
  `.github/workflows/ci.yml`.
- Templates are read with CRLF line endings converted to LF.

The catalog loader enforces these rules, and its tests load the shipped
catalog, so a mistake fails `make test`.

## The manifest

`component.yaml` has exactly these keys. Unknown keys are rejected, so a typo
fails loading instead of being ignored.

| Key | Kinds | Meaning |
|---|---|---|
| `name` | all | Must equal the directory name and match `^[a-z][a-z0-9-]*$`. |
| `title` | all | Required. A short display name, such as `SQLite`. |
| `description` | all | Optional. One sentence. |
| `slots` | architecture only | The slots the architecture declares (at least one). |
| `provides` | all | Traits this component provides: trait → value. |
| `requires` | all | Traits this component needs: trait → list of accepted values. |

Each slot has these keys:

| Key | Meaning |
|---|---|
| `dir` | Required. The slot's directory, relative to the project root. It is a template over `.Project`, for example `cmd/{{.Project.Name}}`. |
| `package` | Optional. The Go package name. It defaults to the last element of the rendered `dir`; set it when that is not a valid package name, as the entry point does with `package: main`. |
| `filled_by` | Required. The kind that fills the slot: `architecture`, `framework`, `database`, or `data`. |
| `required` | Optional, default `false`. The selection must fill the slot. |

Slot names match `^[a-z][a-z0-9_]*$`, because templates use them as field
names (`.Slots.handlers`), and `root` is reserved.

This is a complete architecture manifest. The documentation test loads it with
the catalog loader, so it stays correct:

<!-- example-manifest -->
```yaml
name: mvc
title: Model-View-Controller
description: Models, controllers, and a data store in separate packages, wired together in cmd/<name>.
slots:
  entry:    { dir: "cmd/{{.Project.Name}}", package: main, filled_by: architecture }
  models:   { dir: internal/models,         filled_by: architecture }
  handlers: { dir: internal/controllers,    filled_by: framework, required: true }
  database: { dir: internal/database,       filled_by: database,  required: true }
  store:    { dir: internal/store,          filled_by: data,      required: true }
```

A component does not list the slots it fills. They follow from its layout: a
template under `files/<slot>/` fills `<slot>`.

## Slots and contracts

Each slot is one directory, which means one Go package, and it has exactly one
owner. Two components can therefore never write into the same package or the
same file, and their identifiers cannot clash. The owner may put its own
subpackages below the slot directory; sqlc's generated code lives in
`internal/store/sqlcgen`, for example. Other slots never import them.

The architecture owns the entry point and the *contract*: the shared types and
the functions each slot must provide. For `mvc`:

| Slot | Contract |
|---|---|
| `models` (architecture) | `type Item struct{ ID int64; Name string; CreatedAt time.Time }` and `type ItemStore interface{ List(ctx) ([]Item, error); Create(ctx, name string) (Item, error) }` |
| `database` | `func Open(ctx context.Context, dsn string) (T, error)`, where `T` has a `Close` method |
| `store` | `func New(db T) models.ItemStore` |
| `handlers` | `func NewRouter(items models.ItemStore) http.Handler` |

The entry point calls these functions, and the handlers import only the
`models` package. The Go compiler checks the contract in every generated
project, and gobento's integration test checks that the handlers package does
not depend on the database or store packages.

## Traits

A component never names another component. It depends on traits instead,
which are plain key/value data. `sqlite` provides:

```yaml
provides:
  sql-dialect: sqlite
  conn: database/sql
  sql-schema: schema.sql
```

and `sqlc` accepts it:

```yaml
requires:
  sql-dialect: [sqlite]
  conn: [database/sql]
  sql-schema: [schema.sql]
```

A selection is refused if a required trait is missing or has a value the
component does not accept, or if two components provide different values for
the same trait. Traits also carry the few facts one component needs about
another's files: `sqlc.yaml` finds the schema at
`{{.Slots.database.Dir}}/{{index .Traits "sql-schema"}}`.

## Template data

Templates use Go's [`text/template`](https://pkg.go.dev/text/template). Every
template, including slot directories and notes, gets the same data:

| Field | Example | Meaning |
|---|---|---|
| `.Project.Name` | `shop` | The project name |
| `.Project.Module` | `example.com/shop` | The module path |
| `.Project.GoVersion` | `1.27.0` | The `go` directive of the generated `go.mod` |
| `.Slots.<name>.Dir` | `internal/controllers` | The slot's directory |
| `.Slots.<name>.Package` | `controllers` | The slot's package name |
| `.Slots.<name>.Import` | `example.com/shop/internal/controllers` | The slot's import path |
| `.Traits` | `{{index .Traits "sql-dialect"}}` | The traits of the selection |
| `.Versions` | `{{index .Versions "modernc.org/sqlite"}}` | Chosen version per module |

Use `index` for keys that contain `-`, `.`, or `/`. A missing key is an error,
not an empty string, so `.Slots.handler` (a typo) fails composition.

A handler template therefore starts like this:

```go
package {{.Slots.handlers.Package}}

import "{{.Slots.models.Import}}"

func NewRouter(items {{.Slots.models.Package}}.ItemStore) http.Handler {
```

Composition also:

- formats every `.go` output with `go/format`, and reports a failure with the
  component and the file;
- appends each component's rendered `notes.md.tmpl` to the root `README.md`,
  in kind order;
- reports every problem at once: unknown or unfilled slots, slots with two
  owners or an owner of the wrong kind, two slots in one directory, file
  conflicts, root files inside a slot directory, components that produce
  `go.mod` or `go.sum` (gobento writes those), unmet traits, and templates that
  fail to render.

Never write a version number in a template. Use `.Versions`, which comes from
the rules file.

## Versions

[`components/versions.yaml`](../components/versions.yaml) holds every version a
generated project uses:

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
```

- A range is a space-separated list of comparators (`>=`, `>`, `<=`, `<`, `=`)
  that must all hold.
- `role: library` modules are required in the generated `go.mod`; users may
  choose any version inside the range with `--pin`. `role: tool` modules, such
  as sqlc, are run with `go run` and never added to `go.mod`; their range is a
  single exact version, fixed per gobento release.
- A constraint applies when the module's chosen version is inside `when`, and
  then requires `go` or another module to be inside a range.
- gobento rejects the generation if `go mod tidy` would raise any chosen
  version or the `go` directive, and reports the missing constraint.

### Bumping a version

1. Change `default` (and `range`, if needed) in `versions.yaml`.
2. Run `make test` and `go test ./internal/generate -run TestIntegration`
   (it needs network access). The integration test generates every supported
   selection with the defaults and with the range floors.
3. If a template's output changes, update the golden files with
   `go test ./internal/compose -run TestGolden -update` and review the diff.

## Adding a component

1. Open a [component request](https://github.com/Tales-of-zuna/gobento/issues/new?template=component_request.yml)
   and, for a new kind of behavior, an OpenSpec change (see
   [CONTRIBUTING.md](../CONTRIBUTING.md)).
2. Create `components/<kind>/<name>/component.yaml` with the traits it provides
   and requires.
3. Add templates under `files/<slot>/` for the slot its kind fills, and under
   `files/root/` for project-root files. Implement the slot's contract and
   import other slots only through the `models` contract.
4. Add tests to the templates (`*_test.go.tmpl`); they run in every generated
   project. Tests that use files close them in `t.Cleanup`, because Windows
   cannot remove open files.
5. Add the modules it uses to `versions.yaml`, with `components` naming it.
6. Run `make lint test`, then the integration test. Every supported selection
   containing the new component is generated, built, vetted, and tested.

### sqlc output

The `sqlc` component ships sqlc's generated code instead of running sqlc
during generation. After changing `queries.sql` or the schema, regenerate it
in a generated project with the pinned sqlc version, copy the files in
`internal/store/sqlcgen/` back to `components/data/sqlc/files/store/sqlcgen/`
with the `.tmpl` suffix, and update the golden files. CI checks that the
shipped output matches what the pinned sqlc generates.
