package catalog_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/Tales-of-zuna/gobento/components"
	"github.com/Tales-of-zuna/gobento/internal/catalog"
	"github.com/Tales-of-zuna/gobento/internal/compat"
)

// TestShippedCatalog loads the catalog compiled into gobento. Load checks
// every manifest, the .tmpl-only file rule, the rules file, and that every
// component referenced by the rules exists.
func TestShippedCatalog(t *testing.T) {
	t.Parallel()

	sources := map[string]fs.FS{
		"embedded": components.FS(),
		// The directory on disk also contains files the embed patterns
		// would skip, so the file rules are checked for those too.
		"on disk": os.DirFS(filepath.Join("..", "..", "components")),
	}
	for name, fsys := range sources {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cat, err := catalog.Load(fsys)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			for _, m := range cat.Rules().Modules {
				for _, id := range m.Components {
					if !hasComponent(cat, id) {
						t.Errorf("rules module %s references missing component %s", m.Path, id)
					}
				}
			}

			supported := cat.Candidates()
			if len(supported) == 0 {
				t.Fatal("Candidates() is empty")
			}
			for _, s := range supported {
				var ids []string
				for _, c := range s.Components() {
					ids = append(ids, c.ID())
				}
				if _, err := cat.Rules().Resolve(compat.Request{Components: ids}); err != nil {
					t.Errorf("Resolve(defaults) for %s error = %v", s, err)
				}
			}
		})
	}
}

func TestShippedCatalogCandidates(t *testing.T) {
	t.Parallel()

	cat, err := catalog.Load(components.FS())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	var got []string
	for _, s := range cat.Candidates() {
		got = append(got, s.String())
	}
	want := []string{"mvc/chi/sqlite/sqlc", "mvc/gin/sqlite/sqlc", "mvc/nethttp/sqlite/sqlc"}
	if !slices.Equal(got, want) {
		t.Errorf("Candidates() = %v, want %v", got, want)
	}
}

// apiTest is the framework-independent API test every framework component
// ships in its handlers slot.
const apiTest = "router_test.go"

// TestFrameworksShareTheAPITest enforces the spec requirement "Framework
// components serve one API contract": every framework component ships the
// same API test, byte for byte, so the generated projects of all frameworks
// are held to the same behavior.
func TestFrameworksShareTheAPITest(t *testing.T) {
	t.Parallel()

	cat, err := catalog.Load(components.FS())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	var ref *catalog.Component
	var refTest string
	for _, c := range cat.Components(catalog.KindFramework) {
		i := slices.IndexFunc(c.Files, func(f catalog.File) bool { return f.Slot == "handlers" && f.Path == apiTest })
		if i < 0 {
			t.Errorf("%s has no handlers/%s.tmpl; every framework ships the shared API test", c.ID(), apiTest)
			continue
		}
		if ref == nil {
			ref, refTest = c, c.Files[i].Template
			continue
		}
		if c.Files[i].Template != refTest {
			t.Errorf("handlers/%s.tmpl of %s differs from %s; framework components must share one API test", apiTest, c.ID(), ref.ID())
		}
	}
}

func hasComponent(cat *catalog.Catalog, id string) bool {
	for _, kind := range catalog.Kinds() {
		for _, c := range cat.Components(kind) {
			if c.ID() == id {
				return true
			}
		}
	}
	return false
}
