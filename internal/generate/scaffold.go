package generate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Tales-of-zuna/gobento/internal/catalog"
	"github.com/Tales-of-zuna/gobento/internal/compat"
)

// Errors of the scaffold step, wrapped with details. Test for them with
// errors.Is.
var (
	// ErrNoNode reports that the node or npm command was not found.
	ErrNoNode = errors.New("node or npm command not found")
	// ErrNodeVersion reports a Node.js version outside the rules' range.
	ErrNodeVersion = errors.New("unsupported Node.js version")
	// ErrNPMCommand reports a failed npm command, such as npm create
	// without network access.
	ErrNPMCommand = errors.New("npm command failed")
)

// lockFile is the npm lock file the scaffold step writes.
const lockFile = "package-lock.json"

// stagingPrefix starts the name of the directory a scaffolder runs in,
// inside the temporary project directory.
const stagingPrefix = ".gobento-scaffold-"

// Scaffold is one project directory an npm scaffolder creates during
// generation.
type Scaffold struct {
	// Component is the ID of the component that declares it, for example
	// frontend/react.
	Component string
	// Package is the scaffolder's npm package, for example create-vite.
	Package string
	// Version is the scaffolder's exact version, for example 9.3.3.
	Version string
	// Template is the scaffolder's template, for example react-ts.
	Template string
	// Dir is the directory it creates, relative to the project root, with
	// forward slashes, for example frontend.
	Dir string
	// Remove lists scaffolded paths, relative to Dir, to delete.
	Remove []string
	// Dependencies and DevDependencies are npm packages, as name@version,
	// to add to the scaffolded project.
	Dependencies, DevDependencies []string
}

// CreateArgs returns the arguments of the npm command that runs the
// scaffolder into the directory base, for example
// create --yes vite@9.3.3 frontend -- --no-interactive --no-immediate --template react-ts.
func (s Scaffold) CreateArgs(base string) []string {
	return []string{
		"create", "--yes", strings.TrimPrefix(s.Package, "create-") + "@" + s.Version, base,
		"--", "--no-interactive", "--no-immediate", "--template", s.Template,
	}
}

// Command returns the scaffolder command as users would type it, for
// messages: npm create vite@9.3.3 frontend -- --template react-ts.
func (s Scaffold) Command() string {
	return "npm create " + strings.TrimPrefix(s.Package, "create-") + "@" + s.Version + " " + s.Dir + " -- --template " + s.Template
}

// scaffolds returns the scaffold steps of the selection's components.
func scaffolds(sel catalog.Selection, res compat.Resolved) []Scaffold {
	var deps, devDeps []string
	for _, p := range res.NPM {
		switch p.Role {
		case compat.NPMDependency:
			deps = append(deps, p.Name+"@"+p.Version)
		case compat.NPMDevDependency:
			devDeps = append(devDeps, p.Name+"@"+p.Version)
		}
	}
	versions := res.NPMVersions()
	var out []Scaffold
	for _, c := range sel.Components() {
		if c.Scaffold == nil {
			continue
		}
		out = append(out, Scaffold{
			Component:       c.ID(),
			Package:         c.Scaffold.With,
			Version:         versions[c.Scaffold.With],
			Template:        c.Scaffold.Template,
			Dir:             c.Scaffold.Dir,
			Remove:          c.Scaffold.Remove,
			Dependencies:    deps,
			DevDependencies: devDeps,
		})
	}
	return out
}

// node describes the local Node.js installation.
type node struct {
	npm string // absolute path of the npm command
}

