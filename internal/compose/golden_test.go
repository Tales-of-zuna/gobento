package compose

import (
	"bytes"
	"flag"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/Tales-of-zuna/gobento/components"
	"github.com/Tales-of-zuna/gobento/internal/catalog"
	"github.com/Tales-of-zuna/gobento/internal/compat"
)

var update = flag.Bool("update", false, "rewrite testdata/golden from the current templates")

// goldens maps each golden directory to the reference selection it holds:
// one per architecture, with the same framework, database, and data access.
var goldens = []struct {
	dir string
	sel [4]string // architecture, framework, database, data
}{
	{dir: "testdata/golden", sel: [4]string{"mvc", "nethttp", "sqlite", "sqlc"}},
	{dir: "testdata/golden-vertical-slice", sel: [4]string{"vertical-slice", "nethttp", "sqlite", "sqlc"}},
}

// TestGolden composes the reference selection of each architecture from the
// shipped catalog with default versions and compares every planned file
// with its golden directory. Run go test ./internal/compose -run TestGolden
// -update to accept intended changes, and review the diff.
func TestGolden(t *testing.T) {
	for _, g := range goldens {
		t.Run(g.sel[0], func(t *testing.T) {
			checkGolden(t, g.dir, referencePlan(t, g.sel))
		})
	}
}

// checkGolden compares plan with the files in goldenDir, after rewriting
// them with -update.
func checkGolden(t *testing.T, goldenDir string, plan *Plan) {
	t.Helper()
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
		want, err := os.ReadFile(filepath.Join(goldenDir, filepath.FromSlash(f.Path))) //nolint:gosec // G304: golden testdata path
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(f.Content, want) {
			t.Errorf("%s differs from %s; run with -update if the change is intended", f.Path, goldenDir)
		}
	}
}

// referencePlan composes the named selection of the shipped catalog for
// the example project shop, with default versions.
func referencePlan(t *testing.T, names [4]string) *Plan {
	t.Helper()
	cat, err := catalog.Load(components.FS())
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}
	sel := pick(t, cat, names[0], names[1], names[2], names[3])
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

// TestShippedFrameworkIsLayoutIndependent covers the spec scenario "Shipped
// components are reused unchanged": every framework file generated for
// vertical-slice equals the one generated for mvc once the slot directories,
// package names, and import paths of mvc are replaced by vertical-slice's.
func TestShippedFrameworkIsLayoutIndependent(t *testing.T) {
	t.Parallel()

	mvc := referencePlan(t, goldens[0].sel)
	vslice := referencePlan(t, goldens[1].sel)
	// Swap import paths and directories verbatim, and package names only
	// where Go code or a doc comment names the package.
	var pairs []string
	type rename struct {
		re *regexp.Regexp
		to string
	}
	var renames []rename
	for _, name := range slices.Sorted(maps.Keys(mvc.Slots)) {
		from, to := mvc.Slots[name], vslice.Slots[name]
		pairs = append(pairs, from.Import, to.Import, from.Dir+"/", to.Dir+"/")
		renames = append(renames,
			rename{regexp.MustCompile(`\b` + from.Package + `\.`), to.Package + "."},
			rename{regexp.MustCompile(`\b([Pp]ackage) ` + from.Package + `\b`), "$1 " + to.Package})
	}
	paths := strings.NewReplacer(pairs...)
	swap := func(s string) string {
		s = paths.Replace(s)
		for _, r := range renames {
			s = r.re.ReplaceAllString(s, r.to)
		}
		return s
	}

	var compared int
	for _, f := range mvc.Files {
		if f.Component != "framework/nethttp" {
			continue
		}
		want := swap(string(f.Content))
		got, ok := vslice.File(paths.Replace(f.Path))
		if !ok {
			t.Errorf("vertical-slice plan has no counterpart of %s", f.Path)
			continue
		}
		compared++
		if string(got.Content) != want {
			t.Errorf("%s differs from mvc's %s beyond slot names", got.Path, f.Path)
		}
	}
	if compared == 0 {
		t.Fatal("no framework files compared")
	}
}

// TestShippedCatalogSupported checks that the shipped catalog supports
// exactly the reference selection of each architecture.
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
	want := []string{"mvc/nethttp/sqlite/sqlc", "vertical-slice/nethttp/sqlite/sqlc"}
	if !slices.Equal(got, want) {
		t.Errorf("Supported() = %v, want %v", got, want)
	}
}
