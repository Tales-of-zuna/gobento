package compose

import (
	"bytes"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Tales-of-zuna/gobento/components"
	"github.com/Tales-of-zuna/gobento/internal/catalog"
	"github.com/Tales-of-zuna/gobento/internal/compat"
)

// postgresFixture is a test-only PostgreSQL database component, shared with
// the integration test in internal/generate.
const postgresFixture = "../generate/testdata/postgres"

// shippedWithPostgres loads the shipped catalog with the PostgreSQL
// database fixture added.
func shippedWithPostgres(t *testing.T) *catalog.Catalog {
	t.Helper()
	m := fstest.MapFS{}
	for _, fsys := range []fs.FS{components.FS(), os.DirFS(postgresFixture)} {
		err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			data, err := fs.ReadFile(fsys, p)
			m[p] = &fstest.MapFile{Data: data}
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	cat, err := catalog.Load(m)
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}
	return cat
}

// composeDefaults composes sel for the example project shop with the
// default versions of its modules.
func composeDefaults(t *testing.T, cat *catalog.Catalog, sel catalog.Selection) (*Plan, compat.Resolved) {
	t.Helper()
	var ids []string
	for _, c := range sel.Components() {
		ids = append(ids, c.ID())
	}
	res, err := cat.Rules().Resolve(compat.Request{Components: ids, Traits: sel.Traits()})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	plan, err := Compose(Input{
		Project:   Project{Name: "shop", Module: "example.com/shop", GoVersion: "1.27"},
		Selection: sel,
		Versions:  res.Versions(),
	})
	if err != nil {
		t.Fatalf("Compose(%s) error = %v", sel, err)
	}
	return plan, res
}

func planFile(t *testing.T, plan *Plan, path string) string {
	t.Helper()
	for _, f := range plan.Files {
		if f.Path == path {
			return string(f.Content)
		}
	}
	t.Fatalf("plan has no %s; it has %v", path, plan.Paths())
	return ""
}

// TestGORMDialectFromTraits covers the spec "Dialect-specific code is
// chosen from traits": the GORM store and its modules follow sql-dialect.
func TestGORMDialectFromTraits(t *testing.T) {
	t.Parallel()

	cat := shippedWithPostgres(t)
	tests := []struct {
		database   string
		want       []string
		notWant    []string
		wantModule string
		notModule  string
	}{
		{
			database:   "sqlite",
			want:       []string{`"github.com/libtnb/sqlite"`, "sqlite.New(sqlite.Config{Conn: db})", `filepath.Join(t.TempDir(), "test.db")`},
			notWant:    []string{"postgres", "DATABASE_URL", "AutoMigrate"},
			wantModule: "github.com/libtnb/sqlite",
			notModule:  "gorm.io/driver/postgres",
		},
		{
			database:   "postgres",
			want:       []string{`"gorm.io/driver/postgres"`, "postgres.New(postgres.Config{Conn: db})", `os.Getenv("DATABASE_URL")`},
			notWant:    []string{"libtnb", "AutoMigrate", "t.TempDir()"},
			wantModule: "gorm.io/driver/postgres",
			notModule:  "github.com/libtnb/sqlite",
		},
	}
	for _, tt := range tests {
		t.Run(tt.database, func(t *testing.T) {
			t.Parallel()

			plan, res := composeDefaults(t, cat, pick(t, cat, "mvc", "nethttp", tt.database, "gorm"))
			code := planFile(t, plan, "internal/store/store.go") + planFile(t, plan, "internal/store/store_test.go")
			for _, w := range tt.want {
				if !strings.Contains(code, w) {
					t.Errorf("store code lacks %q", w)
				}
			}
			for _, nw := range tt.notWant {
				if strings.Contains(code, nw) {
					t.Errorf("store code contains %q", nw)
				}
			}
			versions := res.Versions()
			if _, ok := versions[tt.wantModule]; !ok {
				t.Errorf("modules %v lack %s", versions, tt.wantModule)
			}
			if _, ok := versions[tt.notModule]; ok {
				t.Errorf("modules %v include %s", versions, tt.notModule)
			}
			if !strings.Contains(planFile(t, plan, "README.md"), "## Data access with GORM") {
				t.Error("README.md lacks the GORM notes")
			}
		})
	}
}

// TestShippedSwappingDataLeavesFrameworkUntouched covers the spec scenario
// "Swapping one component leaves others untouched" for the shipped sqlc and
// gorm components.
func TestShippedSwappingDataLeavesFrameworkUntouched(t *testing.T) {
	t.Parallel()

	cat, err := catalog.Load(components.FS())
	if err != nil {
		t.Fatal(err)
	}
	frameworkFiles := func(data string) map[string][]byte {
		plan, _ := composeDefaults(t, cat, pick(t, cat, "mvc", "nethttp", "sqlite", data))
		out := make(map[string][]byte)
		for _, f := range plan.Files {
			if f.Component == "framework/nethttp" {
				out[f.Path] = f.Content
			}
		}
		return out
	}
	withSQLC, withGORM := frameworkFiles("sqlc"), frameworkFiles("gorm")
	if len(withSQLC) == 0 || len(withSQLC) != len(withGORM) {
		t.Fatalf("framework files: %d with sqlc, %d with gorm", len(withSQLC), len(withGORM))
	}
	for p, content := range withSQLC {
		if !bytes.Equal(content, withGORM[p]) {
			t.Errorf("%s differs between sqlc and gorm", p)
		}
	}
}
