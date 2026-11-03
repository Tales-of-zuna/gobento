package compat

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"sort"

	"go.yaml.in/yaml/v3"
	"golang.org/x/mod/module"
)

// GoTarget is the constraint target that refers to the Go version rather than
// to a module.
const GoTarget = "go"

// Role says how a module is used by a generated project.
type Role string

// The module roles.
const (
	// RoleLibrary modules are required in the generated go.mod. Users may
	// pin them to any version inside the supported range.
	RoleLibrary Role = "library"
	// RoleTool modules are run as tools and never appear in go.mod. Their
	// version is fixed per gobento release, so their range is exact.
	RoleTool Role = "tool"
)

// Rules is the parsed and validated version-compatibility rules file.
type Rules struct {
	// Go describes the supported Go versions.
	Go GoRule
	// Modules lists every module used by some component, sorted by path.
	Modules []Module
}

// GoRule describes the supported Go versions for generated projects.
type GoRule struct {
	// Range is the supported range of Go versions.
	Range Range
	// Default is the Go version used when the user does not choose one,
	// written as in go.mod (for example 1.27).
	Default string
}

// Module describes the supported versions of one module.
type Module struct {
	// Path is the module path, for example modernc.org/sqlite.
	Path string
	// Components lists the components that use the module, as
	// "<kind>/<name>" identifiers such as database/sqlite.
	Components []string
	// Role says whether the module is a library or a tool.
	Role Role
	// Range is the supported range of versions.
	Range Range
	// Default is the version used when the user does not pin one.
	Default string
	// Constraints restrict other versions depending on this module's
	// chosen version.
	Constraints []Constraint
}

// Constraint applies its requirements whenever the owning module's chosen
// version lies inside When.
type Constraint struct {
	// When is the range of the owning module's versions the constraint
	// applies to.
	When Range
	// Requires lists the versions other targets must then have.
	Requires []Requirement
}

// Requirement restricts one target, GoTarget or a module path, to a range.
type Requirement struct {
	// Target is GoTarget or the path of another module in the rules.
	Target string
	// Range is the range the target's chosen version must lie in.
	Range Range
}

// usedBy reports whether any of the selected component identifiers uses m.
func (m Module) usedBy(components []string) bool {
	for _, c := range m.Components {
		if slices.Contains(components, c) {
			return true
		}
	}
	return false
}

// rulesFile mirrors the YAML layout of the rules file.
type rulesFile struct {
	Go struct {
		Range   string `yaml:"range"`
		Default string `yaml:"default"`
	} `yaml:"go"`
	Modules map[string]moduleFile `yaml:"modules"`
}

type moduleFile struct {
	Components  []string         `yaml:"components"`
	Role        Role             `yaml:"role"`
	Range       string           `yaml:"range"`
	Default     string           `yaml:"default"`
	Constraints []constraintFile `yaml:"constraints"`
}

type constraintFile struct {
	When    string            `yaml:"when"`
	Require map[string]string `yaml:"require"`
}

// ParseRules parses and validates a rules file. It reports every problem it
// finds, joined with errors.Join: unknown fields, malformed ranges or
// versions, defaults outside their range, tool modules without an exact
// range, and constraints that refer to unknown targets. References to
// components are checked separately with CheckComponents, because the
// component catalog is loaded elsewhere.
func ParseRules(data []byte) (*Rules, error) {
	var file rulesFile
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&file); err != nil {
		return nil, fmt.Errorf("parse rules: %w", err)
	}

	var errs []error
	rules := &Rules{}

	goRange, err := ParseGoRange(file.Go.Range)
	if err != nil {
		errs = append(errs, fmt.Errorf("go: range: %w", err))
	}
	rules.Go = GoRule{Range: goRange, Default: file.Go.Default}
	if def, err := NormalizeGo(file.Go.Default); err != nil {
		errs = append(errs, fmt.Errorf("go: default: %w", err))
	} else if goRange.comparators != nil && !goRange.Contains(def) {
		errs = append(errs, fmt.Errorf("go: default %s is outside its range %q", file.Go.Default, goRange))
	}

	paths := make([]string, 0, len(file.Modules))
	for p := range file.Modules {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, p := range paths {
		m, merrs := parseModule(p, file.Modules[p], file.Modules)
		rules.Modules = append(rules.Modules, m)
		errs = append(errs, merrs...)
	}

	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("invalid rules: %w", err)
	}
	return rules, nil
}

