package catalog

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

func loadFixture(t *testing.T, name string) *Catalog {
	t.Helper()
	cat, err := Load(os.DirFS(filepath.Join("testdata", name)))
	if err != nil {
		t.Fatalf("Load(%s) error = %v", name, err)
	}
	return cat
}

func TestLoadValidCatalog(t *testing.T) {
	t.Parallel()

	cat := loadFixture(t, "catalog")

	arch, ok := cat.Component(KindArchitecture, "layered")
	if !ok {
		t.Fatal("architecture/layered not found")
	}
	if arch.ID() != "architecture/layered" || arch.Title != "Layered" || arch.Description == "" {
		t.Errorf("architecture = %+v", arch)
	}
	wantSlots := []Slot{
		{Name: "entry", Dir: "cmd/{{.Project.Name}}", Package: "main", FilledBy: KindArchitecture},
		{Name: "handlers", Dir: "internal/handlers", FilledBy: KindFramework, Required: true},
		{Name: "store", Dir: "internal/store", FilledBy: KindData, Required: true},
	}
	if !slices.Equal(arch.Slots, wantSlots) {
		t.Errorf("Slots = %+v, want %+v (sorted by name)", arch.Slots, wantSlots)
	}

	// The dot- prefix maps to "." in every path element, and .tmpl is
	// removed; files are sorted by slot and path.
	var got []string
	for _, f := range arch.Files {
		got = append(got, f.Slot+":"+f.Path)
	}
	want := []string{"entry:main.go", "root:.github/workflows/ci.yml", "root:.gitignore"}
	if !slices.Equal(got, want) {
		t.Errorf("Files = %v, want %v", got, want)
	}
	if src := arch.Files[0].Source; src != "architecture/layered/files/entry/main.go.tmpl" {
		t.Errorf("Files[0].Source = %q", src)
	}
	if fs := arch.FilledSlots(); !slices.Equal(fs, []string{"entry"}) {
		t.Errorf("FilledSlots() = %v, want [entry]", fs)
	}

	db, _ := cat.Component(KindDatabase, "lite")
	if db.Provides["sql-dialect"] != "sqlite" || db.Provides["conn"] != "database/sql" {
		t.Errorf("database Provides = %v", db.Provides)
	}
	query, _ := cat.Component(KindData, "query")
	if !slices.Equal(query.Requires["sql-dialect"], []string{"sqlite", "postgres"}) {
		t.Errorf("data Requires = %v", query.Requires)
	}
	if query.Notes != "## Query notes\n" {
		t.Errorf("Notes = %q", query.Notes)
	}

	var names []string
	for _, c := range cat.Components(KindData) {
		names = append(names, c.Name)
	}
	if !slices.Equal(names, []string{"pgonly", "query"}) {
		t.Errorf("Components(data) = %v, want [pgonly query]", names)
	}
	if _, ok := cat.Component(KindData, "missing"); ok {
		t.Error("Component(data, missing) found")
	}
	if m, ok := cat.Rules().Module("example.com/sqlitedriver"); !ok || m.Default != "v1.2.0" {
		t.Errorf("Rules().Module() = %+v, %v", m, ok)
	}
}

func TestLoadInvalidCatalogs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		fixture string
		want    []string
	}{
		{"non-tmpl-file", []string{"framework/web/files/handlers/router.go", "must end in .tmpl"}},
		{"dotfile", []string{"files/root/.gitignore.tmpl", `use the "dot-" prefix`}},
		{"unknown-kind", []string{"frontend: unknown component kind"}},
		{"root-file", []string{"README.md: unexpected file at the catalog root"}},
		{"name-mismatch", []string{"framework/web", `manifest name "http" must equal the directory name "web"`}},
		{"unknown-field", []string{"framework/web", "field provide not found"}},
		{"missing-title", []string{"framework/web: title is required"}},
		{"missing-manifest", []string{"framework/web", "component.yaml"}},
		{"slots-outside-architecture", []string{"framework/web: only architectures declare slots"}},
		{"bad-slots", []string{
			`slot root: "root" is reserved`,
			`slot handlers: filled_by "frontend"`,
			"slot store: dir is required",
			`slot models: package "Models"`,
			"slot entry: dir:",
		}},
		{"undeclared-slot-files", []string{`files/models: the architecture declares no slot "models"`}},
		{"no-slots", []string{"architecture/arch: an architecture must declare at least one slot"}},
		{"file-outside-slot", []string{"framework/web/files/router.go.tmpl", "templates belong in a slot directory"}},
		{"unexpected-component-entry", []string{`framework/web: unexpected entry "README.md"`}},
		{"bad-traits", []string{
			`provides: trait name "Dialect"`,
			"provides: trait conn has an empty value",
			"requires: trait sql-dialect lists no accepted values",
		}},
		{"unknown-rules-component", []string{"versions.yaml", `module example.com/lib: unknown component "database/missing"`}},
		{"unknown-rules-trait", []string{"versions.yaml", `module example.com/lib: trait condition on "sql-dialekt", which no component provides`}},
		{"missing-rules", []string{"versions.yaml"}},
		{"several-problems", []string{"framework/web: title is required", `database/db: manifest name "other"`}},
	}
	for _, tt := range tests {
		t.Run(tt.fixture, func(t *testing.T) {
			t.Parallel()

			_, err := Load(os.DirFS(filepath.Join("testdata", "invalid", tt.fixture)))
			if err == nil {
				t.Fatal("Load() error = nil, want an error")
			}
			for _, w := range tt.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("Load() error = %q\nwant it to contain %q", err, w)
				}
			}
		})
	}
}

func TestLoadNormalizesCRLF(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"versions.yaml":                              {Data: []byte("go: {range: \">=1.26\", default: \"1.27\"}\nmodules: {}\n")},
		"framework/web/component.yaml":               {Data: []byte("name: web\r\ntitle: Web\r\n")},
		"framework/web/files/handlers/a.go.tmpl":     {Data: []byte("package a\r\n\r\nvar x = 1\r\n")},
		"framework/web/notes.md.tmpl":                {Data: []byte("line one\r\nline two\r\n")},
		"framework/web/files/handlers/sub/b.go.tmpl": {Data: []byte("package sub\n")},
	}
	cat, err := Load(fsys)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	web, _ := cat.Component(KindFramework, "web")
	if got := web.Files[0].Template; got != "package a\n\nvar x = 1\n" {
		t.Errorf("Template = %q, want LF line endings", got)
	}
	if web.Notes != "line one\nline two\n" {
		t.Errorf("Notes = %q, want LF line endings", web.Notes)
	}
	if got := web.Files[1].Path; got != "sub/b.go" {
		t.Errorf("nested file Path = %q, want sub/b.go", got)
	}
}

func TestOutputPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{in: "main.go.tmpl", want: "main.go"},
		{in: "dot-gitignore.tmpl", want: ".gitignore"},
		{in: "dot-github/workflows/ci.yml.tmpl", want: ".github/workflows/ci.yml"},
		{in: "docs/dot-keep.tmpl", want: "docs/.keep"},
		{in: "README.md.tmpl", want: "README.md"},
		{in: "dotfile.tmpl", want: "dotfile"},
		{in: "main.go", wantErr: true},
		{in: ".tmpl", wantErr: true},
		{in: "dot-.tmpl", wantErr: true},
		{in: "dot-dot-/x.tmpl", want: ".dot-/x"},
		{in: "a//b.tmpl", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			got, err := OutputPath(tt.in)
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Fatalf("OutputPath(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("OutputPath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
