package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/Tales-of-zuna/gobento/components"
	"github.com/Tales-of-zuna/gobento/internal/generate"
)

// app holds what the commands depend on, so tests can replace the catalog,
// the generation step, and the output streams.
type app struct {
	catalog  fs.FS
	generate func(context.Context, generate.Request) error
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
	// interactive reports whether stdin and stdout are terminals, so that
	// gobento may ask questions. nil means never.
	interactive func() bool
	// getenv reads an environment variable. nil means none is set.
	getenv func(string) string
	// asker asks the questions. nil means huh on stdin and stdout.
	asker asker
}

func defaultApp() *app {
	return &app{
		catalog:     components.FS(),
		generate:    generate.Generate,
		stdin:       os.Stdin,
		stdout:      os.Stdout,
		stderr:      os.Stderr,
		interactive: stdioIsTerminal,
		getenv:      os.Getenv,
	}
}

// stdioIsTerminal reports whether both stdin and stdout are terminals.
func stdioIsTerminal() bool {
	return isTerminal(os.Stdin) && isTerminal(os.Stdout)
}

// isTerminal reports whether f is a terminal, including a Windows console.
func isTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

// env returns the environment variable key, or "" when a has no getenv.
func (a *app) env(key string) string {
	if a.getenv == nil {
		return ""
	}
	return a.getenv(key)
}

// accessible reports whether questions use huh's accessible mode: when
// ACCESSIBLE is set to a non-empty value, or the terminal is dumb.
func (a *app) accessible() bool {
	return a.env("ACCESSIBLE") != "" || a.env("TERM") == "dumb"
}

// questions returns the asker for gobento's questions.
func (a *app) questions() asker {
	if a.asker != nil {
		return a.asker
	}
	return huhAsker{in: a.stdin, out: a.stdout, accessible: a.accessible(), width: formWidth(a.stdout)}
}

// minFormWidth is the narrowest terminal the full-screen form adapts to.
const minFormWidth = 20

// formWidth returns 0, which lets the full-screen form follow the terminal
// width, unless out is a terminal that reports no usable size (for example a
// pseudo-terminal sized 0x0, which the form cannot draw in). The form then
// gets a fixed width of 80 columns.
func formWidth(out io.Writer) int {
	f, ok := out.(*os.File)
	if !ok {
		return 0
	}
	cols, _, err := term.GetSize(int(f.Fd()))
	if err != nil || cols < minFormWidth {
		return 80
	}
	return 0
}

// NewRootCommand returns the gobento root command with its subcommands.
// version is reported by `gobento --version`.
func NewRootCommand(version string) *cobra.Command {
	return newRootCommand(version, defaultApp())
}

func newRootCommand(version string, a *app) *cobra.Command {
	root := &cobra.Command{
		Use:   "gobento",
		Short: "Generate a new Go project from selectable components",
		Long: `gobento generates a new Go project from a selection of components:
an architecture, an HTTP framework, a database, and a data-access library.
Dependency versions come from gobento's compatibility rules, and every
supported combination is built and tested in gobento's CI.`,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(a.stdout)
	root.SetErr(a.stderr)
	root.AddCommand(newNewCommand(a))
	return root
}

// Execute runs the root command with the process arguments and returns the
// first error a command reports.
func Execute(ctx context.Context, version string) error {
	return NewRootCommand(version).ExecuteContext(ctx)
}

// ExitInterrupted is the exit status after an interrupt, following the shell
// convention of 128 plus the signal number of SIGINT.
const ExitInterrupted = 130

// Run runs gobento with the process arguments, reports any error on stderr,
// and returns the process exit status: 0 on success, ExitInterrupted when
// ctx was cancelled by a signal or a question was cancelled, and 1
// otherwise.
func Run(ctx context.Context, version string) int {
	a := defaultApp()
	return a.exitStatus(ctx, newRootCommand(version, a).ExecuteContext(ctx))
}

// exitStatus reports err and maps it to an exit status.
func (a *app) exitStatus(ctx context.Context, err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, errAborted):
		_, _ = fmt.Fprintln(a.stderr, "gobento: interrupted; nothing was left behind")
		return ExitInterrupted
	case ctx.Err() != nil && errors.Is(err, context.Canceled):
		if errors.Is(err, generate.ErrCleanup) {
			_, _ = fmt.Fprintf(a.stderr, "gobento: interrupted, and some files could not be removed:\n%v\n", err)
		} else {
			_, _ = fmt.Fprintln(a.stderr, "gobento: interrupted; nothing was left behind")
		}
		return ExitInterrupted
	default:
		_, _ = fmt.Fprintln(a.stderr, "gobento:", err)
		return 1
	}
}
