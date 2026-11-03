package cli

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Tales-of-zuna/gobento/components"
	"github.com/Tales-of-zuna/gobento/internal/generate"
)

var update = flag.Bool("update", false, "rewrite the snapshot files in testdata")

// fakeGenerator records the requests it receives instead of generating.
type fakeGenerator struct {
	calls []generate.Request
	err   error
}

func (g *fakeGenerator) generate(_ context.Context, req generate.Request) error {
	g.calls = append(g.calls, req)
	return g.err
}

// run executes gobento with args against catalog and returns stdout, stderr,
// and the command's error.
func run(t *testing.T, catalog fs.FS, gen *fakeGenerator, args ...string) (string, string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	a := &app{catalog: catalog, generate: gen.generate, stdout: &stdout, stderr: &stderr}
	cmd := newRootCommand("1.2.3", a)
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(context.Background())
	return stdout.String(), stderr.String(), err
}

func TestRootCommandVersion(t *testing.T) {
	t.Parallel()

	out, _, err := run(t, components.FS(), &fakeGenerator{}, "--version")
	if err != nil {
		t.Fatalf("--version error = %v", err)
	}
	if !strings.Contains(out, "1.2.3") {
		t.Errorf("--version output = %q, want 1.2.3", out)
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		wantTarget string
		wantModule string
		wantGo     string
		wantOut    []string
	}{
		{
			// Spec: Fully specified by flags.
			name:       "fully specified by flags",
			args:       []string{"new", "shop", "--module", "example.com/shop", "--arch", "mvc", "--framework", "nethttp", "--db", "sqlite", "--data", "sqlc"},
			wantTarget: "shop",
			wantModule: "example.com/shop",
			wantGo:     "1.27",
			wantOut:    []string{"Created example.com/shop in shop.", "cd shop", "go test ./...", "go run ./cmd/shop"},
		},
		{
			// Spec: Module path defaults to the project name, and
			// Automatic selection of the only option.
			name:       "defaults",
			args:       []string{"new", "shop"},
			wantTarget: "shop",
			wantModule: "shop",
			wantGo:     "1.27",
		},
		{
			name:       "versions and directory",
			args:       []string{"new", "my-app", "--go", "1.26", "--pin", "modernc.org/sqlite=v1.61.0", "--dir", "out/app"},
			wantTarget: "out/app",
			wantModule: "my-app",
			wantGo:     "1.26",
			wantOut:    []string{"cd out/app", "go run ./cmd/my-app"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gen := &fakeGenerator{}
			out, _, err := run(t, components.FS(), gen, tt.args...)
			if err != nil {
				t.Fatalf("gobento %v error = %v", tt.args, err)
			}
			if len(gen.calls) != 1 {
				t.Fatalf("generate called %d times, want 1", len(gen.calls))
			}
			req := gen.calls[0]
			if req.Target != tt.wantTarget || req.Module != tt.wantModule || req.GoVersion != tt.wantGo {
				t.Errorf("request target %q module %q go %q, want %q %q %q",
					req.Target, req.Module, req.GoVersion, tt.wantTarget, tt.wantModule, tt.wantGo)
			}
			if req.Plan == nil || len(req.Plan.Files) == 0 || req.Warn == nil {
				t.Errorf("request has no plan or warning writer: %+v", req)
			}
			for _, w := range tt.wantOut {
				if !strings.Contains(out, w) {
					t.Errorf("output does not contain %q:\n%s", w, out)
				}
			}
		})
	}
}

func TestNewPinVersionIsUsed(t *testing.T) {
	t.Parallel()

	gen := &fakeGenerator{}
	if _, _, err := run(t, components.FS(), gen, "new", "shop", "--pin", "modernc.org/sqlite=v1.61.0"); err != nil {
		t.Fatal(err)
	}
	for _, m := range gen.calls[0].Modules {
		if m.Path == "modernc.org/sqlite" && m.Version != "v1.61.0" {
			t.Errorf("modernc.org/sqlite = %s, want v1.61.0", m.Version)
		}
	}
}

func TestNewErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		catalog fs.FS
		args    []string
		want    []string
	}{
		{
			// Spec: Invalid project name.
			name: "invalid name",
			args: []string{"new", "My App"},
			want: []string{`invalid project name "My App"`, "lowercase letters, digits, - and _"},
		},
		{
			// Spec: Invalid module path.
			name: "invalid module",
			args: []string{"new", "shop", "--module", "Not A Path"},
			want: []string{`invalid module path "Not A Path"`},
		},
		{
			// Spec: Several options require a choice.
			name:    "several options",
			catalog: twoFrameworkFS(t),
			args:    []string{"new", "shop"},
			want:    []string{"--framework: choose a framework; supported: chi, nethttp"},
		},
		{
			name: "tool pin",
			args: []string{"new", "shop", "--pin", "github.com/sqlc-dev/sqlc=v1.30.0"},
			want: []string{"github.com/sqlc-dev/sqlc", "fixed per gobento release"},
		},
		{
			name: "no name",
			args: []string{"new"},
			want: []string{"expected exactly one project name, for example: gobento new shop"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cat := tt.catalog
			if cat == nil {
				cat = components.FS()
			}
			gen := &fakeGenerator{}
			_, _, err := run(t, cat, gen, tt.args...)
			if err == nil {
				t.Fatal("error = nil, want an error")
			}
			for _, w := range tt.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error = %q\nwant it to contain %q", err, w)
				}
			}
			if len(gen.calls) != 0 {
				t.Errorf("generate called %d times, want 0", len(gen.calls))
			}
		})
	}
}

