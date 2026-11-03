package generate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang.org/x/mod/modfile"

	"github.com/Tales-of-zuna/gobento/internal/compat"
	"github.com/Tales-of-zuna/gobento/internal/compose"
)

// ErrRaised reports that dependency resolution changed a chosen version,
// which means the rules file is missing a constraint.
var ErrRaised = errors.New("dependency resolution changed a chosen version")

// ErrGoCommand reports a failed go command, such as go get without network
// access.
var ErrGoCommand = errors.New("go command failed")

// tempPrefix starts the name of the temporary directory a project is built
// in, next to the target.
const tempPrefix = ".gobento-tmp-"

// Request describes one generation.
type Request struct {
	// Plan is the composed project.
	Plan *compose.Plan
	// Module is the module path written to go.mod.
	Module string
	// GoVersion is the selected Go version, for example 1.27. go.mod gets
	// its full form, 1.27.0.
	GoVersion string
	// Modules lists the chosen module versions. Library modules are
	// required in go.mod; tool modules are ignored.
	Modules []compat.ModuleVersion
	// Target is the directory to create or fill.
	Target string
	// GoCommand is the go command to run: a name looked up on PATH, or a
	// path to the binary. The default is "go".
	GoCommand string
	// Env is the environment for go commands. The default is the process
	// environment. GOWORK and GOFLAGS are always overridden; everything
	// else, such as GOPROXY and GOPRIVATE, is passed through unchanged.
	Env []string
	// Warn receives warnings. The default discards them.
	Warn io.Writer
}

// withDefaults fills in the defaults of r and makes Target absolute.
func (r Request) withDefaults() (Request, error) {
	if r.GoCommand == "" {
		r.GoCommand = "go"
	}
	if r.Env == nil {
		r.Env = os.Environ()
	}
	if r.Warn == nil {
		r.Warn = io.Discard
	}
	abs, err := filepath.Abs(r.Target)
	if err != nil {
		return r, fmt.Errorf("%w: %w", ErrTarget, err)
	}
	r.Target = abs
	return r, nil
}

// Generate creates the project described by req in req.Target:
//
//  1. Preflight checks the go command and the target.
//  2. The plan and go.mod are written into a temporary directory: next to
//     a missing target, or inside an existing empty one.
//  3. go get fetches the chosen library versions, then go mod tidy runs.
//  4. go.mod is checked: the go directive and every chosen version must be
//     unchanged, and a toolchain line is removed.
//  5. The result is moved into place.
//
// If any step fails or ctx is cancelled, everything Generate created is
// removed and a pre-existing target is left as it was.
func Generate(ctx context.Context, req Request) (err error) {
	req, err = req.withDefaults()
	if err != nil {
		return err
	}
	tc, err := preflight(ctx, req)
	if err != nil {
		return err
	}
	goDirective, err := compat.GoDirective(req.GoVersion)
	if err != nil {
		return err
	}
	exists, err := targetExists(req.Target)
	if err != nil {
		return err
	}

	tmp, err := workDir(req.Target, exists)
	if err != nil {
		return err
	}
	var moved []string
	defer func() {
		if err == nil {
			return
		}
		if cerr := cleanup(tmp, moved); cerr != nil {
			err = errors.Join(err, cerr)
		}
	}()

	if err := writePlan(tmp, req.Plan, req.Module, goDirective); err != nil {
		return err
	}
	if err := resolve(ctx, tc.path, tmp, req); err != nil {
		return err
	}
	if err := recheck(ctx, tc.path, tmp, req, goDirective); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	moved, err = place(ctx, tmp, req.Target, exists)
	return err
}

// writePlan writes every planned file and go.mod into dir.
func writePlan(dir string, plan *compose.Plan, module, goDirective string) error {
	for _, f := range plan.Files {
		rel := filepath.FromSlash(f.Path)
		if !filepath.IsLocal(rel) {
			return fmt.Errorf("write %s: the path leaves the project directory", f.Path)
		}
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), dirMode); err != nil {
			return fmt.Errorf("write %s: %w", f.Path, err)
		}
		if err := os.WriteFile(p, f.Content, f.Mode.Perm()); err != nil {
			return fmt.Errorf("write %s: %w", f.Path, err)
		}
	}
	gomod := fmt.Sprintf("module %s\n\ngo %s\n", modfile.AutoQuote(module), goDirective)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644); err != nil { //nolint:gosec // G306: project files are meant to be readable
		return fmt.Errorf("write go.mod: %w", err)
	}
	return nil
}

