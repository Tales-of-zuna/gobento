// Command gobento generates new Go projects from a selection of components.
//
// The command-line interface lives in package internal/cli; this package only
// provides the entry point, the build version, and the process exit status.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/Tales-of-zuna/gobento/internal/cli"
)

// version is the gobento release version. GoReleaser sets it at build time
// with -ldflags "-X main.version=<version>"; local builds report "dev".
var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	err := cli.Execute(ctx, version)
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, "gobento:", err)
		os.Exit(1)
	}
}
