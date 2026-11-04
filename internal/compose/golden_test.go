package compose

import (
	"bytes"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/Tales-of-zuna/gobento/components"
	"github.com/Tales-of-zuna/gobento/internal/catalog"
	"github.com/Tales-of-zuna/gobento/internal/compat"
)

var update = flag.Bool("update", false, "rewrite testdata/golden from the current templates")

// goldenDir holds the expected plan for the reference selection, and
// goldenComponentsDir the files each component contributes, in
// <kind>/<name>/ subdirectories.
const (
	goldenDir           = "testdata/golden"
	goldenComponentsDir = "testdata/golden-components"
)

// TestGolden composes the reference selection mvc/nethttp/sqlite/sqlc from
// the shipped catalog with default versions and compares every planned file
// with testdata/golden. Run go test ./internal/compose -run TestGolden
// -update to accept intended changes, and review the diff.
func TestGolden(t *testing.T) {
	cat := shippedCatalog(t)
	plan := planFor(t, cat, pick(t, cat, "mvc", "nethttp", "sqlite", "sqlc"))
	checkGolden(t, goldenDir, plan.Files)
}

// TestGoldenComponents composes, for every shipped component, the first
// supported selection that contains it, and compares the files that
// component contributes with testdata/golden-components/<kind>/<name>. This
// reviews the output of every component without one golden project per
// combination.
func TestGoldenComponents(t *testing.T) {
	cat := shippedCatalog(t)
	supported := Supported(cat)
	var ids []string
	for _, kind := range catalog.Kinds() {
		for _, c := range cat.Components(kind) {
			ids = append(ids, c.ID())
			t.Run(c.ID(), func(t *testing.T) {
				i := slices.IndexFunc(supported, func(sel catalog.Selection) bool {
					return slices.Contains(sel.Components(), c)
				})
				if i < 0 {
					t.Fatalf("%s is in no supported selection", c.ID())
				}
				var files []File
				for _, f := range planFor(t, cat, supported[i]).Files {
					if f.Component == c.ID() {
						files = append(files, f)
					}
				}
				checkGolden(t, filepath.Join(goldenComponentsDir, filepath.FromSlash(c.ID())), files)
			})
		}
	}

	// Every golden directory must belong to a shipped component, so the
	// files of a removed or renamed component do not linger.
	dirs, err := filepath.Glob(filepath.Join(goldenComponentsDir, "*", "*"))
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range dirs {
		rel, err := filepath.Rel(goldenComponentsDir, dir)
		if err != nil {
			t.Fatal(err)
		}
		if id := filepath.ToSlash(rel); !slices.Contains(ids, id) {
			if *update {
				if err := os.RemoveAll(dir); err != nil {
					t.Fatal(err)
				}
				continue
			}
			t.Errorf("%s has golden files but is not a shipped component; run with -update to remove them", id)
		}
	}
}

// checkGolden compares files with the golden directory dir, or rewrites
// dir from files with -update.
func checkGolden(t *testing.T, dir string, files []File) {
	t.Helper()
	if *update {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
		for _, f := range files {
			p := filepath.Join(dir, filepath.FromSlash(f.Path))
			if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, f.Content, 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}

	var golden []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		golden = append(golden, filepath.ToSlash(rel))
		return err
	})
	if err != nil {
		t.Fatalf("read golden files: %v (run with -update to create them)", err)
	}
	slices.Sort(golden)
	var got []string
	for _, f := range files {
		got = append(got, f.Path)
	}
	if !slices.Equal(got, golden) {
		t.Fatalf("planned paths differ from %s:\nplan:   %v\ngolden: %v", dir, got, golden)
	}
	for _, f := range files {
		want, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(f.Path))) //nolint:gosec // G304: path under testdata
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(f.Content, want) {
			t.Errorf("%s differs from %s; run with -update if the change is intended", f.Path, dir)
		}
	}
}

// shippedCatalog loads the catalog compiled into gobento.
func shippedCatalog(t *testing.T) *catalog.Catalog {
	t.Helper()
	cat, err := catalog.Load(components.FS())
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}
	return cat
}

// planFor composes sel with default versions for the example project shop.
func planFor(t *testing.T, cat *catalog.Catalog, sel catalog.Selection) *Plan {
	t.Helper()
	var ids []string
	for _, c := range sel.Components() {
		ids = append(ids, c.ID())
	}
	res, err := cat.Rules().Resolve(compat.Request{Components: ids})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	goVersion, err := compat.GoDirective(res.Go)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := Compose(Input{
		Project:   Project{Name: "shop", Module: "example.com/shop", GoVersion: goVersion},
		Selection: sel,
		Versions:  res.Versions(),
	})
	if err != nil {
		t.Fatalf("Compose() error = %v", err)
	}
	return plan
}

// TestShippedCatalogSupported lists the selections the shipped catalog
// supports.
func TestShippedCatalogSupported(t *testing.T) {
	t.Parallel()

	var got []string
	for _, s := range Supported(shippedCatalog(t)) {
		got = append(got, s.String())
	}
	want := []string{"mvc/chi/sqlite/sqlc", "mvc/gin/sqlite/sqlc", "mvc/nethttp/sqlite/sqlc"}
	if !slices.Equal(got, want) {
		t.Errorf("Supported() = %v, want %v", got, want)
	}
}
