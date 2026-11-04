package compat

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"path"
	"regexp"
	"slices"
	"sort"
	"strings"

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
	// Assets lists every vendored asset, sorted by name.
	Assets []Asset
	// Binaries lists every platform binary, sorted by name.
	Binaries []Binary
	// Tooling lists the versions of the project tooling that are not Go
	// modules: container images, GitHub Actions, and tools.
	Tooling Tooling
	// Node describes the Node.js versions that scaffolding needs and the
	// release line of the generated tooling. It is zero without npm
	// packages.
	Node NodeRule
	// NPM lists every npm package components use, sorted by name.
	NPM []NPMPackage
}

// Asset is a file that components ship unmodified from an upstream release,
// such as a JavaScript library served by the generated project. Its version
// is fixed per gobento release, like a tool's, and cannot be pinned.
type Asset struct {
	// Name identifies the asset, for example htmx.org (its npm package).
	Name string
	// Components lists the components that ship the asset, as
	// "<kind>/<name>" identifiers.
	Components []string
	// Version is the upstream release, for example 2.0.11.
	Version string
	// Source is the https URL of the upstream release archive (a .tgz).
	Source string
	// Path is the asset's path inside the archive, for example
	// package/dist/htmx.min.js.
	Path string
	// SHA256 is the asset file's SHA-256 checksum in lowercase hex.
	SHA256 string
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
	// Traits optionally restricts the module to selections whose traits
	// have one of the listed values, for example sql-dialect: [postgresql]
	// for a dialect-specific driver. Empty means no restriction.
	Traits map[string][]string
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

// usedBy reports whether a selection with the given component identifiers
// and traits uses m: one of its components is selected and every trait
// condition holds. Nil traits mean the traits are not known yet, and the
// conditions are not checked.
func (m Module) usedBy(components []string, traits map[string]string) bool {
	if !slices.ContainsFunc(m.Components, func(c string) bool { return slices.Contains(components, c) }) {
		return false
	}
	if traits == nil {
		return true
	}
	for trait, allowed := range m.Traits {
		v, ok := traits[trait]
		if !ok || !slices.Contains(allowed, v) {
			return false
		}
	}
	return true
}

// traitCondition describes the module's trait conditions, for example
// "sql-dialect postgresql" or "sql-dialect sqlite or postgresql".
func (m Module) traitCondition() string {
	var parts []string
	for _, trait := range slices.Sorted(maps.Keys(m.Traits)) {
		parts = append(parts, trait+" "+strings.Join(m.Traits[trait], " or "))
	}
	return strings.Join(parts, " and ")
}

// rulesFile mirrors the YAML layout of the rules file.
type rulesFile struct {
	Go struct {
		Range   string `yaml:"range"`
		Default string `yaml:"default"`
	} `yaml:"go"`
	Modules  map[string]moduleFile `yaml:"modules"`
	Assets   map[string]assetFile  `yaml:"assets"`
	Binaries map[string]binaryFile `yaml:"binaries"`
	Tooling  toolingFile           `yaml:"tooling"`
	Node     nodeFile              `yaml:"node"`
	NPM      map[string]npmFile    `yaml:"npm"`
}

type assetFile struct {
	Components []string `yaml:"components"`
	Version    string   `yaml:"version"`
	Source     string   `yaml:"source"`
	Path       string   `yaml:"path"`
	SHA256     string   `yaml:"sha256"`
}

type moduleFile struct {
	Components  []string            `yaml:"components"`
	Traits      map[string][]string `yaml:"traits"`
	Role        Role                `yaml:"role"`
	Range       string              `yaml:"range"`
	Default     string              `yaml:"default"`
	Constraints []constraintFile    `yaml:"constraints"`
}

type constraintFile struct {
	When    string            `yaml:"when"`
	Require map[string]string `yaml:"require"`
}

// ParseRules parses and validates a rules file. It reports every problem it
// finds, joined with errors.Join: unknown fields, malformed ranges or
// versions, defaults outside their range, tool modules without an exact
// range, constraints that refer to unknown targets, assets without
// components or with a malformed version, source, path, or checksum,
// binaries without components or platforms or with a malformed version,
// platform, URL, or checksum, and malformed tooling entries. References to components are checked
// separately with CheckComponents, because the component catalog is loaded
// elsewhere.
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

	for _, name := range slices.Sorted(maps.Keys(file.Assets)) {
		a, aerrs := parseAsset(name, file.Assets[name])
		rules.Assets = append(rules.Assets, a)
		errs = append(errs, aerrs...)
		if _, ok := file.Modules[name]; ok {
			errs = append(errs, fmt.Errorf("asset %s: the name is also a module; use another one", name))
		}
	}

	for _, name := range slices.Sorted(maps.Keys(file.Binaries)) {
		b, berrs := parseBinary(name, file.Binaries[name])
		rules.Binaries = append(rules.Binaries, b)
		errs = append(errs, berrs...)
		_, isModule := file.Modules[name]
		_, isAsset := file.Assets[name]
		if isModule || isAsset {
			errs = append(errs, fmt.Errorf("binary %s: the name is also a module or an asset; use another one", name))
		}
	}
	npm, nerrs := parseNPM(file.NPM)
	rules.NPM = npm
	errs = append(errs, nerrs...)
	node, nodeErrs := parseNode(file.Node, len(npm) > 0)
	rules.Node = node
	errs = append(errs, nodeErrs...)

	tooling, terrs := parseTooling(file.Tooling)
	rules.Tooling = tooling
	errs = append(errs, terrs...)
	if node.IsZero() {
		for _, key := range slices.Sorted(maps.Keys(tooling.Images)) {
			if strings.Contains(tooling.Images[key].Tag, NodePlaceholder) {
				errs = append(errs, fmt.Errorf("tooling: image %s: the tag uses %s, but the rules have no node section", key, NodePlaceholder))
			}
		}
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

	m := Module{Path: path, Components: mf.Components, Traits: mf.Traits, Role: mf.Role, Default: mf.Default}
	for _, trait := range slices.Sorted(maps.Keys(mf.Traits)) {
		switch values := mf.Traits[trait]; {
		case trait == "":
			fail("traits: empty trait name")
		case len(values) == 0:
			fail("traits: %s lists no values", trait)
		case slices.Contains(values, ""):
			fail("traits: %s lists an empty value", trait)
		}
	}

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

var (
	// assetVersionPattern matches an asset version such as 2.0.11, as
	// upstream projects outside Go write it.
	assetVersionPattern = regexp.MustCompile(`^v?[0-9]+\.[0-9]+\.[0-9]+([-+][0-9A-Za-z.-]+)?$`)
	// sha256Pattern matches a SHA-256 checksum in lowercase hex.
	sha256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// parseAsset converts and validates one asset entry.
func parseAsset(name string, af assetFile) (Asset, []error) {
	var errs []error
	fail := func(format string, args ...any) {
		errs = append(errs, fmt.Errorf("asset %s: "+format, append([]any{name}, args...)...))
	}
	if strings.TrimSpace(name) == "" {
		fail("empty name")
	}
	if len(af.Components) == 0 {
		fail("no components listed")
	}
	if !assetVersionPattern.MatchString(af.Version) {
		fail("version %q must be a release such as 2.0.11", af.Version)
	}
	if u, err := url.Parse(af.Source); err != nil || u.Scheme != "https" || u.Host == "" {
		fail("source %q must be an https URL of the release archive", af.Source)
	}
	if af.Path == "" || path.IsAbs(af.Path) || path.Clean(af.Path) != af.Path || strings.HasPrefix(af.Path, "../") {
		fail("path %q must be a clean relative path inside the source archive", af.Path)
	}
	if !sha256Pattern.MatchString(af.SHA256) {
		fail("sha256 %q must be 64 lowercase hexadecimal digits", af.SHA256)
	}
	return Asset{Name: name, Components: af.Components, Version: af.Version, Source: af.Source, Path: af.Path, SHA256: af.SHA256}, errs
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
// exists returns false, naming the module, asset, or binary and the
// component.
func (r *Rules) CheckComponents(exists func(id string) bool) error {
	var errs []error
	for _, m := range r.Modules {
		for _, c := range m.Components {
			if !exists(c) {
				errs = append(errs, fmt.Errorf("module %s: unknown component %q", m.Path, c))
			}
		}
	}
	for _, a := range r.Assets {
		for _, c := range a.Components {
			if !exists(c) {
				errs = append(errs, fmt.Errorf("asset %s: unknown component %q", a.Name, c))
			}
		}
	}
	for _, b := range r.Binaries {
		for _, c := range b.Components {
			if !exists(c) {
				errs = append(errs, fmt.Errorf("binary %s: unknown component %q", b.Name, c))
			}
		}
	}
	for _, p := range r.NPM {
		for _, c := range p.Components {
			if !exists(c) {
				errs = append(errs, fmt.Errorf("npm %s: unknown component %q", p.Name, c))
			}
		}
	}
	return errors.Join(errs...)
}

// Asset returns the asset with the given name.
func (r *Rules) Asset(name string) (Asset, bool) {
	i := slices.IndexFunc(r.Assets, func(a Asset) bool { return a.Name == name })
	if i < 0 {
		return Asset{}, false
	}
	return r.Assets[i], true
}

// CheckTraits reports every trait named in a module's trait conditions for
// which known returns false, naming the module and the trait.
func (r *Rules) CheckTraits(known func(trait string) bool) error {
	var errs []error
	for _, m := range r.Modules {
		for _, trait := range slices.Sorted(maps.Keys(m.Traits)) {
			if !known(trait) {
				errs = append(errs, fmt.Errorf("module %s: trait condition on %q, which no component provides", m.Path, trait))
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
