package compose

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestComposePlan(t *testing.T) {
	t.Parallel()

	cat := loadCatalog(t, baseFixture())
	plan, err := Compose(input(pick(t, cat, "layered", "web", "lite", "query")))
	if err != nil {
		t.Fatalf("Compose() error = %v", err)
	}

	wantPaths := []string{
		".gitignore",
		"README.md",
		"cmd/shop/main.go",
		"internal/controllers/router.go",
		"internal/controllers/router_test.go",
		"internal/database/db.go",
		"internal/models/models.go",
		"internal/store/store.go",
		"query.yaml",
	}
	if got := plan.Paths(); !slices.Equal(got, wantPaths) {
		t.Errorf("Paths() =\n%v\nwant\n%v", got, wantPaths)
	}
	for _, f := range plan.Files {
		if f.Mode != 0o644 {
			t.Errorf("%s: Mode = %v, want 0644", f.Path, f.Mode)
		}
	}

	tests := []struct {
		path      string
		component string
		contains  []string
	}{
		{
			path:      "cmd/shop/main.go",
			component: "architecture/layered",
			contains: []string{
				`"example.com/shop/internal/database"`,
				`"example.com/shop/internal/controllers"`,
				"database.Open()",
				"controllers.NewRouter(store.New(db))",
			},
		},
		{
			// .Versions and .Traits, including names with dashes.
			path:      "internal/database/db.go",
			component: "database/lite",
			contains:  []string{"// Driver v1.2.0 speaks sqlite over database/sql."},
		},
		{
			path:      "README.md",
			component: "architecture/layered",
			contains:  []string{"# shop", "Module example.com/shop, Go 1.27."},
		},
		{path: "query.yaml", component: "data/query", contains: []string{"package: store"}},
	}
	for _, tt := range tests {
		f, ok := plan.File(tt.path)
		if !ok {
			t.Errorf("File(%q) not found", tt.path)
			continue
		}
		if f.Component != tt.component {
			t.Errorf("%s: Component = %q, want %q", tt.path, f.Component, tt.component)
		}
		for _, w := range tt.contains {
			if !strings.Contains(string(f.Content), w) {
				t.Errorf("%s does not contain %q:\n%s", tt.path, w, f.Content)
			}
		}
	}
	if _, ok := plan.File("missing.go"); ok {
		t.Error(`File("missing.go") found`)
	}
}

