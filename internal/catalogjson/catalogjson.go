package catalogjson

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"slices"
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
	Assets     []Asset             `json:"assets"`
	Binaries   []Binary            `json:"binaries"`
	Tooling    Tooling             `json:"tooling"`
	Node       NodeRule            `json:"node"`
	NPM        []NPMPackage        `json:"npm"`
	Supported  []map[string]string `json:"supported"`
	Rules      Rules               `json:"rules"`
}

// Kind describes a component kind. A supported selection has no entry for
// an optional kind it leaves out.
type Kind struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Optional    bool   `json:"optional"`
}

// Asset is a file that components vendor from an upstream release, with a
// version fixed per gobento release.
type Asset struct {
	Name       string   `json:"name"`
	Components []string `json:"components"`
	Version    string   `json:"version"`
	Source     string   `json:"source"`
}

// Binary is a program that generated projects download on demand, one
// file per platform, with a version fixed per gobento release.
type Binary struct {
	Name       string   `json:"name"`
	Components []string `json:"components"`
	Version    string   `json:"version"`
	// Platforms lists the supported platforms as "<GOOS>/<GOARCH>",
	// sorted.
	Platforms []string `json:"platforms"`
}

// NodeRule is the Node.js range that scaffolding needs and the release
// line of the generated tooling; both are "" without npm packages.
type NodeRule struct {
	Range   string `json:"range"`
	Default string `json:"default"`
}

// NPMPackage is an npm package that components use, with a version fixed
// per gobento release. Role is scaffolder, dependency, or dev-dependency.
type NPMPackage struct {
	Name       string   `json:"name"`
	Components []string `json:"components"`
	Role       string   `json:"role"`
	Version    string   `json:"version"`
}

// Tooling lists the versions of the project tooling that are not Go modules,
// fixed per gobento release: each list is sorted by key or name.
type Tooling struct {
	Images  []Image  `json:"images"`
	Actions []Action `json:"actions"`
	Tools   []Tool   `json:"tools"`
}

// Image is a container image of the generated Dockerfile or CI workflow.
type Image struct {
	// Key is the name templates use, as in {{.Images.runtime}}.
	Key   string `json:"key"`
	Image string `json:"image"`
	// Tag may contain "{go}", the selected Go version.
	Tag string `json:"tag"`
	// Digest is "" for an image that is not pinned by digest.
	Digest string `json:"digest"`
}

// Action is a GitHub Action of the generated workflow, pinned by commit.
type Action struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	SHA     string `json:"sha"`
}

// Tool is a tool the generated Makefile and workflow run.
type Tool struct {
	Name    string `json:"name"`
	Version string `json:"version"`
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
	catalog.KindFrontend:     {Title: "Frontend", Description: "Web pages served next to the API, which then moves under /api."},
	catalog.KindCSS:          {Title: "CSS framework", Description: "Styles the frontend's pages; offered with a frontend that supports it."},
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
		Assets:     []Asset{},
		Binaries:   []Binary{},
		NPM:        []NPMPackage{},
		Tooling:    Tooling{Images: []Image{}, Actions: []Action{}, Tools: []Tool{}},
		Supported:  []map[string]string{},
	}

	for _, k := range catalog.Kinds() {
		info, ok := kindInfo[k]
		if !ok {
			return nil, fmt.Errorf("kind %s has no title and description", k)
		}
		info.ID = string(k)
		info.Optional = k.Optional()
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

	for _, a := range rules.Assets {
		out.Assets = append(out.Assets, Asset{Name: a.Name, Components: a.Components, Version: a.Version, Source: a.Source})
	}

	for _, b := range rules.Binaries {
		eb := Binary{Name: b.Name, Components: b.Components, Version: b.Version, Platforms: []string{}}
		for _, p := range b.Platforms {
			eb.Platforms = append(eb.Platforms, p.Key())
		}
		out.Binaries = append(out.Binaries, eb)
	}

	if !rules.Node.IsZero() {
		out.Node = NodeRule{Range: rules.Node.Range.String(), Default: rules.Node.Default}
	}
	for _, p := range rules.NPM {
		out.NPM = append(out.NPM, NPMPackage{Name: p.Name, Components: p.Components, Role: string(p.Role), Version: p.Version})
	}

	for _, key := range slices.Sorted(maps.Keys(rules.Tooling.Images)) {
		img := rules.Tooling.Images[key]
		out.Tooling.Images = append(out.Tooling.Images, Image{Key: key, Image: img.Name, Tag: img.Tag, Digest: img.Digest})
	}
	for _, name := range slices.Sorted(maps.Keys(rules.Tooling.Actions)) {
		a := rules.Tooling.Actions[name]
		out.Tooling.Actions = append(out.Tooling.Actions, Action{Name: name, Version: a.Version, SHA: a.SHA})
	}
	for _, name := range slices.Sorted(maps.Keys(rules.Tooling.Tools)) {
		out.Tooling.Tools = append(out.Tooling.Tools, Tool{Name: name, Version: rules.Tooling.Tools[name]})
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
