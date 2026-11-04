package compose

import (
	"io/fs"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Tales-of-zuna/gobento/components"
	"github.com/Tales-of-zuna/gobento/internal/catalog"
)

// viteStylesheets maps each shipped Vite frontend to the stylesheet its
// entry imports, which includes the point css-imports.
var viteStylesheets = map[string]string{
	"react":  "frontend/src/index.css",
	"svelte": "frontend/src/app.css",
	"vue":    "frontend/src/style.css",
}

// shippedCatalogWith returns the shipped catalog with the given files
// replaced.
func shippedCatalogWith(t *testing.T, changes map[string]string) *catalog.Catalog {
	t.Helper()
	fsys := fstest.MapFS{}
	err := fs.WalkDir(components.FS(), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(components.FS(), p)
		fsys[p] = &fstest.MapFile{Data: data}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range changes {
		fsys[name] = &fstest.MapFile{Data: []byte(data)}
	}
	cat, err := catalog.Load(fsys)
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}
	return cat
}

// vitePlan composes mvc/nethttp/sqlite/sqlc with the named frontend and
// CSS framework ("" for none).
func vitePlan(t *testing.T, cat *catalog.Catalog, frontend, css string) (*Plan, error) {
	t.Helper()
	sel := pick(t, cat, "mvc", "nethttp", "sqlite", "sqlc")
	fe, ok := cat.Component(catalog.KindFrontend, frontend)
	if !ok {
		t.Fatalf("no frontend %s", frontend)
	}
	sel = sel.With(fe)
	if css != "" {
		c, _ := cat.Component(catalog.KindCSS, css)
		sel = sel.With(c)
	}
	return planForErr(t, cat, sel)
}

// TestViteHook covers the spec scenario "Vite frontend with Tailwind": with
// each shipped Vite frontend, css/tailwind adds the Vite plugin to
// frontend/vite.config.ts and the Tailwind import to the top of the
// stylesheet through the frontend kind's points, and adds none of its files
// for go builds. Its npm packages come from the rules file
// (TestPrepareFrontend).
func TestViteHook(t *testing.T) {
	t.Parallel()

	cat, err := catalog.Load(components.FS())
	if err != nil {
		t.Fatal(err)
	}
	supported := Supported(cat)
	for frontend, stylesheet := range viteStylesheets {
		t.Run(frontend, func(t *testing.T) {
			t.Parallel()
			if !slices.ContainsFunc(supported, func(s catalog.Selection) bool {
				return s.String() == "mvc/nethttp/sqlite/sqlc/"+frontend+"/tailwind"
			}) {
				t.Fatalf("mvc/nethttp/sqlite/sqlc/%s/tailwind is not supported", frontend)
			}
			plan, err := vitePlan(t, cat, frontend, "tailwind")
			if err != nil {
				t.Fatalf("Compose() error = %v", err)
			}
			plain, err := vitePlan(t, cat, frontend, "")
			if err != nil {
				t.Fatalf("Compose() without a CSS framework error = %v", err)
			}
			content := func(p *Plan, path string) string {
				f, ok := p.File(path)
				if !ok {
					t.Fatalf("%s is not planned", path)
				}
				return string(f.Content)
			}

			// The plugin's import follows the frontend's imports, and its
			// call ends the plugins array.
			config, plainConfig := content(plan, "frontend/vite.config.ts"), content(plain, "frontend/vite.config.ts")
			const imp = "import { defineConfig } from 'vite'\nimport tailwindcss from '@tailwindcss/vite'\n"
			if !strings.Contains(config, imp) {
				t.Errorf("vite.config.ts does not import the plugin after the frontend's imports:\n%s", config)
			}
			if !strings.Contains(config, "    tailwindcss(),\n  ],\n") {
				t.Errorf("vite.config.ts does not call the plugin last in plugins:\n%s", config)
			}
			if strings.Contains(plainConfig, "tailwind") {
				t.Errorf("vite.config.ts without a CSS framework mentions Tailwind:\n%s", plainConfig)
			}
			if got, want := content(plan, stylesheet), "@import \"tailwindcss\";\n\n"+content(plain, stylesheet); got != want {
				t.Errorf("%s = %.80q..., want %.80q...: the plain stylesheet after the Tailwind import", stylesheet, got, want)
			}
			if readme := content(plan, "README.md"); !strings.Contains(readme, "through its Vite plugin, `@tailwindcss/vite`") {
				t.Errorf("README.md does not describe Tailwind through Vite:\n%s", readme)
			}
			for _, p := range []string{"Makefile", ".github/workflows/ci.yml", "Dockerfile"} {
				if content(plan, p) != content(plain, p) {
					t.Errorf("%s differs with Tailwind; the Vite build compiles it", p)
				}
			}
			for _, p := range plan.Paths() {
				if p == "tailwind.css" || strings.HasPrefix(p, "internal/tools/") || strings.HasSuffix(p, "static/styles.css") {
					t.Errorf("%s is planned with a Vite frontend", p)
				}
			}
		})
	}
}

// TestViteHookWithoutPoints covers the spec scenario "Vite frontend without
// the points": a Vite frontend whose templates do not include the points
// loses the CSS framework's contributions, so its selections with Tailwind
// fail to compose and are not supported, while the other frontends' still
// are.
func TestViteHookWithoutPoints(t *testing.T) {
	t.Parallel()

	cat := shippedCatalogWith(t, map[string]string{
		"frontend/react/files/root/frontend/vite.config.ts.tmpl": "export default {}\n",
	})
	_, err := vitePlan(t, cat, "react", "tailwind")
	if err == nil {
		t.Fatal("Compose() error = nil, want the lost contributions reported")
	}
	for _, want := range []string{
		"css/tailwind/contrib/vite-plugins.tmpl", `frontend/react includes no contributions to the point "vite-plugins"`,
		"css/tailwind/contrib/vite-config-imports.tmpl",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err, want)
		}
	}
	if _, err := vitePlan(t, cat, "react", ""); err != nil {
		t.Errorf("Compose() without a CSS framework error = %v", err)
	}
	var tailwind []string
	for _, s := range Supported(cat) {
		if s.CSS != nil && s.String() == "mvc/nethttp/sqlite/sqlc/"+s.Frontend.Name+"/tailwind" {
			tailwind = append(tailwind, s.Frontend.Name)
		}
	}
	if want := []string{"htmx", "svelte", "vue"}; !slices.Equal(tailwind, want) {
		t.Errorf("frontends supported with Tailwind = %v, want %v", tailwind, want)
	}
}
