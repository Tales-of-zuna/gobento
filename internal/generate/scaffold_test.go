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

// scaffoldRequest is testRequest with one scaffold step into frontend/ and
// gobento's own frontend/src/App.tsx in the plan.
func scaffoldRequest(t *testing.T, target string, env ...string) Request {
	t.Helper()
	req := testRequest(t, target, env...)
	req.Plan.Files = append(req.Plan.Files, compose.File{Path: "frontend/src/App.tsx", Mode: 0o644, Content: []byte("gobento\n")})
	rng, err := compat.ParseRange(">=v22.12.0")
	if err != nil {
		t.Fatal(err)
	}
	req.Node = compat.NodeRule{Range: rng, Default: "26"}
	req.NodeCommand, req.NPMCommand = fakeNodePath, fakeNPMPath
	req.Scaffolds = []Scaffold{{
		Component: "frontend/react", Package: "create-vite", Version: "9.3.3",
		Template: "react-ts", Dir: "frontend", Remove: []string{"src/App.css"},
	}}
	return req
}

// TestGenerateScaffold covers the spec requirement "Scaffold step": the
// scaffolder runs without questions, the remove list is applied, gobento's
// files win, and the lock file is written without node_modules.
func TestGenerateScaffold(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	target := filepath.Join(parent, "shop")
	log := filepath.Join(t.TempDir(), "npm.log")
	req := scaffoldRequest(t, target, "FAKE_NPM_LOG="+log)
	req.Scaffolds[0].DevDependencies = []string{"@scope/plugin@4.1.0"}
	if err := Generate(context.Background(), req); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	dir := filepath.Join(target, "frontend")
	if got := readFile(t, filepath.Join(dir, "src", "App.tsx")); got != "gobento\n" {
		t.Errorf("src/App.tsx = %q, want gobento's file", got)
	}
	for _, p := range []string{"package.json", "index.html", "src/assets/logo.svg", "package-lock.json"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(p))); err != nil {
			t.Errorf("%s missing: %v", p, err)
		}
	}
	for _, p := range []string{"src/App.css", "node_modules"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(p))); err == nil {
			t.Errorf("%s exists", p)
		}
	}
	entries, _ := os.ReadDir(target)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), stagingPrefix) {
			t.Errorf("staging directory %s left in the project", e.Name())
		}
	}
	if !strings.Contains(readFile(t, filepath.Join(dir, "package-lock.json")), "--save-dev @scope/plugin@4.1.0") {
		t.Errorf("package-lock.json does not record the dev dependency install")
	}
	calls := readFile(t, log)
	for _, want := range []string{
		"npm create --yes vite@9.3.3 frontend -- --no-interactive --no-immediate --template react-ts | dir=" + stagingPrefix,
		"npm install --package-lock-only --ignore-scripts --no-audit --no-fund --save-dev @scope/plugin@4.1.0 | dir=frontend yes=true",
	} {
		if !strings.Contains(calls, want) {
			t.Errorf("npm calls do not contain %q:\n%s", want, calls)
		}
	}
	assertNoTempDirs(t, parent)
}

// TestGenerateScaffoldAllowRemote covers npm 12's refusal of registry
// tarballs with bundled dependencies in lock-only installs (npm/cli#9800):
// the install is retried with --allow-remote=all, and the lock file must
// resolve every package on the configured registry's host.
func TestGenerateScaffoldAllowRemote(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, resolved, wantErr string
	}{
		{name: "registry tarballs", resolved: ""},
		{name: "tarball from another host", resolved: "https://evil.example/a-1.0.0.tgz", wantErr: "node_modules/a to https://evil.example/a-1.0.0.tgz, which is not on the npm registry registry.npmjs.org"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			parent := t.TempDir()
			target := filepath.Join(parent, "shop")
			log := filepath.Join(t.TempDir(), "npm.log")
			env := []string{"FAKE_NPM=allowremote", "FAKE_NPM_LOG=" + log}
			if tc.resolved != "" {
				env = append(env, "FAKE_NPM_RESOLVED="+tc.resolved)
			}
			req := scaffoldRequest(t, target, env...)
			req.Scaffolds[0].DevDependencies = []string{"tailwindcss@4.3.4"}
			err := Generate(context.Background(), req)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Generate() error = %v", err)
				}
				if calls := readFile(t, log); !strings.Contains(calls, "--save-dev tailwindcss@4.3.4 --allow-remote=all") {
					t.Errorf("npm was not retried with --allow-remote=all:\n%s", calls)
				}
				return
			}
			if !errors.Is(err, ErrNPMCommand) || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("Generate() error = %v, want %q", err, tc.wantErr)
			}
			if _, statErr := os.Stat(target); statErr == nil {
				t.Error("the project was left behind")
			}
			assertNoTempDirs(t, parent)
		})
	}
}

