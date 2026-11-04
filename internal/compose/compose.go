package compose

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"go/token"
	"io/fs"
	"maps"
	"path"
	"slices"
	"strings"
	"text/template"

	"github.com/Tales-of-zuna/gobento/internal/catalog"
	"github.com/Tales-of-zuna/gobento/internal/compat"
)

// Errors returned by Compose, wrapped with details. Test for them with
// errors.Is.
var (
	// ErrSelection reports a selection without exactly one component of
	// each kind.
	ErrSelection = errors.New("incomplete selection")
	// ErrUnknownSlot reports templates for a slot the architecture does not
	// declare.
	ErrUnknownSlot = errors.New("unknown slot")
	// ErrUnfilledSlot reports a required slot that no component fills.
	ErrUnfilledSlot = errors.New("required slot unfilled")
	// ErrSlotOwner reports a slot filled by more than one component, or by
	// a component of a kind other than the one the architecture names.
	ErrSlotOwner = errors.New("slot ownership conflict")
	// ErrPathConflict reports two templates that produce the same file, or
	// a root file that is reserved or lies inside a slot directory.
	ErrPathConflict = errors.New("path conflict")
	// ErrTraits reports unmet or conflicting traits.
	ErrTraits = errors.New("incompatible traits")
	// ErrTemplate reports a template that fails to parse, render, or
	// format, or a slot directory or package name that is invalid.
	ErrTemplate = errors.New("template error")
)

// readmePath is the project README that component notes are appended to.
const readmePath = "README.md"

// isReserved reports whether p is a root file that gobento creates itself
// during generation.
func isReserved(p string) bool {
	return p == "go.mod" || p == "go.sum"
}

// fileMode is the permission of every generated file.
const fileMode fs.FileMode = 0o644

// Input is everything Compose needs to build a Plan.
type Input struct {
	// Project describes the generated project.
	Project Project
	// Selection is the chosen component of each kind.
	Selection catalog.Selection
	// Versions maps module paths to the chosen versions.
	Versions map[string]string
	// Base is the base layer composed into every selection, or nil for
	// none (catalog.Catalog.Base).
	Base *catalog.Component
	// Tooling holds the tooling versions: images, actions, and tools.
	Tooling compat.ToolingVersions
}

// Plan is the composed project: every file to write, in path order. It
// does not contain go.mod and go.sum, which are produced during generation.
type Plan struct {
	// Files lists the planned files, sorted by path.
	Files []File
	// Slots locates every slot of the architecture, as given to templates.
	Slots map[string]SlotData
	// Contributions lists the rendered contributions of the selected
	// components to the base layer's points, in kind order and then by
	// point. It is empty without a base layer.
	Contributions []Contribution
}

// Contribution is one component's rendered contribution to a point.
type Contribution struct {
	// Component is the ID of the contributing component.
	Component string
	// Point is the name of the contribution point.
	Point string
	// Content is the rendered contribution, without leading and trailing
	// newlines.
	Content string
}

// File is one planned file.
type File struct {
	// Path is the file path relative to the project root, with forward
	// slashes.
	Path string
	// Mode is the file's permission bits.
	Mode fs.FileMode
	// Content is the rendered, and for Go files formatted, content.
	Content []byte
	// Component is the ID of the component whose template produced the
	// file, for example "framework/nethttp".
	Component string
}

// Paths returns the paths of all planned files, in order.
func (p *Plan) Paths() []string {
	out := make([]string, len(p.Files))
	for i, f := range p.Files {
		out[i] = f.Path
	}
	return out
}

// File returns the planned file with the given path.
func (p *Plan) File(path string) (File, bool) {
	i := slices.IndexFunc(p.Files, func(f File) bool { return f.Path == path })
	if i < 0 {
		return File{}, false
	}
	return p.Files[i], true
}

// Compose checks the selection and renders every template of every selected
// component, and of the base layer with the components' contributions, into
// a Plan. It reports every problem it finds, joined with
// errors.Join: an incomplete selection, unmet traits, unknown or unfilled
// slots, slots with more than one owner, path conflicts, and templates that
// fail to render or, for .go files, to format. Nothing is written to disk.
func Compose(in Input) (*Plan, error) {
	c := &composer{in: in}
	c.run()
	if err := errors.Join(c.errs...); err != nil {
		return nil, err
	}
	slices.SortFunc(c.files, func(a, b File) int { return strings.Compare(a.Path, b.Path) })
	return &Plan{Files: c.files, Slots: c.data.Slots, Contributions: c.contribs}, nil
}

