package generate

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Tales-of-zuna/gobento/internal/compat"
	"github.com/Tales-of-zuna/gobento/internal/compose"
)

func testRequest(t *testing.T, target string, env ...string) Request {
	t.Helper()
	return Request{
		Plan: &compose.Plan{Files: []compose.File{
			{Path: "README.md", Mode: 0o644, Content: []byte("# shop\n")},
			{Path: "cmd/shop/main.go", Mode: 0o644, Content: []byte("package main\n\nfunc main() {}\n")},
		}},
		Module:    "example.com/shop",
		GoVersion: "1.27",
		Modules: []compat.ModuleVersion{
			{Path: "example.com/lib", Version: "v1.2.0", Role: compat.RoleLibrary},
			{Path: "example.com/other", Version: "v0.5.0", Role: compat.RoleLibrary},
			{Path: "example.com/tool", Version: "v1.31.1", Role: compat.RoleTool},
		},
		Target:    target,
		GoCommand: fakeGoPath,
		Env:       fakeEnv(env...),
	}
}

func TestGenerateIntoMissingTarget(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	target := filepath.Join(parent, "shop")
	log := filepath.Join(t.TempDir(), "go.log")
	req := testRequest(t, target, "FAKE_GO_LOG="+log, "FAKE_GO_TOOLCHAIN_LINE=1",
		"GOPROXY=https://proxy.example.test", "GOWORK=/some/go.work", "GOFLAGS=-mod=vendor")

	if err := Generate(context.Background(), req); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	for _, p := range []string{"README.md", "cmd/shop/main.go"} {
		if _, err := os.Stat(filepath.Join(target, filepath.FromSlash(p))); err != nil {
			t.Errorf("%s missing: %v", p, err)
		}
	}
	gomod := readFile(t, filepath.Join(target, "go.mod"))
	for _, want := range []string{"module example.com/shop", "go 1.27.0", "example.com/lib v1.2.0", "example.com/other v0.5.0"} {
		if !strings.Contains(gomod, want) {
			t.Errorf("go.mod does not contain %q:\n%s", want, gomod)
		}
	}
	for _, unwanted := range []string{"toolchain", "example.com/tool"} {
		if strings.Contains(gomod, unwanted) {
			t.Errorf("go.mod contains %q:\n%s", unwanted, gomod)
		}
	}
	assertNoTempDirs(t, parent)

	calls := readFile(t, log)
	for _, want := range []string{
		"go get example.com/lib@v1.2.0 example.com/other@v0.5.0 |",
		"go mod tidy |",
		"go mod edit -toolchain=none |",
		"GOWORK=off GOFLAGS=-mod=mod GOPROXY=https://proxy.example.test",
	} {
		if !strings.Contains(calls, want) {
			t.Errorf("go invocations do not contain %q:\n%s", want, calls)
		}
	}
	if strings.Count(calls, "go get") != 1 {
		t.Errorf("want exactly one go get:\n%s", calls)
	}
}

// TestGenerateIntoEmptyTarget covers spec "Generate into the current empty
// directory": files land directly in the existing directory.
func TestGenerateIntoEmptyTarget(t *testing.T) {
	t.Parallel()

	target := t.TempDir()
	if err := Generate(context.Background(), testRequest(t, target)); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "cmd", "shop", "main.go")); err != nil {
		t.Errorf("main.go missing: %v", err)
	}
	assertNoTempDirs(t, filepath.Dir(target))
	assertNoTempDirs(t, target)
}

// TestGenerateRaisedVersion covers spec "Resolution raises a version".
func TestGenerateRaisedVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		env  string
		want []string
	}{
		{
			name: "module raised",
			env:  "FAKE_GO_RAISE=example.com/lib@v1.5.0",
			want: []string{"example.com/lib was raised from v1.2.0 to v1.5.0", "missing a constraint"},
		},
		{
			name: "go directive raised",
			env:  "FAKE_GO_RAISE_GO=1.27.5",
			want: []string{"go directive was raised from 1.27.0 to 1.27.5", "missing a constraint"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parent := t.TempDir()
			target := filepath.Join(parent, "shop")
			err := Generate(context.Background(), testRequest(t, target, tt.env))
			if !errors.Is(err, ErrRaised) {
				t.Fatalf("Generate() error = %v, want ErrRaised", err)
			}
			for _, w := range tt.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("Generate() error = %q, want it to contain %q", err, w)
				}
			}
			assertGone(t, target)
			assertNoTempDirs(t, parent)
		})
	}
}

