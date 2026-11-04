package cli

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/Tales-of-zuna/gobento/internal/catalog"
	"github.com/Tales-of-zuna/gobento/internal/generate"
)

// errAborted reports that the user cancelled a question, for example with
// Ctrl-C or Esc. Like an interrupt, it exits with ExitInterrupted.
var errAborted = errors.New("cancelled at a question")

// errInputEnded reports that the input ended before every question was
// answered.
var errInputEnded = errors.New("input ended before every question was answered; nothing was written")

// Question fields that are not a component kind.
const (
	fieldName   = "name"
	fieldModule = "module"
)

// question is one thing gobento asks. It is a choice when options is not
// empty, and a line of text otherwise.
type question struct {
	// field is fieldName, fieldModule, or the component kind.
	field string
	// title is the question, for example "HTTP framework".
	title string
	// help is a one-line hint shown below the title.
	help string
	// options lists the choices, in display order.
	options []option
	// value is the initial text of a text question; an empty answer is
	// replaced with it.
	value string
	// validate checks a text answer, after an empty answer was replaced
	// with value.
	validate func(string) error
}

// option is one choice of a question.
type option struct {
	label string
	value string
}

// asker asks one question and returns the answer: an option's value, or
// the text entered. It returns errAborted when the user cancels,
// errInputEnded when the input ends, and the context's error when ctx is
// done.
type asker interface {
	ask(ctx context.Context, q question) (string, error)
}

// newInput is the input of `gobento new` that questions can complete.
type newInput struct {
	name       string
	hasName    bool
	module     string
	hasModule  bool
	components map[catalog.Kind]string
}

// canAsk reports whether the given input leaves something to ask and is
// valid as far as it goes: a given name and module path follow the rules,
// the pins parse, and the component flags match a supported selection.
// Otherwise questions would be wasted, and the problem report comes first.
func canAsk(supported []catalog.Selection, in newInput, pinErrs []error) bool {
	if len(pinErrs) > 0 {
		return false
	}
	if in.hasName && generate.CheckName(in.name) != nil {
		return false
	}
	if in.hasModule && generate.CheckModule(in.module, in.name) != nil {
		return false
	}
	return len(matching(supported, in.components)) > 0
}

// interview asks for what in leaves open: the name when it is missing, then
// the module path, and each kind without a flag that has several options
// among the supported selections that match the flags and the earlier
// answers; for an optional kind, catalog.None is one of the options. It fills in the answers, including kinds left with one option,
// and reports whether it asked anything.
func interview(ctx context.Context, supported []catalog.Selection, in *newInput, a asker) (bool, error) {
	asked := false
	if !in.hasName {
		name, err := a.ask(ctx, nameQuestion())
		if err != nil {
			return asked, err
		}
		in.name, in.hasName, asked = name, true, true
		if !in.hasModule {
			mod, err := a.ask(ctx, moduleQuestion(in.name))
			if err != nil {
				return asked, err
			}
			if mod == "" {
				mod = in.name
			}
			if mod != in.name {
				in.module, in.hasModule = mod, true
			}
		}
	}

	if in.components == nil {
		in.components = make(map[catalog.Kind]string)
	}
	for _, kind := range catalog.Kinds() {
		if in.components[kind] != "" {
			continue
		}
		left := matching(supported, in.components)
		opts := componentOptions(left, kind)
		switch len(opts) {
		case 0:
			// No supported selection matches: Prepare reports it.
			return asked, nil
		case 1:
			// "none" alone is what an omitted optional flag means anyway.
			if opts[0].value != catalog.None {
				in.components[kind] = opts[0].value
			}
			continue
		}
		value, err := a.ask(ctx, kindQuestion(kind, opts))
		if err != nil {
			return asked, err
		}
		in.components[kind] = value
		asked = true
	}
	return asked, nil
}

// nameQuestion asks for the project name.
func nameQuestion() question {
	return question{
		field:    fieldName,
		title:    "Project name",
		help:     "Used for the directory, the default module path, and cmd/<name>; " + generate.NameRule + ".",
		validate: generate.CheckName,
	}
}

// moduleQuestion asks for the module path of project name.
func moduleQuestion(name string) question {
	return question{
		field: fieldModule,
		title: "Go module path",
		help:  "For example example.com/" + name + "; press Enter to use " + name + ". Skip this question with --module.",
		value: name,
		validate: func(path string) error {
			return generate.CheckModule(path, name)
		},
	}
}

