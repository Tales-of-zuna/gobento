package compose

import (
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Tales-of-zuna/gobento/components"
	"github.com/Tales-of-zuna/gobento/internal/catalog"
)

// viteCatalog returns the shipped catalog plus the fixture Vite frontend in
// testdata/vite-frontend, with the given files changed or, with empty
// values, removed. It stands in for the Vite frontends that are not shipped
// yet, to test the hook of the css kind for frontend-build: vite.
func viteCatalog(t *testing.T, changes map[string]string) *catalog.Catalog {
	t.Helper()
	fsys := fstest.MapFS{}
	add := func(src fs.FS) {
		err := fs.WalkDir(src, ".", func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			data, err := fs.ReadFile(src, p)
			fsys[p] = &fstest.MapFile{Data: data}
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	add(components.FS())
	add(os.DirFS("testdata/vite-frontend"))
	for name, data := range changes {
		if data == "" {
			delete(fsys, name)
			continue
		}
		fsys[name] = &fstest.MapFile{Data: []byte(data)}
	}
	cat, err := catalog.Load(fsys)
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}
	return cat
}

// vitePlan composes mvc/nethttp/sqlite/sqlc with the fixture Vite frontend
// and the named CSS framework ("" for none).
func vitePlan(t *testing.T, cat *catalog.Catalog, css string) (*Plan, error) {
	t.Helper()
	sel := pick(t, cat, "mvc", "nethttp", "sqlite", "sqlc")
	vite, _ := cat.Component(catalog.KindFrontend, "vite")
	sel = sel.With(vite)
	if css != "" {
		c, _ := cat.Component(catalog.KindCSS, css)
		sel = sel.With(c)
	}
	return planForErr(t, cat, sel)
}

// TestViteHook covers the spec scenario "Vite frontend with Tailwind": with
// a frontend whose frontend-build is vite, css/tailwind contributes the Vite
// plugin, the stylesheet import, and the npm packages of the pinned release
// to the frontend kind's points, and adds none of its files for go builds.
func TestViteHook(t *testing.T) {
	t.Parallel()

	cat := viteCatalog(t, nil)
	if !slices.ContainsFunc(Supported(cat), func(s catalog.Selection) bool { return s.String() == "mvc/nethttp/sqlite/sqlc/vite/tailwind" }) {
		t.Fatal("mvc/nethttp/sqlite/sqlc/vite/tailwind is not supported")
	}
	plan, err := vitePlan(t, cat, "tailwind")
	if err != nil {
		t.Fatalf("Compose() error = %v", err)
	}
	version := cat.Rules().Binaries[0].Version
	for p, want := range map[string]string{
		"frontend/vite.config.js": `import { defineConfig } from "vite";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [
    tailwindcss(),
  ],
});
`,
		"frontend/src/index.css":        "@import \"tailwindcss\";\n\nbody {\n  margin: 0;\n}\n",
		"frontend/dev-dependencies.txt": "tailwindcss@" + version + "\n@tailwindcss/vite@" + version + "\n",
		"internal/views/handler.go":     "",
		"README.md":                     "through its Vite plugin, `@tailwindcss/vite`",
		".github/workflows/ci.yml":      "",
		"Makefile":                      "",
	} {
		f, ok := plan.File(p)
		if !ok {
			t.Errorf("%s is not planned; paths: %v", p, plan.Paths())
			continue
		}
		got := string(f.Content)
		switch {
		case strings.HasSuffix(p, ".js") || strings.HasSuffix(p, ".css") || strings.HasSuffix(p, ".txt"):
			if got != want {
				t.Errorf("%s =\n%s\nwant\n%s", p, got, want)
			}
		case want != "" && !strings.Contains(got, want):
			t.Errorf("%s does not contain %q", p, want)
		}
		if p == "Makefile" || p == ".github/workflows/ci.yml" {
			if strings.Contains(got, "tailwind") {
				t.Errorf("%s mentions Tailwind with a Vite frontend:\n%s", p, got)
			}
		}
	}
	for _, p := range plan.Paths() {
		if p == "tailwind.css" || strings.HasPrefix(p, "internal/tools/") || strings.HasPrefix(p, "internal/views/static/") {
			t.Errorf("%s is planned with a Vite frontend", p)
		}
	}

	// Without a CSS framework, the frontend's files have no Tailwind, and
	// a template that renders nothing produces no file.
	plain, err := vitePlan(t, cat, "")
	if err != nil {
		t.Fatalf("Compose() without a CSS framework error = %v", err)
	}
	if f, _ := plain.File("frontend/vite.config.js"); strings.Contains(string(f.Content), "tailwind") {
		t.Errorf("vite.config.js without a CSS framework:\n%s", f.Content)
	}
	if _, ok := plain.File("frontend/dev-dependencies.txt"); ok {
		t.Error("frontend/dev-dependencies.txt is planned, but its template renders nothing without a CSS framework")
	}
}

// TestViteHookWithoutPoints covers the spec scenario "Vite frontend without
// the points": a Vite frontend whose templates do not include the points
// loses the CSS framework's contributions, so its selections with Tailwind
// fail to compose and are not supported.
func TestViteHookWithoutPoints(t *testing.T) {
	t.Parallel()

	cat := viteCatalog(t, map[string]string{
		"frontend/vite/files/root/frontend/vite.config.js.tmpl": "export default {};\n",
	})
	_, err := vitePlan(t, cat, "tailwind")
	if err == nil {
		t.Fatal("Compose() error = nil, want the lost contributions reported")
	}
	for _, want := range []string{"css/tailwind/contrib/vite-plugins.tmpl", `includes no contributions to the point "vite-plugins"`, "css/tailwind/contrib/vite-config-imports.tmpl"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err, want)
		}
	}
	if slices.ContainsFunc(Supported(cat), func(s catalog.Selection) bool {
		return path.Base(s.String()) == "tailwind" && s.Frontend.Name == "vite"
	}) {
		t.Error("vite with tailwind is supported although the frontend drops its contributions")
	}
}
