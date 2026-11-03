// Command gobento generates new Go projects from a selection of components.
//
// The command-line interface lives in package internal/cli; this package only
// provides the entry point, the build version, and signal handling.
package main

import (
	"context"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"

	"github.com/Tales-of-zuna/gobento/internal/cli"
)

// version is the gobento release version. GoReleaser sets it at build time
// with -ldflags "-X main.version=<version>". Other builds keep "dev", and
// buildVersion then falls back to the module version Go recorded.
var version = "dev"

func main() {
	// Ctrl-C and SIGTERM cancel the context; gobento then removes
	// everything it created and exits with status 130.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := cli.Run(ctx, buildVersion(version, debug.ReadBuildInfo))
	stop()
	os.Exit(code)
}

// buildVersion returns the version `gobento --version` reports. A version
// set with -ldflags wins. Otherwise it is the main module version from the
// build info, without the leading "v": `go install ...@v0.1.0` records
// "v0.1.0", and builds from a checkout may record a VCS pseudo-version.
// Without either it stays "dev" (for example, `go run` or a build with
// -buildvcs=false, which record "(devel)").
func buildVersion(ldflagsVersion string, readBuildInfo func() (*debug.BuildInfo, bool)) string {
	if ldflagsVersion != "dev" {
		return ldflagsVersion
	}
	info, ok := readBuildInfo()
	if !ok || info == nil {
		return ldflagsVersion
	}
	switch v := info.Main.Version; v {
	case "", "(devel)":
		return ldflagsVersion
	default:
		return strings.TrimPrefix(v, "v")
	}
}
