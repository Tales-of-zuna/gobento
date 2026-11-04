package catalog

import (
	"maps"
	"strings"
	"testing"
	"testing/fstest"
)

// baseCatalog returns a minimal valid catalog with a base layer and a
// database component that contributes to one of its points. Tests add or,
// with empty values, remove files.
func baseCatalog(changes map[string]string) fstest.MapFS {
	files := map[string]string{
		"versions.yaml": "go: {range: \">=1.26\", default: \"1.27\"}\nmodules: {}\n",
		"architecture/mvc/component.yaml": `name: mvc
title: MVC
slots:
  entry: { dir: "cmd/{{.Project.Name}}", package: main, filled_by: architecture }
`,
		"base/base.yaml": `title: Project tooling
contributions:
  ci-services: Service containers of the CI job.
  docker-build: Instructions in the build stage.
`,
		"base/notes.md.tmpl":                               "## Development\n",
		"base/files/root/Makefile.tmpl":                    "build:\n",
		"base/files/root/dot-github/workflows/ci.yml.tmpl": "name: CI\n",
		"database/pg/component.yaml":                       "name: pg\ntitle: PG\n",
		"database/pg/contrib/ci-services.tmpl":             "postgres:\n  image: postgres\n",
	}
	maps.Copy(files, changes)
	fsys := fstest.MapFS{}
	for name, data := range files {
		if data != "" {
			fsys[name] = &fstest.MapFile{Data: []byte(data)}
		}
	}
	return fsys
}

func TestLoadBaseLayer(t *testing.T) {
	t.Parallel()

	cat, err := Load(baseCatalog(nil))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	base := cat.Base()
	if base == nil {
		t.Fatal("Base() = nil")
	}
	if base.ID() != BaseID || !base.IsBase() || base.Title != "Project tooling" || base.Notes != "## Development\n" {
		t.Errorf("Base() = %+v", base)
	}
	var paths []string
	for _, f := range base.Files {
		if f.Slot != RootSlot {
			t.Errorf("%s: Slot = %q, want %q", f.Source, f.Slot, RootSlot)
		}
		paths = append(paths, f.Path)
	}
	if got, want := strings.Join(paths, ","), ".github/workflows/ci.yml,Makefile"; got != want {
		t.Errorf("base paths = %s, want %s", got, want)
	}
	if len(base.Points) != 2 || base.Points[0].Name != "ci-services" || base.Points[1].Description != "Instructions in the build stage." {
		t.Errorf("Points = %+v", base.Points)
	}
	if _, ok := base.Point("docker-build"); !ok {
		t.Error(`Point("docker-build") not found`)
	}

	pg, _ := cat.Component(KindDatabase, "pg")
	if got := pg.Contributions["ci-services"]; got != "postgres:\n  image: postgres\n" {
		t.Errorf("pg contribution = %q", got)
	}
	// The base layer is not a component of any kind.
	for _, k := range Kinds() {
		for _, c := range cat.Components(k) {
			if c.IsBase() {
				t.Errorf("Components(%s) contains the base layer", k)
			}
		}
	}
}

func TestLoadWithoutBaseLayer(t *testing.T) {
	t.Parallel()

	cat, err := Load(baseCatalog(map[string]string{
		"base/base.yaml": "", "base/notes.md.tmpl": "", "base/files/root/Makefile.tmpl": "",
		"base/files/root/dot-github/workflows/ci.yml.tmpl": "", "database/pg/contrib/ci-services.tmpl": "",
	}))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cat.Base() != nil {
		t.Errorf("Base() = %+v, want nil", cat.Base())
	}
}

func TestLoadBaseLayerErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		changes map[string]string
		want    []string
	}{
		{
			// Spec scenario "Unknown contribution point".
			name:    "undeclared point",
			changes: map[string]string{"database/pg/contrib/ci-steps.tmpl": "- run: make\n"},
			want:    []string{"database/pg/contrib/ci-steps.tmpl", `no contribution point "ci-steps"`, "declared points: ci-services, docker-build"},
		},
		{
			name: "contribution without a base layer",
			changes: map[string]string{
				"base/base.yaml": "", "base/notes.md.tmpl": "", "base/files/root/Makefile.tmpl": "",
				"base/files/root/dot-github/workflows/ci.yml.tmpl": "",
			},
			want: []string{"database/pg/contrib/ci-services.tmpl", "no base layer"},
		},
		{
			name:    "slot files in the base layer",
			changes: map[string]string{"base/files/entry/main.go.tmpl": "package main\n"},
			want:    []string{"base/files/entry/main.go.tmpl", "no slots"},
		},
		{
			name:    "unexpected entry",
			changes: map[string]string{"base/component.yaml": "name: base\n"},
			want:    []string{"base", `unexpected entry "component.yaml"`},
		},
		{
			name:    "missing manifest",
			changes: map[string]string{"base/base.yaml": ""},
			want:    []string{"base", "base.yaml"},
		},
		{
			name:    "unknown manifest field",
			changes: map[string]string{"base/base.yaml": "title: T\nslots: {}\n"},
			want:    []string{"base", "slots"},
		},
		{
			name:    "point without description",
			changes: map[string]string{"base/base.yaml": "title: T\ncontributions:\n  ci-services: \"\"\n"},
			want:    []string{"point ci-services needs a description"},
		},
		{
			name:    "point name",
			changes: map[string]string{"base/base.yaml": "title: T\ncontributions:\n  CI: Services.\n"},
			want:    []string{`point name "CI"`},
		},
		{
			name:    "missing title",
			changes: map[string]string{"base/base.yaml": "contributions:\n  ci-services: Services.\n"},
			want:    []string{"base: title is required"},
		},
		{
			name:    "contribution that is not a template",
			changes: map[string]string{"database/pg/contrib/ci-services.yml": "x\n"},
			want:    []string{"database/pg/contrib/ci-services.yml", "<point>.tmpl"},
		},
		{
			name:    "contribution in a subdirectory",
			changes: map[string]string{"database/pg/contrib/ci/services.tmpl": "x\n"},
			want:    []string{"database/pg/contrib/ci", "<point>.tmpl"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := Load(baseCatalog(tt.changes))
			if err == nil {
				t.Fatal("Load() error = nil")
			}
			for _, w := range tt.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q does not contain %q", err, w)
				}
			}
		})
	}
}
