package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
	"testing/fstest"

	"github.com/Tales-of-zuna/gobento/components"
	"github.com/Tales-of-zuna/gobento/internal/catalog"
)

// websiteCatalog is the committed export the website reads.
const websiteCatalog = "../../../website/src/generated/catalog.json"

// TestCommittedCatalogIsCurrent fails when the catalog or the exporter
// changed without regenerating the website's copy. Fix it with
// `make catalog`.
func TestCommittedCatalogIsCurrent(t *testing.T) {
	t.Parallel()

	var got bytes.Buffer
	if err := run(nil, &got); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	want, err := os.ReadFile(websiteCatalog)
	if err != nil {
		t.Fatalf("read the committed catalog: %v", err)
	}
	if !bytes.Equal(got.Bytes(), want) {
		t.Errorf("%s is out of date; run `make catalog` and commit the result", websiteCatalog)
	}
}

func TestExportShipped(t *testing.T) {
	t.Parallel()

	o, err := Export(components.FS())
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	var kinds []string
	for _, k := range o.Kinds {
		kinds = append(kinds, k.ID)
	}
	if want := []string{"architecture", "framework", "database", "data"}; !reflect.DeepEqual(kinds, want) {
		t.Errorf("kinds = %v, want %v", kinds, want)
	}

	ids := make(map[string]bool)
	for _, c := range o.Components {
		if c.ID != c.Kind+"/"+c.Name {
			t.Errorf("component id %q, want %s/%s", c.ID, c.Kind, c.Name)
		}
		if c.Provides == nil || c.Requires == nil {
			t.Errorf("%s: provides and requires must be empty maps, not null", c.ID)
		}
		ids[c.ID] = true
	}
	for _, m := range o.Modules {
		for _, id := range m.Components {
			if !ids[id] {
				t.Errorf("module %s: unknown component %q", m.Path, id)
			}
		}
		if m.Constraints == nil {
			t.Errorf("module %s: constraints must be an empty list, not null", m.Path)
		}
	}

	if len(o.Supported) == 0 {
		t.Fatal("no supported selections")
	}
	for _, s := range o.Supported {
		if len(s) != len(catalog.Kinds()) {
			t.Errorf("supported selection %v does not name one component per kind", s)
		}
		for kind, name := range s {
			if !ids[kind+"/"+name] {
				t.Errorf("supported selection %v: unknown component %s/%s", s, kind, name)
			}
		}
	}

	for name, pattern := range map[string]string{
		"projectName": o.Rules.ProjectName.Pattern,
		"modulePath":  o.Rules.ModulePath.Pattern,
	} {
		if _, err := regexp.Compile(pattern); err != nil {
			t.Errorf("rules.%s: %v", name, err)
		}
	}
	if len(o.Rules.ProjectName.Reserved) == 0 {
		t.Error("rules.projectName.reserved is empty")
	}
}

// TestExportFlattensConstraints checks that a constraint with several
// requirements becomes one exported constraint per requirement, that
// ranges keep their written form, and that trait conditions are exported
// (as an empty object when there are none).
func TestExportFlattensConstraints(t *testing.T) {
	t.Parallel()

	fsys := copyFS(t, os.DirFS("../../catalog/testdata/catalog"))
	fsys["versions.yaml"] = &fstest.MapFile{Data: []byte(`go:
  range: ">=1.26   <1.28"
  default: "1.27"
modules:
  example.com/a:
    components: [database/lite]
    traits: { sql-dialect: [sqlite, postgresql] }
    role: library
    range: ">=v1.0.0 <v2.0.0"
    default: v1.2.0
    constraints:
      - when: ">=v1.1.0"
        require: { go: ">=1.27", example.com/b: ">=v0.2.0" }
  example.com/b:
    components: [data/query]
    role: tool
    range: "=v0.3.0"
    default: v0.3.0
`)}

	o, err := Export(fsys)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}
	if o.Go.Range != ">=1.26 <1.28" {
		t.Errorf("go range = %q, want spaces collapsed", o.Go.Range)
	}
	want := []Constraint{
		{When: ">=v1.1.0", Requires: Requirement{Target: "example.com/b", Range: ">=v0.2.0"}},
		{When: ">=v1.1.0", Requires: Requirement{Target: "go", Range: ">=1.27"}},
	}
	if got := o.Modules[0].Constraints; !reflect.DeepEqual(got, want) {
		t.Errorf("constraints = %+v, want %+v", got, want)
	}
	if got := o.Modules[1]; got.Role != "tool" || got.Range != "=v0.3.0" || len(got.Constraints) != 0 {
		t.Errorf("tool module = %+v", got)
	}
	if got, want := o.Modules[0].Traits, map[string][]string{"sql-dialect": {"sqlite", "postgresql"}}; !reflect.DeepEqual(got, want) {
		t.Errorf("traits = %v, want %v", got, want)
	}
	if got := o.Modules[1].Traits; got == nil || len(got) != 0 {
		t.Errorf("traits of an unconditioned module = %#v, want an empty, non-nil map", got)
	}
}

func TestExportInvalidCatalog(t *testing.T) {
	t.Parallel()

	if _, err := Export(fstest.MapFS{}); err == nil {
		t.Error("Export(empty) error = nil, want an error")
	}
}

func TestKindInfo(t *testing.T) {
	t.Parallel()

	for _, k := range catalog.Kinds() {
		info, ok := kindInfo[k]
		if !ok || info.Title == "" || info.Description == "" {
			t.Errorf("kind %s needs a title and a description in kindInfo", k)
		}
	}
}

func TestSentence(t *testing.T) {
	t.Parallel()

	for in, want := range map[string]string{
		"start here":  "Start here.",
		"done.":       "Done.",
		"élan vital":  "Élan vital.",
		"Already Big": "Already Big.",
	} {
		if got := sentence(in); got != want {
			t.Errorf("sentence(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRunWritesFile(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	if err := run(nil, &stdout); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	name := filepath.Join(t.TempDir(), "catalog.json")
	if err := os.WriteFile(name, []byte("stale"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-o", name}, nil); err != nil {
		t.Fatalf("run(-o) error = %v", err)
	}
	got, err := os.ReadFile(filepath.Clean(name))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, stdout.Bytes()) {
		t.Error("the file written with -o differs from the stdout output")
	}
	if entries, _ := os.ReadDir(filepath.Dir(name)); len(entries) != 1 {
		t.Errorf("directory holds %d entries, want only catalog.json", len(entries))
	}

	if err := run([]string{"extra"}, nil); err == nil {
		t.Error("run(extra) error = nil, want an error")
	}
}

// copyFS reads every file of fsys into a MapFS that a test can modify.
func copyFS(t *testing.T, fsys fs.FS) fstest.MapFS {
	t.Helper()
	out := fstest.MapFS{}
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(fsys, p)
		out[p] = &fstest.MapFile{Data: data}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}
