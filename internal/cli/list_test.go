package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/Tales-of-zuna/gobento/components"
	"github.com/Tales-of-zuna/gobento/internal/catalog"
	"github.com/Tales-of-zuna/gobento/internal/catalogjson"
	"github.com/Tales-of-zuna/gobento/internal/compose"
)

// listFixture is a frozen catalog document, a copy of the website's export
// at the time `gobento list` was added, extended with the optional frontend
// kind (htmx, its templ module and vendored asset, and a selection with htmx
// for every mvc and vertical-slice selection). The snapshot tests render it,
// so they do not change whenever a component ships.
const listFixture = "testdata/list/catalog.json"

// runList executes gobento with args against the document d (or the
// document built from the embedded catalog when d is nil), in a terminal whose questions fail the
// test, and returns stdout, stderr, and the command's error.
func runList(t *testing.T, d *catalogjson.Document, args ...string) (string, string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	a := &app{
		catalog:     components.FS(),
		generate:    (&fakeGenerator{}).generate,
		stdout:      &stdout,
		stderr:      &stderr,
		interactive: func() bool { return true },
		asker:       failingAsker{t},
		doc:         d,
	}
	cmd := newRootCommand("1.2.3", a)
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(context.Background())
	return stdout.String(), stderr.String(), err
}

// failingAsker fails the test when gobento asks a question (spec: No
// terminal needed).
type failingAsker struct{ t *testing.T }

func (f failingAsker) ask(context.Context, question) (string, error) {
	f.t.Error("gobento list asked a question")
	return "", errAborted
}

func loadListFixture(t *testing.T) *catalogjson.Document {
	t.Helper()
	data, err := os.ReadFile(listFixture)
	if err != nil {
		t.Fatal(err)
	}
	var d catalogjson.Document
	if err := json.Unmarshal(data, &d); err != nil {
		t.Fatal(err)
	}
	return &d
}

// loadShipped builds the document of the embedded catalog once: building
// it composes every candidate selection, which is slow under the race
// detector.
var loadShipped = sync.OnceValues(func() (*catalogjson.Document, error) {
	return catalogjson.Load(components.FS())
})

// shippedDocument returns the document of the embedded catalog. Tests must
// not modify it.
func shippedDocument(t *testing.T) *catalogjson.Document {
	t.Helper()
	d, err := loadShipped()
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// TestListSnapshots compares the output of every form of `gobento list` for
// the frozen fixture with testdata/list. Rewrite the snapshots with
// `go test ./internal/cli -run TestListSnapshots -update`.
func TestListSnapshots(t *testing.T) {
	t.Parallel()

	d := loadListFixture(t)
	tests := []struct {
		file string
		args []string
	}{
		{"list.txt", []string{"list"}},
		{"components.txt", []string{"list", "components"}},
		{"components-kind-framework.txt", []string{"list", "components", "--kind", "framework"}},
		{"components-kind-database.json", []string{"list", "components", "--kind", "database", "--json"}},
		{"selections.txt", []string{"list", "selections"}},
		{"selections-data-gorm.txt", []string{"list", "selections", "--data", "gorm"}},
		{"selections-arch-mvc-db-postgres.json", []string{"list", "selections", "--arch", "mvc", "--db", "postgres", "--json"}},
		{"versions.txt", []string{"list", "versions"}},
		{"versions.json", []string{"list", "versions", "--json"}},
	}
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			t.Parallel()

			out, stderr, err := runList(t, d, tt.args...)
			if err != nil {
				t.Fatalf("%v: %v\n%s", tt.args, err, stderr)
			}
			name := filepath.Join("testdata", "list", tt.file)
			if *update {
				if err := os.WriteFile(name, []byte(out), 0o600); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(name) //nolint:gosec // a snapshot file named by the test table
			if err != nil {
				t.Fatalf("%v; run go test ./internal/cli -run TestListSnapshots -update", err)
			}
			if out != string(want) {
				t.Errorf("gobento %s differs from %s:\n%s", strings.Join(tt.args, " "), name, out)
			}
		})
	}
}