// TestComposeSameFragmentDifferentArchitecture covers the spec scenario
// "Same fragment, different architecture layout": the framework's handler
// file lands in each architecture's handlers directory, with the matching
// package name and import paths.
func TestComposeSameFragmentDifferentArchitecture(t *testing.T) {
	t.Parallel()

	cat := loadCatalog(t, baseFixture())
	tests := []struct {
		arch       string
		path       string
		wantPkg    string
		wantImport string
	}{
		{
			arch:       "layered",
			path:       "internal/controllers/router.go",
			wantPkg:    "package controllers\n",
			wantImport: `"example.com/shop/internal/models"`,
		},
		{
			arch:       "hexagonal",
			path:       "internal/http/handlers/router.go",
			wantPkg:    "package handlers\n",
			wantImport: `"example.com/shop/internal/domain"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.arch, func(t *testing.T) {
			t.Parallel()

			plan, err := Compose(input(pick(t, cat, tt.arch, "web", "lite", "query")))
			if err != nil {
				t.Fatalf("Compose() error = %v", err)
			}
			f, ok := plan.File(tt.path)
			if !ok {
				t.Fatalf("File(%q) not found; paths: %v", tt.path, plan.Paths())
			}
			content := string(f.Content)
			if !strings.HasPrefix(content, tt.wantPkg) {
				t.Errorf("%s does not start with %q:\n%s", tt.path, tt.wantPkg, content)
			}
			if !strings.Contains(content, tt.wantImport) {
				t.Errorf("%s does not import %s:\n%s", tt.path, tt.wantImport, content)
			}
			models := "models.ItemStore"
			if tt.arch == "hexagonal" {
				models = "domain.ItemStore"
			}
			if !strings.Contains(content, models) {
				t.Errorf("%s does not use %s:\n%s", tt.path, models, content)
			}
		})
	}
}

func TestComposeNormalizesCRLF(t *testing.T) {
	t.Parallel()

	cat := loadCatalog(t, withFiles(map[string]string{
		"framework/web/files/root/web.txt.tmpl": "line one\r\nline two {{.Project.Name}}\r\n",
	}))
	plan, err := Compose(input(pick(t, cat, "layered", "web", "lite", "query")))
	if err != nil {
		t.Fatalf("Compose() error = %v", err)
	}
	f, _ := plan.File("web.txt")
	if got := string(f.Content); got != "line one\nline two shop\n" {
		t.Errorf("web.txt = %q, want LF line endings", got)
	}
}

// TestComposeErrors covers the spec scenarios of "One owner per slot",
// "Traits connect compatible components", and "Reject selections that
// cannot be composed", plus the other composition checks.
func TestComposeErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		files   map[string]string
		sel     [4]string // architecture, framework, database, data
		project string    // project name; "shop" when empty
		wantIs  []error
		wantMsg []string
	}{
		{
			// Spec: Two components target one slot.
			name:    "two owners",
			files:   map[string]string{"framework/web/files/store/cache.go.tmpl": "package {{.Slots.store.Package}}\n"},
			wantIs:  []error{ErrSlotOwner},
			wantMsg: []string{"slot store is filled by both framework/web and data/query"},
		},
		{
			name: "wrong kind of owner",
			files: map[string]string{
				"database/lite/files/database/db.go.tmpl": "",
				"framework/web/files/database/db.go.tmpl": "package {{.Slots.database.Package}}\n",
			},
			wantIs:  []error{ErrSlotOwner},
			wantMsg: []string{"slot database is filled by the database component, but framework/web has templates for it"},
		},
		{
			// Spec: Unknown slot.
			name:    "unknown slot",
			files:   map[string]string{"data/query/files/cache/cache.go.tmpl": "package cache\n"},
			wantIs:  []error{ErrUnknownSlot},
			wantMsg: []string{"data/query has templates for slot cache, which architecture/layered does not declare"},
		},
		{
			// Spec: Required slot unfilled.
			name:    "required slot unfilled",
			files:   map[string]string{"data/query/files/store/store.go.tmpl": ""},
			wantIs:  []error{ErrUnfilledSlot},
			wantMsg: []string{"slot store (internal/store) must be filled by the data component"},
		},
		{
			// Spec: Unmet trait.
			name: "unmet trait",
			files: map[string]string{"data/query/component.yaml": `name: query
title: Query
requires:
  sql-dialect: [postgres]
`},
			wantIs:  []error{ErrTraits},
			wantMsg: []string{"trait sql-dialect", "data/query", "database/lite", `"sqlite"`},
		},
		{
			name:    "root path conflict",
			files:   map[string]string{"framework/web/files/root/query.yaml.tmpl": "x\n"},
			wantIs:  []error{ErrPathConflict},
			wantMsg: []string{"root file query.yaml is produced by both framework/web and data/query"},
		},
		{
			name:    "reserved root path",
			files:   map[string]string{"data/query/files/root/go.mod.tmpl": "module x\n"},
			wantIs:  []error{ErrPathConflict},
			wantMsg: []string{"data/query may not produce go.mod"},
		},
		{
			name:    "root file inside a slot",
			files:   map[string]string{"data/query/files/root/internal/controllers/extra.go.tmpl": "package controllers\n"},
			wantIs:  []error{ErrPathConflict},
			wantMsg: []string{"root file internal/controllers/extra.go from data/query lies inside slot handlers"},
		},
		{
			name:    "incomplete selection",
			sel:     [4]string{"layered", "web", "lite", ""},
			wantIs:  []error{ErrSelection, ErrUnfilledSlot},
			wantMsg: []string{"no data component selected", "slot store"},
		},
		{
			name:    "invalid package name",
			files:   map[string]string{"architecture/layered/component.yaml": strings.Replace(baseFixture()["architecture/layered/component.yaml"], "package: main, ", "", 1)},
			project: "my-app",
			wantIs:  []error{ErrTemplate},
			wantMsg: []string{`slot entry: "my-app" is not a valid Go package name`},
		},
		{
			name: "every problem is reported",
			files: map[string]string{
				"data/query/files/cache/cache.go.tmpl":    "package cache\n",
				"framework/web/files/store/cache.go.tmpl": "package store\n",
				"data/query/component.yaml":               "name: query\ntitle: Query\nrequires:\n  conn: [pgx]\n",
			},
			wantIs:  []error{ErrUnknownSlot, ErrSlotOwner, ErrTraits},
			wantMsg: []string{"slot cache", "slot store is filled by both", "trait conn"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cat := loadCatalog(t, withFiles(tt.files))
			names := tt.sel
			if names == [4]string{} {
				names = [4]string{"layered", "web", "lite", "query"}
			}
			in := input(pick(t, cat, names[0], names[1], names[2], names[3]))
			if tt.project != "" {
				in.Project.Name = tt.project
			}
			plan, err := Compose(in)
			if err == nil {
				t.Fatalf("Compose() error = nil, want an error; paths: %v", plan.Paths())
			}
			if plan != nil {
				t.Errorf("Compose() plan = %v, want nil on error", plan.Paths())
			}
			for _, target := range tt.wantIs {
				if !errors.Is(err, target) {
					t.Errorf("Compose() error = %q\nwant errors.Is(%v)", err, target)
				}
			}
			for _, w := range tt.wantMsg {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("Compose() error = %q\nwant it to contain %q", err, w)
				}
			}
		})
	}
}

func TestComposeFormatsGo(t *testing.T) {
	t.Parallel()

	cat := loadCatalog(t, withFiles(map[string]string{
		"framework/web/files/handlers/ugly.go.tmpl": "package {{.Slots.handlers.Package}}\nfunc   Ugly( ) int {return 1}\n",
		"framework/web/files/root/ugly.txt.tmpl":    "func   Ugly( ) int {return 1}\n",
	}))
	plan, err := Compose(input(pick(t, cat, "layered", "web", "lite", "query")))
	if err != nil {
		t.Fatalf("Compose() error = %v", err)
	}
	goFile, _ := plan.File("internal/controllers/ugly.go")
	if want := "package controllers\n\nfunc Ugly() int { return 1 }\n"; string(goFile.Content) != want {
		t.Errorf("ugly.go = %q, want %q", goFile.Content, want)
	}
	txt, _ := plan.File("ugly.txt")
	if want := "func   Ugly( ) int {return 1}\n"; string(txt.Content) != want {
		t.Errorf("ugly.txt = %q, want it unchanged", txt.Content)
	}
}

// TestComposeTemplateErrors covers task 4.3: a template whose output is not
// valid Go, and templates that fail to parse or execute, are reported with
// the component and the file.
func TestComposeTemplateErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		files   map[string]string
		wantMsg []string
	}{
		{
			name:    "go output does not format",
			files:   map[string]string{"framework/web/files/handlers/broken.go.tmpl": "package {{.Slots.handlers.Package}}\nfunc Broken( {\n"},
			wantMsg: []string{"framework/web", "framework/web/files/handlers/broken.go.tmpl", "rendered Go code does not format"},
		},
		{
			name:    "template does not parse",
			files:   map[string]string{"framework/web/files/handlers/bad.go.tmpl": "package {{.Slots.handlers.Package\n"},
			wantMsg: []string{"framework/web", "framework/web/files/handlers/bad.go.tmpl"},
		},
		{
			name:    "unknown slot in template data",
			files:   map[string]string{"framework/web/files/handlers/typo.go.tmpl": "package {{.Slots.handler.Package}}\n"},
			wantMsg: []string{"framework/web/files/handlers/typo.go.tmpl", `map has no entry for key "handler"`},
		},
		{
			name:    "notes do not render",
			files:   map[string]string{"data/query/notes.md.tmpl": "{{.Project.Title}}\n"},
			wantMsg: []string{"data/query: notes.md.tmpl", "Title"},
		},
		{
			name: "notes without a README",
			files: map[string]string{
				"architecture/layered/files/root/README.md.tmpl": "",
			},
			wantMsg: []string{"data/query have notes, but no component produces README.md"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cat := loadCatalog(t, withFiles(tt.files))
			_, err := Compose(input(pick(t, cat, "layered", "web", "lite", "query")))
			if !errors.Is(err, ErrTemplate) {
				t.Fatalf("Compose() error = %v, want ErrTemplate", err)
			}
			for _, w := range tt.wantMsg {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("Compose() error = %q\nwant it to contain %q", err, w)
				}
			}
		})
	}
}

func TestComposeAppendsNotesInKindOrder(t *testing.T) {
	t.Parallel()

	cat := loadCatalog(t, withFiles(map[string]string{
		"database/lite/notes.md.tmpl": "## Lite\n\nThe database is a file.\n\n\n",
	}))
	plan, err := Compose(input(pick(t, cat, "layered", "web", "lite", "query")))
	if err != nil {
		t.Fatalf("Compose() error = %v", err)
	}
	readme, _ := plan.File("README.md")
	want := "# shop\n\nModule example.com/shop, Go 1.27.\n\n" +
		"## Lite\n\nThe database is a file.\n\n" +
		"## Query\n\nRegenerate with query shop.\n"
	if string(readme.Content) != want {
		t.Errorf("README.md =\n%q\nwant\n%q", readme.Content, want)
	}
}

// TestComposeSwappingDataLeavesFrameworkUntouched covers the spec scenario
// "Swapping one component leaves others untouched".
func TestComposeSwappingDataLeavesFrameworkUntouched(t *testing.T) {
	t.Parallel()

	cat := loadCatalog(t, baseFixture())
	frameworkFiles := func(data string) map[string][]byte {
		plan, err := Compose(input(pick(t, cat, "layered", "web", "lite", data)))
		if err != nil {
			t.Fatalf("Compose(data=%s) error = %v", data, err)
		}
		out := make(map[string][]byte)
		for _, f := range plan.Files {
			if f.Component == "framework/web" {
				out[f.Path] = f.Content
			}
		}
		return out
	}

	withQuery, withORM := frameworkFiles("query"), frameworkFiles("orm")
	if len(withQuery) == 0 {
		t.Fatal("the framework contributed no files")
	}
	if len(withQuery) != len(withORM) {
		t.Fatalf("framework files differ: %d with query, %d with orm", len(withQuery), len(withORM))
	}
	for p, content := range withQuery {
		if !bytes.Equal(content, withORM[p]) {
			t.Errorf("%s differs between data components:\nquery: %q\norm:   %q", p, content, withORM[p])
		}
	}
}
