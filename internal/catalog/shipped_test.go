package catalog_test

import (
	"io/fs"
	"os"
	"path/filepath"
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

			supported := cat.Supported()
			if len(supported) == 0 {
				t.Fatal("Supported() is empty")
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

func TestShippedCatalogFirstSlice(t *testing.T) {
	t.Parallel()

	cat, err := catalog.Load(components.FS())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	var got []string
	for _, s := range cat.Supported() {
		got = append(got, s.String())
	}
	if len(got) != 1 || got[0] != "mvc/nethttp/sqlite/sqlc" {
		t.Errorf("Supported() = %v, want [mvc/nethttp/sqlite/sqlc]", got)
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
