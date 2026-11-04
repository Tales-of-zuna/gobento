package catalog

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Kind is a component kind. A selection contains exactly one component of
// each required kind and at most one of each optional kind.
type Kind string

// The component kinds, in the order used for listing, notes, and errors.
const (
	KindArchitecture Kind = "architecture"
	KindFramework    Kind = "framework"
	KindDatabase     Kind = "database"
	KindData         Kind = "data"
	// KindFrontend is optional: a project without a frontend is API-only.
	KindFrontend Kind = "frontend"
)

// None is the name that selects no component of an optional kind, for
// example --frontend none. No component of an optional kind may use it.
const None = "none"

// Kinds returns every component kind in canonical order.
func Kinds() []Kind {
	return []Kind{KindArchitecture, KindFramework, KindDatabase, KindData, KindFrontend}
}

// valid reports whether k is one of the known kinds.
func (k Kind) valid() bool {
	return slices.Contains(Kinds(), k)
}

// Optional reports whether a selection may leave out a component of kind k.
func (k Kind) Optional() bool {
	return k == KindFrontend
}

// RootSlot is the pseudo-slot for files placed at the project root. Its
// templates live under files/root/ in a component directory.
const RootSlot = "root"

// Component is one loaded and validated component.
type Component struct {
	// Kind is the component's kind, taken from its directory.
	Kind Kind
	// Name is the component's name, equal to its directory name.
	Name string
	// Title is a short human-readable name, such as "SQLite".
	Title string
	// Description is an optional one-sentence summary.
	Description string
	// Slots lists the slots an architecture declares, sorted by name.
	// Components of other kinds declare no slots.
	Slots []Slot
	// Provides maps trait names to the values this component provides.
	Provides map[string]string
	// Requires maps trait names to the values this component accepts. A
	// selection must provide one of them for every listed trait.
	Requires map[string][]string
	// Files lists the component's templates, sorted by slot and path.
	Files []File
	// Notes is the template of the component's notes.md.tmpl, or "" if it
	// has none. Rendered notes are appended to the project README.
	Notes string
	// Contributions maps contribution points of the base layer to the
	// component's template for that point, from contrib/<point>.tmpl.
	Contributions map[string]string
	// Points lists the contribution points the base layer declares,
	// sorted by name. Only the base layer declares points.
	Points []Point
}

// Point is a contribution point of the base layer: a named place in its
// files where the selected components' contributions are inserted.
type Point struct {
	// Name identifies the point, for example ci-services. It is also the
	// file name, without .tmpl, of a contribution in contrib/.
	Name string
	// Description says what the point accepts and where it ends up.
	Description string
}

// BaseID is the ID of the base layer, which has no kind.
const BaseID = "base"

// ID returns the component's identifier, "<kind>/<name>", as used in the
// rules file, or BaseID for the base layer.
func (c *Component) ID() string {
	if c.Kind == "" {
		return c.Name
	}
	return string(c.Kind) + "/" + c.Name
}

// IsBase reports whether c is the base layer.
func (c *Component) IsBase() bool {
	return c.Kind == "" && c.Name == BaseID
}

// Point returns the contribution point with the given name declared by c.
func (c *Component) Point(name string) (Point, bool) {
	i := slices.IndexFunc(c.Points, func(p Point) bool { return p.Name == name })
	if i < 0 {
		return Point{}, false
	}
	return c.Points[i], true
}

// Slot returns the slot with the given name declared by c.
func (c *Component) Slot(name string) (Slot, bool) {
	i := slices.IndexFunc(c.Slots, func(s Slot) bool { return s.Name == name })
	if i < 0 {
		return Slot{}, false
	}
	return c.Slots[i], true
}

// FilledSlots returns the names of the slots c has template files for, in
// sorted order, excluding RootSlot.
func (c *Component) FilledSlots() []string {
	var out []string
	for _, f := range c.Files {
		if f.Slot != RootSlot && !slices.Contains(out, f.Slot) {
			out = append(out, f.Slot)
		}
	}
	slices.Sort(out)
	return out
}

// Slot is a directory of the generated project, which is one Go package,
// declared by an architecture and filled by one component.
type Slot struct {
	// Name identifies the slot, for example handlers. It is also the
	// directory name under files/ that holds the templates for the slot.
	Name string
	// Dir is the slot's directory relative to the project root. It is a
	// template evaluated with the project data, for example
	// "cmd/{{.Project.Name}}".
	Dir string
	// Package is the Go package name of the slot. If empty, it is the last
	// element of the rendered Dir.
	Package string
	// FilledBy is the kind of component that fills the slot.
	FilledBy Kind
	// Required reports whether a selection must fill the slot.
	Required bool
}

