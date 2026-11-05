package generate

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"golang.org/x/mod/modfile"

	"github.com/Tales-of-zuna/gobento/internal/catalog"
	"github.com/Tales-of-zuna/gobento/internal/compat"
	"github.com/Tales-of-zuna/gobento/internal/compose"
)

// TestIntegration generates the selections CI covers (compose.Coverage:
// every supported selection without a frontend, and a pairwise covering set
// of those with one) twice, with the default versions and with the floor of
// every range, using the real go command, and checks that each project
// builds (also with cgo disabled), vets, passes its tests, is formatted, has
// exactly the chosen versions, keeps the handlers and web packages
// independent of the database and store packages, and keeps the contract
// package independent of every other package. With the default versions it
// also checks the generated tooling: actionlint accepts the workflow,
// golangci-lint passes with the generated configuration, and, where make is
// installed, the Makefile's targets work. It needs network access to the
// module proxy and is skipped with -short.
func TestIntegration(t *testing.T) {
	skipIntegration(t)
	cat := shippedCatalog(t)
	supported := compose.Supported(cat)
	if len(supported) == 0 {
		t.Fatal("no supported selections")
	}
	tools := installTools(t, cat)
	covered := compose.Coverage(supported)
	t.Logf("generating %d of %d supported selections", len(covered), len(supported))
	for _, sel := range covered {
		checkSelection(t, cat, sel, tools)
	}
}

// actionlintVersion is the actionlint version that checks generated
// workflows; gobento's own CI uses the same one.
const actionlintVersion = "v1.7.12"

// lintTools are the linters for generated projects, installed once per test
// run.
type lintTools struct {
	actionlint   string
	golangciLint string
	make         string // "" if make is not installed
}

// installTools installs actionlint and the golangci-lint version of the
// rules file into a temporary directory, and looks up make.
func installTools(t *testing.T, cat *catalog.Catalog) lintTools {
	t.Helper()
	version, ok := cat.Rules().Tooling.Tools["golangci-lint"]
	if !ok {
		t.Fatal("versions.yaml has no golangci-lint tool")
	}
	bin := t.TempDir()
	env := []string{"GOBIN=" + bin}
	runInEnv(t, bin, env, "go", "install", "github.com/rhysd/actionlint/cmd/actionlint@"+actionlintVersion)
	runInEnv(t, bin, env, "go", "install", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@"+version)
	tools := lintTools{
		actionlint:   filepath.Join(bin, "actionlint"+exeSuffix()),
		golangciLint: filepath.Join(bin, "golangci-lint"+exeSuffix()),
	}
	if p, err := exec.LookPath("make"); err == nil {
		tools.make = p
	}
	return tools
}

// exeSuffix returns the file name suffix of executables.
func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// checkTooling covers the spec scenarios "Valid workflow", "Fresh project
// lints clean", "Build with make", and "Default target".
func checkTooling(t *testing.T, dir string, tools lintTools) {
	t.Helper()
	runIn(t, dir, tools.actionlint, filepath.Join(".github", "workflows", "ci.yml"))
	if tools.make == "" {
		// Without make, run what make lint runs.
		runIn(t, dir, tools.golangciLint, "run")
		return
	}
	if out := runIn(t, dir, tools.make); !strings.Contains(out, "docker-build") {
		t.Errorf("make lists no docker-build target:\n%s", out)
	}
	runIn(t, dir, tools.make, "build")
	if _, err := os.Stat(filepath.Join(dir, "bin", "shop"+exeSuffix())); err != nil {
		t.Errorf("make build: %v", err)
	}
	// Forward slashes: make on Windows runs recipes with sh, which would
	// drop the backslashes of a Windows path.
	runIn(t, dir, tools.make, "lint", "GOLANGCI_LINT="+filepath.ToSlash(tools.golangciLint))
}

// checkTemplFresh covers the spec scenarios "Regenerate the pages with make"
// and "Fresh templ code passes CI": it runs the command of the CI workflow's
// templ step, go generate in the web package, which runs the pinned templ,
// and checks that every planned file is unchanged, and, where make is
// installed, that make generate runs the same command.
func checkTemplFresh(t *testing.T, dir string, p *Prepared, webDir string, tools lintTools) {
	t.Helper()
	runIn(t, dir, "go", "generate", "./"+webDir)
	for _, f := range p.Plan.Files {
		got, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(f.Path))) //nolint:gosec // G304: a planned path in the generated project
		if err != nil {
			t.Errorf("after go generate: %v", err)
			continue
		}
		if string(got) != string(f.Content) {
			t.Errorf("go generate ./%s changed %s: the shipped templ output is stale", webDir, f.Path)
		}
	}
	entries, err := os.ReadDir(filepath.Join(dir, filepath.FromSlash(webDir)))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if _, ok := p.Plan.File(webDir + "/" + e.Name()); !ok && !e.IsDir() {
			t.Errorf("go generate ./%s created %s", webDir, e.Name())
		}
	}
	if tools.make == "" {
		return
	}
	if out := runIn(t, dir, tools.make, "--dry-run", "generate"); !strings.Contains(out, "go generate ./"+webDir) {
		t.Errorf("make generate does not run go generate ./%s:\n%s", webDir, out)
	}
}