// TestListComponentsShipped checks that every component of the embedded
// catalog is listed under its kind with its title (spec: All components, A
// component is added).
func TestListComponentsShipped(t *testing.T) {
	t.Parallel()

	d := shippedDocument(t)
	all, _, err := runList(t, shippedDocument(t), "list", "components")
	if err != nil {
		t.Fatal(err)
	}
	var sections []string
	for _, k := range d.Kinds {
		out, _, err := runList(t, shippedDocument(t), "list", "components", "--kind", k.ID)
		if err != nil {
			t.Fatal(err)
		}
		sections = append(sections, out)
		flag := kindFlag(catalog.Kind(k.ID))
		if k.Optional {
			flag += ", optional"
			if !strings.Contains(out, "  "+catalog.None+"  ") {
				t.Errorf("--kind %s does not list %s:\n%s", k.ID, catalog.None, out)
			}
		}
		heading := k.Title + " (" + flag + ")\n" + k.Description + "\n"
		if !strings.HasPrefix(out, heading) {
			t.Errorf("--kind %s does not start with %q:\n%s", k.ID, heading, out)
		}
		for _, c := range d.Components {
			if c.Kind == k.ID && !strings.Contains(out, "  "+c.Name+"  ") {
				t.Errorf("%s is not listed under %s", c.ID, k.Title)
			}
			if c.Kind == k.ID && !strings.Contains(out, c.Title+"\n") {
				t.Errorf("the title of %s is missing", c.ID)
			}
		}
	}
	if want := strings.Join(sections, "\n"); all != want {
		t.Errorf("gobento list components is not the kinds' lists in order:\n%s", all)
	}
	// A few known entries.
	for _, want := range []string{
		"Database (--db)",
		"  postgres  PostgreSQL\n",
		"provides: conn=database/sql, sql-dialect=postgresql, sql-schema=schema.sql\n",
		"requires: conn=database/sql, sql-dialect=sqlite|postgresql, sql-schema=schema.sql\n",
	} {
		if !strings.Contains(all, want) {
			t.Errorf("output lacks %q:\n%s", want, all)
		}
	}
}

func TestListComponentsKind(t *testing.T) {
	t.Parallel()

	d := shippedDocument(t)
	out, _, err := runList(t, shippedDocument(t), "list", "components", "--kind", "framework")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "HTTP framework (--framework)\n") {
		t.Errorf("output does not start with the framework kind:\n%s", out)
	}
	for _, c := range d.Components {
		listed := strings.Contains(out, "  "+c.Name+" ")
		if listed != (c.Kind == "framework") {
			t.Errorf("%s listed = %v, want %v", c.ID, listed, c.Kind == "framework")
		}
	}
}

// selectionRows returns the rows of the selections table in out, each as
// the names in kind order joined by "/", like catalog.Selection.String.
func selectionRows(t *testing.T, out string) []string {
	t.Helper()
	lines := strings.Split(out, "\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "ARCHITECTURE") {
		t.Fatalf("no table header in:\n%s", out)
	}
	var rows []string
	for _, line := range lines[1:] {
		if line == "" {
			break
		}
		rows = append(rows, strings.Join(strings.Fields(line), "/"))
	}
	return rows
}

// selectionRow returns sel as selectionRows reads a row of the table: the
// names of its components joined by slashes, with catalog.None for a
// missing frontend.
func selectionRow(sel catalog.Selection) string {
	row := sel.String()
	if sel.Component(catalog.KindFrontend) == nil {
		row += "/" + catalog.None
	}
	return row
}

// TestListSelectionsShipped checks that the listed selections are exactly
// the ones compose.Supported computes, in its order (spec: Every supported
// selection).
func TestListSelectionsShipped(t *testing.T) {
	t.Parallel()

	cat, err := catalog.Load(components.FS())
	if err != nil {
		t.Fatal(err)
	}
	var want []string
	for _, sel := range compose.Supported(cat) {
		want = append(want, selectionRow(sel))
	}

	out, _, err := runList(t, shippedDocument(t), "list", "selections")
	if err != nil {
		t.Fatal(err)
	}
	if got := selectionRows(t, out); !slices.Equal(got, want) {
		t.Errorf("rows = %v, want %v", got, want)
	}
	if !strings.Contains(out, "\n"+strconv.Itoa(len(want))+" supported selections.\n") {
		t.Errorf("no count line in:\n%s", out)
	}

	// Spec: Narrowed by component.
	out, _, err = runList(t, shippedDocument(t), "list", "selections", "--data", "gorm", "--db", "postgres")
	if err != nil {
		t.Fatal(err)
	}
	var narrowed []string
	for _, s := range want {
		if strings.Contains(s, "/postgres/gorm/") {
			narrowed = append(narrowed, s)
		}
	}
	if len(narrowed) == 0 {
		t.Fatal("no supported selection with postgres and gorm")
	}
	if got := selectionRows(t, out); !slices.Equal(got, narrowed) {
		t.Errorf("rows with --db postgres --data gorm = %v, want %v", got, narrowed)
	}
	if !strings.Contains(out, "\n"+strconv.Itoa(len(narrowed))+" of "+strconv.Itoa(len(want))+" supported selections.\n") {
		t.Errorf("no count line in:\n%s", out)
	}
}

