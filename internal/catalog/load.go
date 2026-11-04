package catalog

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"
	"text/template"

	"go.yaml.in/yaml/v3"

	"github.com/Tales-of-zuna/gobento/internal/compat"
)

// File names with a fixed meaning in the catalog file system.
const (
	// RulesFile is the version-compatibility rules file at the catalog root.
	RulesFile = "versions.yaml"
	// ManifestFile is the manifest in every component directory.
	ManifestFile = "component.yaml"
	// NotesFile is the optional notes template in a component directory.
	NotesFile = "notes.md.tmpl"
	// TemplateSuffix ends every template file name.
	TemplateSuffix = ".tmpl"
	// DotPrefix at the start of a path element maps to ".", so that
	// templates for dotfiles are not dotfiles themselves.
	DotPrefix = "dot-"

	// embedFile is the Go file that embeds the catalog; it is allowed at
	// the catalog root and otherwise ignored.
	embedFile = "embed.go"
	// filesDir holds a component's templates, one subdirectory per slot.
	filesDir = "files"
)

var (
	// namePattern matches component names and trait names.
	namePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	// slotPattern matches slot names, which are used as template field
	// names (.Slots.<name>) and therefore must be identifiers.
	slotPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	// packagePattern matches Go package names gobento accepts.
	packagePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
)

// Catalog is a loaded and validated component catalog together with its
// version-compatibility rules.
type Catalog struct {
	components []*Component // sorted by kind order, then name
	rules      *compat.Rules
}

// manifest mirrors the YAML layout of component.yaml.
type manifest struct {
	Name        string                  `yaml:"name"`
	Title       string                  `yaml:"title"`
	Description string                  `yaml:"description"`
	Slots       map[string]slotManifest `yaml:"slots"`
	Provides    map[string]string       `yaml:"provides"`
	Requires    map[string][]string     `yaml:"requires"`
}

type slotManifest struct {
	Dir      string `yaml:"dir"`
	Package  string `yaml:"package"`
	FilledBy Kind   `yaml:"filled_by"`
	Required bool   `yaml:"required"`
}

// Load reads a catalog from fsys, whose root holds versions.yaml and one
// directory per kind (architecture/, framework/, database/, data/,
// frontend/), each with one directory per component.
//
// Load validates every manifest, enforces that every file other than
// component.yaml, versions.yaml, and embed.go ends in .tmpl and that no
// file or directory name starts with ".", parses the rules file, and checks
// that every component referenced by the rules exists. It reports every
// problem found, joined with errors.Join.
func Load(fsys fs.FS) (*Catalog, error) {
	var errs []error
	cat := &Catalog{}

	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read catalog: %w", err)
	}
	for _, e := range entries {
		name := e.Name()
		switch {
		case !e.IsDir() && (name == RulesFile || name == embedFile):
		case e.IsDir() && Kind(name).valid():
		case e.IsDir():
			errs = append(errs, fmt.Errorf("%s: unknown component kind; want one of %s", name, kindList()))
		default:
			errs = append(errs, fmt.Errorf("%s: unexpected file at the catalog root", name))
		}
	}

	for _, kind := range Kinds() {
		comps, kerrs := loadKind(fsys, kind)
		cat.components = append(cat.components, comps...)
		errs = append(errs, kerrs...)
	}

	data, err := fs.ReadFile(fsys, RulesFile)
	if err != nil {
		errs = append(errs, fmt.Errorf("%s: %w", RulesFile, err))
	} else if rules, err := compat.ParseRules(data); err != nil {
		errs = append(errs, fmt.Errorf("%s: %w", RulesFile, err))
	} else {
		exists := func(id string) bool {
			kind, name, _ := strings.Cut(id, "/")
			_, ok := cat.Component(Kind(kind), name)
			return ok
		}
		if err := rules.CheckComponents(exists); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", RulesFile, err))
		}
		provided := func(trait string) bool {
			return slices.ContainsFunc(cat.components, func(c *Component) bool {
				_, ok := c.Provides[trait]
				return ok
			})
		}
		if err := rules.CheckTraits(provided); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", RulesFile, err))
		}
		cat.rules = rules
	}

	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("invalid catalog: %w", err)
	}
	return cat, nil
}

// loadKind loads every component directory of one kind. A missing kind
// directory means the catalog has no components of that kind.
func loadKind(fsys fs.FS, kind Kind) ([]*Component, []error) {
	entries, err := fs.ReadDir(fsys, string(kind))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, []error{fmt.Errorf("%s: %w", kind, err)}
	}
	var (
		comps []*Component
		errs  []error
	)
	for _, e := range entries {
		dir := path.Join(string(kind), e.Name())
		if !e.IsDir() {
			errs = append(errs, fmt.Errorf("%s: unexpected file; components are directories", dir))
			continue
		}
		c, cerrs := loadComponent(fsys, kind, dir)
		if len(cerrs) > 0 {
			errs = append(errs, cerrs...)
			continue
		}
		comps = append(comps, c)
	}
	return comps, errs
}

