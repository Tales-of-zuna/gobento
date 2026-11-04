package generate

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/mod/module"

	"github.com/Tales-of-zuna/gobento/internal/catalog"
	"github.com/Tales-of-zuna/gobento/internal/compat"
	"github.com/Tales-of-zuna/gobento/internal/compose"
)

// Errors returned by Prepare, wrapped with details. Test for them with
// errors.Is.
var (
	// ErrInvalidName reports a project name that breaks the naming rules.
	ErrInvalidName = errors.New("invalid project name")
	// ErrInvalidModule reports an invalid module path.
	ErrInvalidModule = errors.New("invalid module path")
)

// Project name and module path rules. They are exported so that tools
// outside the CLI, such as the website's configurator (through
// internal/tools/catalogexport), apply the same rules.
const (
	// NamePattern is the regular expression a project name must match. It
	// uses only syntax that Go's regexp and JavaScript's RegExp share.
	NamePattern = `^[a-z][a-z0-9_-]*$`
	// NameRule explains NamePattern; it completes the message of an
	// ErrInvalidName error.
	NameRule = "start with a lowercase letter and use only lowercase letters, digits, - and _"
	// ReservedNameRule explains why a name in ReservedNames is rejected.
	ReservedNameRule = "the name is reserved on Windows and cannot be used for files or directories; choose another name"

	// ModulePathPattern approximates module.CheckImportPath, which Prepare
	// uses to validate module paths, as a regular expression for clients
	// that cannot call it. It accepts every path CheckImportPath accepts:
	// non-empty elements of ASCII letters, digits, and "-._~+", separated
	// by single slashes, with no element ending in "." and no leading "-".
	// It does not reject the path elements CheckImportPath forbids for
	// Windows (device names such as "con" or "con.go", and short names such
	// as "abc~1"), which the CLI then still reports.
	ModulePathPattern = `^(?:[A-Za-z0-9_~+]|[A-Za-z0-9._~+][A-Za-z0-9._~+-]*[A-Za-z0-9_~+-])(?:/[A-Za-z0-9._~+-]*[A-Za-z0-9_~+-])*$`
	// ModulePathRule explains ModulePathPattern.
	ModulePathRule = "use a Go module path such as example.com/shop: elements of letters, digits, '.', '_', '~', '+' and '-', separated by '/', none ending in '.'"
)

// namePattern is the compiled NamePattern.
var namePattern = regexp.MustCompile(NamePattern)

// ReservedNames returns the project names Windows reserves as device names
// for files and directories, which gobento rejects on every system: con,
// prn, aux, nul, com1 to com9, and lpt1 to lpt9.
func ReservedNames() []string {
	names := []string{"con", "prn", "aux", "nul"}
	for _, prefix := range []string{"com", "lpt"} {
		for d := '1'; d <= '9'; d++ {
			names = append(names, prefix+string(d))
		}
	}
	return names
}

// CheckName reports whether name follows the project name rules, with an
// error wrapping ErrInvalidName that explains the rule it breaks. Prepare
// applies the same check.
func CheckName(name string) error {
	switch {
	case !namePattern.MatchString(name):
		return fmt.Errorf("%w %q: %s", ErrInvalidName, name, NameRule)
	case isWindowsReserved(name):
		return fmt.Errorf("%w %q: %s", ErrInvalidName, name, ReservedNameRule)
	}
	return nil
}

// CheckModule reports whether path is a valid module path, with an error
// wrapping ErrInvalidModule that suggests a path built from the project
// name. Prepare applies the same check.
func CheckModule(path, name string) error {
	if err := module.CheckImportPath(path); err != nil {
		return fmt.Errorf("%w %q: %w; use a path such as example.com/%s", ErrInvalidModule, path, err, safeExample(name))
	}
	return nil
}

// Spec is the user's description of a project, before validation.
type Spec struct {
	// Name is the project name, for example shop.
	Name string
	// Module is the module path; "" means the project name.
	Module string
	// Components maps each kind to the chosen component name. A missing or
	// empty entry selects the kind's only supported component.
	Components map[catalog.Kind]string
	// Go is the chosen Go version, or "" for the default.
	Go string
	// Pins maps module paths to chosen versions.
	Pins map[string]string
}

// Prepared is a validated, composed project, ready to be generated.
type Prepared struct {
	// Name is the project name.
	Name string
	// Module is the module path.
	Module string
	// Selection is the resolved selection.
	Selection catalog.Selection
	// Versions holds the chosen Go and module versions.
	Versions compat.Resolved
	// Plan is the composed project.
	Plan *compose.Plan
}

// Request returns the generation request for p in target.
func (p *Prepared) Request(target string) Request {
	return Request{
		Plan:      p.Plan,
		Module:    p.Module,
		GoVersion: p.Versions.Go,
		Modules:   p.Versions.Modules,
		Target:    target,
	}
}

// SelectionError reports a component kind whose choice is missing or not
// supported.
type SelectionError struct {
	// Kind is the component kind.
	Kind catalog.Kind
	// Value is the chosen name, or "" if none was given.
	Value string
	// Supported lists the supported names for the kind.
	Supported []string
}

