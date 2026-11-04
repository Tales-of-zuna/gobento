package catalogjson

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Tales-of-zuna/gobento/internal/catalog"
	"github.com/Tales-of-zuna/gobento/internal/compose"
	"github.com/Tales-of-zuna/gobento/internal/generate"
)

// Document is the exported document, the root of the website's catalog
// schema.
type Document struct {
	Kinds      []Kind              `json:"kinds"`
	Components []Component         `json:"components"`
	Go         GoRule              `json:"go"`
	Modules    []Module            `json:"modules"`
	Supported  []map[string]string `json:"supported"`
	Rules      Rules               `json:"rules"`
}

// Kind describes a component kind.
type Kind struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Component is one component of the catalog.
type Component struct {
	// ID is "<kind>/<name>", as used by Module.Components.
	ID          string              `json:"id"`
	Kind        string              `json:"kind"`
	Name        string              `json:"name"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Provides    map[string]string   `json:"provides"`
	Requires    map[string][]string `json:"requires"`
}

// GoRule is the supported range of Go versions and the default.
type GoRule struct {
	Range   string `json:"range"`
	Default string `json:"default"`
}

// Module is one module of the version rules.
type Module struct {
	Path       string   `json:"path"`
	Components []string `json:"components"`
	// Traits restricts the module to selections whose traits have one of
	// the listed values; empty means no restriction.
	Traits      map[string][]string `json:"traits"`
	Role        string              `json:"role"`
	Range       string              `json:"range"`
	Default     string              `json:"default"`
	Constraints []Constraint        `json:"constraints"`
}

// Constraint restricts one target while the owning module's version lies in
// When. A rules-file constraint with several requirements becomes one
// Constraint per requirement, which means the same.
type Constraint struct {
	When     string      `json:"when"`
	Requires Requirement `json:"requires"`
}

// Requirement restricts Target ("go" or a module path) to Range.
type Requirement struct {
	Target string `json:"target"`
	Range  string `json:"range"`
}

// Rules are the CLI's rules for the project name and the module path.
type Rules struct {
	ProjectName NameRule `json:"projectName"`
	ModulePath  Rule     `json:"modulePath"`
}

// Rule is a regular expression a value must match, with the explanation to
// show when it does not.
type Rule struct {
	Pattern string `json:"pattern"`
	Message string `json:"message"`
}

// NameRule is the project name rule: a pattern and a list of names that
// match it but are reserved.
type NameRule struct {
	Rule
	Reserved        []string `json:"reserved"`
	ReservedMessage string   `json:"reservedMessage"`
}

// kindInfo holds the title and description of each kind. The catalog itself
// has no text for kinds; TestKindInfo checks that every kind has an entry.
var kindInfo = map[catalog.Kind]Kind{
	catalog.KindArchitecture: {Title: "Architecture", Description: "How the project's packages are laid out and wired together."},
	catalog.KindFramework:    {Title: "HTTP framework", Description: "Routes HTTP requests to handlers."},
	catalog.KindDatabase:     {Title: "Database", Description: "Where the project stores its data."},
	catalog.KindData:         {Title: "Data access", Description: "How the code reads and writes the database."},
}

// Load loads the catalog in fsys and converts it with FromCatalog.
func Load(fsys fs.FS) (*Document, error) {
	cat, err := catalog.Load(fsys)
	if err != nil {
		return nil, err
	}
	return FromCatalog(cat)
}

// FromCatalog converts cat to the document. Kinds are in canonical order,
// components and supported selections in catalog order, and modules sorted
// by path, as the rules hold them.
func FromCatalog(cat *catalog.Catalog) (*Document, error) {
	out := &Document{
		Components: []Component{},
		Modules:    []Module{},
		Supported:  []map[string]string{},
	}

	for _, k := range catalog.Kinds() {
		info, ok := kindInfo[k]
		if !ok {
			return nil, fmt.Errorf("kind %s has no title and description", k)
		}
		info.ID = string(k)
		out.Kinds = append(out.Kinds, info)
		for _, c := range cat.Components(k) {
			out.Components = append(out.Components, exportComponent(c))
		}
	}

	rules := cat.Rules()
	out.Go = GoRule{Range: rules.Go.Range.String(), Default: rules.Go.Default}
	for _, m := range rules.Modules {
		em := Module{
			Path:        m.Path,
			Components:  m.Components,
			Traits:      map[string][]string{},
			Role:        string(m.Role),
			Range:       m.Range.String(),
			Default:     m.Default,
			Constraints: []Constraint{},
		}
		maps.Copy(em.Traits, m.Traits)
		for _, c := range m.Constraints {
			for _, r := range c.Requires {
				em.Constraints = append(em.Constraints, Constraint{
					When:     c.When.String(),
					Requires: Requirement{Target: r.Target, Range: r.Range.String()},
				})
			}
		}
		out.Modules = append(out.Modules, em)
	}

	for _, sel := range compose.Supported(cat) {
		s := make(map[string]string, len(catalog.Kinds()))
		for _, c := range sel.Components() {
			s[string(c.Kind)] = c.Name
		}
		out.Supported = append(out.Supported, s)
	}

	out.Rules = Rules{
		ProjectName: NameRule{
			Rule:            Rule{Pattern: generate.NamePattern, Message: sentence(generate.NameRule)},
			Reserved:        generate.ReservedNames(),
			ReservedMessage: sentence(generate.ReservedNameRule),
		},
		ModulePath: Rule{Pattern: generate.ModulePathPattern, Message: sentence(generate.ModulePathRule)},
	}
	return out, nil
}

// exportComponent converts c, with empty rather than nil maps.
func exportComponent(c *catalog.Component) Component {
	ec := Component{
		ID:          c.ID(),
		Kind:        string(c.Kind),
		Name:        c.Name,
		Title:       c.Title,
		Description: c.Description,
		Provides:    map[string]string{},
		Requires:    map[string][]string{},
	}
	maps.Copy(ec.Provides, c.Provides)
	maps.Copy(ec.Requires, c.Requires)
	return ec
}

// sentence turns a CLI message fragment, which follows a colon, into a
// sentence for a form field: capitalized and ending in a period.
func sentence(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	s = string(unicode.ToUpper(r)) + s[size:]
	if !strings.HasSuffix(s, ".") {
		s += "."
	}
	return s
}

// Encode writes v, a Document or a part of one, as indented JSON with a
// trailing newline. HTML escaping is off so that ranges such as ">=1.26"
// stay readable.
func Encode(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
