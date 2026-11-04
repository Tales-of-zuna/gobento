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
// data access, and the first one again with a frontend, which shows how the
// architecture mounts it.
var goldens = []struct {
	dir      string
	sel      [4]string // architecture, framework, database, data
	frontend string    // "" for none
}{
	{dir: "testdata/golden", sel: [4]string{"mvc", "nethttp", "sqlite", "sqlc"}},
	{dir: "testdata/golden-vertical-slice", sel: [4]string{"vertical-slice", "nethttp", "sqlite", "sqlc"}},
	{dir: "testdata/golden-ddd", sel: [4]string{"ddd", "nethttp", "sqlite", "sqlc"}},
	{dir: "testdata/golden-htmx", sel: [4]string{"mvc", "nethttp", "sqlite", "sqlc"}, frontend: "htmx"},
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
		t.Run(filepath.Base(g.dir), func(t *testing.T) {
			checkGolden(t, g.dir, referencePlanWith(t, g.sel, g.frontend).Files)
		})
	}
}

// TestGoldenComponents composes, for every shipped component, the first
// supported selection that contains it, and compares the files that
// component contributes with testdata/golden-components/<kind>/<name>. This
// reviews the output of every component without one golden project per
// combination. Components other than architectures are composed with the
// reference architecture (that of goldens[0]) when it supports them, so
// adding an architecture leaves their goldens as they are.
func TestGoldenComponents(t *testing.T) {
	cat := shippedCatalog(t)
	supported := Supported(cat)
	refArch := goldens[0].sel[0]
	var ids []string
	for _, kind := range catalog.Kinds() {
		for _, c := range cat.Components(kind) {
			ids = append(ids, c.ID())
			t.Run(c.ID(), func(t *testing.T) {
				contains := func(sel catalog.Selection) bool {
					return slices.ContainsFunc(sel.Components(), func(s *catalog.Component) bool { return s.ID() == c.ID() })
				}
				i := slices.IndexFunc(supported, func(sel catalog.Selection) bool {
					return contains(sel) && (kind == catalog.KindArchitecture || sel.Architecture.Name == refArch)
				})
				if i < 0 {
					i = slices.IndexFunc(supported, contains)
				}
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

// referencePlan composes the named selection of the shipped catalog, without
// a frontend, for the example project shop, with default versions.
func referencePlan(t *testing.T, names [4]string) *Plan {
	t.Helper()
	return referencePlanWith(t, names, "")
}

// referencePlanWith is referencePlan with the named frontend, or none for "".
func referencePlanWith(t *testing.T, names [4]string, frontend string) *Plan {
	t.Helper()
	cat := shippedCatalog(t)
	sel := pick(t, cat, names[0], names[1], names[2], names[3])
	if frontend != "" {
		ui, ok := cat.Component(catalog.KindFrontend, frontend)
		if !ok {
			t.Fatalf("frontend %s is not shipped", frontend)
		}
		sel = sel.With(ui)
	}
	return planFor(t, cat, sel)
}

// planFor composes sel with default versions for the example project shop.
func planFor(t *testing.T, cat *catalog.Catalog, sel catalog.Selection) *Plan {
	t.Helper()
	var ids []string
	for _, c := range sel.Components() {
		ids = append(ids, c.ID())
	}
	res, err := cat.Rules().Resolve(compat.Request{Components: ids, Traits: sel.Traits()})
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
		Assets:    res.AssetVersions(),
	})
	if err != nil {
		t.Fatalf("Compose() error = %v", err)
	}
	return plan
}

// TestShippedFrameworksAreLayoutIndependent covers the spec scenarios
// "Shipped components are reused unchanged" and "Shipped components are
// reused unchanged in ddd": for every framework and every architecture other
// than mvc, each file generated for that architecture equals the one
// generated for mvc once the slot directories, package names, and import
// paths of mvc are replaced by the architecture's.
func TestShippedFrameworksAreLayoutIndependent(t *testing.T) {
	t.Parallel()

	cat := shippedCatalog(t)
	for _, arch := range cat.Components(catalog.KindArchitecture) {
		if arch.Name == "mvc" {
			continue
		}
		for _, fw := range cat.Components(catalog.KindFramework) {
			t.Run(arch.Name+"/"+fw.Name, func(t *testing.T) {
				t.Parallel()
				checkLayoutIndependent(t, arch.Name, fw)
			})
		}
	}
}

// checkLayoutIndependent compares the files of framework fw under the
// architecture arch with those under mvc, as described at
// TestShippedFrameworksAreLayoutIndependent.
func checkLayoutIndependent(t *testing.T, arch string, fw *catalog.Component) {
	t.Helper()

	mvc := referencePlan(t, [4]string{"mvc", fw.Name, "sqlite", "sqlc"})
	other := referencePlan(t, [4]string{arch, fw.Name, "sqlite", "sqlc"})
	// Swap import paths and directories verbatim, and package names
	// only where Go code or a doc comment names the package.
	var pairs []string
	type rename struct {
		re *regexp.Regexp
		to string
	}
	var renames []rename
	for _, name := range slices.Sorted(maps.Keys(mvc.Slots)) {
		from, to := mvc.Slots[name], other.Slots[name]
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
		got, ok := other.File(paths.Replace(f.Path))
		if !ok {
			t.Errorf("%s plan has no counterpart of %s", arch, f.Path)
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

// TestShippedCatalogSupported lists the selections the shipped catalog
// supports.
func TestShippedCatalogSupported(t *testing.T) {
	t.Parallel()

	var got []string
	for _, s := range Supported(shippedCatalog(t)) {
		got = append(got, s.String())
	}
	var want []string
	for _, arch := range []string{"ddd", "mvc", "vertical-slice"} {
		for _, fw := range []string{"chi", "gin", "nethttp"} {
			for _, db := range []string{"postgres", "sqlite"} {
				for _, data := range []string{"gorm", "sqlc"} {
					sel := arch + "/" + fw + "/" + db + "/" + data
					want = append(want, sel)
					if arch != "ddd" {
						want = append(want, sel+"/htmx")
					}
				}
			}
		}
	}
	if !slices.Equal(got, want) {
		t.Errorf("Supported() = %v, want %v", got, want)
	}
}

// TestShippedDatabasesShareEntryAndHandlers covers the spec scenario "Entry
// point is the same for every database": the Go files of the architecture
// and the framework do not depend on the database.
func TestShippedDatabasesShareEntryAndHandlers(t *testing.T) {
	t.Parallel()

	sqlite, postgres := shippedPlan(t, "sqlite"), shippedPlan(t, "postgres")
	compared := 0
	for _, f := range sqlite.Files {
		if !strings.HasSuffix(f.Path, ".go") || (f.Component != "architecture/mvc" && f.Component != "framework/nethttp") {
			continue
		}
		g, ok := postgres.File(f.Path)
		if !ok || !bytes.Equal(f.Content, g.Content) {
			t.Errorf("%s differs between the sqlite and postgres selections", f.Path)
		}
		compared++
	}
	if compared == 0 {
		t.Fatal("no architecture or framework Go files compared")
	}
}

// TestShippedSQLcDialects covers the spec scenarios "sqlc on PostgreSQL" and
// "sqlc on SQLite" as far as composition goes: the engine and the
// placeholder follow the sql-dialect trait. The CI job sqlc-fresh checks the
// generated code against sqlc itself.
func TestShippedSQLcDialects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		database, engine, placeholder, driver string
	}{
		{"sqlite", "engine: sqlite", "VALUES (?1)", `_ "modernc.org/sqlite"`},
		{"postgres", "engine: postgresql", "VALUES ($1)", `_ "github.com/jackc/pgx/v5/stdlib"`},
	}
	for _, tt := range tests {
		t.Run(tt.database, func(t *testing.T) {
			t.Parallel()

			plan := shippedPlan(t, tt.database)
			for path, want := range map[string]string{
				"sqlc.yaml":                             tt.engine,
				"internal/store/sqlcgen/queries.sql.go": tt.placeholder,
				"internal/database/database.go":         tt.driver,
				"internal/store/store_test.go":          `"example.com/shop/internal/database/dbtest"`,
			} {
				f, ok := plan.File(path)
				if !ok {
					t.Errorf("%s is not planned", path)
					continue
				}
				if !strings.Contains(string(f.Content), want) {
					t.Errorf("%s does not contain %q:\n%s", path, want, f.Content)
				}
			}
			for _, path := range []string{"internal/database/dbtest/dbtest.go", "internal/database/schema.sql"} {
				if _, ok := plan.File(path); !ok {
					t.Errorf("%s is not planned", path)
				}
			}
		})
	}
}

// shippedPlan composes mvc/nethttp/<database>/sqlc from the shipped catalog
// with default versions for the example project shop. With "sqlite", it is
// the reference selection.
func shippedPlan(t *testing.T, database string) *Plan {
	t.Helper()
	return referencePlan(t, [4]string{"mvc", "nethttp", database, "sqlc"})
}