// tailwindArgs are the arguments of the Tailwind runner that rebuild the
// stylesheet of a project with a go frontend build, as in its Makefile.
func tailwindArgs(webDir string) []string {
	return []string{"run", "./internal/tools/tailwindcss", "--input", "tailwind.css", "--output", webDir + "/static/styles.css", "--minify"}
}

// checkTailwindFresh covers the spec scenario "Regeneration changes
// nothing" for Tailwind CSS: rebuilding the stylesheet with the pinned
// standalone CLI, which the runner downloads and verifies, changes no
// planned file, and make generate runs the same command.
func checkTailwindFresh(t *testing.T, dir string, p *Prepared, tools lintTools) {
	t.Helper()
	webDir := p.Plan.Slots["web"].Dir
	runIn(t, dir, "go", tailwindArgs(webDir)...)
	for _, f := range p.Plan.Files {
		got, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(f.Path))) //nolint:gosec // G304: a planned path in the generated project
		if err != nil {
			t.Errorf("after the Tailwind CLI: %v", err)
			continue
		}
		if string(got) != string(f.Content) {
			t.Errorf("the Tailwind CLI changed %s: the shipped Tailwind output is stale", f.Path)
		}
	}
	if tools.make == "" {
		return
	}
	want := "go run ./internal/tools/tailwindcss --input tailwind.css --output " + webDir + "/static/styles.css --minify"
	if out := runIn(t, dir, tools.make, "--dry-run", "generate"); !strings.Contains(out, want) {
		t.Errorf("make generate does not run %s:\n%s", want, out)
	}
}

// TestIntegrationTailwindProjectName covers the spec scenario "Project name
// is a utility": a project whose name and module path are Tailwind
// utilities gets the same shipped stylesheet, and rebuilding it changes
// nothing, because Tailwind scans only the pages, which hold neither.
func TestIntegrationTailwindProjectName(t *testing.T) {
	skipIntegration(t)
	cat := shippedCatalog(t)
	p, err := Prepare(cat, Spec{Name: "container", Module: "grid", Components: map[catalog.Kind]string{
		catalog.KindArchitecture: "vertical-slice", catalog.KindFramework: "chi", catalog.KindDatabase: "sqlite",
		catalog.KindData: "sqlc", catalog.KindFrontend: "htmx", catalog.KindCSS: "tailwind",
	}})
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	dir := filepath.Join(t.TempDir(), "container")
	if err := Generate(context.Background(), p.Request(dir)); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	checkTailwindFresh(t, dir, p, lintTools{})
	runIn(t, dir, "go", "test", "-count=1", "./internal/web/...", "./internal/tools/...")
}

func skipIntegration(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test: needs the go command and network access")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not found")
	}
}

// checkSelection generates sel with the defaults and with the floors and
// checks each project.
func checkSelection(t *testing.T, cat *catalog.Catalog, sel catalog.Selection, tools lintTools) {
	t.Helper()
	for _, variant := range []string{"defaults", "floors"} {
		t.Run(strings.ReplaceAll(sel.String(), "/", "-")+"/"+variant, func(t *testing.T) {
			spec := Spec{Name: "shop", Module: "example.com/shop", Components: selectionNames(sel)}
			if variant == "floors" {
				spec.Go, spec.Pins = floors(t, cat, sel)
			}
			p, err := Prepare(cat, spec)
			if err != nil {
				t.Fatalf("Prepare() error = %v", err)
			}
			dir := filepath.Join(t.TempDir(), "shop")
			if err := Generate(context.Background(), p.Request(dir)); err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			checkScaffolds(t, dir, p)
			if variant == "defaults" {
				// The floors keep the unbuilt placeholder, so both states of
				// the web package are built and tested.
				buildFrontends(t, dir, p)
			}

			for _, args := range [][]string{
				{"build", "./..."},
				{"vet", "./..."},
				{"test", "-count=1", "./..."},
				{"mod", "verify"},
			} {
				runIn(t, dir, "go", args...)
			}
			// Spec: Build without cgo.
			runInEnv(t, dir, []string{"CGO_ENABLED=0"}, "go", "build", "./...")
			if out := runIn(t, dir, "gofmt", "-l", "."); strings.TrimSpace(out) != "" {
				t.Errorf("gofmt -l lists files:\n%s", out)
			}
			checkPins(t, dir, p.Versions)
			checkContract(t, dir, p, "handlers")
			if p.Plan.Slots["web"].Filled {
				checkContract(t, dir, p, "web")
				checkNoFramework(t, dir, p, cat)
			}
			checkLayers(t, dir, p)
			if variant == "defaults" {
				checkTooling(t, dir, tools)
				if web := p.Plan.Slots["web"]; web.Filled && sel.Traits()["frontend-build"] == "go" {
					checkTemplFresh(t, dir, p, web.Dir, tools)
				}
				if p.Selection.CSS != nil && p.Selection.Traits()["frontend-build"] == "go" {
					checkTailwindFresh(t, dir, p, tools)
				}
			}
		})
	}
}

