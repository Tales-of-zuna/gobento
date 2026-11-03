package generate

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPreflightToolchain covers the spec "Go toolchain preflight" scenarios.
func TestPreflightToolchain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		goCommand string // "" means the fake go command
		env       []string
		goVersion string
		wantIs    error
		wantMsg   []string
		wantWarn  string
	}{
		{
			// Spec: Go is not installed.
			name:      "go not installed",
			goCommand: filepath.Join(t.TempDir(), "go"),
			goVersion: "1.27",
			wantIs:    ErrNoGo,
			wantMsg:   []string{"https://go.dev/dl/"},
		},
		{
			// Spec: Selected Go is newer and downloads are disabled.
			name:      "newer and GOTOOLCHAIN=local",
			env:       []string{"FAKE_GO_VERSION=go1.26.2", "FAKE_GOTOOLCHAIN=local"},
			goVersion: "1.27",
			wantIs:    ErrToolchain,
			wantMsg:   []string{"Go 1.27 is selected", "Go 1.26.2", "GOTOOLCHAIN=local"},
		},
		{
			name:      "newer and GOTOOLCHAIN=path",
			env:       []string{"FAKE_GO_VERSION=go1.26.2", "FAKE_GOTOOLCHAIN=path"},
			goVersion: "1.27",
			wantIs:    ErrToolchain,
		},
		{
			name:      "newer with downloads allowed warns",
			env:       []string{"FAKE_GO_VERSION=go1.26.2", "FAKE_GOTOOLCHAIN=auto"},
			goVersion: "1.27",
			wantWarn:  "will download Go 1.27",
		},
		{
			name:      "same release",
			env:       []string{"FAKE_GO_VERSION=go1.27.0-X:nodwarf5", "FAKE_GOTOOLCHAIN=local"},
			goVersion: "1.27",
		},
		{
			name:      "older selection",
			env:       []string{"FAKE_GO_VERSION=go1.27.3", "FAKE_GOTOOLCHAIN=local"},
			goVersion: "1.26",
		},
		{
			name:      "development toolchain is not compared",
			env:       []string{"FAKE_GO_VERSION=devel go1.28-abcdef", "FAKE_GOTOOLCHAIN=local"},
			goVersion: "1.27",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			goCommand := tt.goCommand
			if goCommand == "" {
				goCommand = fakeGoPath
			}
			var warn bytes.Buffer
			err := Preflight(context.Background(), Request{
				GoVersion: tt.goVersion,
				Target:    filepath.Join(t.TempDir(), "shop"),
				GoCommand: goCommand,
				Env:       fakeEnv(tt.env...),
				Warn:      &warn,
			})
			if tt.wantIs == nil && err != nil {
				t.Fatalf("Preflight() error = %v, want nil", err)
			}
			if tt.wantIs != nil && !errors.Is(err, tt.wantIs) {
				t.Fatalf("Preflight() error = %v, want errors.Is(%v)", err, tt.wantIs)
			}
			for _, w := range tt.wantMsg {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("Preflight() error = %q, want it to contain %q", err, w)
				}
			}
			if tt.wantWarn == "" && warn.Len() > 0 {
				t.Errorf("warning = %q, want none", warn.String())
			}
			if !strings.Contains(warn.String(), tt.wantWarn) {
				t.Errorf("warning = %q, want it to contain %q", warn.String(), tt.wantWarn)
			}
		})
	}
}

// TestPreflightTarget covers the spec "Output location" checks.
func TestPreflightTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) string // returns the target
		wantErr string
	}{
		{
			name:  "missing target",
			setup: func(t *testing.T) string { return filepath.Join(t.TempDir(), "shop") },
		},
		{
			// Spec: Generate into the current empty directory.
			name:  "empty directory",
			setup: func(t *testing.T) string { return t.TempDir() },
		},
		{
			// Spec: Target directory is not empty.
			name: "non-empty directory",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "keep.txt"), "mine")
				return dir
			},
			wantErr: "is not empty",
		},
		{
			name: "target is a file",
			setup: func(t *testing.T) string {
				p := filepath.Join(t.TempDir(), "shop")
				writeFile(t, p, "")
				return p
			},
			wantErr: "is not a directory",
		},
		{
			name:    "missing parent",
			setup:   func(t *testing.T) string { return filepath.Join(t.TempDir(), "a", "b") },
			wantErr: "does not exist",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			target := tt.setup(t)
			before := snapshot(t, filepath.Dir(target))
			err := Preflight(context.Background(), Request{
				GoVersion: "1.27",
				Target:    target,
				GoCommand: fakeGoPath,
				Env:       fakeEnv(),
			})
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Preflight() error = %v, want nil", err)
				}
			} else if !errors.Is(err, ErrTarget) || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Preflight() error = %v, want ErrTarget containing %q", err, tt.wantErr)
			}
			if after := snapshot(t, filepath.Dir(target)); after != before {
				t.Errorf("Preflight changed the file system:\nbefore %s\nafter  %s", before, after)
			}
		})
	}
}

func TestReleaseVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in     string
		want   string
		wantOK bool
	}{
		{"go1.27.3", "v1.27.3", true},
		{"go1.27.3-X:nodwarf5", "v1.27.3", true},
		{"go1.28rc1", "v1.28.0", true},
		{"go1.27", "v1.27.0", true},
		{"devel go1.28-abc", "", false},
		{"", "", false},
	}
	for _, tt := range tests {
		got, ok := releaseVersion(tt.in)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("releaseVersion(%q) = %q, %v; want %q, %v", tt.in, got, ok, tt.want, tt.wantOK)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

// snapshot lists every path under dir with its size, for comparing a
// directory before and after an operation.
func snapshot(t *testing.T, dir string) string {
	t.Helper()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "(missing)"
	}
	var b strings.Builder
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		info, err := d.Info()
		if err != nil {
			return err
		}
		size := int64(0)
		if !d.IsDir() {
			size = info.Size()
		}
		b.WriteString(filepath.ToSlash(rel))
		b.WriteString(":")
		b.WriteString(strings.Repeat("#", int(min(size, 64))))
		b.WriteString(" ")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return b.String()
}
