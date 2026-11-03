package compose

import (
	"maps"
	"testing"
	"testing/fstest"

	"github.com/Tales-of-zuna/gobento/internal/catalog"
)

// baseFixture is a small catalog with two architectures that place the same
// slots in different directories, one framework, one database, and two
// interchangeable data components. Tests copy it and add or remove files.
func baseFixture() map[string]string {
	const arch1 = `name: layered
title: Layered
slots:
  entry:    { dir: "cmd/{{.Project.Name}}", package: main, filled_by: architecture }
  models:   { dir: internal/models,         filled_by: architecture }
  handlers: { dir: internal/controllers,    filled_by: framework, required: true }
  database: { dir: internal/database,       filled_by: database,  required: true }
  store:    { dir: internal/store,          filled_by: data,      required: true }
`
	const arch2 = `name: hexagonal
title: Hexagonal
slots:
  entry:    { dir: "cmd/{{.Project.Name}}", package: main, filled_by: architecture }
  models:   { dir: internal/domain,         filled_by: architecture }
  handlers: { dir: internal/http/handlers,  filled_by: framework, required: true }
  database: { dir: internal/adapters/db,    filled_by: database,  required: true }
  store:    { dir: internal/adapters/repo,  filled_by: data,      required: true }
`
	const entry = `package main

import (
	"{{.Slots.database.Import}}"
	"{{.Slots.handlers.Import}}"
	"{{.Slots.store.Import}}"
)

func main() {
	db, _ := {{.Slots.database.Package}}.Open()
	_ = {{.Slots.handlers.Package}}.NewRouter({{.Slots.store.Package}}.New(db))
}
`
	const models = `package {{.Slots.models.Package}}

type Item struct{ Name string }

type ItemStore interface{ List() []Item }
`
	const readme = `# {{.Project.Name}}

Module {{.Project.Module}}, Go {{.Project.GoVersion}}.
`
	return map[string]string{
		"versions.yaml": `go: { range: ">=1.26 <1.28", default: "1.27" }
modules:
  example.com/driver:
    components: [database/lite]
    role: library
    range: ">=v1.0.0"
    default: v1.2.0
`,
		"architecture/layered/component.yaml":                  arch1,
		"architecture/layered/files/entry/main.go.tmpl":        entry,
		"architecture/layered/files/models/models.go.tmpl":     models,
		"architecture/layered/files/root/README.md.tmpl":       readme,
		"architecture/layered/files/root/dot-gitignore.tmpl":   "/bin/\n",
		"architecture/hexagonal/component.yaml":                arch2,
		"architecture/hexagonal/files/entry/main.go.tmpl":      entry,
		"architecture/hexagonal/files/models/models.go.tmpl":   models,
		"architecture/hexagonal/files/root/README.md.tmpl":     readme,
		"architecture/hexagonal/files/root/dot-gitignore.tmpl": "/bin/\n",

		"framework/web/component.yaml": "name: web\ntitle: Web\n",
		"framework/web/files/handlers/router.go.tmpl": `package {{.Slots.handlers.Package}}

import (
	"net/http"

	"{{.Slots.models.Import}}"
)

// NewRouter serves items.
func NewRouter(items {{.Slots.models.Package}}.ItemStore) http.Handler {
	_ = items
	return http.NewServeMux()
}
`,
		"framework/web/files/handlers/router_test.go.tmpl": "package {{.Slots.handlers.Package}}\n",

		"database/lite/component.yaml": `name: lite
title: Lite
provides:
  sql-dialect: sqlite
  conn: database/sql
`,
		"database/lite/files/database/db.go.tmpl": `package {{.Slots.database.Package}}

// Driver {{index .Versions "example.com/driver"}} speaks {{index .Traits "sql-dialect"}} over {{.Traits.conn}}.
type DB struct{}

func Open() (*DB, error) { return &DB{}, nil }
`,

		"data/query/component.yaml": `name: query
title: Query
requires:
  sql-dialect: [sqlite]
  conn: [database/sql]
`,
		"data/query/notes.md.tmpl":              "## Query\n\nRegenerate with query {{.Project.Name}}.\n",
		"data/query/files/root/query.yaml.tmpl": "package: {{.Slots.store.Package}}\n",
		"data/query/files/store/store.go.tmpl": `package {{.Slots.store.Package}}

import "{{.Slots.models.Import}}"

type Store struct{}

func New(any) Store { return Store{} }

func (Store) List() []{{.Slots.models.Package}}.Item { return nil }
`,

		"data/orm/component.yaml": `name: orm
title: ORM
requires:
  conn: [database/sql]
`,
		"data/orm/files/store/repo.go.tmpl": `package {{.Slots.store.Package}}

type Repo struct{}

func New(any) Repo { return Repo{} }
`,
	}
}

// loadCatalog loads files as a catalog.
func loadCatalog(t *testing.T, files map[string]string) *catalog.Catalog {
	t.Helper()
	fsys := make(fstest.MapFS, len(files))
	for name, data := range files {
		fsys[name] = &fstest.MapFile{Data: []byte(data)}
	}
	cat, err := catalog.Load(fsys)
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}
	return cat
}

// withFiles returns a copy of baseFixture with the given files added or,
// for empty values, removed.
func withFiles(changes map[string]string) map[string]string {
	files := baseFixture()
	maps.Copy(files, changes)
	for name, data := range changes {
		if data == "" {
			delete(files, name)
		}
	}
	return files
}

// pick returns the selection of the named components.
func pick(t *testing.T, cat *catalog.Catalog, arch, framework, database, data string) catalog.Selection {
	t.Helper()
	get := func(kind catalog.Kind, name string) *catalog.Component {
		if name == "" {
			return nil
		}
		c, ok := cat.Component(kind, name)
		if !ok {
			t.Fatalf("component %s/%s not in fixture", kind, name)
		}
		return c
	}
	return catalog.Selection{
		Architecture: get(catalog.KindArchitecture, arch),
		Framework:    get(catalog.KindFramework, framework),
		Database:     get(catalog.KindDatabase, database),
		Data:         get(catalog.KindData, data),
	}
}

// input returns the Input for a selection with a fixed example project.
func input(sel catalog.Selection) Input {
	return Input{
		Project:   Project{Name: "shop", Module: "example.com/shop", GoVersion: "1.27"},
		Selection: sel,
		Versions:  map[string]string{"example.com/driver": "v1.2.0"},
	}
}
