package generate

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Tales-of-zuna/gobento/internal/compat"
	"github.com/Tales-of-zuna/gobento/internal/compose"
)

func skipOnWindows(t *testing.T, why string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip(why)
	}
}

func skipIfRoot(t *testing.T) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("permission checks do not apply to root")
	}
}

func TestGenerateDirectoryModes(t *testing.T) {
	t.Parallel()
	skipOnWindows(t, "Windows has no Unix permission bits")

	// A directory created like gobento's subdirectories, under the same
	// umask, is the reference for subdirectory modes.
	ref := filepath.Join(t.TempDir(), "ref")
	if err := os.Mkdir(ref, 0o755); err != nil { //nolint:gosec // G301: the reference uses the same mode as gobento
		t.Fatal(err)
	}
	refInfo, err := os.Stat(ref)
	if err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(t.TempDir(), "shop")
	if err := Generate(context.Background(), testRequest(t, target)); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	root, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if got := root.Mode().Perm(); got != 0o755 {
		t.Errorf("project root mode = %v, want 0755", got)
	}
	sub, err := os.Stat(filepath.Join(target, "cmd", "shop"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := sub.Mode().Perm(), refInfo.Mode().Perm(); got != want {
		t.Errorf("subdirectory mode = %v, want %v (0755 minus the umask)", got, want)
	}
}

// TestGenerateIntoEmptyTargetWithReadOnlyParent shows that an existing
// target is filled from inside, without writing to its parent.
func TestGenerateIntoEmptyTargetWithReadOnlyParent(t *testing.T) {
	t.Parallel()
	skipOnWindows(t, "Windows has no Unix permission bits")
	skipIfRoot(t)

	parent := t.TempDir()
	target := filepath.Join(parent, "shop")
	if err := os.Mkdir(target, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(parent, 0o555); err != nil { //nolint:gosec // G302: test makes the directory read-only
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(parent, 0o755) }) //nolint:gosec // G302: restore for TempDir cleanup

	if err := Generate(context.Background(), testRequest(t, target)); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "go.mod")); err != nil {
		t.Errorf("go.mod missing: %v", err)
	}
	assertNoTempDirs(t, target)
}

// TestPlaceStopsAtExistingEntry shows that placing into an existing target
// stops when an entry already exists there, and that the rollback removes
// only what gobento moved.
func TestPlaceStopsAtExistingEntry(t *testing.T) {
	t.Parallel()

	target := t.TempDir()
	tmp := filepath.Join(target, tempPrefix+"x")
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		writeFile(t, filepath.Join(tmp, name), "new")
	}
	writeFile(t, filepath.Join(target, "b.txt"), "mine")

	moved, err := place(context.Background(), tmp, target, true)
	if err == nil || !strings.Contains(err.Error(), "b.txt already exists") {
		t.Fatalf("place() error = %v, want b.txt already exists", err)
	}
	if want := []string{filepath.Join(target, "a.txt")}; !slices.Equal(moved, want) {
		t.Errorf("moved = %v, want %v", moved, want)
	}

	if err := cleanup(tmp, moved); err != nil {
		t.Fatalf("cleanup() error = %v", err)
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "b.txt" {
		t.Errorf("target entries = %v, want only the pre-existing b.txt", entries)
	}
	if got := readFile(t, filepath.Join(target, "b.txt")); got != "mine" {
		t.Errorf("b.txt = %q, want it unchanged", got)
	}
}

func TestCheckTargetSymlinks(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	emptyDir := filepath.Join(dir, "empty")
	file := filepath.Join(dir, "file")
	if err := os.Mkdir(emptyDir, 0o750); err != nil {
		t.Fatal(err)
	}
	writeFile(t, file, "")

	tests := []struct {
		name    string
		points  string
		wantErr string
	}{
		{name: "link to empty directory", points: emptyDir},
		{name: "link to file", points: file, wantErr: "is not a directory"},
		{name: "broken link", points: filepath.Join(dir, "missing"), wantErr: "cannot be followed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			link := filepath.Join(t.TempDir(), "link")
			if err := os.Symlink(tt.points, link); err != nil {
				t.Skipf("cannot create symbolic links: %v", err)
			}
			err := checkTarget(link)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("checkTarget() error = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, ErrTarget) || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("checkTarget() error = %v, want ErrTarget containing %q", err, tt.wantErr)
			}
		})
	}
}

