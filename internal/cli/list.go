package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/Tales-of-zuna/gobento/internal/catalog"
	"github.com/Tales-of-zuna/gobento/internal/catalogjson"
	"github.com/Tales-of-zuna/gobento/internal/generate"
)

// listHint is the sentence the help of other commands uses to point at
// `gobento list`.
const listHint = "Run gobento list to see the supported components, selections, and versions."

// componentsJSON is the JSON of `gobento list components`: the kinds and
// components fields of the catalog document.
type componentsJSON struct {
	Kinds      []catalogjson.Kind      `json:"kinds"`
	Components []catalogjson.Component `json:"components"`
}

// selectionsJSON is the JSON of `gobento list selections`: the supported
// field of the catalog document.
type selectionsJSON struct {
	Supported []map[string]string `json:"supported"`
}

// versionsJSON is the JSON of `gobento list versions`: the go, modules,
// assets, and tooling fields of the catalog document.
type versionsJSON struct {
	Go      catalogjson.GoRule   `json:"go"`
	Modules []catalogjson.Module `json:"modules"`
	Assets  []catalogjson.Asset  `json:"assets"`
	Tooling catalogjson.Tooling  `json:"tooling"`
}

// newListCommand returns the `gobento list` command and its subcommands.
func newListCommand(a *app) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the supported components, selections, and versions",
		Long: `List what gobento supports, from the catalog and version rules built into
this binary: the components of each kind, the supported selections, and the
version rules. Without a subcommand, all three are listed.

With --json, the output is one JSON document instead of text. For gobento list
it is the catalog document the gobento website is built from; each subcommand
prints only the fields it lists ("kinds" and "components", "supported", or
"go", "modules", and "assets"), in the same format.

gobento list never asks a question, writes a file, or uses the network.`,
		Example: `  # Everything gobento supports
  gobento list

  # The HTTP frameworks, with their descriptions
  gobento list components --kind framework

  # The supported selections that use GORM
  gobento list selections --data gorm

  # The Go and module versions, as JSON for a script
  gobento list versions --json`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			d, err := a.document()
			if err != nil {
				return err
			}
			if asJSON {
				return catalogjson.Encode(a.stdout, d)
			}
			return writeAll(a.stdout, d)
		},
	}
	cmd.PersistentFlags().BoolVar(&asJSON, "json", false, "print JSON instead of text")
	cmd.AddCommand(
		newListComponentsCommand(a, &asJSON),
		newListSelectionsCommand(a, &asJSON),
		newListVersionsCommand(a, &asJSON),
	)
	return cmd
}

func newListComponentsCommand(a *app, asJSON *bool) *cobra.Command {
	var kind string
	cmd := &cobra.Command{
		Use:   "components",
		Short: "List the components of each kind",
		Long: `List each component kind with the gobento new flag that selects it, and each
component with its name, title, description, and the traits it provides and
requires. An optional kind, such as the frontend, also lists none, which
selects no component of the kind. --kind lists one kind only.`,
		Example: `  gobento list components
  gobento list components --kind database --json`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			d, err := a.document()
			if err != nil {
				return err
			}
			v, err := componentsOf(d, kind)
			if err != nil {
				return err
			}
			if *asJSON {
				return catalogjson.Encode(a.stdout, v)
			}
			return writeComponents(a.stdout, v)
		},
	}
	cmd.Flags().StringVar(&kind, "kind", "", "list only this kind: "+strings.Join(kindIDs(), ", "))
	return cmd
}

