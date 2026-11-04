package compose

import (
	"errors"
	"strings"
	"testing"

	"github.com/Tales-of-zuna/gobento/internal/catalog"
	"github.com/Tales-of-zuna/gobento/internal/compat"
)

// baseLayerFiles is a base layer for the fixture catalog: a workflow and a
// Makefile that include contributions, and README notes.
func baseLayerFiles() map[string]string {
	return map[string]string{
		"base/base.yaml": `title: Tooling
contributions:
  ci-services: Services.
  ci-steps: Steps.
  make: Makefile lines.
`,
		"base/notes.md.tmpl": "## Development\n\nLint with {{index .Tools \"lint\"}}.{{with contrib \"make\"}} Make: {{.}}{{end}}\n",
		"base/files/root/dot-github/workflows/ci.yml.tmpl": `jobs:
  check:
    runs-on: {{.Images.runner}}
{{- with contrib "ci-services"}}
    services:
{{indent 6 .}}
{{- end}}
    steps:
      - uses: {{index .Actions "actions/checkout"}}
{{- with contrib "ci-steps"}}
{{indent 6 .}}
{{- end}}
      - run: go build ./{{.Slots.entry.Dir}}
`,
		"base/files/root/Makefile.tmpl": "build:\n{{prefix \"\\t\" \"go build\\n\\ngo vet\"}}\n",
	}
}

// baseInput returns the Input for sel with the fixture's base layer and
// tooling versions.
func baseInput(cat *catalog.Catalog, sel catalog.Selection) Input {
	in := input(sel)
	in.Base = cat.Base()
	in.Tooling = compat.ToolingVersions{
		Images:  map[string]string{"runner": "ubuntu:24.04"},
		Actions: map[string]string{"actions/checkout": "actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1"},
		Tools:   map[string]string{"lint": "v2.15.0"},
	}
	return in
}

func withBase(changes map[string]string) map[string]string {
	files := baseLayerFiles()
	for k, v := range changes {
		files[k] = v
	}
	return withFiles(files)
}

func TestComposeBaseLayer(t *testing.T) {
	t.Parallel()

	cat := loadCatalog(t, withBase(map[string]string{
		// The database contributes a service and a step, the data
		// component a step; steps appear in kind order.
		"database/lite/contrib/ci-services.tmpl": "\n\nlite:\n  image: lite:{{index .Versions \"example.com/driver\"}}\n\n",
		"database/lite/contrib/ci-steps.tmpl":    "- run: lite-check {{index .Traits \"sql-dialect\"}}\n",
		"data/query/contrib/ci-steps.tmpl":       "- name: query\n  run: query diff\n",
		"data/query/contrib/make.tmpl":           "   \n",
	}))
	plan, err := Compose(baseInput(cat, pick(t, cat, "hexagonal", "web", "lite", "query")))
	if err != nil {
		t.Fatalf("Compose() error = %v", err)
	}

	ci, ok := plan.File(".github/workflows/ci.yml")
	if !ok {
		t.Fatalf("no workflow planned; paths: %v", plan.Paths())
	}
	if ci.Component != catalog.BaseID {
		t.Errorf("workflow Component = %q, want %q", ci.Component, catalog.BaseID)
	}
	want := `jobs:
  check:
    runs-on: ubuntu:24.04
    services:
      lite:
        image: lite:v1.2.0
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
      - run: lite-check sqlite
      - name: query
        run: query diff
      - run: go build ./cmd/shop
`
	if string(ci.Content) != want {
		t.Errorf("workflow =\n%s\nwant\n%s", ci.Content, want)
	}

	mk, _ := plan.File("Makefile")
	if got := string(mk.Content); got != "build:\n\tgo build\n\n\tgo vet\n" {
		t.Errorf("Makefile = %q", got)
	}

	// The base layer's notes come after every component's notes, and a
	// contribution that is only whitespace counts as none.
	readme, _ := plan.File("README.md")
	if !strings.HasSuffix(string(readme.Content), "## Query\n\nRegenerate with query shop.\n\n## Development\n\nLint with v2.15.0.\n") {
		t.Errorf("README does not end with the query notes and then the base notes:\n%s", readme.Content)
	}

	var got []string
	for _, k := range plan.Contributions {
		got = append(got, k.Component+" "+k.Point)
	}
	if want := "database/lite ci-services,database/lite ci-steps,data/query ci-steps"; strings.Join(got, ",") != want {
		t.Errorf("Contributions = %v, want %s", got, want)
	}
	if plan.Contributions[0].Content != "lite:\n  image: lite:v1.2.0" {
		t.Errorf("contribution content = %q, want it without surrounding newlines", plan.Contributions[0].Content)
	}
}

