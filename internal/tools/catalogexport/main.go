// Command catalogexport writes gobento's component catalog, version rules,
// supported selections, and input rules as JSON for the website, which reads
// them from website/src/generated/catalog.json.
//
// The output follows the schema in website/lib/catalog/schema.ts; the types
// in this package mirror it, and the two change together. The exporter uses
// the same packages as the CLI (internal/catalog, internal/compat,
// internal/compose, and the name rules of internal/generate), so the website
// shows exactly what `gobento new` accepts.
//
// Usage:
//
//	go run ./internal/tools/catalogexport                  # print to stdout
//	go run ./internal/tools/catalogexport -o catalog.json  # write a file
//
// `make catalog` regenerates the website's copy and `make catalog-check`
// fails if it is out of date.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Tales-of-zuna/gobento/components"
	"github.com/Tales-of-zuna/gobento/internal/catalog"
	"github.com/Tales-of-zuna/gobento/internal/compose"
	"github.com/Tales-of-zuna/gobento/internal/generate"
)

// Output is the exported document, the root of the website's catalog schema.
type Output struct {
	Kinds      []Kind              `json:"kinds"`
	Components []Component         `json:"components"`
	Go         GoRule              `json:"go"`
	Modules    []Module            `json:"modules"`
	Assets     []Asset             `json:"assets"`
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

// kindInfo holds the website's title and description of each kind. The
// catalog itself has no text for kinds; TestKindInfo checks that every kind
// has an entry.
var kindInfo = map[catalog.Kind]Kind{
	catalog.KindArchitecture: {Title: "Architecture", Description: "How the project's packages are laid out and wired together."},
	catalog.KindFramework:    {Title: "HTTP framework", Description: "Routes HTTP requests to handlers."},
	catalog.KindDatabase:     {Title: "Database", Description: "Where the project stores its data."},
	catalog.KindData:         {Title: "Data access", Description: "How the code reads and writes the database."},
	catalog.KindFrontend:     {Title: "Frontend", Description: "Web pages served next to the API, which then moves under /api."},
}

// Export loads the catalog in fsys and converts it to the website's schema.
func Export(fsys fs.FS) (*Output, error) {
	cat, err := catalog.Load(fsys)
	if err != nil {
		return nil, err
	}
	out := &Output{
		Components: []Component{},
		Modules:    []Module{},
		Assets:     []Asset{},
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
	for k, v := range c.Provides {
		ec.Provides[k] = v
	}
	for k, v := range c.Requires {
		ec.Requires[k] = v
	}
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

// Encode writes o as indented JSON with a trailing newline. HTML escaping is
// off so that ranges such as ">=1.26" stay readable.
func Encode(w io.Writer, o *Output) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(o)
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "catalogexport:", err)
		os.Exit(1)
	}
}

// run exports the embedded catalog to stdout, or to the file named by -o.
func run(args []string, stdout io.Writer) error {
	fl := flag.NewFlagSet("catalogexport", flag.ContinueOnError)
	output := fl.String("o", "", "write to this file instead of stdout")
	if err := fl.Parse(args); err != nil {
		return err
	}
	if fl.NArg() > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(fl.Args(), " "))
	}

	o, err := Export(components.FS())
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := Encode(&buf, o); err != nil {
		return err
	}
	if *output == "" {
		_, err := stdout.Write(buf.Bytes())
		return err
	}
	return writeFile(*output, buf.Bytes())
}

// writeFile replaces name with data through a temporary file in the same
// directory, so a failed run never leaves a truncated file behind.
func writeFile(name string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(name), ".catalogexport-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) //nolint:errcheck // already renamed on success; best effort otherwise
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil { //nolint:gosec // a source file meant to be committed and read by everyone
		return err
	}
	return os.Rename(tmp.Name(), name)
}
