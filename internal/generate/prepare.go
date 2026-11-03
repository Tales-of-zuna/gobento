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

// namePattern is the rule for project names.
var namePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

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
	if e.Value == "" {
		return fmt.Sprintf("choose a %s; supported: %s", e.Kind, supported)
	}
	return fmt.Sprintf("unknown %s %q; supported: %s", e.Kind, e.Value, supported)
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

	if !namePattern.MatchString(spec.Name) {
		errs = append(errs, fmt.Errorf("%w %q: start with a lowercase letter and use only lowercase letters, digits, - and _", ErrInvalidName, spec.Name))
	}
	if err := module.CheckImportPath(p.Module); err != nil {
		errs = append(errs, fmt.Errorf("%w %q: %w; use a path such as example.com/%s", ErrInvalidModule, p.Module, err, safeExample(spec.Name)))
	}

	sel, ids, selErrs := selectComponents(cat, spec.Components)
	errs = append(errs, selErrs...)

	// Versions are checked even when a kind is not resolved yet; every
	// component of such a kind then counts as possibly selected.
	res, err := cat.Rules().Resolve(compat.Request{Components: ids, Go: spec.Go, Pins: spec.Pins})
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

// safeExample returns name if it is a valid path element, or "app".
func safeExample(name string) string {
	if namePattern.MatchString(name) {
		return name
	}
	return "app"
}