// checkNode locates node and npm and checks the Node.js version against
// req.Node. It returns every problem together.
func checkNode(ctx context.Context, req Request) (node, error) {
	needs := needsNode(req)
	var errs []error
	nodePath, err := exec.LookPath(req.NodeCommand)
	if err != nil {
		errs = append(errs, fmt.Errorf("%w: the node command was not found (%w); %s; install Node.js, which includes npm, from https://nodejs.org/", ErrNoNode, err, needs))
	} else {
		cmd := exec.CommandContext(ctx, nodePath, "--version") //nolint:gosec // G204: nodePath is the node command found on PATH
		cmd.Dir = os.TempDir()
		cmd.Env = req.Env
		out, err := cmd.Output()
		version := strings.TrimSpace(string(out))
		if err != nil {
			errs = append(errs, fmt.Errorf("%w: %s --version: %w", ErrNoNode, nodePath, err))
		} else if ok, err := req.Node.CheckNodeVersion(version); err != nil || !ok {
			errs = append(errs, fmt.Errorf("%w: the local Node.js is %s; %s; install a newer Node.js from https://nodejs.org/", ErrNodeVersion, version, needs))
		}
	}
	npmPath, err := exec.LookPath(req.NPMCommand)
	if err != nil && nodePath != "" {
		errs = append(errs, fmt.Errorf("%w: the npm command was not found (%w); %s; install npm with Node.js from https://nodejs.org/", ErrNoNode, err, needs))
	}
	return node{npm: npmPath}, errors.Join(errs...)
}

// needsNode explains which components need Node.js, for preflight errors.
func needsNode(req Request) string {
	var parts []string
	for _, s := range req.Scaffolds {
		parts = append(parts, fmt.Sprintf("%s is created with %s", s.Component, s.Command()))
	}
	return strings.Join(parts, " and ") + fmt.Sprintf(", which needs Node.js %s and npm", req.Node.Range)
}

// scaffold runs every scaffold step of req in dir, the temporary project
// directory that already holds the plan's files.
func scaffold(ctx context.Context, n node, dir string, req Request) error {
	for _, s := range req.Scaffolds {
		if err := scaffoldOne(ctx, n, dir, req, s); err != nil {
			return fmt.Errorf("scaffold %s with %s: %w", s.Dir, s.Command(), err)
		}
	}
	return nil
}

