package compat

import (
	"errors"
	"fmt"
	"slices"
	"sort"
)

// Errors returned by Resolve, wrapped with details. Test for them with
// errors.Is.
var (
	// ErrOutOfRange reports a chosen version outside its supported range.
	ErrOutOfRange = errors.New("version outside the supported range")
	// ErrNotSelected reports a pin for a module that no selected component
	// uses.
	ErrNotSelected = errors.New("module not used by the selection")
	// ErrNotPinnable reports a pin for a tool module, a vendored asset, or
	// a platform binary, whose version is fixed per gobento release.
	ErrNotPinnable = errors.New("version is fixed")
	// ErrConstraint reports a violated constraint between chosen versions.
	ErrConstraint = errors.New("incompatible versions")
)

// Request is the input to Resolve.
type Request struct {
	// Components lists the selected components as "<kind>/<name>"
	// identifiers.
	Components []string
	// Go is the Go version chosen by the user, or "" for the default.
	Go string
	// Pins maps module paths to versions chosen by the user.
	Pins map[string]string
	// Traits maps trait names to the values the selection provides; it
	// decides whether modules with trait conditions are used. Nil means the
	// traits are not known yet (for example, while a component kind is
	// unresolved): trait conditions are then ignored, so every module of a
	// selected component counts as used.
	Traits map[string]string
}

// Resolved is the outcome of Resolve: the exact versions a generated project
// uses.
type Resolved struct {
	// Go is the chosen Go version as written in go.mod, for example 1.27.
	Go string
	// Modules lists every module the selection uses, sorted by path.
	Modules []ModuleVersion
	// Assets lists every asset the selected components ship, sorted by
	// name.
	Assets []AssetVersion
	// Binaries lists every platform binary the selected components use,
	// sorted by name.
	Binaries []Binary
	// Tooling holds the tooling versions for the chosen Go version.
	Tooling ToolingVersions
	// NPM lists every npm package the selected components use, sorted by
	// name.
	NPM []NPMVersion
	// Node is the Node.js release line of the generated tooling, or "" if
	// the rules have none.
	Node string
}

// NPMVersions returns the npm package versions keyed by package name.
func (r Resolved) NPMVersions() map[string]string {
	out := make(map[string]string, len(r.NPM))
	for _, p := range r.NPM {
		out[p.Name] = p.Version
	}
	return out
}

// AssetVersion is one vendored asset with its version.
type AssetVersion struct {
	// Name is the asset's name in the rules file.
	Name string
	// Version is the vendored version.
	Version string
}

// AssetVersions returns the vendored versions keyed by asset name.
func (r Resolved) AssetVersions() map[string]string {
	out := make(map[string]string, len(r.Assets))
	for _, a := range r.Assets {
		out[a.Name] = a.Version
	}
	return out
}

// ModuleVersion is one module with its chosen version.
type ModuleVersion struct {
	// Path is the module path.
	Path string
	// Version is the chosen version.
	Version string
	// Role says whether the module is a library or a tool.
	Role Role
}

// Versions returns the chosen versions keyed by module path.
func (r Resolved) Versions() map[string]string {
	out := make(map[string]string, len(r.Modules))
	for _, m := range r.Modules {
		out[m.Path] = m.Version
	}
	return out
}