// TestGenerateNoPartialProjects covers spec "No partial projects": a failing
// go get and an interrupt both leave nothing behind and leave a
// pre-existing target as it was.
func TestGenerateNoPartialProjects(t *testing.T) {
	t.Parallel()

	for _, existing := range []bool{false, true} {
		name := "missing target"
		if existing {
			name = "existing empty target"
		}

		t.Run("dependency fetch fails/"+name, func(t *testing.T) {
			t.Parallel()

			parent, target := newTarget(t, existing)
			err := Generate(context.Background(), testRequest(t, target, "FAKE_GO_GET=fail"))
			if !errors.Is(err, ErrGoCommand) || !strings.Contains(err.Error(), "no such host") {
				t.Fatalf("Generate() error = %v, want ErrGoCommand with the go output", err)
			}
			assertTarget(t, target, existing)
			assertNoTempDirs(t, parent)
		})

		t.Run("interrupted/"+name, func(t *testing.T) {
			t.Parallel()

			parent, target := newTarget(t, existing)
			started := filepath.Join(t.TempDir(), "started")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			done := make(chan error, 1)
			go func() {
				done <- Generate(ctx, testRequest(t, target, "FAKE_GO_GET=hang", "FAKE_GO_STARTED="+started))
			}()
			waitForFile(t, started)
			cancel()

			select {
			case err := <-done:
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("Generate() error = %v, want context.Canceled", err)
				}
			case <-time.After(30 * time.Second):
				t.Fatal("Generate() did not return after cancellation")
			}
			assertTarget(t, target, existing)
			assertNoTempDirs(t, parent)
		})
	}
}

func TestGeneratePreflightFailureWritesNothing(t *testing.T) {
	t.Parallel()

	target := t.TempDir()
	writeFile(t, filepath.Join(target, "keep.txt"), "mine")
	err := Generate(context.Background(), testRequest(t, target))
	if !errors.Is(err, ErrTarget) {
		t.Fatalf("Generate() error = %v, want ErrTarget", err)
	}
	if got := readFile(t, filepath.Join(target, "keep.txt")); got != "mine" {
		t.Errorf("keep.txt = %q, want it unchanged", got)
	}
	entries, _ := os.ReadDir(target)
	if len(entries) != 1 {
		t.Errorf("target has %d entries, want only keep.txt", len(entries))
	}
	assertNoTempDirs(t, filepath.Dir(target))
}

func TestCleanupKeepsTarget(t *testing.T) {
	t.Parallel()

	target := t.TempDir()
	tmp := filepath.Join(t.TempDir(), tempPrefix+"x")
	writeFile(t, filepath.Join(tmp, "left.txt"), "")
	moved := filepath.Join(target, "cmd")
	writeFile(t, filepath.Join(moved, "main.go"), "")

	if err := cleanup(tmp, []string{moved}); err != nil {
		t.Fatalf("cleanup() error = %v", err)
	}
	assertGone(t, tmp)
	assertGone(t, moved)
	if _, err := os.Stat(target); err != nil {
		t.Errorf("target removed: %v", err)
	}
}

func TestRetry(t *testing.T) {
	t.Parallel()

	errBusy := errors.New("busy")
	errFatal := errors.New("fatal")
	retryable := func(err error) bool { return errors.Is(err, errBusy) }
	delays := []time.Duration{0, 0, 0}

	tests := []struct {
		name      string
		results   []error
		wantErr   error
		wantCalls int
	}{
		{name: "succeeds at once", results: []error{nil}, wantCalls: 1},
		{name: "succeeds after retries", results: []error{errBusy, errBusy, nil}, wantCalls: 3},
		{name: "gives up", results: []error{errBusy, errBusy, errBusy, errBusy}, wantErr: errBusy, wantCalls: 4},
		{name: "does not retry other errors", results: []error{errFatal}, wantErr: errFatal, wantCalls: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			calls := 0
			err := retry(context.Background(), func() error {
				err := tt.results[calls]
				calls++
				return err
			}, retryable, delays)
			if !errors.Is(err, tt.wantErr) || (tt.wantErr == nil && err != nil) {
				t.Errorf("retry() error = %v, want %v", err, tt.wantErr)
			}
			if calls != tt.wantCalls {
				t.Errorf("calls = %d, want %d", calls, tt.wantCalls)
			}
		})
	}
}

func newTarget(t *testing.T, existing bool) (parent, target string) {
	t.Helper()
	parent = t.TempDir()
	target = filepath.Join(parent, "shop")
	if existing {
		if err := os.Mkdir(target, 0o750); err != nil {
			t.Fatal(err)
		}
	}
	return parent, target
}

// assertTarget checks that a missing target is still missing and an
// existing one still exists and is empty.
func assertTarget(t *testing.T, target string, existing bool) {
	t.Helper()
	if !existing {
		assertGone(t, target)
		return
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		t.Fatalf("pre-existing target: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("pre-existing target has %d leftover entries", len(entries))
	}
}

func assertGone(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("%s exists (err = %v), want it removed", path, err)
	}
}

func assertNoTempDirs(t *testing.T, dir string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, tempPrefix+"*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) > 0 {
		t.Errorf("temporary directories left behind: %v", matches)
	}
}

func waitForFile(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("%s was not created", path)
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path) //nolint:gosec // G304: test-controlled path
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