// loadComponent loads and validates the component in dir.
func loadComponent(fsys fs.FS, kind Kind, dir string) (*Component, []error) {
	var errs []error
	fail := func(format string, args ...any) {
		errs = append(errs, fmt.Errorf("%s: "+format, append([]any{dir}, args...)...))
	}

	c := &Component{Kind: kind, Name: path.Base(dir)}

	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, []error{fmt.Errorf("%s: %w", dir, err)}
	}
	for _, e := range entries {
		switch name := e.Name(); {
		case name == ManifestFile && !e.IsDir():
		case name == NotesFile && !e.IsDir():
			notes, err := readTemplate(fsys, path.Join(dir, name))
			if err != nil {
				fail("%w", err)
			}
			c.Notes = notes
		case name == filesDir && e.IsDir():
			files, ferrs := loadFiles(fsys, path.Join(dir, filesDir))
			c.Files = files
			errs = append(errs, ferrs...)
		default:
			fail("unexpected entry %q; a component holds only %s, %s, and %s/", name, ManifestFile, NotesFile, filesDir)
		}
	}

	m, err := readManifest(fsys, path.Join(dir, ManifestFile))
	if err != nil {
		fail("%w", err)
		return nil, errs
	}
	c.Title, c.Description, c.Provides, c.Requires = m.Title, m.Description, m.Provides, m.Requires

	if m.Name != c.Name {
		fail("manifest name %q must equal the directory name %q", m.Name, c.Name)
	}
	if !namePattern.MatchString(c.Name) {
		fail("name %q must match %s", c.Name, namePattern)
	}
	if kind.Optional() && c.Name == None {
		fail("name %q is reserved: it selects no %s component", None, kind)
	}
	if strings.TrimSpace(m.Title) == "" {
		fail("title is required")
	}
	for _, k := range sortedKeys(m.Provides) {
		if !namePattern.MatchString(k) {
			fail("provides: trait name %q must match %s", k, namePattern)
		}
		if m.Provides[k] == "" {
			fail("provides: trait %s has an empty value", k)
		}
	}
	for _, k := range sortedKeys(m.Requires) {
		if !namePattern.MatchString(k) {
			fail("requires: trait name %q must match %s", k, namePattern)
		}
		if len(m.Requires[k]) == 0 {
			fail("requires: trait %s lists no accepted values", k)
		}
	}

	if kind == KindArchitecture {
		if len(m.Slots) == 0 {
			fail("an architecture must declare at least one slot")
		}
		for _, name := range sortedKeys(m.Slots) {
			s, serrs := checkSlot(name, m.Slots[name])
			for _, err := range serrs {
				fail("slot %s: %w", name, err)
			}
			c.Slots = append(c.Slots, s)
		}
		for _, filled := range c.FilledSlots() {
			if _, ok := c.Slot(filled); !ok {
				fail("%s/%s: the architecture declares no slot %q", filesDir, filled, filled)
			}
		}
	} else if len(m.Slots) > 0 {
		fail("only architectures declare slots")
	}

	if len(errs) > 0 {
		return nil, errs
	}
	return c, nil
}

func readManifest(fsys fs.FS, name string) (manifest, error) {
	var m manifest
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return m, err
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&m); err != nil {
		return m, fmt.Errorf("%s: %w", path.Base(name), err)
	}
	return m, nil
}

// checkSlot validates one slot declaration.
func checkSlot(name string, sm slotManifest) (Slot, []error) {
	var errs []error
	s := Slot{Name: name, Dir: sm.Dir, Package: sm.Package, FilledBy: sm.FilledBy, Required: sm.Required}
	if name == RootSlot {
		errs = append(errs, fmt.Errorf("%q is reserved for files at the project root", RootSlot))
	} else if !slotPattern.MatchString(name) {
		errs = append(errs, fmt.Errorf("name must match %s", slotPattern))
	}
	if strings.TrimSpace(sm.Dir) == "" {
		errs = append(errs, errors.New("dir is required"))
	} else if _, err := template.New(name).Parse(sm.Dir); err != nil {
		errs = append(errs, fmt.Errorf("dir: %w", err))
	}
	if sm.Package != "" && !packagePattern.MatchString(sm.Package) {
		errs = append(errs, fmt.Errorf("package %q must match %s", sm.Package, packagePattern))
	}
	if !sm.FilledBy.valid() {
		errs = append(errs, fmt.Errorf("filled_by %q must be one of %s", sm.FilledBy, kindList()))
	}
	return s, errs
}