// resolve fetches every chosen library version with one go get, then tidies
// the module.
func resolve(ctx context.Context, goPath, dir string, req Request) error {
	var args []string
	for _, m := range req.Modules {
		if m.Role == compat.RoleLibrary {
			args = append(args, m.Path+"@"+m.Version)
		}
	}
	if len(args) > 0 {
		if err := runGo(ctx, goPath, dir, req.Env, append([]string{"get"}, args...)...); err != nil {
			return err
		}
	}
	return runGo(ctx, goPath, dir, req.Env, "mod", "tidy")
}

// recheck removes a toolchain line added by the local go command and
// verifies that go.mod has exactly the chosen go directive and versions.
func recheck(ctx context.Context, goPath, dir string, req Request, goDirective string) error {
	f, err := readModFile(dir)
	if err != nil {
		return err
	}
	if f.Toolchain != nil {
		if err := runGo(ctx, goPath, dir, req.Env, "mod", "edit", "-toolchain=none"); err != nil {
			return err
		}
		if f, err = readModFile(dir); err != nil {
			return err
		}
	}

	var errs []error
	if f.Go == nil || f.Go.Version != goDirective {
		got := "none"
		if f.Go != nil {
			got = f.Go.Version
		}
		errs = append(errs, fmt.Errorf("%w: the go directive was raised from %s to %s; the rules file is missing a constraint on the Go version",
			ErrRaised, goDirective, got))
	}
	required := make(map[string]string)
	for _, r := range f.Require {
		if !r.Indirect {
			required[r.Mod.Path] = r.Mod.Version
		}
	}
	for _, m := range req.Modules {
		if m.Role != compat.RoleLibrary {
			continue
		}
		got, ok := required[m.Path]
		switch {
		case !ok:
			errs = append(errs, fmt.Errorf("%w: %s %s is not a direct requirement after go mod tidy; no generated code imports it",
				ErrRaised, m.Path, m.Version))
		case got != m.Version:
			errs = append(errs, fmt.Errorf("%w: %s was raised from %s to %s; the rules file is missing a constraint for %s %s",
				ErrRaised, m.Path, m.Version, got, m.Path, m.Version))
		}
	}
	return errors.Join(errs...)
}

func readModFile(dir string) (*modfile.File, error) {
	p := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(p) //nolint:gosec // G304: the path is inside gobento's own temporary directory
	if err != nil {
		return nil, fmt.Errorf("read go.mod: %w", err)
	}
	f, err := modfile.Parse(p, data, nil)
	if err != nil {
		return nil, fmt.Errorf("parse go.mod: %w", err)
	}
	return f, nil
}

// runGo runs the go command in dir with the controlled environment. It is
// killed when ctx is cancelled.
func runGo(ctx context.Context, goPath, dir string, env []string, args ...string) error {
	cmd := exec.CommandContext(ctx, goPath, args...) //nolint:gosec // G204: goPath is the located go command and args are module versions from the rules
	cmd.Dir = dir
	cmd.Env = goEnv(env)
	cmd.WaitDelay = 5 * time.Second
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("go %s: %w", strings.Join(args, " "), ctxErr)
		}
		return fmt.Errorf("%w: go %s: %w\n%s", ErrGoCommand, strings.Join(args, " "), err, strings.TrimSpace(out.String()))
	}
	return nil
}

// goEnv returns env with the settings gobento controls: no workspace, and
// go commands may update go.mod. The user's other GOFLAGS are kept; only a
// -mod flag is replaced.
func goEnv(env []string) []string {
	out := make([]string, 0, len(env)+2)
	var flags []string
	for _, kv := range env {
		key, value, _ := strings.Cut(kv, "=")
		switch {
		case envKeyIs(key, "GOWORK"):
			continue
		case envKeyIs(key, "GOFLAGS"):
			flags = nil // the last setting wins, as in the go command
			for _, f := range strings.Fields(value) {
				if !strings.HasPrefix(f, "-mod=") && !strings.HasPrefix(f, "--mod=") {
					flags = append(flags, f)
				}
			}
			continue
		}
		out = append(out, kv)
	}
	flags = append(flags, "-mod=mod")
	return append(out, "GOWORK=off", "GOFLAGS="+strings.Join(flags, " "))
}

// envKeyIs reports whether the environment variable name key is name.
// Windows treats names case-insensitively.
func envKeyIs(key, name string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(key, name)
	}
	return key == name
}
