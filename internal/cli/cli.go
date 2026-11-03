package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/spf13/cobra"

	"github.com/Tales-of-zuna/gobento/components"
	"github.com/Tales-of-zuna/gobento/internal/generate"
)

// app holds what the commands depend on, so tests can replace the catalog,
// the generation step, and the output streams.
type app struct {
	catalog  fs.FS
	generate func(context.Context, generate.Request) error
	stdout   io.Writer
	stderr   io.Writer
}

func defaultApp() *app {
	return &app{
		catalog:  components.FS(),
		generate: generate.Generate,
		stdout:   os.Stdout,
		stderr:   os.Stderr,
	}
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
// ctx was cancelled by a signal, and 1 otherwise.
func Run(ctx context.Context, version string) int {
	a := defaultApp()
	return a.exitStatus(ctx, newRootCommand(version, a).ExecuteContext(ctx))
}

// exitStatus reports err and maps it to an exit status.
func (a *app) exitStatus(ctx context.Context, err error) int {
	switch {
	case err == nil:
		return 0
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
