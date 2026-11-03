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

// goldenDir holds the expected plan for the reference selection.
const goldenDir = "testdata/golden"

// TestGolden composes the reference selection mvc/nethttp/sqlite/sqlc from
// the shipped catalog with default versions and compares every planned file
// with testdata/golden. Run go test ./internal/compose -run TestGolden
// -update to accept intended changes, and review the diff.
func TestGolden(t *testing.T) {
	plan := referencePlan(t)

	if *update {
		if err := os.RemoveAll(goldenDir); err != nil {
			t.Fatal(err)
		}
		for _, f := range plan.Files {
			p := filepath.Join(goldenDir, filepath.FromSlash(f.Path))
			if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, f.Content, 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}

	var golden []string
	err := filepath.WalkDir(goldenDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(goldenDir, p)
		golden = append(golden, filepath.ToSlash(rel))
		return err
	})
	if err != nil {
		t.Fatalf("read golden files: %v (run with -update to create them)", err)
	}
	slices.Sort(golden)
	if got := plan.Paths(); !slices.Equal(got, golden) {
		t.Fatalf("planned paths differ from %s:\nplan:   %v\ngolden: %v", goldenDir, got, golden)
	}
	for _, f := range plan.Files {
		want, err := os.ReadFile(filepath.Join(goldenDir, filepath.FromSlash(f.Path)))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(f.Content, want) {
			t.Errorf("%s differs from %s; run with -update if the change is intended", f.Path, goldenDir)
		}
	}
}

// referencePlan composes the reference selection of the shipped catalog for
// the example project shop.
func referencePlan(t *testing.T) *Plan {
	t.Helper()
	cat, err := catalog.Load(components.FS())
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}
	sel := pick(t, cat, "mvc", "nethttp", "sqlite", "sqlc")
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

// TestShippedCatalogSupported checks that the shipped catalog supports
// exactly the first slice.
func TestShippedCatalogSupported(t *testing.T) {
	t.Parallel()

	cat, err := catalog.Load(components.FS())
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, s := range Supported(cat) {
		got = append(got, s.String())
	}
	if !slices.Equal(got, []string{"mvc/nethttp/sqlite/sqlc"}) {
		t.Errorf("Supported() = %v, want [mvc/nethttp/sqlite/sqlc]", got)
	}
}
