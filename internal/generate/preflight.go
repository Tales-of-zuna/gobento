package generate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/Tales-of-zuna/gobento/internal/compat"
)

// Errors returned by Generate before anything is written, wrapped with details. Test for them with
// errors.Is.
var (
	// ErrNoGo reports that the go command was not found.
	ErrNoGo = errors.New("go command not found")
	// ErrToolchain reports that the selected Go version is newer than the
	// local toolchain and GOTOOLCHAIN forbids downloading it.
	ErrToolchain = errors.New("go toolchain too old")
	// ErrTarget reports a target that is not an empty or missing directory.
	ErrTarget = errors.New("unusable target directory")
)

// toolchain describes the local go command.
type toolchain struct {
	path      string // absolute path of the go command
	version   string // GOVERSION, for example go1.27.3
	goToolEnv string // GOTOOLCHAIN, for example auto or local
}

// preflight checks, before anything is written, that the go command exists,
// that it can provide the selected Go version, and that the target is a
// missing or empty directory whose parent exists. Warnings, such as a Go
// toolchain that will be downloaded, are written to req.Warn. All problems
// are returned together.
func preflight(ctx context.Context, req Request) (toolchain, error) {
	var errs []error
	tc, err := findToolchain(ctx, req)
	if err != nil {
		errs = append(errs, err)
	} else if err := checkGoVersion(tc, req.GoVersion, req.Warn); err != nil {
		errs = append(errs, err)
	}
	if err := checkTarget(req.Target); err != nil {
		errs = append(errs, err)
	}
	return tc, errors.Join(errs...)
}

// findToolchain locates the go command and asks it for its version and its
// GOTOOLCHAIN setting.
func findToolchain(ctx context.Context, req Request) (toolchain, error) {
	path, err := exec.LookPath(req.GoCommand)
	if err != nil {
		return toolchain{}, fmt.Errorf("%w: %w; install Go from https://go.dev/dl/ and make sure go is on your PATH", ErrNoGo, err)
	}
	cmd := exec.CommandContext(ctx, path, "env", "GOVERSION", "GOTOOLCHAIN") //nolint:gosec // G204: path is the go command found on PATH
	// Run outside any module, so a go.mod in the working directory cannot
	// trigger a toolchain switch.
	cmd.Dir = os.TempDir()
	cmd.Env = goEnv(req.Env)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return toolchain{}, fmt.Errorf("%w: %s env: %w: %s", ErrNoGo, path, err, strings.TrimSpace(stderr.String()))
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	tc := toolchain{path: path}
	if len(lines) > 0 {
		tc.version = strings.TrimSpace(lines[0])
	}
	if len(lines) > 1 {
		tc.goToolEnv = strings.TrimSpace(lines[1])
	}
	return tc, nil
}

// checkGoVersion compares the selected Go version with the local toolchain.
// An unparsable local version, such as a development build, is not an
// error; the go command then decides.
func checkGoVersion(tc toolchain, selected string, warn io.Writer) error {
	want, err := compat.NormalizeGo(selected)
	if err != nil {
		return err
	}
	local, ok := releaseVersion(tc.version)
	if !ok || semver.Compare(want, local) <= 0 {
		return nil
	}
	localShown := strings.TrimPrefix(local, "v")
	if downloadsDisabled(tc.goToolEnv) {
		return fmt.Errorf("%w: Go %s is selected, but the local toolchain is Go %s and GOTOOLCHAIN=%s prevents downloading a newer one; install Go %s or choose an older version with --go",
			ErrToolchain, selected, localShown, tc.goToolEnv, selected)
	}
	// A warning that cannot be written is not a reason to stop.
	_, _ = fmt.Fprintf(warn, "warning: Go %s is selected and the local toolchain is Go %s; the go command will download Go %s\n",
		selected, localShown, selected)
	return nil
}

// releaseVersion extracts the release from a GOVERSION value such as
// go1.27.3 or go1.27.3-X:nodwarf5, as a semantic version.
func releaseVersion(goversion string) (string, bool) {
	v, ok := strings.CutPrefix(goversion, "go")
	if !ok {
		return "", false
	}
	if i := strings.IndexFunc(v, func(r rune) bool { return (r < '0' || r > '9') && r != '.' }); i >= 0 {
		v = v[:i]
	}
	norm, err := compat.NormalizeGo(v)
	if err != nil {
		return "", false
	}
	return norm, true
}

// downloadsDisabled reports whether a GOTOOLCHAIN value forbids the go
// command from downloading toolchains.
func downloadsDisabled(goToolchain string) bool {
	return goToolchain == "local" || goToolchain == "path" || strings.HasSuffix(goToolchain, "+path")
}

// checkTarget accepts a missing directory whose parent exists, or an empty
// directory. A symbolic link to an empty directory is accepted; a broken
// link is not.
func checkTarget(target string) error {
	linfo, err := os.Lstat(target)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		parent := filepath.Dir(target)
		if pinfo, err := os.Stat(parent); err != nil || !pinfo.IsDir() {
			return fmt.Errorf("%w: %s: the parent directory %s does not exist", ErrTarget, target, parent)
		}
		return nil
	case err != nil:
		return fmt.Errorf("%w: %w", ErrTarget, err)
	}
	info := linfo
	if linfo.Mode()&fs.ModeSymlink != 0 {
		if info, err = os.Stat(target); err != nil {
			return fmt.Errorf("%w: %s is a symbolic link that cannot be followed: %w", ErrTarget, target, err)
		}
	}
	if !info.IsDir() {
		return fmt.Errorf("%w: %s exists and is not a directory", ErrTarget, target)
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrTarget, err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("%w: %s is not empty; choose a new or empty directory with --dir", ErrTarget, target)
	}
	return nil
}