// composer holds the state of one Compose call.
type composer struct {
	in    Input
	data  Data
	errs  []error
	files []File
	// owner maps each output path to the component that produced it.
	owner map[string]string
	// contribs holds the rendered contributions, in kind order.
	contribs []Contribution
}

// components returns the selected components in kind order, followed by
// the base layer if there is one.
func (c *composer) components() []*catalog.Component {
	comps := c.in.Selection.Components()
	if c.in.Base != nil {
		comps = append(comps, c.in.Base)
	}
	return comps
}

func (c *composer) fail(format string, args ...any) {
	c.errs = append(c.errs, fmt.Errorf(format, args...))
}

func (c *composer) run() {
	sel := c.in.Selection
	c.checkSelection()
	if sel.Architecture == nil {
		return
	}
	if err := sel.CheckTraits(); err != nil {
		c.fail("%w: %w", ErrTraits, err)
	}

	c.data = Data{
		Project:  c.in.Project,
		Slots:    make(map[string]SlotData),
		Traits:   sel.Traits(),
		Versions: orEmpty(c.in.Versions),
		Images:   orEmpty(c.in.Tooling.Images),
		Actions:  orEmpty(c.in.Tooling.Actions),
		Tools:    orEmpty(c.in.Tooling.Tools),
	}
	c.resolveSlots()
	c.checkOwnership()
	c.checkRootFiles()
	c.renderContributions()

	c.owner = make(map[string]string)
	for _, comp := range c.components() {
		for _, f := range comp.Files {
			c.render(comp, f)
		}
	}
	c.appendNotes()
}

// orEmpty returns m, or an empty map if m is nil, so that templates report
// a missing key instead of rendering "<no value>".
func orEmpty(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	return m
}

// renderContributions renders every selected component's contributions to
// the base layer's points, in kind order and then by point. Without a base
// layer nothing would include them, so none are rendered.
func (c *composer) renderContributions() {
	base := c.in.Base
	if base == nil {
		return
	}
	for _, comp := range c.in.Selection.Components() {
		for _, point := range slices.Sorted(maps.Keys(comp.Contributions)) {
			source := comp.ID() + "/" + catalog.ContribDir + "/" + point + catalog.TemplateSuffix
			if _, ok := base.Point(point); !ok {
				c.fail("%w: %s: the base layer declares no contribution point %q", ErrTemplate, source, point)
				continue
			}
			content, err := c.execute(source, comp.Contributions[point], nil)
			if err != nil {
				c.fail("%w: %s: %w", ErrTemplate, source, err)
				continue
			}
			text := strings.Trim(string(content), "\n")
			if strings.TrimSpace(text) == "" {
				continue
			}
			c.contribs = append(c.contribs, Contribution{Component: comp.ID(), Point: point, Content: text})
		}
	}
}

// contrib returns the contributions to point, joined with newlines, or ""
// if no selected component contributes. It is the template function
// "contrib" of the base layer's templates.
func (c *composer) contrib(point string) (string, error) {
	if _, ok := c.in.Base.Point(point); !ok {
		return "", fmt.Errorf("the base layer declares no contribution point %q in %s", point, catalog.BaseManifestFile)
	}
	var parts []string
	for _, k := range c.contribs {
		if k.Point == point {
			parts = append(parts, k.Content)
		}
	}
	return strings.Join(parts, "\n"), nil
}

// funcsFor returns the extra template functions of comp's templates: the
// base layer's templates may include contributions.
func (c *composer) funcsFor(comp *catalog.Component) template.FuncMap {
	if comp != c.in.Base {
		return nil
	}
	return template.FuncMap{"contrib": c.contrib}
}

// checkSelection reports kinds without a component and components in the
// wrong field.
func (c *composer) checkSelection() {
	sel := c.in.Selection
	fields := []struct {
		kind catalog.Kind
		comp *catalog.Component
	}{
		{catalog.KindArchitecture, sel.Architecture},
		{catalog.KindFramework, sel.Framework},
		{catalog.KindDatabase, sel.Database},
		{catalog.KindData, sel.Data},
	}
	for _, f := range fields {
		switch {
		case f.comp == nil:
			c.fail("%w: no %s component selected", ErrSelection, f.kind)
		case f.comp.Kind != f.kind:
			c.fail("%w: %s is not a %s component", ErrSelection, f.comp.ID(), f.kind)
		}
	}
}