func newListSelectionsCommand(a *app, asJSON *bool) *cobra.Command {
	var f newFlags
	cmd := &cobra.Command{
		Use:   "selections",
		Short: "List the supported selections",
		Long: `List every supported selection: the components that gobento new accepts
together, one of each kind and at most one frontend. A selection without a
frontend shows none in its frontend column. The component flags of gobento
new narrow the list to the selections that contain the named component;
--frontend none keeps the selections without a frontend.`,
		Example: `  gobento list selections
  gobento list selections --arch mvc --db postgres
  gobento list selections --frontend htmx`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			d, err := a.document()
			if err != nil {
				return err
			}
			chosen := map[catalog.Kind]string{
				catalog.KindArchitecture: f.arch,
				catalog.KindFramework:    f.framework,
				catalog.KindDatabase:     f.db,
				catalog.KindData:         f.data,
				catalog.KindFrontend:     f.frontend,
			}
			v, err := selectionsOf(d, chosen)
			if err != nil {
				return err
			}
			if *asJSON {
				return catalogjson.Encode(a.stdout, v)
			}
			filtered := f.arch != "" || f.framework != "" || f.db != "" || f.data != "" || f.frontend != ""
			return writeSelections(a.stdout, d, v, filtered)
		},
	}
	fl := cmd.Flags()
	fl.StringVar(&f.arch, "arch", "", "only selections with this architecture")
	fl.StringVar(&f.framework, "framework", "", "only selections with this HTTP framework")
	fl.StringVar(&f.db, "db", "", "only selections with this database")
	fl.StringVar(&f.data, "data", "", "only selections with this data-access library")
	fl.StringVar(&f.frontend, "frontend", "", "only selections with this frontend, or none for those without one")
	return cmd
}

func newListVersionsCommand(a *app, asJSON *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "versions",
		Short: "List the supported Go and module versions",
		Long: `List the supported Go range and the default Go version, and each module of
the version rules with its role, range, default, the components that use it,
the trait values it is limited to, and its constraints on other versions,
followed by the files components vendor from an upstream release (such as
htmx), whose versions are fixed per gobento release. Choose versions inside
the module ranges with gobento new --go and --pin.`,
		Example: `  gobento list versions`,
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			d, err := a.document()
			if err != nil {
				return err
			}
			if *asJSON {
				return catalogjson.Encode(a.stdout, versionsJSON{Go: d.Go, Modules: d.Modules, Assets: d.Assets, Tooling: d.Tooling})
			}
			return writeVersions(a.stdout, d)
		},
	}
}

// document loads the catalog document of a's catalog, or returns a.doc when
// a test set one.
func (a *app) document() (*catalogjson.Document, error) {
	if a.doc != nil {
		return a.doc, nil
	}
	return catalogjson.Load(a.catalog)
}

// kindIDs returns the IDs of every kind in canonical order.
func kindIDs() []string {
	var out []string
	for _, k := range catalog.Kinds() {
		out = append(out, string(k))
	}
	return out
}

// componentsOf returns the kinds and components of d, restricted to kind
// unless it is empty.
func componentsOf(d *catalogjson.Document, kind string) (componentsJSON, error) {
	v := componentsJSON{Kinds: []catalogjson.Kind{}, Components: []catalogjson.Component{}}
	if kind != "" && !slices.ContainsFunc(d.Kinds, func(k catalogjson.Kind) bool { return k.ID == kind }) {
		var ids []string
		for _, k := range d.Kinds {
			ids = append(ids, k.ID)
		}
		return v, fmt.Errorf("--kind: unknown kind %q; supported: %s", kind, strings.Join(ids, ", "))
	}
	for _, k := range d.Kinds {
		if kind == "" || k.ID == kind {
			v.Kinds = append(v.Kinds, k)
		}
	}
	for _, c := range d.Components {
		if kind == "" || c.Kind == kind {
			v.Components = append(v.Components, c)
		}
	}
	return v, nil
}

// selectionsOf returns the supported selections of d that contain the
// component chosen for each kind; kinds without a name match any component,
// and catalog.None matches the selections without a component of an
// optional kind. A name that is not a component of its kind is an error.
func selectionsOf(d *catalogjson.Document, chosen map[catalog.Kind]string) (selectionsJSON, error) {
	v := selectionsJSON{Supported: []map[string]string{}}
	var errs []error
	for _, k := range catalog.Kinds() {
		name := chosen[k]
		if name == "" {
			continue
		}
		var names []string
		for _, c := range d.Components {
			if c.Kind == string(k) {
				names = append(names, c.Name)
			}
		}
		if k.Optional() && name == catalog.None {
			continue
		}
		if !slices.Contains(names, name) {
			slices.Sort(names)
			if k.Optional() {
				names = append([]string{catalog.None}, names...)
			}
			errs = append(errs, &generate.SelectionError{Kind: k, Value: name, Supported: names})
		}
	}
	if len(errs) > 0 {
		var lines []string
		for _, err := range errs {
			lines = append(lines, problemLines(err)...)
		}
		if len(lines) == 1 {
			return v, errors.New(lines[0])
		}
		return v, fmt.Errorf("cannot list selections:\n  - %s", strings.Join(lines, "\n  - "))
	}
	for _, s := range d.Supported {
		match := true
		for k, name := range chosen {
			if name != "" && selected(s, k) != name {
				match = false
			}
		}
		if match {
			v.Supported = append(v.Supported, s)
		}
	}
	return v, nil
}

