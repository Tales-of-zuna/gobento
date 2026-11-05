package catalog_test

import (
	"io/fs"
	"maps"
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
				if _, err := cat.Rules().Resolve(compat.Request{Components: ids, Traits: s.Traits()}); err != nil {
					t.Errorf("Resolve(defaults) for %s error = %v", s, err)
				}
			}
		})
	}
}

// TestShippedCatalogCandidates checks the shipped candidates: every
// architecture with every framework, database, and data access, each
// without a frontend and with every frontend, with and without Tailwind
// CSS; never Tailwind CSS without a frontend.
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
	var want []string
	for _, arch := range []string{"ddd", "mvc", "vertical-slice"} {
		for _, fw := range []string{"chi", "gin", "nethttp"} {
			for _, db := range []string{"postgres", "sqlite"} {
				for _, data := range []string{"gorm", "sqlc"} {
					sel := arch + "/" + fw + "/" + db + "/" + data
					want = append(want, sel)
					for _, ui := range []string{"htmx", "react", "svelte", "vue"} {
						want = append(want, sel+"/"+ui, sel+"/"+ui+"/tailwind")
					}
				}
			}
		}
	}
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

// viteShared lists the templates every Vite frontend ships with the same
// text, as slot:path; each frontend's stylesheet is compared separately,
// because its name follows the scaffolded entry's import.
var viteShared = []string{
	"web:handler.go",
	"web:handler_test.go",
	"web:dist/.gitignore",
	"root:frontend/src/api.ts",
}

// viteOwn lists, per Vite frontend, its page and the stylesheet its
// scaffolded entry imports, which gobento replaces.
var viteOwn = map[string][2]string{
	"react":  {"frontend/src/App.tsx", "frontend/src/index.css"},
	"svelte": {"frontend/src/App.svelte", "frontend/src/app.css"},
	"vue":    {"frontend/src/App.vue", "frontend/src/style.css"},
}

// TestViteFrontendsAlike covers the spec scenario "The three frontends are
// alike": every Vite frontend provides frontend-build: vite, is scaffolded
// by the same scaffolder into frontend/, ships the same web package, API
// client, stylesheet, and contributions, byte for byte, and owns the same
// set of files apart from its page.
func TestViteFrontendsAlike(t *testing.T) {
	t.Parallel()

	cat, err := catalog.Load(components.FS())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	files := func(c *catalog.Component) map[string]string {
		out := make(map[string]string)
		for _, f := range c.Files {
			out[f.Slot+":"+f.Path] = f.Template
		}
		return out
	}
	ref, _ := cat.Component(catalog.KindFrontend, "react")
	refFiles := files(ref)
	refCSS := refFiles["root:"+viteOwn["react"][1]]
	if refCSS == "" {
		t.Fatal("react has no stylesheet")
	}
	for name, own := range viteOwn {
		c, ok := cat.Component(catalog.KindFrontend, name)
		if !ok {
			t.Errorf("frontend %s not found", name)
			continue
		}
		if c.Provides["frontend-build"] != "vite" {
			t.Errorf("%s provides frontend-build %q, want vite", name, c.Provides["frontend-build"])
		}
		if sc := c.Scaffold; sc == nil || sc.With != "create-vite" || sc.Dir != "frontend" || sc.Template != name+"-ts" {
			t.Errorf("%s scaffold = %+v, want create-vite with %s-ts into frontend", name, sc, name)
		}
		got := files(c)
		for _, p := range viteShared {
			if got[p] == "" || got[p] != refFiles[p] {
				t.Errorf("%s: %s differs from react's or is missing", name, p)
			}
		}
		if got["root:"+own[1]] != refCSS {
			t.Errorf("%s: %s differs from react's stylesheet", name, own[1])
		}
		if got["root:"+own[0]] == "" || got["root:frontend/vite.config.ts"] == "" {
			t.Errorf("%s: missing %s or frontend/vite.config.ts", name, own[0])
		}
		if len(got) != len(refFiles) {
			t.Errorf("%s has %d templates, react %d; the Vite frontends own the same files", name, len(got), len(refFiles))
		}
		if !maps.Equal(c.Contributions, ref.Contributions) {
			t.Errorf("%s: contributions differ from react's", name)
		}
	}
	htmx, _ := cat.Component(catalog.KindFrontend, "htmx")
	if htmx.Provides["frontend-build"] != "go" {
		t.Errorf("htmx provides frontend-build %q, want go", htmx.Provides["frontend-build"])
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

// TestTailwindReleasedTogether covers the spec scenario "Tailwind's npm
// packages from the rules": the npm packages Vite frontends install with
// css/tailwind have the version of the standalone CLI, since Tailwind
// releases them together.
func TestTailwindReleasedTogether(t *testing.T) {
	t.Parallel()

	cat, err := catalog.Load(components.FS())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	rules := cat.Rules()
	cli, ok := rules.Binary("tailwindcss")
	if !ok {
		t.Fatal("no tailwindcss binary in the rules")
	}
	for _, name := range []string{"tailwindcss", "@tailwindcss/vite"} {
		p, ok := rules.NPMPackage(name)
		switch {
		case !ok:
			t.Errorf("npm package %s is not in the rules", name)
		case p.Version != cli.Version || p.Role != compat.NPMDevDependency || !slices.Equal(p.Components, []string{"css/tailwind"}):
			t.Errorf("npm %s = %+v, want a dev-dependency of css/tailwind at the CLI's version %s", name, p, cli.Version)
		}
	}
}
