// Command gobento generates new Go projects from a selection of components.
//
// The command-line interface lives in package internal/cli; this package only
// provides the entry point, the build version, and signal handling.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Tales-of-zuna/gobento/internal/cli"
)

// version is the gobento release version. GoReleaser sets it at build time
// with -ldflags "-X main.version=<version>"; local builds report "dev".
var version = "dev"

func main() {
	// Ctrl-C and SIGTERM cancel the context; gobento then removes
	// everything it created and exits with status 130.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := cli.Run(ctx, version)
	stop()
	os.Exit(code)
}