// TestListVersionsShipped checks the Go rule and that every module of the
// rules is listed with its range and default (spec: Module with a trait
// condition and a constraint).
func TestListVersionsShipped(t *testing.T) {
	t.Parallel()

	d := shippedDocument(t)
	out, _, err := runList(t, shippedDocument(t), "list", "versions")
	if err != nil {
		t.Fatal(err)
	}
	if want := "Go " + d.Go.Range + " (default " + d.Go.Default + ")\n"; !strings.HasPrefix(out, want) {
		t.Errorf("output does not start with %q:\n%s", want, out)
	}
	for _, m := range d.Modules {
		var row string
		for _, line := range strings.Split(out, "\n") {
			if strings.HasPrefix(line, m.Path+" ") {
				row = line
			}
		}
		if row == "" {
			t.Errorf("module %s is not listed", m.Path)
			continue
		}
		for _, want := range append([]string{m.Role, m.Range, m.Default}, m.Components...) {
			if !strings.Contains(row, want) {
				t.Errorf("row of %s lacks %q: %s", m.Path, want, row)
			}
		}
		for _, c := range m.Constraints {
			if want := "  " + m.Path + " " + c.When + " needs " + c.Requires.Target + " " + c.Requires.Range + "\n"; !strings.Contains(out, want) {
				t.Errorf("constraint %q is not listed", want)
			}
		}
	}
	if !strings.Contains(out, "sql-dialect=sqlite\n") {
		t.Errorf("no module with a trait condition in:\n%s", out)
	}
}

// TestListOverview checks that `gobento list` contains every section
// (spec: Overview).
func TestListOverview(t *testing.T) {
	t.Parallel()

	all, _, err := runList(t, shippedDocument(t), "list")
	if err != nil {
		t.Fatal(err)
	}
	last := 0
	for _, sub := range []string{"components", "selections", "versions"} {
		out, _, err := runList(t, shippedDocument(t), "list", sub)
		if err != nil {
			t.Fatal(err)
		}
		i := strings.Index(all, out)
		if i < last {
			t.Errorf("gobento list does not contain the output of gobento list %s after the previous section", sub)
		}
		last = i
	}
}

// TestListJSON checks that `gobento list --json` is the exporter's
// document, byte for byte, that it decodes back into the same value, and
// that each subcommand prints only its fields (spec: Same data as the
// website, Scripts read the components).
func TestListJSON(t *testing.T) {
	t.Parallel()

	d := shippedDocument(t)
	var want bytes.Buffer
	if err := catalogjson.Encode(&want, d); err != nil {
		t.Fatal(err)
	}
	out, _, err := runList(t, nil, "list", "--json") // the real path: build the document from the catalog
	if err != nil {
		t.Fatal(err)
	}
	if out != want.String() {
		t.Error("gobento list --json differs from the exported document")
	}

	dec := json.NewDecoder(strings.NewReader(out))
	dec.DisallowUnknownFields()
	var back catalogjson.Document
	if err := dec.Decode(&back); err != nil {
		t.Fatal(err)
	}
	var again bytes.Buffer
	if err := catalogjson.Encode(&again, &back); err != nil {
		t.Fatal(err)
	}
	if again.String() != out {
		t.Error("the decoded document encodes differently")
	}

	fields := map[string][]string{
		"components": {"components", "kinds"},
		"selections": {"supported"},
		"versions":   {"assets", "go", "modules"},
	}
	for sub, want := range fields {
		out, _, err := runList(t, shippedDocument(t), "list", sub, "--json")
		if err != nil {
			t.Fatal(err)
		}
		var obj map[string]json.RawMessage
		if err := json.Unmarshal([]byte(out), &obj); err != nil {
			t.Fatal(err)
		}
		keys := sortedKeys(obj)
		if !slices.Equal(keys, want) {
			t.Errorf("gobento list %s --json has fields %v, want %v", sub, keys, want)
		}
		// Every field decodes like the same field of the full document.
		dec := json.NewDecoder(strings.NewReader(out))
		dec.DisallowUnknownFields()
		var part catalogjson.Document
		if err := dec.Decode(&part); err != nil {
			t.Errorf("gobento list %s --json: %v", sub, err)
		}
		switch sub {
		case "components":
			if !slices.EqualFunc(part.Components, d.Components, componentEqual) {
				t.Error("components differ from the document")
			}
		case "selections":
			if len(part.Supported) != len(d.Supported) {
				t.Error("supported differs from the document")
			}
		case "versions":
			if part.Go != d.Go || len(part.Modules) != len(d.Modules) || len(part.Assets) != len(d.Assets) {
				t.Error("go, modules, or assets differ from the document")
			}
		}
	}

	out, _, err = runList(t, shippedDocument(t), "list", "components", "--kind", "database", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var comps componentsJSON
	if err := json.Unmarshal([]byte(out), &comps); err != nil {
		t.Fatal(err)
	}
	if len(comps.Kinds) != 1 || comps.Kinds[0].ID != "database" {
		t.Errorf("kinds = %+v, want only database", comps.Kinds)
	}
	for _, c := range comps.Components {
		if c.Kind != "database" {
			t.Errorf("component %s listed with --kind database", c.ID)
		}
	}
	if !slices.ContainsFunc(comps.Components, func(c catalogjson.Component) bool { return c.ID == "database/postgres" }) {
		t.Error("database/postgres is missing")
	}

	out, _, err = runList(t, shippedDocument(t), "list", "selections", "--db", "mysql-nope", "--json")
	if err == nil {
		t.Errorf("unknown --db with --json succeeded: %s", out)
	}
}

func componentEqual(a, b catalogjson.Component) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return bytes.Equal(x, y)
}

func TestListErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			// Spec: Unknown kind.
			name: "unknown kind",
			args: []string{"list", "components", "--kind", "db"},
			want: `--kind: unknown kind "db"; supported: architecture, framework, database, data, frontend`,
		},
		{
			// Spec: Unknown component.
			name: "unknown component",
			args: []string{"list", "selections", "--db", "mysql"},
			want: `--db: unknown database "mysql"; supported: postgres, sqlite`,
		},
		{
			name: "several unknown components",
			args: []string{"list", "selections", "--framework", "express", "--data", "ent"},
			want: "cannot list selections:\n" +
				`  - --framework: unknown framework "express"; supported: chi, gin, nethttp` + "\n" +
				`  - --data: unknown data-access library "ent"; supported: gorm, sqlc`,
		},
		{
			name: "unknown frontend",
			args: []string{"list", "selections", "--frontend", "react"},
			want: `--frontend: unknown frontend "react"; supported: none, htmx`,
		},
		{
			name: "argument",
			args: []string{"list", "extra"},
			want: `unknown command "extra" for "gobento list"`,
		},
	}
	d := loadListFixture(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			out, _, err := runList(t, d, tt.args...)
			if err == nil {
				t.Fatalf("%v succeeded: %s", tt.args, out)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %q, want %q", err, tt.want)
			}
			if out != "" {
				t.Errorf("printed %q on failure", out)
			}
		})
	}
}

// TestListSelectionsFrontend checks the --frontend filter: a frontend keeps
// the selections with it, and none keeps those without a frontend.
func TestListSelectionsFrontend(t *testing.T) {
	t.Parallel()

	d := loadListFixture(t)
	for _, tt := range []struct {
		value string
		want  func(map[string]string) bool
	}{
		{"htmx", func(s map[string]string) bool { return s["frontend"] == "htmx" }},
		{catalog.None, func(s map[string]string) bool { _, ok := s["frontend"]; return !ok }},
	} {
		out, _, err := runList(t, d, "list", "selections", "--frontend", tt.value, "--json")
		if err != nil {
			t.Fatal(err)
		}
		var got selectionsJSON
		if err := json.Unmarshal([]byte(out), &got); err != nil {
			t.Fatal(err)
		}
		var want []map[string]string
		for _, s := range d.Supported {
			if tt.want(s) {
				want = append(want, s)
			}
		}
		if len(want) == 0 || !slices.EqualFunc(got.Supported, want, maps.Equal) {
			t.Errorf("--frontend %s = %v, want %v", tt.value, got.Supported, want)
		}

		text, _, err := runList(t, d, "list", "selections", "--frontend", tt.value)
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range selectionRows(t, text) {
			if !strings.HasSuffix(row, "/"+tt.value) {
				t.Errorf("--frontend %s lists %s", tt.value, row)
			}
		}
	}
}

// TestListNoMatch checks the message for valid filters without a match.
func TestListNoMatch(t *testing.T) {
	t.Parallel()

	d := loadListFixture(t)
	d.Supported = d.Supported[:1]
	other := "sqlc"
	if d.Supported[0]["data"] == other {
		other = "gorm"
	}
	out, _, err := runList(t, d, "list", "selections", "--data", other)
	if err != nil {
		t.Fatal(err)
	}
	if out != "No supported selection matches.\n" {
		t.Errorf("output = %q", out)
	}
	d.Supported = nil
	out, _, _ = runList(t, d, "list", "selections")
	if out != "No supported selections.\n" {
		t.Errorf("output = %q", out)
	}
}

// TestHelpPointsAtList checks the help of the root command and of new
// (spec: Help for new).
func TestHelpPointsAtList(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{{"--help"}, {"new", "--help"}} {
		out, _, err := runList(t, nil, args...)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, "gobento list") {
			t.Errorf("gobento %s does not mention gobento list:\n%s", strings.Join(args, " "), out)
		}
	}
}