// resolveSlots renders every slot directory and fills in c.data.Slots.
func (c *composer) resolveSlots() {
	arch := c.in.Selection.Architecture
	dirs := make(map[string]string) // rendered dir -> slot name
	for _, s := range arch.Slots {
		var buf bytes.Buffer
		tmpl, err := newTemplate(arch.ID()+": slot "+s.Name+" dir", s.Dir, nil)
		if err == nil {
			err = tmpl.Execute(&buf, c.data)
		}
		if err != nil {
			c.fail("%w: %s: slot %s: dir: %w", ErrTemplate, arch.ID(), s.Name, err)
			continue
		}
		dir := buf.String()
		if err := checkRelPath(dir); err != nil {
			c.fail("%w: %s: slot %s: dir %q: %w", ErrTemplate, arch.ID(), s.Name, dir, err)
			continue
		}
		if other, ok := dirs[dir]; ok {
			c.fail("%w: %s: slots %s and %s share the directory %s", ErrSlotOwner, arch.ID(), other, s.Name, dir)
			continue
		}
		dirs[dir] = s.Name

		pkg := s.Package
		if pkg == "" {
			pkg = path.Base(dir)
		}
		if !token.IsIdentifier(pkg) || token.IsKeyword(pkg) {
			c.fail("%w: %s: slot %s: %q is not a valid Go package name; set the slot's package key", ErrTemplate, arch.ID(), s.Name, pkg)
			continue
		}
		c.data.Slots[s.Name] = SlotData{Dir: dir, Package: pkg, Import: c.in.Project.Module + "/" + dir}
	}
}

// checkOwnership reports templates for undeclared slots, slots with more
// than one owner or an owner of the wrong kind, and unfilled required slots.
func (c *composer) checkOwnership() {
	sel := c.in.Selection
	arch := sel.Architecture
	fillers := make(map[string][]*catalog.Component)
	for _, comp := range c.components() {
		for _, name := range comp.FilledSlots() {
			if _, ok := arch.Slot(name); !ok {
				c.fail("%w: %s has templates for slot %s, which %s does not declare", ErrUnknownSlot, comp.ID(), name, arch.ID())
				continue
			}
			fillers[name] = append(fillers[name], comp)
		}
	}
	for _, s := range arch.Slots {
		owners := fillers[s.Name]
		switch {
		case len(owners) > 1:
			ids := make([]string, len(owners))
			for i, o := range owners {
				ids[i] = o.ID()
			}
			c.fail("%w: slot %s is filled by both %s; a slot has exactly one owner", ErrSlotOwner, s.Name, strings.Join(ids, " and "))
		case len(owners) == 1 && owners[0].Kind != s.FilledBy:
			c.fail("%w: slot %s is filled by the %s component, but %s has templates for it", ErrSlotOwner, s.Name, s.FilledBy, owners[0].ID())
		case len(owners) == 0 && s.Required:
			c.fail("%w: slot %s (%s) must be filled by the %s component, but no selected component fills it",
				ErrUnfilledSlot, s.Name, slotDir(c.data.Slots, s), s.FilledBy)
		}
	}
}

// checkRootFiles reports root files that more than one component produces,
// that gobento reserves, or that lie inside a slot directory.
func (c *composer) checkRootFiles() {
	owners := make(map[string][]string)
	var order []string
	for _, comp := range c.components() {
		for _, f := range comp.Files {
			if f.Slot != catalog.RootSlot {
				continue
			}
			if len(owners[f.Path]) == 0 {
				order = append(order, f.Path)
			}
			owners[f.Path] = append(owners[f.Path], comp.ID())
		}
	}
	for _, p := range order {
		ids := owners[p]
		if len(ids) > 1 {
			c.fail("%w: root file %s is produced by both %s", ErrPathConflict, p, strings.Join(ids, " and "))
		}
		if isReserved(p) {
			c.fail("%w: %s may not produce %s; gobento writes it during generation", ErrPathConflict, ids[0], p)
		}
		for name, s := range c.data.Slots {
			if strings.HasPrefix(p, s.Dir+"/") {
				c.fail("%w: root file %s from %s lies inside slot %s (%s)", ErrPathConflict, p, ids[0], name, s.Dir)
			}
		}
	}
}