// Resolve chooses the Go version and the version of every module the
// selected components use. It starts from the defaults, applies the user's
// Go version and pins, checks every chosen version against its range, and
// then evaluates every applicable constraint. All problems are returned
// together, joined with errors.Join.
func (r *Rules) Resolve(req Request) (Resolved, error) {
	var errs []error
	res := Resolved{Go: r.Go.Default}

	if req.Go != "" {
		res.Go = req.Go
		norm, err := NormalizeGo(req.Go)
		switch {
		case err != nil:
			errs = append(errs, fmt.Errorf("go version: %w", err))
		case !r.Go.Range.Contains(norm):
			errs = append(errs, fmt.Errorf("%w: go %s; supported range is %q", ErrOutOfRange, req.Go, r.Go.Range))
		}
	}

	chosen := make(map[string]int) // module path -> index in res.Modules
	for _, m := range r.Modules {
		if !m.usedBy(req.Components, req.Traits) {
			continue
		}
		chosen[m.Path] = len(res.Modules)
		res.Modules = append(res.Modules, ModuleVersion{Path: m.Path, Version: m.Default, Role: m.Role})
	}

	for _, a := range r.Assets {
		if slices.ContainsFunc(a.Components, func(c string) bool { return slices.Contains(req.Components, c) }) {
			res.Assets = append(res.Assets, AssetVersion{Name: a.Name, Version: a.Version})
		}
	}

	for _, b := range r.Binaries {
		if slices.ContainsFunc(b.Components, func(c string) bool { return slices.Contains(req.Components, c) }) {
			res.Binaries = append(res.Binaries, b)
		}
	}

	for _, p := range r.NPM {
		if slices.ContainsFunc(p.Components, func(c string) bool { return slices.Contains(req.Components, c) }) {
			res.NPM = append(res.NPM, NPMVersion{Name: p.Name, Role: p.Role, Version: p.Version})
		}
	}

	for _, path := range sortedKeys(req.Pins) {
		version := req.Pins[path]
		m, known := r.Module(path)
		i, used := chosen[path]
		if a, ok := r.Asset(path); ok && !known {
			errs = append(errs, fmt.Errorf("%w: cannot pin %s, vendored asset versions are fixed per gobento release (this release ships %s)", ErrNotPinnable, path, a.Version))
			continue
		}
		if b, ok := r.Binary(path); ok && !known {
			errs = append(errs, fmt.Errorf("%w: cannot pin %s, binary versions are fixed per gobento release (this release uses %s)", ErrNotPinnable, path, b.Version))
			continue
		}
		if p, ok := r.NPMPackage(path); ok && !known {
			errs = append(errs, fmt.Errorf("%w: cannot pin %s, npm package versions are fixed per gobento release (this release uses %s)", ErrNotPinnable, path, p.Version))
			continue
		}
		switch {
		case known && !used && m.usedBy(req.Components, nil):
			errs = append(errs, fmt.Errorf("%w: cannot pin %s, it is only used with %s", ErrNotSelected, path, m.traitCondition()))
		case !known || !used:
			errs = append(errs, fmt.Errorf("%w: cannot pin %s, no selected component uses it", ErrNotSelected, path))
		case m.Role == RoleTool:
			errs = append(errs, fmt.Errorf("%w: cannot pin %s, tool versions are fixed per gobento release (this release uses %s)", ErrNotPinnable, path, m.Default))
		default:
			if err := CheckModuleVersion(version); err != nil {
				errs = append(errs, fmt.Errorf("pin %s: %w", path, err))
				continue
			}
			if !m.Range.Contains(version) {
				errs = append(errs, fmt.Errorf("%w: %s %s; supported range is %q", ErrOutOfRange, path, version, m.Range))
				continue
			}
			res.Modules[i].Version = version
		}
	}

	// Constraints are only meaningful once every chosen version is valid.
	if len(errs) == 0 {
		errs = append(errs, r.checkConstraints(res, chosen)...)
	}
	if err := errors.Join(errs...); err != nil {
		return Resolved{}, err
	}
	res.Node = r.Node.Default
	res.Tooling = r.Tooling.Resolve(res.Go, res.Node)
	return res, nil
}

// checkConstraints evaluates every constraint of every chosen module whose
// When range contains the chosen version. A requirement on a module that is
// not part of the selection does not apply.
func (r *Rules) checkConstraints(res Resolved, chosen map[string]int) []error {
	goVersion, err := NormalizeGo(res.Go)
	if err != nil {
		return []error{fmt.Errorf("go version: %w", err)}
	}

	var errs []error
	for _, mv := range res.Modules {
		m, _ := r.Module(mv.Path)
		for _, c := range m.Constraints {
			if !c.When.Contains(mv.Version) {
				continue
			}
			for _, req := range c.Requires {
				if req.Target == GoTarget {
					if !req.Range.Contains(goVersion) {
						errs = append(errs, fmt.Errorf("%w: %s %s requires go %s, but go %s is selected",
							ErrConstraint, mv.Path, mv.Version, req.Range, res.Go))
					}
					continue
				}
				i, ok := chosen[req.Target]
				if !ok {
					continue
				}
				other := res.Modules[i]
				if !req.Range.Contains(other.Version) {
					errs = append(errs, fmt.Errorf("%w: %s %s requires %s %s, but %s %s is selected",
						ErrConstraint, mv.Path, mv.Version, req.Target, req.Range, other.Path, other.Version))
				}
			}
		}
	}
	return errs
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