// TestGenerateScaffoldFailures covers the spec scenario "Scaffolder fails"
// and a scaffolder output without a file of the remove list.
func TestGenerateScaffoldFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, env string
		want      []string
		is        error
	}{
		{"offline", "FAKE_NPM=fail", []string{"npm create --yes vite@9.3.3", "ENOTFOUND"}, ErrNPMCommand},
		{"remove list", "FAKE_NPM_NO_REMOVE=1", []string{"did not create src/App.css", "scaffold.remove"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parent := t.TempDir()
			target := filepath.Join(parent, "shop")
			err := Generate(context.Background(), scaffoldRequest(t, target, tt.env))
			if err == nil || (tt.is != nil && !errors.Is(err, tt.is)) {
				t.Fatalf("Generate() error = %v, want %v", err, tt.is)
			}
			for _, w := range tt.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q does not contain %q", err, w)
				}
			}
			assertGone(t, target)
			assertNoTempDirs(t, parent)
		})
	}
}

// TestGenerateScaffoldInterrupted covers the spec scenario "Interrupted
// while scaffolding": cancelling stops npm and the process it started, and
// nothing is left behind, not even later writes of that process.
func TestGenerateScaffoldInterrupted(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	target := filepath.Join(parent, "shop")
	started := filepath.Join(t.TempDir(), "started")
	late := filepath.Join(parent, "late.txt")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- Generate(ctx, scaffoldRequest(t, target, "FAKE_NPM=hang", "FAKE_NPM_STARTED="+started, "FAKE_NPM_LATE="+late))
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
	assertGone(t, target)
	assertNoTempDirs(t, parent)
	time.Sleep(3 * time.Second)
	if _, err := os.Stat(late); err == nil {
		t.Error("the process npm started kept running after cancellation")
	}
}

// TestNodePreflight covers the spec scenarios "Node.js is missing" and
// "Node.js is too old", and that nothing is written.
func TestNodePreflight(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		edit  func(*Request)
		is    error
		wants []string
	}{
		{"missing node", func(r *Request) { r.NodeCommand = filepath.Join(t.TempDir(), "node") }, ErrNoNode,
			[]string{"node command was not found", "frontend/react is created with npm create vite@9.3.3 frontend -- --template react-ts", ">=v22.12.0", "https://nodejs.org/"}},
		{"missing npm", func(r *Request) { r.NPMCommand = filepath.Join(t.TempDir(), "npm") }, ErrNoNode,
			[]string{"npm command was not found"}},
		{"old node", func(r *Request) { r.Env = append(r.Env, "FAKE_NODE_VERSION=v20.19.0") }, ErrNodeVersion,
			[]string{"the local Node.js is v20.19.0", ">=v22.12.0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parent := t.TempDir()
			target := filepath.Join(parent, "shop")
			req := scaffoldRequest(t, target)
			tt.edit(&req)
			err := Generate(context.Background(), req)
			if !errors.Is(err, tt.is) {
				t.Fatalf("Generate() error = %v, want %v", err, tt.is)
			}
			for _, w := range tt.wants {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q does not contain %q", err, w)
				}
			}
			assertGone(t, target)
			assertNoTempDirs(t, parent)
		})
	}

	// Spec scenario "API-only projects need no Node.js".
	req := testRequest(t, filepath.Join(t.TempDir(), "shop"))
	req.NodeCommand = filepath.Join(t.TempDir(), "node")
	if err := Generate(context.Background(), req); err != nil {
		t.Errorf("Generate() without scaffold steps and without node: %v", err)
	}
}