// render renders one template into c.files. Files of slots that could not
// be resolved are skipped; their problem has already been reported.
func (c *composer) render(comp *catalog.Component, f catalog.File) {
	out := f.Path
	if f.Slot != catalog.RootSlot {
		s, ok := c.data.Slots[f.Slot]
		if !ok {
			return
		}
		out = s.Dir + "/" + f.Path
		if inner, ok := c.nestedSlot(f.Slot, out); ok {
			c.fail("%w: %s: %s would be written to %s, inside slot %s (%s); only that slot's owner writes there",
				ErrSlotOwner, comp.ID(), f.Source, out, inner, c.data.Slots[inner].Dir)
			return
		}
	}

	content, err := c.execute(f.Source, f.Template, c.funcsFor(comp))
	if err != nil {
		c.fail("%w: %s: %s: %w", ErrTemplate, comp.ID(), f.Source, err)
		return
	}
	if strings.HasSuffix(out, ".go") {
		formatted, err := format.Source(content)
		if err != nil {
			c.fail("%w: %s: %s: rendered Go code does not format: %w", ErrTemplate, comp.ID(), f.Source, err)
			return
		}
		content = formatted
	}

	if prev, ok := c.owner[out]; ok {
		// Root conflicts were reported by checkRootFiles.
		if f.Slot != catalog.RootSlot || prev == comp.ID() {
			c.fail("%w: %s and %s both produce %s", ErrPathConflict, prev, comp.ID(), out)
		}
		return
	}
	c.owner[out] = comp.ID()
	c.files = append(c.files, File{Path: out, Mode: fileMode, Content: content, Component: comp.ID()})
}

// nestedSlot reports the slot whose directory lies inside the directory of
// slot owner and contains out, the output path of one of owner's files. A
// slot directory may be nested inside another one, but the outer slot's
// files must stay out of it, so that every package has exactly one owner.
func (c *composer) nestedSlot(owner, out string) (string, bool) {
	outer := c.data.Slots[owner].Dir
	for _, name := range slices.Sorted(maps.Keys(c.data.Slots)) {
		dir := c.data.Slots[name].Dir
		if name != owner && strings.HasPrefix(dir, outer+"/") && strings.HasPrefix(out, dir+"/") {
			return name, true
		}
	}
	return "", false
}

// appendNotes renders every component's notes, in kind order and then the
// base layer's, and appends them to the README.
func (c *composer) appendNotes() {
	var notes []string
	var from []string
	for _, comp := range c.components() {
		if comp.Notes == "" {
			continue
		}
		content, err := c.execute(comp.ID()+"/"+catalog.NotesFile, comp.Notes, c.funcsFor(comp))
		if err != nil {
			c.fail("%w: %s: %s: %w", ErrTemplate, comp.ID(), catalog.NotesFile, err)
			continue
		}
		notes = append(notes, strings.TrimSpace(string(content)))
		from = append(from, comp.ID())
	}
	if len(notes) == 0 {
		return
	}
	i := slices.IndexFunc(c.files, func(f File) bool { return f.Path == readmePath })
	if i < 0 {
		c.fail("%w: %s have notes, but no component produces %s to append them to", ErrTemplate, strings.Join(from, ", "), readmePath)
		return
	}
	readme := strings.TrimRight(string(c.files[i].Content), "\n")
	c.files[i].Content = []byte(readme + "\n\n" + strings.Join(notes, "\n\n") + "\n")
}

// execute parses and renders one template with the composition data and the
// extra functions funcs.
func (c *composer) execute(name, text string, funcs template.FuncMap) ([]byte, error) {
	tmpl, err := newTemplate(name, text, funcs)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, c.data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// newTemplate parses a template that fails on missing map keys, so a typo
// such as .Slots.handler is an error instead of "<no value>". Every
// template may use the functions indent and prefix, and the functions in
// funcs.
func newTemplate(name, text string, funcs template.FuncMap) (*template.Template, error) {
	return template.New(name).Option("missingkey=error").
		Funcs(template.FuncMap{"indent": indent, "prefix": prefix}).
		Funcs(funcs).
		Parse(text)
}

// indent returns s with n spaces before every line that is not empty, for
// placing a contribution written at column zero inside YAML.
func indent(n int, s string) string {
	return prefix(strings.Repeat(" ", n), s)
}

// prefix returns s with p before every line that is not empty, for example
// a tab before every line of a Make recipe.
func prefix(p, s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		if l != "" {
			lines[i] = p + l
		}
	}
	return strings.Join(lines, "\n")
}

// checkRelPath reports whether p is a clean, relative, slash-separated
// directory path inside the project.
func checkRelPath(p string) error {
	switch {
	case p == "" || p == ".":
		return errors.New("must not be empty")
	case strings.Contains(p, `\`):
		return errors.New("must use forward slashes")
	case path.IsAbs(p):
		return errors.New("must be relative to the project root")
	case path.Clean(p) != p:
		return fmt.Errorf("must be clean (%s)", path.Clean(p))
	case p == ".." || strings.HasPrefix(p, "../"):
		return errors.New("must stay inside the project")
	}
	return nil
}

// slotDir returns the rendered directory of s, or its template if it could
// not be rendered.
func slotDir(slots map[string]SlotData, s catalog.Slot) string {
	if d, ok := slots[s.Name]; ok {
		return d.Dir
	}
	return s.Dir
}