// selected returns the name of the component of kind k in the supported
// selection s, or catalog.None when s has no component of k.
func selected(s map[string]string, k catalog.Kind) string {
	if name, ok := s[string(k)]; ok {
		return name
	}
	return catalog.None
}

// writeTable writes rows with their cells aligned in columns, two spaces
// apart. The table is laid out in memory, so only the final write can fail.
func writeTable(w io.Writer, rows [][]string) error {
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	for _, row := range rows {
		_, _ = io.WriteString(tw, strings.Join(row, "\t")+"\n") // writes to a bytes.Buffer never fail
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	_, err := w.Write(buf.Bytes())
	return err
}

// writeAll writes the components, selections, and versions of d under one
// heading each.
func writeAll(w io.Writer, d *catalogjson.Document) error {
	comps, err := componentsOf(d, "")
	if err != nil {
		return err
	}
	sections := []struct {
		title string
		write func() error
	}{
		{"Components", func() error { return writeComponents(w, comps) }},
		{"Selections", func() error { return writeSelections(w, d, selectionsJSON{Supported: d.Supported}, false) }},
		{"Versions", func() error { return writeVersions(w, d) }},
	}
	for i, s := range sections {
		sep := ""
		if i > 0 {
			sep = "\n"
		}
		rule := strings.Repeat("=", len(s.title))
		if _, err := fmt.Fprintf(w, "%s%s\n%s\n\n", sep, s.title, rule); err != nil {
			return err
		}
		if err := s.write(); err != nil {
			return err
		}
	}
	return nil
}

// writeComponents writes each kind of v with its flag and description, and
// its components with their titles, descriptions, and traits.
func writeComponents(w io.Writer, v componentsJSON) error {
	for i, k := range v.Kinds {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		flag := kindFlag(catalog.Kind(k.ID))
		if k.Optional {
			flag += ", optional"
		}
		if _, err := fmt.Fprintf(w, "%s (%s)\n%s\n\n", k.Title, flag, k.Description); err != nil {
			return err
		}
		var rows [][]string
		if k.Optional {
			rows = append(rows, []string{"  " + catalog.None, "No " + strings.ToLower(k.Title) + " (the default)"})
		}
		for _, c := range v.Components {
			if c.Kind != k.ID {
				continue
			}
			rows = append(rows, []string{"  " + c.Name, c.Title})
			if c.Description != "" {
				rows = append(rows, []string{"  ", c.Description})
			}
			if p := providedTraits(c.Provides); p != "" {
				rows = append(rows, []string{"  ", "provides: " + p})
			}
			if r := requiredTraits(c.Requires); r != "" {
				rows = append(rows, []string{"  ", "requires: " + r})
			}
		}
		if err := writeTable(w, rows); err != nil {
			return err
		}
	}
	return nil
}

// providedTraits formats traits as name=value, sorted by name.
func providedTraits(traits map[string]string) string {
	var out []string
	for _, k := range sortedKeys(traits) {
		out = append(out, k+"="+traits[k])
	}
	return strings.Join(out, ", ")
}

// requiredTraits formats traits as name=a|b, sorted by name, with the
// accepted values in their catalog order.
func requiredTraits(traits map[string][]string) string {
	var out []string
	for _, k := range sortedKeys(traits) {
		out = append(out, k+"="+strings.Join(traits[k], "|"))
	}
	return strings.Join(out, ", ")
}

// sortedKeys returns the keys of m in sorted order.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// writeSelections writes the selections of v as a table with one column per
// kind of d, and their number. filtered reports whether v was narrowed by
// component flags.
func writeSelections(w io.Writer, d *catalogjson.Document, v selectionsJSON, filtered bool) error {
	n := len(v.Supported)
	switch {
	case n == 0 && filtered:
		_, err := fmt.Fprintln(w, "No supported selection matches.")
		return err
	case n == 0:
		_, err := fmt.Fprintln(w, "No supported selections.")
		return err
	}
	var header []string
	for _, k := range d.Kinds {
		header = append(header, strings.ToUpper(k.ID))
	}
	rows := [][]string{header}
	for _, s := range v.Supported {
		var row []string
		for _, k := range d.Kinds {
			row = append(row, selected(s, catalog.Kind(k.ID)))
		}
		rows = append(rows, row)
	}
	if err := writeTable(w, rows); err != nil {
		return err
	}
	noun := "selections"
	if n == 1 {
		noun = "selection"
	}
	if filtered {
		_, err := fmt.Fprintf(w, "\n%d of %d supported selections.\n", n, len(d.Supported))
		return err
	}
	_, err := fmt.Fprintf(w, "\n%d supported %s.\n", n, noun)
	return err
}

// writeVersions writes the Go rule, a table of the modules, and their
// constraints.
func writeVersions(w io.Writer, d *catalogjson.Document) error {
	if _, err := fmt.Fprintf(w, "Go %s (default %s)\n\n", d.Go.Range, d.Go.Default); err != nil {
		return err
	}
	rows := [][]string{{"MODULE", "ROLE", "RANGE", "DEFAULT", "USED BY", "ONLY WHEN"}}
	for _, m := range d.Modules {
		when := requiredTraits(m.Traits)
		if when == "" {
			when = "-"
		}
		rows = append(rows, []string{m.Path, m.Role, m.Range, m.Default, strings.Join(m.Components, ", "), when})
	}
	if err := writeTable(w, rows); err != nil {
		return err
	}

	var lines []string
	for _, m := range d.Modules {
		for _, c := range m.Constraints {
			lines = append(lines, fmt.Sprintf("  %s %s needs %s %s", m.Path, c.When, c.Requires.Target, c.Requires.Range))
		}
	}
	if len(lines) > 0 {
		if _, err := fmt.Fprintf(w, "\nConstraints:\n%s\n", strings.Join(lines, "\n")); err != nil {
			return err
		}
	}
	if len(d.Assets) > 0 {
		if _, err := fmt.Fprint(w, "\nVendored assets (fixed per gobento release):\n\n"); err != nil {
			return err
		}
		rows = [][]string{{"ASSET", "VERSION", "USED BY", "SOURCE"}}
		for _, a := range d.Assets {
			rows = append(rows, []string{a.Name, a.Version, strings.Join(a.Components, ", "), a.Source})
		}
		if err := writeTable(w, rows); err != nil {
			return err
		}
	}
	return writeTooling(w, d.Tooling)
}

// writeTooling writes a table of the tooling versions: images, actions, and
// tools, in that order. An image's version is its reference, with {go} for
// the selected Go version; PINNED BY is its digest or an action's commit.
func writeTooling(w io.Writer, t catalogjson.Tooling) error {
	if len(t.Images)+len(t.Actions)+len(t.Tools) == 0 {
		return nil
	}
	if _, err := fmt.Fprint(w, "\nProject tooling (fixed per gobento release; {go} is the selected Go version):\n\n"); err != nil {
		return err
	}
	rows := [][]string{{"TYPE", "NAME", "VERSION", "PINNED BY"}}
	for _, img := range t.Images {
		pin := img.Digest
		if pin == "" {
			pin = "-"
		}
		rows = append(rows, []string{"image", img.Key, img.Image + ":" + img.Tag, pin})
	}
	for _, a := range t.Actions {
		rows = append(rows, []string{"action", a.Name, a.Version, a.SHA})
	}
	for _, tool := range t.Tools {
		rows = append(rows, []string{"tool", tool.Name, tool.Version, "-"})
	}
	return writeTable(w, rows)
}
