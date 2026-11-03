package generate

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"golang.org/x/mod/modfile"
)

// The test binary doubles as a fake go command. TestMain copies it into a
// temporary directory as go (go.exe on Windows); when it runs with
// GOBENTO_FAKE_GO=1 it behaves as the go command, configured by these
// environment variables:
//
//	FAKE_GO_VERSION        output of go env GOVERSION (default go1.27.3)
//	FAKE_GOTOOLCHAIN       output of go env GOTOOLCHAIN (default auto)
//	FAKE_GO_LOG            file that receives each invocation and its env
//	FAKE_GO_GET=fail       go get fails as if offline
//	FAKE_GO_GET=hang       go get creates FAKE_GO_STARTED and blocks
//	FAKE_GO_RAISE=mod@ver  go get requires mod at ver instead
//	FAKE_GO_RAISE_GO=ver   go get raises the go directive to ver
//	FAKE_GO_TOOLCHAIN_LINE go get adds a toolchain line
var fakeGoPath string

func TestMain(m *testing.M) {
	if os.Getenv("GOBENTO_FAKE_GO") == "1" {
		os.Exit(fakeGo(os.Args[1:]))
	}
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	dir, err := os.MkdirTemp("", "gobento-fake-go-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer func() { _ = os.RemoveAll(dir) }()
	if fakeGoPath, err = installFakeGo(dir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return m.Run()
}

func installFakeGo(dir string) (string, error) {
	self, err := os.Executable()
	if err != nil {
		return "", err
	}
	name := "go"
	if runtime.GOOS == "windows" {
		name = "go.exe"
	}
	dst := filepath.Join(dir, name)
	src, err := os.Open(self) //nolint:gosec // G304: copying the running test binary
	if err != nil {
		return "", err
	}
	defer func() { _ = src.Close() }()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755) //nolint:gosec // G302: the fake go command must be executable
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(out, src); err != nil {
		_ = out.Close()
		return "", err
	}
	return dst, out.Close()
}

func fakeGo(args []string) int {
	if log := os.Getenv("FAKE_GO_LOG"); log != "" {
		f, err := os.OpenFile(log, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600) //nolint:gosec // G304: test-controlled path
		if err == nil {
			_, _ = fmt.Fprintf(f, "go %s | GOWORK=%s GOFLAGS=%s GOPROXY=%s\n",
				strings.Join(args, " "), os.Getenv("GOWORK"), os.Getenv("GOFLAGS"), os.Getenv("GOPROXY"))
			_ = f.Close()
		}
	}
	switch {
	case len(args) == 3 && args[0] == "env":
		fmt.Println(envOr("FAKE_GO_VERSION", "go1.27.3"))
		fmt.Println(envOr("FAKE_GOTOOLCHAIN", "auto"))
		return 0
	case len(args) > 0 && args[0] == "get":
		return fakeGet(args[1:])
	case len(args) == 2 && args[0] == "mod" && args[1] == "tidy":
		return 0
	case len(args) == 3 && args[0] == "mod" && args[1] == "edit" && args[2] == "-toolchain=none":
		return editModFile(func(f *modfile.File) error {
			f.DropToolchainStmt()
			return nil
		})
	}
	fmt.Fprintf(os.Stderr, "fake go: unexpected arguments %q\n", args)
	return 2
}

func fakeGet(mods []string) int {
	switch os.Getenv("FAKE_GO_GET") {
	case "fail":
		fmt.Fprintln(os.Stderr, "go: example.com/lib@v1.0.0: dial tcp: lookup proxy.golang.org: no such host")
		return 1
	case "hang":
		_ = os.WriteFile(os.Getenv("FAKE_GO_STARTED"), nil, 0o600) //nolint:gosec // G703: test-controlled path
		time.Sleep(time.Minute)
		return 1
	}
	return editModFile(func(f *modfile.File) error {
		raised := os.Getenv("FAKE_GO_RAISE")
		for _, m := range mods {
			path, version, _ := strings.Cut(m, "@")
			if rp, rv, ok := strings.Cut(raised, "@"); ok && rp == path {
				version = rv
			}
			if err := f.AddRequire(path, version); err != nil {
				return err
			}
		}
		if v := os.Getenv("FAKE_GO_RAISE_GO"); v != "" {
			if err := f.AddGoStmt(v); err != nil {
				return err
			}
		}
		if os.Getenv("FAKE_GO_TOOLCHAIN_LINE") != "" {
			return f.AddToolchainStmt("go1.27.3")
		}
		return nil
	})
}

func editModFile(edit func(*modfile.File) error) int {
	data, err := os.ReadFile("go.mod")
	if err == nil {
		var f *modfile.File
		if f, err = modfile.Parse("go.mod", data, nil); err == nil {
			if err = edit(f); err == nil {
				f.Cleanup()
				if data, err = f.Format(); err == nil {
					err = os.WriteFile("go.mod", data, 0o600) //nolint:gosec // G703: fixed file name in the fake's working directory
				}
			}
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "fake go:", err)
		return 1
	}
	return 0
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// fakeEnv returns the environment for the fake go command with the given
// extra settings.
func fakeEnv(extra ...string) []string {
	return append(append(os.Environ(), "GOBENTO_FAKE_GO=1"), extra...)
}
