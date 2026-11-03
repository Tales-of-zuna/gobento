package cli

import (
	"bufio"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/Tales-of-zuna/gobento/components"
	"github.com/Tales-of-zuna/gobento/internal/catalog"
	"github.com/Tales-of-zuna/gobento/internal/compose"
)

// websiteCommands lists the commands the website's configurator builds for
// every supported selection. The website test
// website/lib/command/cli-commands.test.ts writes it with the configurator's
// own command builder, so neither side duplicates the other's logic.
const websiteCommands = "../../website/lib/command/supported-commands.txt"

// dryRunSelection extracts the selection from the dry-run summary line,
// for example "(mvc/nethttp/sqlite/sqlc, Go 1.27)".
var dryRunSelection = regexp.MustCompile(`\(([a-z0-9/-]+), Go [0-9.]+\)`)

// TestWebsiteCommands runs every command the website offers with --dry-run
// and checks that the CLI accepts it, without generating anything, and that
// together they cover every supported selection (spec: Generated commands
// are accepted by the CLI).
func TestWebsiteCommands(t *testing.T) {
	t.Parallel()

	f, err := os.Open(websiteCommands)
	if err != nil {
		t.Fatalf("open the website's commands: %v", err)
	}
	defer f.Close() //nolint:errcheck // read-only

	var (
		covered []string
		lines   int
	)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines++
		args, ok := strings.CutPrefix(line, "gobento ")
		if !ok || strings.ContainsAny(args, `'"\`) {
			t.Errorf("%q: want a plain `gobento ...` command without quoting", line)
			continue
		}
		gen := &fakeGenerator{}
		out, stderr, err := run(t, components.FS(), gen, append(strings.Fields(args), "--dry-run")...)
		if err != nil {
			t.Errorf("%s --dry-run: %v\n%s", line, err, stderr)
			continue
		}
		if len(gen.calls) != 0 {
			t.Errorf("%s --dry-run generated a project", line)
		}
		m := dryRunSelection.FindStringSubmatch(out)
		if m == nil {
			t.Errorf("%s --dry-run: no selection in the output:\n%s", line, out)
			continue
		}
		if !slices.Contains(covered, m[1]) {
			covered = append(covered, m[1])
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if lines == 0 {
		t.Fatalf("%s lists no commands", websiteCommands)
	}

	cat, err := catalog.Load(components.FS())
	if err != nil {
		t.Fatal(err)
	}
	var want []string
	for _, sel := range compose.Supported(cat) {
		want = append(want, sel.String())
	}
	slices.Sort(covered)
	slices.Sort(want)
	if !slices.Equal(covered, want) {
		t.Errorf("the website's commands cover %v, want every supported selection %v; "+
			"run `make catalog` and `pnpm test -u` in website/", covered, want)
	}
}