// TestGenerateCancelCleanupFails shows that a failed cleanup after an
// interrupt is reported together with the interrupt.
func TestGenerateCancelCleanupFails(t *testing.T) {
	t.Parallel()
	skipOnWindows(t, "Windows has no Unix permission bits")
	skipIfRoot(t)

	parent := t.TempDir()
	target := filepath.Join(parent, "shop")
	started := filepath.Join(t.TempDir(), "started")
	t.Cleanup(func() { _ = os.Chmod(parent, 0o755) }) //nolint:gosec // G302: restore for TempDir cleanup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- Generate(ctx, testRequest(t, target, "FAKE_GO_GET=hang", "FAKE_GO_STARTED="+started))
	}()
	waitForFile(t, started)
	// The temporary directory lives in parent; a read-only parent makes
	// removing it fail.
	if err := os.Chmod(parent, 0o555); err != nil { //nolint:gosec // G302: test makes the directory read-only
		t.Fatal(err)
	}
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) || !errors.Is(err, ErrCleanup) {
			t.Fatalf("Generate() error = %v, want context.Canceled joined with ErrCleanup", err)
		}
		if !strings.Contains(err.Error(), tempPrefix) {
			t.Errorf("Generate() error = %q, want it to name the temporary directory", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("Generate() did not return after cancellation")
	}
}

func TestRecheckAcceptsReplaceDirectives(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		replace string
	}{
		{name: "local directory", replace: "replace example.com/lib => ./lib"},
		{name: "fork", replace: "replace example.com/lib v1.2.0 => example.com/fork v1.9.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			writeFile(t, filepath.Join(dir, "go.mod"),
				"module example.com/shop\n\ngo 1.27.0\n\nrequire example.com/lib v1.2.0\n\n"+tt.replace+"\n")
			req := Request{
				Modules: []compat.ModuleVersion{{Path: "example.com/lib", Version: "v1.2.0", Role: compat.RoleLibrary}},
				Env:     fakeEnv(),
			}
			if err := recheck(context.Background(), fakeGoPath, dir, req, "1.27.0"); err != nil {
				t.Errorf("recheck() error = %v, want nil", err)
			}
		})
	}
}

func TestGoEnv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		env       []string
		wantFlags string
	}{
		{name: "no GOFLAGS", env: []string{"PATH=/bin"}, wantFlags: "GOFLAGS=-mod=mod"},
		{name: "keeps other flags", env: []string{"GOFLAGS=-tags=netgo -trimpath"}, wantFlags: "GOFLAGS=-tags=netgo -trimpath -mod=mod"},
		{name: "replaces -mod", env: []string{"GOFLAGS=-mod=vendor -tags=x"}, wantFlags: "GOFLAGS=-tags=x -mod=mod"},
		{name: "last setting wins", env: []string{"GOFLAGS=-a", "GOFLAGS=-tags=y"}, wantFlags: "GOFLAGS=-tags=y -mod=mod"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := goEnv(append(tt.env, "GOWORK=/x/go.work", "GOPROXY=https://p.example"))
			var flags []string
			for _, kv := range got {
				if strings.HasPrefix(kv, "GOFLAGS=") {
					flags = append(flags, kv)
				}
			}
			if len(flags) != 1 || flags[0] != tt.wantFlags {
				t.Errorf("GOFLAGS entries = %q, want [%q]", flags, tt.wantFlags)
			}
			if !slices.Contains(got, "GOWORK=off") || slices.Contains(got, "GOWORK=/x/go.work") {
				t.Errorf("goEnv() = %q, want only GOWORK=off", got)
			}
			if !slices.Contains(got, "GOPROXY=https://p.example") {
				t.Errorf("goEnv() dropped GOPROXY: %q", got)
			}
		})
	}
}

func TestEnvKeyIs(t *testing.T) {
	t.Parallel()

	if !envKeyIs("GOFLAGS", "GOFLAGS") {
		t.Error(`envKeyIs("GOFLAGS", "GOFLAGS") = false`)
	}
	if got, want := envKeyIs("goflags", "GOFLAGS"), runtime.GOOS == "windows"; got != want {
		t.Errorf(`envKeyIs("goflags", "GOFLAGS") = %v, want %v on %s`, got, want, runtime.GOOS)
	}
}

func TestWritePlanRejectsNonLocalPaths(t *testing.T) {
	t.Parallel()

	for _, p := range []string{"../escape.txt", "/abs.txt", "a/../../b.txt"} {
		plan := &compose.Plan{Files: []compose.File{{Path: p, Mode: 0o644}}}
		err := writePlan(t.TempDir(), plan, "example.com/shop", "1.27.0")
		if err == nil || !strings.Contains(err.Error(), "leaves the project directory") {
			t.Errorf("writePlan(%q) error = %v, want a path error", p, err)
		}
	}
}

func TestRetryStopsOnCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errBusy := errors.New("busy")
	calls := 0
	err := retry(ctx, func() error { calls++; return errBusy }, func(error) bool { return true }, []time.Duration{time.Hour})
	if !errors.Is(err, errBusy) || !errors.Is(err, context.Canceled) || calls != 1 {
		t.Errorf("retry() = %v after %d calls, want busy and canceled after 1 call", err, calls)
	}
}

func TestWithDefaults(t *testing.T) {
	t.Parallel()

	req, err := Request{Target: "shop"}.withDefaults()
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(req.Target) || req.GoCommand != "go" || req.Env == nil || req.Warn == nil {
		t.Errorf("withDefaults() = %+v", req)
	}
}
