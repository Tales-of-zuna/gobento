package compose

import (
	"bytes"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
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
	plan := shippedPlan(t, "sqlite")

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

// shippedPlan composes mvc/nethttp/<database>/sqlc from the shipped catalog
// with default versions for the example project shop. With "sqlite", it is
// the reference selection.
func shippedPlan(t *testing.T, database string) *Plan {
	t.Helper()
	cat, err := catalog.Load(components.FS())
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}
	sel := pick(t, cat, "mvc", "nethttp", database, "sqlc")
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
// sqlc on both databases.
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
	want := []string{"mvc/nethttp/postgres/sqlc", "mvc/nethttp/sqlite/sqlc"}
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
