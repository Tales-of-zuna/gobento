package compose

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strings"
	"testing"

	"github.com/Tales-of-zuna/gobento/internal/catalog"
)

// TestShippedFrontendLeavesFrameworkUntouched covers the spec scenario
// "Frontend leaves the framework untouched": for every framework and
// frontend, the framework's Go files are the same with and without the
// frontend, so the API router and its shared test are unchanged.
func TestShippedFrontendLeavesFrameworkUntouched(t *testing.T) {
	t.Parallel()

	cat := shippedCatalog(t)
	for _, ui := range cat.Components(catalog.KindFrontend) {
		for _, arch := range cat.Components(catalog.KindArchitecture) {
			for _, fw := range cat.Components(catalog.KindFramework) {
				names := [4]string{arch.Name, fw.Name, "sqlite", "sqlc"}
				without, with := referencePlan(t, names), referencePlanWith(t, names, ui.Name)
				compared := 0
				for _, f := range without.Files {
					if f.Component != fw.ID() || !strings.HasSuffix(f.Path, ".go") {
						continue
					}
					g, ok := with.File(f.Path)
					if !ok || !bytes.Equal(f.Content, g.Content) {
						t.Errorf("%s/%s/%s: %s differs with the frontend", ui.Name, arch.Name, fw.Name, f.Path)
					}
					compared++
				}
				if compared == 0 {
					t.Errorf("%s/%s/%s: no framework Go files compared", ui.Name, arch.Name, fw.Name)
				}
			}
		}
	}
}

// TestShippedFrontendMount checks that every shipped architecture mounts the
// API under the API prefix and the web handler at the root when a frontend
// is selected, and leaves the web slot empty otherwise.
func TestShippedFrontendMount(t *testing.T) {
	t.Parallel()

	for _, a := range shippedCatalog(t).Components(catalog.KindArchitecture) {
		arch := a.Name
		names := [4]string{arch, "nethttp", "sqlite", "sqlc"}
		with := referencePlanWith(t, names, "htmx")
		web := with.Slots["web"]
		if !web.Filled {
			t.Fatalf("%s: web slot not filled with htmx", arch)
		}
		main, _ := with.File(with.Slots["entry"].Dir + "/main.go")
		for _, want := range []string{
			`mux.Handle("/api/", http.StripPrefix("/api", `,
			`mux.Handle("/", ` + web.Package + `.NewHandler(`,
			`"` + web.Import + `"`,
		} {
			if !strings.Contains(string(main.Content), want) {
				t.Errorf("%s: main.go with htmx does not contain %q", arch, want)
			}
		}

		without := referencePlan(t, names)
		if without.Slots["web"].Filled {
			t.Errorf("%s: web slot filled without a frontend", arch)
		}
		main, _ = without.File(without.Slots["entry"].Dir + "/main.go")
		if strings.Contains(string(main.Content), web.Import) || strings.Contains(string(main.Content), "/api") {
			t.Errorf("%s: main.go without a frontend mentions the web package or /api", arch)
		}
		for _, p := range without.Paths() {
			if strings.HasPrefix(p, web.Dir+"/") {
				t.Errorf("%s: %s planned without a frontend", arch, p)
			}
		}
	}
}

// TestShippedAssets covers the spec scenario "Vendored file edited": every
// asset in the rules file is shipped, with the recorded checksum, by each
// component listed for it.
func TestShippedAssets(t *testing.T) {
	t.Parallel()

	cat := shippedCatalog(t)
	supported := Supported(cat)
	if len(cat.Rules().Assets) == 0 {
		t.Fatal("the shipped rules list no assets")
	}
	for _, a := range cat.Rules().Assets {
		for _, id := range a.Components {
			i := slices.IndexFunc(supported, func(sel catalog.Selection) bool {
				return slices.ContainsFunc(sel.Components(), func(c *catalog.Component) bool { return c.ID() == id })
			})
			if i < 0 {
				t.Errorf("asset %s: %s is in no supported selection", a.Name, id)
				continue
			}
			found := false
			for _, f := range planFor(t, cat, supported[i]).Files {
				sum := sha256.Sum256(f.Content)
				if f.Component == id && hex.EncodeToString(sum[:]) == a.SHA256 {
					found = true
				}
			}
			if !found {
				t.Errorf("asset %s: %s ships no file with sha256 %s; update the file or the checksum in versions.yaml", a.Name, id, a.SHA256)
			}
		}
	}
}

// TestShippedViteTooling covers the spec scenarios "templ watch files only
// with htmx", "Dependency updates for the frontend", and the Vite tooling of
// "Build with make": the files of the base layer get the Vite frontends'
// contributions, and htmx keeps its own.
func TestShippedViteTooling(t *testing.T) {
	t.Parallel()

	names := [4]string{"mvc", "nethttp", "sqlite", "sqlc"}
	for _, ui := range []string{"react", "svelte", "vue"} {
		plan := referencePlanWith(t, names, ui)
		want := map[string][]string{
			".gitignore":               nil,
			".github/dependabot.yml":   {"- package-ecosystem: npm\n    directory: /frontend\n"},
			".dockerignore":            {"frontend/node_modules\n", "internal/views/dist/app\n"},
			"Makefile":                 {"build: ## Build the binary into bin/\n\t$(NPM) --prefix frontend ci\n\t$(NPM) --prefix frontend run build\n\t$(GO) build", "\ndev: ##"},
			"Dockerfile":               {"FROM node:26-alpine AS frontend\n", "COPY --from=frontend /src/internal/views/dist/app internal/views/dist/app\n"},
			".github/workflows/ci.yml": {"actions/setup-node@", "node-version: 26\n", "run: npm ci\n", "run: npm run build\n"},
		}
		for path, wants := range want {
			f, ok := plan.File(path)
			if !ok {
				t.Errorf("%s: %s not planned", ui, path)
				continue
			}
			for _, w := range wants {
				if !strings.Contains(string(f.Content), w) {
					t.Errorf("%s: %s does not contain %q", ui, path, w)
				}
			}
			if path == ".gitignore" && strings.Contains(string(f.Content), "_templ.txt") {
				t.Errorf("%s: .gitignore lists templ watch files", ui)
			}
		}
	}
	htmx := referencePlanWith(t, names, "htmx")
	if f, _ := htmx.File(".gitignore"); !strings.Contains(string(f.Content), "*_templ.txt") {
		t.Error("htmx: .gitignore does not list *_templ.txt")
	}
	for _, path := range []string{"Makefile", "Dockerfile", ".github/dependabot.yml"} {
		if f, _ := htmx.File(path); strings.Contains(string(f.Content), "npm") || strings.Contains(string(f.Content), "node") {
			t.Errorf("htmx: %s mentions npm or node", path)
		}
	}
}
