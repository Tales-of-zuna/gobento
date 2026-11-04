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

var update = flag.Bool("update", false, "rewrite testdata/golden* from the current templates")

// goldens maps each full-project golden directory to the reference selection
// it holds: one per architecture, with the same framework, database, and
// data access.
var goldens = []struct {
	dir string
	sel [4]string // architecture, framework, database, data
}{
	{dir: "testdata/golden", sel: [4]string{"mvc", "nethttp", "sqlite", "sqlc"}},
	{dir: "testdata/golden-vertical-slice", sel: [4]string{"vertical-slice", "nethttp", "sqlite", "sqlc"}},
}

// goldenComponentsDir holds the files each component contributes, in
// <kind>/<name>/ subdirectories.
const goldenComponentsDir = "testdata/golden-components"

// TestGolden composes the reference selection of each architecture from the
// shipped catalog with default versions and compares every planned file with
// its golden directory. Run go test ./internal/compose -run TestGolden
// -update to accept intended changes, and review the diff.
func TestGolden(t *testing.T) {
	for _, g := range goldens {
		t.Run(g.sel[0], func(t *testing.T) {
			checkGolden(t, g.dir, referencePlan(t, g.sel).Files)
		})
	}
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
					return slices.ContainsFunc(sel.Components(), func(s *catalog.Component) bool { return s.ID() == c.ID() })
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

// checkGolden compares files with the golden directory dir, after rewriting
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

// referencePlan composes the named selection of the shipped catalog for the
// example project shop, with default versions.
func referencePlan(t *testing.T, names [4]string) *Plan {
	t.Helper()
	cat := shippedCatalog(t)
	return planFor(t, cat, pick(t, cat, names[0], names[1], names[2], names[3]))
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

// TestShippedFrameworksAreLayoutIndependent covers the spec scenario
// "Shipped components are reused unchanged": for every framework, each file
// generated for vertical-slice equals the one generated for mvc once the slot
// directories, package names, and import paths of mvc are replaced by
// vertical-slice's.
func TestShippedFrameworksAreLayoutIndependent(t *testing.T) {
	t.Parallel()

	cat := shippedCatalog(t)
	for _, fw := range cat.Components(catalog.KindFramework) {
		t.Run(fw.Name, func(t *testing.T) {
			t.Parallel()

			mvc := referencePlan(t, [4]string{"mvc", fw.Name, "sqlite", "sqlc"})
			vslice := referencePlan(t, [4]string{"vertical-slice", fw.Name, "sqlite", "sqlc"})
			// Swap import paths and directories verbatim, and package names
			// only where Go code or a doc comment names the package.
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
				if f.Component != fw.ID() {
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
		})
	}
}

// TestShippedCatalogSupported lists the selections the shipped catalog
// supports.
func TestShippedCatalogSupported(t *testing.T) {
	t.Parallel()

	var got []string
	for _, s := range Supported(shippedCatalog(t)) {
		got = append(got, s.String())
	}
	want := []string{
		"mvc/chi/sqlite/sqlc", "mvc/gin/sqlite/sqlc", "mvc/nethttp/sqlite/sqlc",
		"vertical-slice/chi/sqlite/sqlc", "vertical-slice/gin/sqlite/sqlc", "vertical-slice/nethttp/sqlite/sqlc",
	}
	if !slices.Equal(got, want) {
		t.Errorf("Supported() = %v, want %v", got, want)
	}
}