// checkScaffolds asserts the spec requirement "Scaffold step": gobento's
// planned files are unchanged by the scaffolder, and each scaffolded
// directory has its package.json and lock file but no node_modules.
func checkScaffolds(t *testing.T, dir string, p *Prepared) {
	t.Helper()
	if len(p.Scaffolds) == 0 {
		return
	}
	for _, f := range p.Plan.Files {
		if got := readFile(t, filepath.Join(dir, filepath.FromSlash(f.Path))); got != string(f.Content) {
			t.Errorf("%s differs from gobento's file after scaffolding", f.Path)
		}
	}
	for _, s := range p.Scaffolds {
		sd := filepath.Join(dir, filepath.FromSlash(s.Dir))
		for _, name := range []string{"package.json", lockFile} {
			if _, err := os.Stat(filepath.Join(sd, name)); err != nil {
				t.Errorf("%s/%s: %v", s.Dir, name, err)
			}
		}
		if _, err := os.Stat(filepath.Join(sd, "node_modules")); err == nil {
			t.Errorf("%s/node_modules exists after generation", s.Dir)
		}
	}
}

// buildFrontends covers the spec scenario "Frontend build in the integration
// test": for each scaffolded frontend it runs what the generated CI workflow
// runs (npm ci, the lint and check scripts if defined, npm run build) and
// checks that the build is in the web package's dist/app, so the Go checks
// that follow embed the real app.
func buildFrontends(t *testing.T, dir string, p *Prepared) {
	t.Helper()
	if len(p.Scaffolds) == 0 {
		return
	}
	npm, err := exec.LookPath("npm")
	if err != nil {
		t.Fatalf("npm not found: %v", err)
	}
	for _, s := range p.Scaffolds {
		sd := filepath.Join(dir, filepath.FromSlash(s.Dir))
		for _, args := range [][]string{
			{"ci", "--no-audit", "--no-fund"},
			{"run", "--if-present", "lint"},
			{"run", "--if-present", "check"},
			{"run", "build"},
		} {
			runIn(t, sd, npm, args...)
		}
	}
	index := filepath.Join(dir, filepath.FromSlash(p.Plan.Slots["web"].Dir), "dist", "app", "index.html")
	if _, err := os.Stat(index); err != nil {
		t.Errorf("npm run build did not write the app into the web package: %v", err)
	}
}

// selectionNames returns the component names of sel by kind.
func selectionNames(sel catalog.Selection) map[catalog.Kind]string {
	names := make(map[catalog.Kind]string)
	for _, c := range sel.Components() {
		names[c.Kind] = c.Name
	}
	return names
}

// floors returns the lowest supported Go version and a pin to the lowest
// supported version of every library module the selection uses.
func floors(t *testing.T, cat *catalog.Catalog, sel catalog.Selection) (string, map[string]string) {
	t.Helper()
	goFloor, ok := cat.Rules().Go.Range.Floor()
	if !ok {
		t.Fatalf("go range %s has no floor", cat.Rules().Go.Range)
	}
	var ids []string
	for _, c := range sel.Components() {
		ids = append(ids, c.ID())
	}
	res, err := cat.Rules().Resolve(compat.Request{Components: ids, Traits: sel.Traits()})
	if err != nil {
		t.Fatal(err)
	}
	pins := make(map[string]string)
	for _, m := range res.Modules {
		if m.Role != compat.RoleLibrary {
			continue
		}
		rule, _ := cat.Rules().Module(m.Path)
		floor, ok := rule.Range.Floor()
		if !ok {
			t.Fatalf("module %s range %s has no floor", m.Path, rule.Range)
		}
		pins[m.Path] = floor
	}
	return strings.TrimPrefix(goFloor, "v"), pins
}

