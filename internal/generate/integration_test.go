package generate

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/mod/modfile"

	"github.com/Tales-of-zuna/gobento/internal/catalog"
	"github.com/Tales-of-zuna/gobento/internal/compat"
	"github.com/Tales-of-zuna/gobento/internal/compose"
)

// TestIntegration generates every supported selection twice, with the
// default versions and with the floor of every range, using the real go
// command, and checks that each project builds (also with cgo disabled),
// vets, passes its tests, is formatted, has exactly the chosen versions, and
// keeps the handlers package independent of the database and store
// packages. It needs network access to the module proxy and is skipped with
// -short.
func TestIntegration(t *testing.T) {
	skipIntegration(t)
	cat := shippedCatalog(t)
	supported := compose.Supported(cat)
	if len(supported) == 0 {
		t.Fatal("no supported selections")
	}
	for _, sel := range supported {
		checkSelection(t, cat, sel)
	}
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
func checkSelection(t *testing.T, cat *catalog.Catalog, sel catalog.Selection) {
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
			checkHandlersContract(t, dir, p)
		})
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

// checkHandlersContract asserts the spec "Handlers depend only on the
// contract": the handlers package imports neither the database nor the
// store package, directly or transitively.
func checkHandlersContract(t *testing.T, dir string, p *Prepared) {
	t.Helper()
	handlers, ok := p.Plan.Slots["handlers"]
	if !ok {
		t.Fatal("handlers slot not found in the plan")
	}
	deps := runIn(t, dir, "go", "list", "-deps", "./"+handlers.Dir)
	for _, forbidden := range []string{"database", "store"} {
		slot, ok := p.Plan.Slots[forbidden]
		if !ok {
			t.Fatalf("%s slot not found in the plan", forbidden)
		}
		for line := range strings.SplitSeq(deps, "\n") {
			if line == slot.Import || strings.HasPrefix(line, slot.Import+"/") {
				t.Errorf("handlers package depends on %s", line)
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