func scaffoldOne(ctx context.Context, n node, dir string, req Request, s Scaffold) error {
	staging, err := os.MkdirTemp(dir, stagingPrefix+"*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(staging) }()

	base := filepath.Base(filepath.FromSlash(s.Dir))
	if err := runNPM(ctx, n.npm, staging, req.Env, s.CreateArgs(base)...); err != nil {
		return err
	}
	created := filepath.Join(staging, base)
	if info, err := os.Stat(created); err != nil || !info.IsDir() {
		return fmt.Errorf("%s did not create %s", s.Package, base)
	}
	for _, r := range s.Remove {
		p := filepath.Join(created, filepath.FromSlash(r))
		if _, err := os.Lstat(p); err != nil {
			return fmt.Errorf("%s@%s did not create %s, which %s removes; update its scaffold.remove list", s.Package, s.Version, r, s.Component)
		}
		if err := os.RemoveAll(p); err != nil {
			return err
		}
	}
	target := filepath.Join(dir, filepath.FromSlash(s.Dir))
	if err := merge(created, target); err != nil {
		return err
	}

	install := []string{"install", "--package-lock-only", "--ignore-scripts", "--no-audit", "--no-fund"}
	if len(s.Dependencies) > 0 {
		if err := lockOnly(ctx, n.npm, target, req.Env, append(append(install, "--save-prod"), s.Dependencies...)); err != nil {
			return err
		}
	}
	if len(s.DevDependencies) > 0 {
		if err := lockOnly(ctx, n.npm, target, req.Env, append(append(install, "--save-dev"), s.DevDependencies...)); err != nil {
			return err
		}
	}
	if len(s.Dependencies) == 0 && len(s.DevDependencies) == 0 {
		if err := lockOnly(ctx, n.npm, target, req.Env, install); err != nil {
			return err
		}
	}
	if _, err := os.Stat(filepath.Join(target, lockFile)); err != nil {
		return fmt.Errorf("npm install did not write %s", lockFile)
	}
	return nil
}

// lockOnly runs the npm install command args, which writes only the lock
// file, in dir. npm 12, whose allow-remote defaults to none, refuses in this
// mode the registry tarballs of a package with bundleDependencies (such as
// @tailwindcss/oxide-wasm32-wasi, an optional dependency of Tailwind CSS),
// although it installs them (npm/cli#9800). Then lockOnly runs the command
// again with --allow-remote=all and checks what npm 12 is meant to ensure:
// that every tarball in the lock file comes from the configured registry's
// host.
func lockOnly(ctx context.Context, npmPath, dir string, env, args []string) error {
	err := runNPM(ctx, npmPath, dir, env, args...)
	if err == nil || !strings.Contains(err.Error(), "EALLOWREMOTE") {
		return err
	}
	out, rerr := npmOutput(ctx, npmPath, dir, env, "config", "get", "registry")
	if rerr != nil {
		return errors.Join(err, rerr)
	}
	registry, perr := url.Parse(strings.TrimSpace(out))
	if perr != nil || registry.Host == "" {
		return errors.Join(err, fmt.Errorf("npm config get registry printed %q, not a URL", strings.TrimSpace(out)))
	}
	if err := runNPM(ctx, npmPath, dir, env, append(slices.Clone(args), "--allow-remote=all")...); err != nil {
		return err
	}
	return checkLockHosts(filepath.Join(dir, lockFile), registry.Host)
}

// checkLockHosts reports a package of the npm lock file at path whose
// tarball URL is not on host.
func checkLockHosts(path, host string) error {
	data, err := os.ReadFile(path) //nolint:gosec // G304: the lock file npm just wrote in the project directory
	if err != nil {
		return err
	}
	var lock struct {
		Packages map[string]struct {
			Resolved string `json:"resolved"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(data, &lock); err != nil {
		return fmt.Errorf("%s: %w", lockFile, err)
	}
	for _, name := range slices.Sorted(maps.Keys(lock.Packages)) {
		resolved := lock.Packages[name].Resolved
		u, err := url.Parse(resolved)
		if resolved == "" || (err == nil && u.Host == host && (u.Scheme == "https" || u.Scheme == "http")) {
			continue
		}
		return fmt.Errorf("%w: %s resolves %s to %s, which is not on the npm registry %s", ErrNPMCommand, lockFile, name, resolved, host)
	}
	return nil
}

// merge moves every file under src into dst, keeping files that already
// exist in dst: gobento's own files replace the scaffolded ones.
func merge(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, dirMode)
		}
		if _, err := os.Lstat(target); err == nil {
			return nil
		} else if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return os.Rename(p, target) //nolint:gosec // G122: both paths are inside gobento's temporary directory; WalkDir does not follow symbolic links and Rename moves a link itself
	})
}

// runNPM runs npm in dir. It is stopped, with every process it started,
// when ctx is cancelled. Questions are impossible: standard input is
// empty and npm's own prompt is answered by --yes.
func runNPM(ctx context.Context, npmPath, dir string, env []string, args ...string) error {
	_, err := npmOutput(ctx, npmPath, dir, env, args...)
	return err
}

// npmOutput is runNPM, returning npm's output.
func npmOutput(ctx context.Context, npmPath, dir string, env []string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, npmPath, args...) //nolint:gosec // G204: npmPath is the npm command found on PATH and args come from the rules
	cmd.Dir = dir
	cmd.Env = append(append([]string{}, env...),
		"npm_config_yes=true", "npm_config_fund=false", "npm_config_audit=false", "npm_config_update_notifier=false")
	cmd.WaitDelay = 5 * time.Second
	killTree(cmd)
	// The error shows both streams in order; the result is stdout only,
	// without npm's warnings.
	var all, stdout bytes.Buffer
	shared := &lockedWriter{w: &all}
	cmd.Stdout = io.MultiWriter(shared, &stdout)
	cmd.Stderr = shared
	if err := cmd.Run(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", fmt.Errorf("npm %s: %w", strings.Join(args, " "), ctxErr)
		}
		return "", fmt.Errorf("%w: npm %s: %w\n%s", ErrNPMCommand, strings.Join(args, " "), err, strings.TrimSpace(all.String()))
	}
	return stdout.String(), nil
}

// lockedWriter serializes writes to w from npm's stdout and stderr.
type lockedWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (l *lockedWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(p)
}