// checkPins asserts that go.mod has exactly the chosen go directive and
// library versions, as direct requirements, and no toolchain line.
func checkPins(t *testing.T, dir string, res compat.Resolved) {
	t.Helper()
	data := readFile(t, filepath.Join(dir, "go.mod"))
	f, err := modfile.Parse("go.mod", []byte(data), nil)
	if err != nil {
		t.Fatal(err)
	}
	wantGo, _ := compat.GoDirective(res.Go)
	if f.Go == nil || f.Go.Version != wantGo {
		t.Errorf("go directive = %v, want %s", f.Go, wantGo)
	}
	if f.Toolchain != nil {
		t.Errorf("go.mod has a toolchain line: %s", f.Toolchain.Name)
	}
	direct := make(map[string]string)
	for _, r := range f.Require {
		if !r.Indirect {
			direct[r.Mod.Path] = r.Mod.Version
		}
	}
	for _, m := range res.Modules {
		_, inMod := direct[m.Path]
		switch {
		case m.Role == compat.RoleTool && inMod:
			t.Errorf("tool module %s is required in go.mod", m.Path)
		case m.Role == compat.RoleLibrary && direct[m.Path] != m.Version:
			t.Errorf("go.mod requires %s %q, want %s", m.Path, direct[m.Path], m.Version)
		}
	}
}

// checkContract asserts the spec scenarios "Handlers depend only on the
// contract" and "Web UI depends only on the contract": the package of slot
// imports neither the database nor the store package, directly or
// transitively.
func checkContract(t *testing.T, dir string, p *Prepared, slot string) {
	t.Helper()
	pkg, ok := p.Plan.Slots[slot]
	if !ok {
		t.Fatalf("%s slot not found in the plan", slot)
	}
	deps := runIn(t, dir, "go", "list", "-deps", "./"+pkg.Dir)
	for _, forbidden := range []string{"database", "store"} {
		other, ok := p.Plan.Slots[forbidden]
		if !ok {
			t.Fatalf("%s slot not found in the plan", forbidden)
		}
		for line := range strings.SplitSeq(deps, "\n") {
			if line == other.Import || strings.HasPrefix(line, other.Import+"/") {
				t.Errorf("%s package depends on %s", slot, line)
			}
		}
	}
}

// checkNoFramework asserts that the web package imports no module of a
// framework component: a frontend works with every framework.
func checkNoFramework(t *testing.T, dir string, p *Prepared, cat *catalog.Catalog) {
	t.Helper()
	deps := runIn(t, dir, "go", "list", "-deps", "./"+p.Plan.Slots["web"].Dir)
	for _, m := range cat.Rules().Modules {
		if !slices.ContainsFunc(m.Components, func(id string) bool { return strings.HasPrefix(id, string(catalog.KindFramework)+"/") }) {
			continue
		}
		for line := range strings.SplitSeq(deps, "\n") {
			if line == m.Path || strings.HasPrefix(line, m.Path+"/") {
				t.Errorf("web package depends on the framework module %s", line)
			}
		}
	}
}

// checkLayers asserts the spec "Domain layers depend inward" as far as it
// applies to every architecture: the contract package (the models slot, the
// domain in ddd) imports no other package of the project. With an
// application slot, it also asserts that the application package imports
// only the contract package, and that no adapter (the web package
// included, when a frontend fills it) imports the application package.
func checkLayers(t *testing.T, dir string, p *Prepared) {
	t.Helper()
	models := p.Plan.Slots["models"]
	module := strings.TrimSuffix(models.Import, "/"+models.Dir)
	projectDeps := func(pkg string) []string {
		var out []string
		for line := range strings.SplitSeq(runIn(t, dir, "go", "list", "-deps", "./"+pkg), "\n") {
			if line == module || strings.HasPrefix(line, module+"/") {
				out = append(out, line)
			}
		}
		return out
	}
	if deps := projectDeps(models.Dir); len(deps) != 1 || deps[0] != models.Import {
		t.Errorf("package %s depends on %v, want no other package of the project", models.Import, deps)
	}
	app, ok := p.Plan.Slots["app"]
	if !ok {
		return
	}
	for _, dep := range projectDeps(app.Dir) {
		if dep != app.Import && dep != models.Import {
			t.Errorf("package %s depends on %s, want only %s", app.Import, dep, models.Import)
		}
	}
	adapters := []string{"handlers", "store", "database"}
	if p.Plan.Slots["web"].Filled {
		adapters = append(adapters, "web")
	}
	for _, adapter := range adapters {
		for _, dep := range projectDeps(p.Plan.Slots[adapter].Dir + "/...") {
			if dep == app.Import {
				t.Errorf("the %s slot depends on %s", adapter, app.Import)
			}
		}
	}
}

// runIn runs a command in dir with gobento's controlled environment and
// fails the test on error.
func runIn(t *testing.T, dir, name string, args ...string) string {
	t.Helper()
	return runInEnv(t, dir, nil, name, args...)
}

// runInEnv is runIn with extra environment variables, which take
// precedence.
func runInEnv(t *testing.T, dir string, env []string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...) //nolint:gosec // G204: the test runs fixed go and gofmt commands
	cmd.Dir = dir
	cmd.Env = append(goEnv(os.Environ()), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, out)
	}
	return string(out)
}
