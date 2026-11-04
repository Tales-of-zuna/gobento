// Command catalogexport writes gobento's component catalog, version rules,
// supported selections, and input rules as JSON for the website, which reads
// them from website/src/generated/catalog.json.
//
// The document and its conversion live in internal/catalogjson, which
// `gobento list --json` uses too, so the website shows exactly what the CLI
// lists and what `gobento new` accepts. Its schema is
// website/lib/catalog/schema.ts.
//
// Usage:
//
//	go run ./internal/tools/catalogexport                  # print to stdout
//	go run ./internal/tools/catalogexport -o catalog.json  # write a file
//
// `make catalog` regenerates the website's copy and `make catalog-check`
// fails if it is out of date.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tales-of-zuna/gobento/components"
	"github.com/Tales-of-zuna/gobento/internal/catalogjson"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "catalogexport:", err)
		os.Exit(1)
	}
}

// run exports the embedded catalog to stdout, or to the file named by -o.
func run(args []string, stdout io.Writer) error {
	fl := flag.NewFlagSet("catalogexport", flag.ContinueOnError)
	output := fl.String("o", "", "write to this file instead of stdout")
	if err := fl.Parse(args); err != nil {
		return err
	}
	if fl.NArg() > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(fl.Args(), " "))
	}

	o, err := catalogjson.Load(components.FS())
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := catalogjson.Encode(&buf, o); err != nil {
		return err
	}
	if *output == "" {
		_, err := stdout.Write(buf.Bytes())
		return err
	}
	return writeFile(*output, buf.Bytes())
}

// writeFile replaces name with data through a temporary file in the same
// directory, so a failed run never leaves a truncated file behind.
func writeFile(name string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(name), ".catalogexport-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) //nolint:errcheck // already renamed on success; best effort otherwise
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil { //nolint:gosec // a source file meant to be committed and read by everyone
		return err
	}
	return os.Rename(tmp.Name(), name)
}
