package compat

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
)

// NodePlaceholder in an image tag stands for the Node.js release line of
// the rules file, for example "{node}-alpine" becomes "26-alpine".
const NodePlaceholder = "{node}"

// NPMRole says how a component uses an npm package.
type NPMRole string

// The npm package roles.
const (
	// NPMScaffolder packages create a project directory at generation
	// time, run as npm create <name without create->@<version>.
	NPMScaffolder NPMRole = "scaffolder"
	// NPMDependency packages are added to the dependencies of every
	// scaffolded frontend of a selection that uses them.
	NPMDependency NPMRole = "dependency"
	// NPMDevDependency packages are added to the devDependencies of every
	// scaffolded frontend of a selection that uses them.
	NPMDevDependency NPMRole = "dev-dependency"
)

// scaffolderPrefix starts the name of every scaffolder package, as npm
// create requires.
const scaffolderPrefix = "create-"

var (
	// npmNamePattern matches npm package names, scoped or not.
	npmNamePattern = regexp.MustCompile(`^(?:@[a-z0-9][a-z0-9._-]*/)?[a-z0-9][a-z0-9._-]*$`)
	// npmVersionPattern matches an exact npm version such as 9.3.3.
	npmVersionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?$`)
	// nodeLinePattern matches a Node.js release line such as 26.
	nodeLinePattern = regexp.MustCompile(`^[1-9][0-9]*$`)
)

// NodeRule describes the Node.js versions of generation and of the
// generated tooling. The zero NodeRule means the rules file has none.
type NodeRule struct {
	// Range is the range of Node.js versions that can run the scaffolders,
	// for example >=v22.12.0, compared with the output of node --version.
	Range Range
	// Default is the Node.js release line of the generated CI workflow
	// and Dockerfile, for example 26.
	Default string
}

// IsZero reports whether the rules file has no Node.js rule.
func (n NodeRule) IsZero() bool {
	return n.Default == "" && n.Range.comparators == nil
}

// NPMPackage is an npm package that components use, with a version fixed
// per gobento release.
type NPMPackage struct {
	// Name is the package name, for example create-vite.
	Name string
	// Components lists the components that use the package, as
	// "<kind>/<name>" identifiers.
	Components []string
	// Traits optionally restricts the package to selections whose traits
	// have one of the listed values, as for modules.
	Traits map[string][]string
	// Role says how the package is used.
	Role NPMRole
	// Version is the exact version, for example 9.3.3.
	Version string
}

// CreateName returns the name npm create takes for a scaffolder: its name
// without the create- prefix, for example vite for create-vite.
func (p NPMPackage) CreateName() string {
	return strings.TrimPrefix(p.Name, scaffolderPrefix)
}

// NPMVersion is one npm package of a selection with its version.
type NPMVersion struct {
	Name    string
	Role    NPMRole
	Version string
}

type nodeFile struct {
	Range   string `yaml:"range"`
	Default string `yaml:"default"`
}

type npmFile struct {
	Components []string            `yaml:"components"`
	Traits     map[string][]string `yaml:"traits"`
	Role       NPMRole             `yaml:"role"`
	Version    string              `yaml:"version"`
}

// parseNode converts and validates the node section. An absent section is
// valid when no npm package needs it.
func parseNode(nf nodeFile, needed bool) (NodeRule, []error) {
	if nf.Range == "" && nf.Default == "" {
		if needed {
			return NodeRule{}, []error{errors.New("node: npm packages are listed, so the Node.js range and default are required")}
		}
		return NodeRule{}, nil
	}
	var errs []error
	rng, err := ParseRange(nf.Range)
	if err != nil {
		errs = append(errs, fmt.Errorf("node: range: %w", err))
	}
	if !nodeLinePattern.MatchString(nf.Default) {
		errs = append(errs, fmt.Errorf("node: default %q must be a release line such as 26", nf.Default))
	} else if err == nil && !rng.Contains(nodeLineVersion(nf.Default)) {
		errs = append(errs, fmt.Errorf("node: default %s is outside its range %q", nf.Default, rng))
	}
	return NodeRule{Range: rng, Default: nf.Default}, errs
}

// nodeLineVersion returns the first version of a release line, for example
// v26.0.0 for 26.
func nodeLineVersion(line string) string {
	return "v" + line + ".0.0"
}

// parseNPM converts and validates the npm section, sorted by name.
func parseNPM(files map[string]npmFile) ([]NPMPackage, []error) {
	var (
		out  []NPMPackage
		errs []error
	)
	for _, name := range slices.Sorted(maps.Keys(files)) {
		f := files[name]
		fail := func(format string, args ...any) {
			errs = append(errs, fmt.Errorf("npm %s: "+format, append([]any{name}, args...)...))
		}
		if !npmNamePattern.MatchString(name) {
			fail("invalid package name")
		}
		if len(f.Components) == 0 {
			fail("no components listed")
		}
		checkTraitConditions(f.Traits, fail)
		switch f.Role {
		case NPMScaffolder:
			if !strings.HasPrefix(name, scaffolderPrefix) || name == scaffolderPrefix {
				fail("a scaffolder's name starts with %q, as npm create requires", scaffolderPrefix)
			}
		case NPMDependency, NPMDevDependency:
		default:
			fail("role %q must be %q, %q, or %q", f.Role, NPMScaffolder, NPMDependency, NPMDevDependency)
		}
		if !npmVersionPattern.MatchString(f.Version) {
			fail("version %q must be an exact version such as 9.3.3", f.Version)
		}
		out = append(out, NPMPackage{Name: name, Components: f.Components, Traits: f.Traits, Role: f.Role, Version: f.Version})
	}
	return out, errs
}

// NPMPackage returns the npm package with the given name.
func (r *Rules) NPMPackage(name string) (NPMPackage, bool) {
	i := slices.IndexFunc(r.NPM, func(p NPMPackage) bool { return p.Name == name })
	if i < 0 {
		return NPMPackage{}, false
	}
	return r.NPM[i], true
}

// CheckNodeVersion reports whether the output of node --version, such as
// v24.22.0, lies in the Node.js range.
func (n NodeRule) CheckNodeVersion(version string) (bool, error) {
	v := strings.TrimSpace(version)
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	if err := CheckModuleVersion(v); err != nil {
		return false, fmt.Errorf("unexpected Node.js version %q", strings.TrimSpace(version))
	}
	return n.Range.Contains(v), nil
}
