package cli

import (
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Tales-of-zuna/gobento/internal/catalog"
	"github.com/Tales-of-zuna/gobento/internal/compose"
	"github.com/Tales-of-zuna/gobento/internal/generate"
)

// newFlags are the flags of `gobento new`.
type newFlags struct {
	module    string
	arch      string
	framework string
	db        string
	data      string
	goVersion string
	pins      []string
	dir       string
	dryRun    bool
	noInput   bool
}

// kindFlag returns the flag that selects components of kind k.
func kindFlag(k catalog.Kind) string {
	switch k {
	case catalog.KindArchitecture:
		return "--arch"
	case catalog.KindFramework:
		return "--framework"
	case catalog.KindDatabase:
		return "--db"
	default:
		return "--data"
	}
}

// newNewCommand returns the `gobento new` command.
func newNewCommand(a *app) *cobra.Command {
	var f newFlags
	cmd := &cobra.Command{
		Use:   "new [name]",
		Short: "Create a new Go project",
		Long: `Create a new Go project named <name> in ./<name>, or in --dir.

A component kind with a single supported option is selected automatically;
for a kind with several, choose one with its flag. Versions default to
gobento's compatibility rules. All input is checked before anything is
written, and every problem is reported at once. If generation fails or is
interrupted, nothing is left behind.

In a terminal, gobento asks for what is missing: the name (and then the module
path) when <name> is omitted, and each component kind with several options
whose flag is omitted. Flags skip their questions. Without a terminal, or with
--no-input, gobento asks nothing: <name> is required, and a kind with several
options needs its flag. Set ACCESSIBLE=1 for plain-text questions.

Generation runs the go command (go get and go mod tidy), so it needs Go and
access to the Go module proxy. --dry-run needs neither.`,
		Example: `  # Create ./shop with mvc, the chi router, and the default components and versions
  gobento new shop --arch mvc --framework chi

  # Organize the code by feature instead of by layer
  gobento new shop --arch vertical-slice --framework nethttp

  # Choose every component and the module path explicitly
  gobento new shop --module example.com/shop --arch mvc --framework nethttp --db sqlite --data sqlc

  # Choose versions within the supported ranges
  gobento new shop --arch mvc --framework nethttp --go 1.26 --pin modernc.org/sqlite=v1.61.0

  # Fill the current, empty directory
  gobento new shop --arch mvc --framework nethttp --dir .

  # Show the files that would be created, without writing anything
  gobento new shop --arch mvc --framework gin --dry-run

  # Never ask a question, even in a terminal (as in scripts)
  gobento new shop --arch mvc --framework nethttp --no-input`,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 1 {
				return nameArgError(len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runNew(cmd, args, f)
		},
	}

	fl := cmd.Flags()
	fl.StringVar(&f.module, "module", "", "Go module path, for example example.com/shop (default: the project name)")
	fl.StringVar(&f.arch, "arch", "", "architecture, for example mvc or vertical-slice (default: the only supported one)")
	fl.StringVar(&f.framework, "framework", "", "HTTP framework, for example chi, gin, or nethttp (default: the only supported one)")
	fl.StringVar(&f.db, "db", "", "database, for example sqlite (default: the only supported one)")
	fl.StringVar(&f.data, "data", "", "data-access library, for example sqlc (default: the only supported one)")
	fl.StringVar(&f.goVersion, "go", "", "Go version for go.mod, for example 1.26 (default: from the compatibility rules)")
	fl.StringArrayVar(&f.pins, "pin", nil, "library version as module=version, for example modernc.org/sqlite=v1.61.0 (repeatable)")
	fl.StringVar(&f.dir, "dir", "", "target directory, for example . for the current one (default: ./<name>)")
	fl.BoolVar(&f.dryRun, "dry-run", false, "check the input and list the files that would be created, without writing anything")
	fl.BoolVar(&f.noInput, "no-input", false, "never ask a question; fail on missing input instead, as without a terminal")
	return cmd
}

// nameArgError reports a wrong number of project name arguments.
func nameArgError(got int) error {
	return fmt.Errorf("expected exactly one project name, for example: gobento new shop (got %d arguments)", got)
}

// mayAsk reports whether gobento may ask questions in this run.
func (a *app) mayAsk(f newFlags) bool {
	return !f.noInput && a.interactive != nil && a.interactive()
}

// runNew asks for missing input in a terminal, validates the input,
// composes the project, and either lists the planned files (--dry-run) or
// generates the project and prints next steps.
func (a *app) runNew(cmd *cobra.Command, args []string, f newFlags) error {
	asking := a.mayAsk(f)
	if len(args) == 0 && !asking {
		return nameArgError(0)
	}
	cat, err := catalog.Load(a.catalog)
	if err != nil {
		return fmt.Errorf("load the component catalog: %w", err)
	}

	in := newInput{
		hasName:   len(args) == 1,
		module:    f.module,
		hasModule: f.module != "", // --module "" means the default, as without it
		components: map[catalog.Kind]string{
			catalog.KindArchitecture: f.arch,
			catalog.KindFramework:    f.framework,
			catalog.KindDatabase:     f.db,
			catalog.KindData:         f.data,
		},
	}
	if in.hasName {
		in.name = args[0]
	}
	pins, pinErrs := parsePins(f.pins)
	if asking {
		supported := compose.Supported(cat)
		if canAsk(supported, in, pinErrs) {
			asked, err := interview(cmd.Context(), supported, &in, a.questions())
			if err != nil {
				return err
			}
			if asked {
				_, _ = fmt.Fprintf(a.stderr, "Equivalent command: %s\n", equivalentCommand(supported, in, f))
			}
		}
	}
	if !in.hasName {
		return nameArgError(0)
	}

	name := in.name
	prepared, err := generate.Prepare(cat, generate.Spec{
		Name:       name,
		Module:     in.module,
		Components: in.components,
		Go:         f.goVersion,
		Pins:       pins,
	})
	if err != nil || len(pinErrs) > 0 {
		return problemsError(name, append(pinErrs, err))
	}

	target := f.dir
	if target == "" {
		target = name
	}
	if f.dryRun {
		return a.printDryRun(prepared, target)
	}

	req := prepared.Request(target)
	req.Warn = a.stderr
	if err := a.generate(cmd.Context(), req); err != nil {
		return problemsError(name, []error{err})
	}
	return a.printNextSteps(prepared, target)
}

// parsePins parses repeated --pin module=version values.
func parsePins(values []string) (map[string]string, []error) {
	pins := make(map[string]string, len(values))
	var errs []error
	for _, v := range values {
		mod, version, ok := strings.Cut(v, "=")
		if !ok || mod == "" || version == "" {
			errs = append(errs, fmt.Errorf("--pin %q: want module=version, for example modernc.org/sqlite=v1.61.0", v))
			continue
		}
		pins[mod] = version
	}
	return pins, errs
}

// problemsError formats every problem on its own line, naming the flag for
// selection problems.
func problemsError(name string, errs []error) error {
	var lines []string
	for _, err := range errs {
		lines = append(lines, problemLines(err)...)
	}
	return &reportError{
		msg:  fmt.Sprintf("cannot create %q:\n  - %s", name, strings.Join(lines, "\n  - ")),
		errs: errs,
	}
}

// problemLines splits the top level of an errors.Join tree into lines and
// names the flag of each selection problem.
func problemLines(err error) []string {
	if err == nil {
		return nil
	}
	if se, ok := err.(*generate.SelectionError); ok { //nolint:errorlint // only a direct element of the report is formatted as a flag problem
		return []string{kindFlag(se.Kind) + ": " + se.Error()}
	}
	if list, ok := joined(err); ok {
		var out []string
		for _, e := range list {
			out = append(out, problemLines(e)...)
		}
		return out
	}
	return strings.Split(err.Error(), "\n")
}

// joined returns the elements of err if err itself is a list made by
// errors.Join, whose message is its elements' messages separated by
// newlines. Errors that merely wrap several errors are not split.
func joined(err error) ([]error, bool) {
	multi, ok := err.(interface{ Unwrap() []error })
	if !ok {
		return nil, false
	}
	list := multi.Unwrap()
	msgs := make([]string, len(list))
	for i, e := range list {
		msgs[i] = e.Error()
	}
	return list, err.Error() == strings.Join(msgs, "\n")
}

// reportError is the multi-line problem report of a failed run.
type reportError struct {
	msg  string
	errs []error
}

func (e *reportError) Error() string   { return e.msg }
func (e *reportError) Unwrap() []error { return e.errs }

// printDryRun lists the files a run would create, including go.mod and
// go.sum, which generation adds.
func (a *app) printDryRun(p *generate.Prepared, target string) error {
	paths := append(p.Plan.Paths(), "go.mod", "go.sum")
	slices.Sort(paths)
	var b strings.Builder
	fmt.Fprintf(&b, "Dry run: nothing was written.\n")
	fmt.Fprintf(&b, "gobento would create %s (%s, Go %s) in %s with these files:\n",
		p.Module, p.Selection, p.Versions.Go, target)
	for _, path := range paths {
		fmt.Fprintf(&b, "  %s\n", path)
	}
	_, err := io.WriteString(a.stdout, b.String())
	return err
}

// printNextSteps reports the created project and the commands to try next.
func (a *app) printNextSteps(p *generate.Prepared, target string) error {
	var b strings.Builder
	if abs, err := filepath.Abs(target); err == nil && isCurrentDir(abs) {
		fmt.Fprintf(&b, "Created %s in the current directory.\n\nNext steps:\n", p.Module)
	} else {
		fmt.Fprintf(&b, "Created %s in %s.\n\nNext steps:\n  cd %s\n", p.Module, target, shellQuote(target))
	}
	b.WriteString("  go test ./...\n")
	if entry, ok := p.Plan.Slots["entry"]; ok {
		fmt.Fprintf(&b, "  go run ./%s\n", entry.Dir)
	}
	_, err := io.WriteString(a.stdout, b.String())
	return err
}

// isCurrentDir reports whether abs is the working directory.
func isCurrentDir(abs string) bool {
	wd, err := filepath.Abs(".")
	return err == nil && wd == abs
}

// shellQuote quotes path for a copy-pasted shell command when it contains
// spaces or other characters a shell would split or expand.
func shellQuote(path string) string {
	if path != "" && !strings.ContainsAny(path, " \t'\"$`\\*?[]{}()<>|&;#~!") {
		return path
	}
	return "'" + strings.ReplaceAll(path, "'", `'\''`) + "'"
}