// kindQuestion asks for a component of kind among opts.
func kindQuestion(kind catalog.Kind, opts []option) question {
	help := "Skip this question with " + kindFlag(kind) + " <name>."
	if kind.Optional() {
		help = "Skip this question with " + kindFlag(kind) + " <name>, or " + kindFlag(kind) + " " + catalog.None + "."
	}
	return question{
		field:   string(kind),
		title:   kindTitle(kind),
		help:    help,
		options: opts,
	}
}

// kindTitle returns the title of a question for kind.
func kindTitle(kind catalog.Kind) string {
	switch kind {
	case catalog.KindArchitecture:
		return "Architecture"
	case catalog.KindFramework:
		return "HTTP framework"
	case catalog.KindDatabase:
		return "Database"
	case catalog.KindFrontend:
		return "Frontend"
	case catalog.KindCSS:
		return "CSS framework"
	default:
		return "Data access"
	}
}

// matching returns the selections whose components have the names chosen
// for their kinds; kinds without a name match any component, and
// catalog.None for an optional kind matches selections without one.
func matching(supported []catalog.Selection, chosen map[catalog.Kind]string) []catalog.Selection {
	var out []catalog.Selection
	for _, sel := range supported {
		ok := true
		for _, kind := range catalog.Kinds() {
			name := chosen[kind]
			c := sel.Component(kind)
			switch {
			case name == "":
			case kind.Optional() && name == catalog.None:
				ok = ok && c == nil
			default:
				ok = ok && c != nil && c.Name == name
			}
		}
		if ok {
			out = append(out, sel)
		}
	}
	return out
}

// noneLabel returns the label of the option of optional kind that selects
// no component.
func noneLabel(kind catalog.Kind) string {
	if kind == catalog.KindCSS {
		return "None (plain CSS)"
	}
	return "None (API only)"
}

// componentOptions returns the distinct components of kind in sels as
// options sorted by name, labelled with their title and name. For an
// optional kind, the option catalog.None comes first when a selection in
// sels has no component of the kind.
func componentOptions(sels []catalog.Selection, kind catalog.Kind) []option {
	var opts []option
	none := false
	for _, sel := range sels {
		c := sel.Component(kind)
		if c == nil {
			none = true
			continue
		}
		if slices.ContainsFunc(opts, func(o option) bool { return o.value == c.Name }) {
			continue
		}
		label := c.Name
		if c.Title != "" && c.Title != c.Name {
			label = c.Title + " (" + c.Name + ")"
		}
		opts = append(opts, option{label: label, value: c.Name})
	}
	slices.SortFunc(opts, func(a, b option) int { return strings.Compare(a.value, b.value) })
	if kind.Optional() && none {
		opts = append([]option{{label: noneLabel(kind), value: catalog.None}}, opts...)
	}
	return opts
}

// equivalentCommand returns the `gobento new` command that repeats a run
// without questions: the name, the module path when it differs from the
// name, every kind that was given or has several supported options, and the
// other flags as given.
func equivalentCommand(supported []catalog.Selection, in newInput, f newFlags) string {
	args := []string{"gobento", "new", shellQuote(in.name)}
	if in.hasModule && in.module != in.name {
		args = append(args, "--module", shellQuote(in.module))
	}
	for _, kind := range catalog.Kinds() {
		value := in.components[kind]
		// A kind with a single supported option, and an optional kind
		// without a component, are left out unless given: gobento selects
		// them by itself.
		given := givenFlag(f, kind) != ""
		if value == "" || !given && (len(componentOptions(supported, kind)) < 2 || kind.Optional() && value == catalog.None) {
			continue
		}
		args = append(args, kindFlag(kind), shellQuote(value))
	}
	if f.goVersion != "" {
		args = append(args, "--go", shellQuote(f.goVersion))
	}
	for _, pin := range f.pins {
		args = append(args, "--pin", shellQuote(pin))
	}
	if f.dir != "" {
		args = append(args, "--dir", shellQuote(f.dir))
	}
	if f.dryRun {
		args = append(args, "--dry-run")
	}
	return strings.Join(args, " ")
}

// givenFlag returns the value of the flag for kind.
func givenFlag(f newFlags, kind catalog.Kind) string {
	switch kind {
	case catalog.KindArchitecture:
		return f.arch
	case catalog.KindFramework:
		return f.framework
	case catalog.KindDatabase:
		return f.db
	case catalog.KindFrontend:
		return f.frontend
	case catalog.KindCSS:
		return f.css
	default:
		return f.data
	}
}