// TestComposeBaseLayerWithoutContributions covers the second half of the
// spec scenario "Database contributes a CI service": a selection whose
// components do not contribute gets no services.
func TestComposeBaseLayerWithoutContributions(t *testing.T) {
	t.Parallel()

	cat := loadCatalog(t, withBase(nil))
	plan, err := Compose(baseInput(cat, pick(t, cat, "layered", "web", "lite", "orm")))
	if err != nil {
		t.Fatalf("Compose() error = %v", err)
	}
	ci, _ := plan.File(".github/workflows/ci.yml")
	if strings.Contains(string(ci.Content), "services:") {
		t.Errorf("workflow has services without contributions:\n%s", ci.Content)
	}
	if len(plan.Contributions) != 0 {
		t.Errorf("Contributions = %v, want none", plan.Contributions)
	}
}

func TestComposeBaseLayerErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		changes map[string]string
		wantErr error
		want    []string
	}{
		{
			// Spec scenario "Component writes a base layer file".
			name:    "component produces a base layer file",
			changes: map[string]string{"data/query/files/root/Makefile.tmpl": "all:\n"},
			wantErr: ErrPathConflict,
			want:    []string{"root file Makefile", "data/query and base"},
		},
		{
			// Spec scenario "Base template names an unknown point".
			name:    "unknown point in a base template",
			changes: map[string]string{"base/files/root/Dockerfile.tmpl": "FROM x\n{{contrib \"docker-build\"}}\n"},
			wantErr: ErrTemplate,
			want:    []string{"base/files/root/Dockerfile.tmpl", `no contribution point "docker-build"`},
		},
		{
			name:    "contrib outside the base layer",
			changes: map[string]string{"data/query/files/root/query.yaml.tmpl": "{{contrib \"make\"}}\n"},
			wantErr: ErrTemplate,
			want:    []string{"data/query/files/root/query.yaml.tmpl", `"contrib" not defined`},
		},
		{
			name:    "contribution fails to render",
			changes: map[string]string{"database/lite/contrib/ci-steps.tmpl": "{{.Images.missing}}\n"},
			wantErr: ErrTemplate,
			want:    []string{"database/lite/contrib/ci-steps.tmpl", "missing"},
		},
		{
			name:    "missing tooling version",
			changes: map[string]string{"base/files/root/Dockerfile.tmpl": "FROM {{.Images.go}}\n"},
			wantErr: ErrTemplate,
			want:    []string{"base/files/root/Dockerfile.tmpl", `"go"`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cat := loadCatalog(t, withBase(tt.changes))
			_, err := Compose(baseInput(cat, pick(t, cat, "layered", "web", "lite", "query")))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Compose() error = %v, want %v", err, tt.wantErr)
			}
			for _, w := range tt.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q does not contain %q", err, w)
				}
			}
		})
	}
}

// TestComposeContributionOfAnyKind covers the spec scenario "New kinds
// contribute without base layer changes" as far as the existing kinds go:
// the base layer walks every selected component, whatever its kind, so a
// framework contributes like a database does.
func TestComposeContributionOfAnyKind(t *testing.T) {
	t.Parallel()

	cat := loadCatalog(t, withBase(map[string]string{
		"framework/web/contrib/ci-steps.tmpl": "- run: web generate\n",
	}))
	plan, err := Compose(baseInput(cat, pick(t, cat, "layered", "web", "lite", "orm")))
	if err != nil {
		t.Fatalf("Compose() error = %v", err)
	}
	ci, _ := plan.File(".github/workflows/ci.yml")
	if !strings.Contains(string(ci.Content), "      - run: web generate\n") {
		t.Errorf("workflow lacks the framework's step:\n%s", ci.Content)
	}
}