// parseModule converts and validates one module entry. known holds every
// module of the file, for checking constraint targets.
func parseModule(path string, mf moduleFile, known map[string]moduleFile) (Module, []error) {
	var errs []error
	fail := func(format string, args ...any) {
		errs = append(errs, fmt.Errorf("module %s: "+format, append([]any{path}, args...)...))
	}

	if err := module.CheckPath(path); err != nil {
		fail("invalid module path: %w", err)
	}
	if len(mf.Components) == 0 {
		fail("no components listed")
	}
	if mf.Role != RoleLibrary && mf.Role != RoleTool {
		fail("role %q must be %q or %q", mf.Role, RoleLibrary, RoleTool)
	}

	m := Module{Path: path, Components: mf.Components, Role: mf.Role, Default: mf.Default}

	rng, rangeErr := ParseRange(mf.Range)
	if rangeErr != nil {
		fail("range: %w", rangeErr)
	}
	m.Range = rng

	if err := CheckModuleVersion(mf.Default); err != nil {
		fail("default: %w", err)
	} else if rangeErr == nil && !rng.Contains(mf.Default) {
		fail("default %s is outside its range %q", mf.Default, rng)
	}

	if mf.Role == RoleTool && rangeErr == nil {
		if _, ok := rng.Exact(); !ok {
			fail("tool range %q must be a single exact version such as =v1.2.3", rng)
		}
	}

	for i, cf := range mf.Constraints {
		c, cerrs := parseConstraint(cf, known)
		for _, err := range cerrs {
			fail("constraint %d: %w", i+1, err)
		}
		m.Constraints = append(m.Constraints, c)
	}
	return m, errs
}

func parseConstraint(cf constraintFile, known map[string]moduleFile) (Constraint, []error) {
	var errs []error
	when, err := ParseRange(cf.When)
	if err != nil {
		errs = append(errs, fmt.Errorf("when: %w", err))
	}
	c := Constraint{When: when}
	if len(cf.Require) == 0 {
		errs = append(errs, errors.New("require lists no targets"))
	}

	targets := make([]string, 0, len(cf.Require))
	for t := range cf.Require {
		targets = append(targets, t)
	}
	sort.Strings(targets)

	for _, target := range targets {
		var (
			rng Range
			err error
		)
		switch _, isModule := known[target]; {
		case target == GoTarget:
			rng, err = ParseGoRange(cf.Require[target])
		case isModule:
			rng, err = ParseRange(cf.Require[target])
		default:
			errs = append(errs, fmt.Errorf("require: unknown target %q: want %q or a module listed in the rules", target, GoTarget))
			continue
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("require %s: %w", target, err))
			continue
		}
		c.Requires = append(c.Requires, Requirement{Target: target, Range: rng})
	}
	return c, errs
}

// CheckComponents reports every component reference in the rules for which
// exists returns false, naming the module and the component.
func (r *Rules) CheckComponents(exists func(id string) bool) error {
	var errs []error
	for _, m := range r.Modules {
		for _, c := range m.Components {
			if !exists(c) {
				errs = append(errs, fmt.Errorf("module %s: unknown component %q", m.Path, c))
			}
		}
	}
	return errors.Join(errs...)
}

// Module returns the rules for the module with the given path.
func (r *Rules) Module(path string) (Module, bool) {
	i := slices.IndexFunc(r.Modules, func(m Module) bool { return m.Path == path })
	if i < 0 {
		return Module{}, false
	}
	return r.Modules[i], true
}