func (e *SelectionError) Error() string {
	supported := strings.Join(e.Supported, ", ")
	if supported == "" {
		supported = "none"
	}
	noun := string(e.Kind)
	if e.Kind == catalog.KindData {
		noun = "data-access library" // "a data" does not read
	}
	if e.Value == "" {
		article := "a"
		if strings.ContainsRune("aeiou", rune(noun[0])) {
			article = "an"
		}
		return fmt.Sprintf("choose %s %s; supported: %s", article, noun, supported)
	}
	return fmt.Sprintf("unknown %s %q; supported: %s", noun, e.Value, supported)
}

// Prepare validates spec against cat and composes the project, entirely in
// memory (step 1 of the generation pipeline). It reports every problem
// together: the name, the module path, each component kind, the versions,
// and composition.
func Prepare(cat *catalog.Catalog, spec Spec) (*Prepared, error) {
	var errs []error
	p := &Prepared{Name: spec.Name, Module: spec.Module}
	if p.Module == "" {
		p.Module = spec.Name
	}

	if err := CheckName(spec.Name); err != nil {
		errs = append(errs, err)
	}
	if err := CheckModule(p.Module, spec.Name); err != nil {
		errs = append(errs, err)
	}

	sel, ids, selErrs := selectComponents(cat, spec.Components)
	errs = append(errs, selErrs...)

	// Versions are checked even when a kind is not resolved yet; every
	// component of such a kind then counts as possibly selected, and so
	// does every module with a trait condition (nil traits).
	var traits map[string]string
	if len(selErrs) == 0 {
		traits = sel.Traits()
	}
	res, err := cat.Rules().Resolve(compat.Request{Components: ids, Go: spec.Go, Pins: spec.Pins, Traits: traits})
	if err != nil || len(selErrs) > 0 {
		return nil, errors.Join(append(errs, err)...)
	}
	p.Selection = sel
	p.Versions = res
	goDirective, err := compat.GoDirective(res.Go)
	if err != nil {
		return nil, errors.Join(append(errs, err)...)
	}
	if len(errs) > 0 {
		// Composition needs a valid name and module path.
		return nil, errors.Join(errs...)
	}

	plan, err := compose.Compose(compose.Input{
		Project:   compose.Project{Name: p.Name, Module: p.Module, GoVersion: goDirective},
		Selection: sel,
		Versions:  res.Versions(),
		Base:      cat.Base(),
		Tooling:   res.Tooling,
	})
	if err != nil {
		return nil, err
	}
	p.Plan = plan
	return p, nil
}

// selectComponents resolves the chosen component names into a supported
// selection, auto-selecting kinds with a single supported component. It also
// returns the IDs of the components whose versions apply: the selected ones,
// and every component of a kind that could not be resolved.
func selectComponents(cat *catalog.Catalog, chosen map[catalog.Kind]string) (catalog.Selection, []string, []error) {
	supported := compose.Supported(cat)
	var (
		errs  []error
		ids   []string
		comps = make(map[catalog.Kind]*catalog.Component)
	)
	for _, kind := range catalog.Kinds() {
		names := supportedNames(supported, kind)
		value := chosen[kind]
		switch {
		case value == "" && len(names) == 1:
			value = names[0]
		case value == "" || !slices.Contains(names, value):
			errs = append(errs, &SelectionError{Kind: kind, Value: value, Supported: names})
			for _, c := range cat.Components(kind) {
				ids = append(ids, c.ID())
			}
			continue
		}
		c, _ := cat.Component(kind, value)
		comps[kind] = c
		ids = append(ids, c.ID())
	}
	if len(errs) > 0 {
		return catalog.Selection{}, ids, errs
	}
	sel := catalog.Selection{
		Architecture: comps[catalog.KindArchitecture],
		Framework:    comps[catalog.KindFramework],
		Database:     comps[catalog.KindDatabase],
		Data:         comps[catalog.KindData],
	}
	if !slices.ContainsFunc(supported, func(s catalog.Selection) bool { return s.String() == sel.String() }) {
		if err := compose.Check(cat, sel); err != nil {
			return catalog.Selection{}, ids, []error{fmt.Errorf("the combination %s is not supported: %w", sel, err)}
		}
	}
	return sel, ids, nil
}

// supportedNames returns the sorted, unique names of kind in the supported
// selections.
func supportedNames(supported []catalog.Selection, kind catalog.Kind) []string {
	var names []string
	for _, s := range supported {
		for _, c := range s.Components() {
			if c.Kind == kind && !slices.Contains(names, c.Name) {
				names = append(names, c.Name)
			}
		}
	}
	slices.Sort(names)
	return names
}

// isWindowsReserved reports whether name is one of ReservedNames.
func isWindowsReserved(name string) bool {
	return slices.Contains(ReservedNames(), name)
}

// safeExample returns name if it is a valid path element, or "app".
func safeExample(name string) string {
	if namePattern.MatchString(name) && !isWindowsReserved(name) {
		return name
	}
	return "app"
}
