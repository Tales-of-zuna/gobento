package compose

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"go/token"
	"io/fs"
	"path"
	"slices"
	"strings"
	"text/template"

	"github.com/Tales-of-zuna/gobento/internal/catalog"
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
}

// Plan is the composed project: every file to write, in path order. It
// does not contain go.mod and go.sum, which are produced during generation.
type Plan struct {
	// Files lists the planned files, sorted by path.
	Files []File
	// Slots locates every slot of the architecture, as given to templates.
	Slots map[string]SlotData
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
// component into a Plan. It reports every problem it finds, joined with
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
	return &Plan{Files: c.files, Slots: c.data.Slots}, nil
}

// composer holds the state of one Compose call.
type composer struct {
	in    Input
	data  Data
	errs  []error
	files []File
	// owner maps each output path to the component that produced it.
	owner map[string]string
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
		Versions: c.in.Versions,
	}
	if c.data.Versions == nil {
		c.data.Versions = map[string]string{}
	}
	c.resolveSlots()
	c.checkOwnership()
	c.checkRootFiles()

	c.owner = make(map[string]string)
	for _, comp := range sel.Components() {
		for _, f := range comp.Files {
			c.render(comp, f)
		}
	}
	c.appendNotes()
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
		tmpl, err := newTemplate(arch.ID()+": slot "+s.Name+" dir", s.Dir)
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
	for _, comp := range sel.Components() {
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
	for _, comp := range c.in.Selection.Components() {
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
	}

	content, err := c.execute(f.Source, f.Template)
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

// appendNotes renders every component's notes, in kind order, and appends
// them to the README.
func (c *composer) appendNotes() {
	var notes []string
	var from []string
	for _, comp := range c.in.Selection.Components() {
		if comp.Notes == "" {
			continue
		}
		content, err := c.execute(comp.ID()+"/"+catalog.NotesFile, comp.Notes)
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

// execute parses and renders one template with the composition data.
func (c *composer) execute(name, text string) ([]byte, error) {
	tmpl, err := newTemplate(name, text)
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
// such as .Slots.handler is an error instead of "<no value>".
func newTemplate(name, text string) (*template.Template, error) {
	return template.New(name).Option("missingkey=error").Parse(text)
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