// File is one template file of a component.
type File struct {
	// Slot is the slot the file belongs to, or RootSlot.
	Slot string
	// Path is the output path relative to the slot directory, with the
	// .tmpl suffix removed and dot- prefixes mapped to ".", using forward
	// slashes.
	Path string
	// Source is the template's path in the catalog file system, for error
	// messages.
	Source string
	// Template is the template text, with CRLF line endings normalized to LF.
	Template string
}

// Selection is one component of each required kind and at most one of each
// optional kind. A nil field of an optional kind selects no component.
type Selection struct {
	Architecture *Component
	Framework    *Component
	Database     *Component
	Data         *Component
	// Frontend is optional; nil means an API-only project.
	Frontend *Component
}

// all returns the fields of s in kind order, including nil ones.
func (s Selection) all() []*Component {
	return []*Component{s.Architecture, s.Framework, s.Database, s.Data, s.Frontend}
}

// Component returns the selected component of kind k, or nil if there is
// none.
func (s Selection) Component(k Kind) *Component {
	i := slices.Index(Kinds(), k)
	if i < 0 {
		return nil
	}
	return s.all()[i]
}

// With returns s with c as its component of c's kind.
func (s Selection) With(c *Component) Selection {
	switch c.Kind {
	case KindArchitecture:
		s.Architecture = c
	case KindFramework:
		s.Framework = c
	case KindDatabase:
		s.Database = c
	case KindData:
		s.Data = c
	case KindFrontend:
		s.Frontend = c
	}
	return s
}

// Components returns the selected components in kind order, skipping kinds
// that have no component.
func (s Selection) Components() []*Component {
	var out []*Component
	for _, c := range s.all() {
		if c != nil {
			out = append(out, c)
		}
	}
	return out
}

// String returns the selection as names in kind order, for example
// "mvc/nethttp/sqlite/sqlc", or "mvc/nethttp/sqlite/sqlc/htmx" with a
// frontend. A missing required kind is shown as "-"; an optional kind
// without a component is left out.
func (s Selection) String() string {
	names := make([]string, 0, len(Kinds()))
	for i, c := range s.all() {
		switch {
		case c != nil:
			names = append(names, c.Name)
		case !Kinds()[i].Optional():
			names = append(names, "-")
		}
	}
	return strings.Join(names, "/")
}

// Traits returns the traits provided by the selected components. If two
// components provide different values for the same trait, the value of the
// earlier one in kind order is kept; CheckTraits reports the conflict.
func (s Selection) Traits() map[string]string {
	out := make(map[string]string)
	for _, c := range s.Components() {
		for k, v := range c.Provides {
			if _, ok := out[k]; !ok {
				out[k] = v
			}
		}
	}
	return out
}

// CheckTraits reports every unmet trait requirement of the selection and
// every trait provided with conflicting values, naming the components and
// traits involved. It returns nil when the traits are consistent.
func (s Selection) CheckTraits() error {
	type provision struct {
		value string
		by    *Component
	}
	provided := make(map[string]provision)
	var errs []error
	for _, c := range s.Components() {
		for _, k := range sortedKeys(c.Provides) {
			v := c.Provides[k]
			if p, ok := provided[k]; ok && p.value != v {
				errs = append(errs, fmt.Errorf("trait %s: %s provides %q, but %s provides %q",
					k, p.by.ID(), p.value, c.ID(), v))
				continue
			}
			provided[k] = provision{value: v, by: c}
		}
	}
	for _, c := range s.Components() {
		for _, k := range sortedKeys(c.Requires) {
			allowed := c.Requires[k]
			p, ok := provided[k]
			switch {
			case !ok:
				errs = append(errs, fmt.Errorf("trait %s: %s requires one of [%s], but no selected component provides it",
					k, c.ID(), strings.Join(allowed, ", ")))
			case !slices.Contains(allowed, p.value):
				errs = append(errs, fmt.Errorf("trait %s: %s requires one of [%s], but %s provides %q",
					k, c.ID(), strings.Join(allowed, ", "), p.by.ID(), p.value))
			}
		}
	}
	return errors.Join(errs...)
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
