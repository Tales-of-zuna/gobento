package cli

import (
	"context"

	"github.com/spf13/cobra"
)

// NewRootCommand returns the gobento root command. version is reported by
// `gobento --version`; subcommands are attached here as they are added.
func NewRootCommand(version string) *cobra.Command {
	return &cobra.Command{
		Use:           "gobento",
		Short:         "Generate a new Go project from selectable components",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// Execute runs the root command with the process arguments and returns the
// first error a command reports.
func Execute(ctx context.Context, version string) error {
	return NewRootCommand(version).ExecuteContext(ctx)
}