// TestNewSeveralProblemsSnapshot covers the spec scenario "Several
// problems" and pins the exact multi-error output in testdata.
func TestNewSeveralProblemsSnapshot(t *testing.T) {
	t.Parallel()

	_, _, err := run(t, components.FS(), &fakeGenerator{},
		"new", "My App", "--module", "Not A Path", "--framework", "express", "--go", "1.20", "--pin", "nonsense")
	if err == nil {
		t.Fatal("error = nil, want the problem report")
	}
	got := "gobento: " + err.Error() + "\n"
	path := filepath.Join("testdata", "several-problems.txt")
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path) //nolint:gosec // G304: fixed testdata path
	if err != nil {
		t.Fatalf("%v (run with -update to create it)", err)
	}
	if got != string(want) {
		t.Errorf("problem report differs from %s (run with -update if intended):\ngot:\n%s\nwant:\n%s", path, got, want)
	}
}

func TestNewGenerateFailure(t *testing.T) {
	t.Parallel()

	gen := &fakeGenerator{err: errors.New("go get: no network")}
	_, _, err := run(t, components.FS(), gen, "new", "shop")
	if err == nil || !strings.Contains(err.Error(), "  - go get: no network") {
		t.Errorf("error = %v, want the generation error in the report", err)
	}
	if !errors.Is(err, gen.err) {
		t.Errorf("error does not wrap the generation error")
	}
}

// TestNewDryRun covers the spec scenarios "Valid dry run" and "Invalid dry
// run".
func TestNewDryRun(t *testing.T) {
	t.Parallel()

	gen := &fakeGenerator{}
	out, _, err := run(t, components.FS(), gen, "new", "drytest", "--dry-run")
	if err != nil {
		t.Fatalf("--dry-run error = %v", err)
	}
	for _, w := range []string{"nothing was written", "cmd/drytest/main.go", "go.mod", "internal/store/sqlcgen/queries.sql.go", "sqlc.yaml"} {
		if !strings.Contains(out, w) {
			t.Errorf("dry-run output does not contain %q:\n%s", w, out)
		}
	}
	if len(gen.calls) != 0 {
		t.Errorf("generate called %d times during a dry run", len(gen.calls))
	}
	if _, err := os.Stat("drytest"); !os.IsNotExist(err) {
		t.Errorf("dry run created a directory (stat error = %v)", err)
	}

	args := []string{"new", "My App", "--module", "Not A Path"}
	_, _, normal := run(t, components.FS(), gen, args...)
	_, _, dry := run(t, components.FS(), gen, append(args, "--dry-run")...)
	if normal == nil || dry == nil || normal.Error() != dry.Error() {
		t.Errorf("dry run reports %v, normal run reports %v; want the same problems", dry, normal)
	}
}

// TestNewHelp covers task 7.2: the help lists every flag, with an example
// for each one.
func TestNewHelp(t *testing.T) {
	t.Parallel()

	out, _, err := run(t, components.FS(), &fakeGenerator{}, "new", "--help")
	if err != nil {
		t.Fatal(err)
	}
	examples := out[strings.Index(out, "Examples:"):strings.Index(out, "Flags:")]
	for _, name := range []string{"module", "arch", "framework", "db", "data", "go", "pin", "dir", "dry-run"} {
		flagLine := ""
		for line := range strings.SplitSeq(out, "\n") {
			if strings.Contains(line, "--"+name+" ") {
				flagLine = line
			}
		}
		if flagLine == "" {
			t.Errorf("help does not list --%s:\n%s", name, out)
			continue
		}
		if !strings.Contains(examples, "--"+name+" ") && !strings.HasSuffix(strings.TrimSpace(examples), "--"+name) && !strings.Contains(examples, "--"+name+"\n") {
			t.Errorf("no example uses --%s:\n%s", name, examples)
		}
	}
}

// twoFrameworkFS is the shipped catalog plus a copy of nethttp named chi.
func twoFrameworkFS(t *testing.T) fs.FS {
	t.Helper()
	m := fstest.MapFS{}
	err := fs.WalkDir(components.FS(), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(components.FS(), p)
		m[p] = &fstest.MapFile{Data: data}
		if rest, ok := strings.CutPrefix(p, "framework/nethttp/"); ok {
			if rest == "component.yaml" {
				data = []byte("name: chi\ntitle: chi\n")
			}
			m["framework/chi/"+rest] = &fstest.MapFile{Data: data}
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// TestNewIntoCurrentDirectory covers spec "Generate into the current empty
// directory": the target is "." and no cd step is printed.
func TestNewIntoCurrentDirectory(t *testing.T) {
	t.Chdir(t.TempDir())

	gen := &fakeGenerator{}
	out, _, err := run(t, components.FS(), gen, "new", "shop", "--dir", ".")
	if err != nil {
		t.Fatal(err)
	}
	if gen.calls[0].Target != "." {
		t.Errorf("Target = %q, want .", gen.calls[0].Target)
	}
	if !strings.Contains(out, "Created shop in the current directory.") || strings.Contains(out, "cd ") {
		t.Errorf("output =\n%s\nwant the current directory and no cd step", out)
	}
}