// loadFiles loads every template under dir (a component's files/
// directory). Its subdirectories are slot names or RootSlot.
func loadFiles(fsys fs.FS, dir string) ([]File, []error) {
	var (
		files []File
		errs  []error
	)
	err := fs.WalkDir(fsys, dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			errs = append(errs, err)
			return nil
		}
		if p == dir {
			return nil
		}
		rel := strings.TrimPrefix(p, dir+"/")
		if strings.HasPrefix(d.Name(), ".") {
			errs = append(errs, fmt.Errorf("%s: names must not start with %q; use the %q prefix instead", p, ".", DotPrefix))
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		slot, sub, nested := strings.Cut(rel, "/")
		if d.IsDir() {
			if !nested && slot != RootSlot && !slotPattern.MatchString(slot) {
				errs = append(errs, fmt.Errorf("%s: slot directory name must match %s or be %q", p, slotPattern, RootSlot))
				return fs.SkipDir
			}
			return nil
		}
		if !nested {
			errs = append(errs, fmt.Errorf("%s: templates belong in a slot directory, such as %s/%s/", p, filesDir, RootSlot))
			return nil
		}
		out, err := OutputPath(sub)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", p, err))
			return nil
		}
		text, err := readTemplate(fsys, p)
		if err != nil {
			errs = append(errs, err)
			return nil
		}
		files = append(files, File{Slot: slot, Path: out, Source: p, Template: text})
		return nil
	})
	if err != nil {
		errs = append(errs, err)
	}
	slices.SortFunc(files, func(a, b File) int {
		return strings.Compare(a.Slot+"\x00"+a.Path, b.Slot+"\x00"+b.Path)
	})
	return files, errs
}

// OutputPath maps a template path, relative to its slot directory, to the
// path of the generated file: the .tmpl suffix is removed from the file name
// and a leading "dot-" in any element becomes ".". For example,
// "dot-gitignore.tmpl" maps to ".gitignore".
func OutputPath(templatePath string) (string, error) {
	if !strings.HasSuffix(templatePath, TemplateSuffix) {
		return "", fmt.Errorf("template file names must end in %s", TemplateSuffix)
	}
	elems := strings.Split(strings.TrimSuffix(templatePath, TemplateSuffix), "/")
	for i, e := range elems {
		if after, ok := strings.CutPrefix(e, DotPrefix); ok {
			e = "." + after
		}
		if e == "" || e == "." || e == ".." {
			return "", fmt.Errorf("path element %q maps to an empty or relative name", elems[i])
		}
		elems[i] = e
	}
	return strings.Join(elems, "/"), nil
}

// readTemplate reads a template file and normalizes CRLF line endings to LF.
func readTemplate(fsys fs.FS, name string) (string, error) {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(string(data), "\r\n", "\n"), nil
}

func kindList() string {
	names := make([]string, 0, len(Kinds()))
	for _, k := range Kinds() {
		names = append(names, string(k))
	}
	return strings.Join(names, ", ")
}

// Rules returns the catalog's version-compatibility rules.
func (c *Catalog) Rules() *compat.Rules {
	return c.rules
}

// Components returns the components of the given kind, sorted by name.
func (c *Catalog) Components(kind Kind) []*Component {
	var out []*Component
	for _, comp := range c.components {
		if comp.Kind == kind {
			out = append(out, comp)
		}
	}
	return out
}

// Component returns the component with the given kind and name.
func (c *Catalog) Component(kind Kind, name string) (*Component, bool) {
	i := slices.IndexFunc(c.components, func(comp *Component) bool {
		return comp.Kind == kind && comp.Name == name
	})
	if i < 0 {
		return nil, false
	}
	return c.components[i], true
}

// Candidates returns every selection of one component per required kind and
// at most one per optional kind whose trait requirements are met, in kind
// order and then by name, with "no component" before the components of an
// optional kind. A candidate is supported only if it also composes;
// compose.Supported checks that.
func (c *Catalog) Candidates() []Selection {
	sels := []Selection{{}}
	for _, kind := range Kinds() {
		choices := c.Components(kind)
		if kind.Optional() {
			choices = append([]*Component{nil}, choices...)
		}
		var next []Selection
		for _, s := range sels {
			for _, comp := range choices {
				if comp == nil {
					next = append(next, s)
					continue
				}
				next = append(next, s.With(comp))
			}
		}
		sels = next
	}
	var out []Selection
	for _, s := range sels {
		if s.CheckTraits() == nil {
			out = append(out, s)
		}
	}
	return out
}
