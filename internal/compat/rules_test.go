package compat

import (
	"strings"
	"testing"
)

// validRules is a small, consistent rules file used as the base for the
// tests in this package.
const validRules = `
go:
  range: ">=1.26 <1.28"
  default: "1.27"
modules:
  example.com/lib:
    components: [database/lib]
    role: library
    range: ">=v1.2.0 <v2.0.0"
    default: v1.3.0
    constraints:
      - when: ">=v1.3.0"
        require: { go: ">=1.27" }
      - when: ">=v1.2.0"
        require: { example.com/other: ">=v0.5.0" }
  example.com/other:
    components: [data/other]
    role: library
    range: ">=v0.4.0 <v1.0.0"
    default: v0.5.0
  example.com/tool:
    components: [data/other]
    role: tool
    range: "=v1.31.1"
    default: v1.31.1
`

func TestParseRulesValid(t *testing.T) {
	t.Parallel()

	r, err := ParseRules([]byte(validRules))
	if err != nil {
		t.Fatalf("ParseRules() error = %v", err)
	}
	if r.Go.Default != "1.27" || r.Go.Range.String() != ">=1.26 <1.28" {
		t.Errorf("Go = %+v, want default 1.27 and range >=1.26 <1.28", r.Go)
	}
	var paths []string
	for _, m := range r.Modules {
		paths = append(paths, m.Path)
	}
	if got, want := strings.Join(paths, ","), "example.com/lib,example.com/other,example.com/tool"; got != want {
		t.Errorf("module paths = %s, want %s (sorted)", got, want)
	}
	lib, ok := r.Module("example.com/lib")
	if !ok {
		t.Fatal(`Module("example.com/lib") not found`)
	}
	if lib.Role != RoleLibrary || lib.Default != "v1.3.0" || len(lib.Constraints) != 2 {
		t.Errorf("lib = %+v", lib)
	}
	if got := lib.Constraints[0].Requires[0]; got.Target != GoTarget || got.Range.String() != ">=1.27" {
		t.Errorf("first requirement = %+v, want go >=1.27", got)
	}
	if _, ok := r.Module("example.com/missing"); ok {
		t.Error(`Module("example.com/missing") found, want not found`)
	}
}

// TestParseRulesErrors covers one case per error in the spec requirement
// "Rules file is checked for consistency", plus the other structural checks.
func TestParseRulesErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		yaml string
		want []string // substrings the error message must contain
	}{
		{
			name: "default outside range",
			yaml: `
go: { range: ">=1.26", default: "1.27" }
modules:
  example.com/lib:
    components: [database/lib]
    role: library
    range: ">=v2.0.0"
    default: v1.5.0
`,
			want: []string{"module example.com/lib", "default v1.5.0 is outside its range", ">=v2.0.0"},
		},
		{
			name: "malformed module range",
			yaml: `
go: { range: ">=1.26", default: "1.27" }
modules:
  example.com/lib:
    components: [database/lib]
    role: library
    range: "~v1.0.0"
    default: v1.0.0
`,
			want: []string{"module example.com/lib", "range", "invalid range"},
		},
		{
			name: "trait condition without values",
			yaml: `
go: { range: ">=1.26", default: "1.27" }
modules:
  example.com/lib:
    components: [database/lib]
    traits: { sql-dialect: [] }
    role: library
    range: ">=v1.0.0"
    default: v1.0.0
`,
			want: []string{"module example.com/lib", "traits: sql-dialect lists no values"},
		},
		{
			name: "trait condition with an empty value",
			yaml: `
go: { range: ">=1.26", default: "1.27" }
modules:
  example.com/lib:
    components: [database/lib]
    traits: { sql-dialect: [""] }
    role: library
    range: ">=v1.0.0"
    default: v1.0.0
`,
			want: []string{"module example.com/lib", "traits: sql-dialect lists an empty value"},
		},
		{
			name: "malformed go range",
			yaml: `go: { range: ">=v1.26", default: "1.27" }`,
			want: []string{"go: range", "invalid range"},
		},
		{
			name: "go default outside range",
			yaml: `go: { range: ">=1.26 <1.28", default: "1.28" }`,
			want: []string{"go: default 1.28 is outside its range"},
		},
		{
			name: "go default malformed",
			yaml: `go: { range: ">=1.26", default: "latest" }`,
			want: []string{"go: default", "invalid version"},
		},
		{
			name: "unknown constraint target",
			yaml: `
go: { range: ">=1.26", default: "1.27" }
modules:
  example.com/lib:
    components: [database/lib]
    role: library
    range: ">=v1.0.0"
    default: v1.0.0
    constraints:
      - when: ">=v1.0.0"
        require: { example.com/nowhere: ">=v1.0.0" }
`,
			want: []string{"module example.com/lib", "constraint 1", `unknown target "example.com/nowhere"`},
		},
		{
			name: "malformed constraint when",
			yaml: `
go: { range: ">=1.26", default: "1.27" }
modules:
  example.com/lib:
    components: [database/lib]
    role: library
    range: ">=v1.0.0"
    default: v1.0.0
    constraints:
      - when: "v1.0.0"
        require: { go: ">=1.26" }
`,
			want: []string{"module example.com/lib", "constraint 1: when", "invalid range"},
		},
		{
			name: "malformed constraint requirement",
			yaml: `
go: { range: ">=1.26", default: "1.27" }
modules:
  example.com/lib:
    components: [database/lib]
    role: library
    range: ">=v1.0.0"
    default: v1.0.0
    constraints:
      - when: ">=v1.0.0"
        require: { go: ">=v1.26.0" }
`,
			want: []string{"module example.com/lib", "constraint 1: require go", "invalid range"},
		},
		{
			name: "constraint without targets",
			yaml: `
go: { range: ">=1.26", default: "1.27" }
modules:
  example.com/lib:
    components: [database/lib]
    role: library
    range: ">=v1.0.0"
    default: v1.0.0
    constraints:
      - when: ">=v1.0.0"
`,
			want: []string{"module example.com/lib", "require lists no targets"},
		},
		{
			name: "unknown role",
			yaml: `
go: { range: ">=1.26", default: "1.27" }
modules:
  example.com/lib:
    components: [database/lib]
    role: plugin
    range: ">=v1.0.0"
    default: v1.0.0
`,
			want: []string{"module example.com/lib", `role "plugin"`},
		},
		{
			name: "tool range not exact",
			yaml: `
go: { range: ">=1.26", default: "1.27" }
modules:
  example.com/tool:
    components: [data/tool]
    role: tool
    range: ">=v1.0.0"
    default: v1.0.0
`,
			want: []string{"module example.com/tool", "single exact version"},
		},
		{
			name: "no components",
			yaml: `
go: { range: ">=1.26", default: "1.27" }
modules:
  example.com/lib:
    role: library
    range: ">=v1.0.0"
    default: v1.0.0
`,
			want: []string{"module example.com/lib", "no components listed"},
		},
		{
			name: "invalid module path",
			yaml: `
go: { range: ">=1.26", default: "1.27" }
modules:
  "Example.com/Bad Path":
    components: [database/lib]
    role: library
    range: ">=v1.0.0"
    default: v1.0.0
`,
			want: []string{"Example.com/Bad Path", "invalid module path"},
		},
		{
			name: "unknown field",
			yaml: `
go: { range: ">=1.26", default: "1.27" }
modules:
  example.com/lib:
    components: [database/lib]
    role: library
    range: ">=v1.0.0"
    version: v1.0.0
`,
			want: []string{"parse rules", "field version not found"},
		},
		{
			name: "every problem is reported",
			yaml: `
go: { range: ">=1.26", default: "1.25" }
modules:
  example.com/a:
    components: [database/a]
    role: library
    range: ">=v2.0.0"
    default: v1.0.0
  example.com/b:
    components: [data/b]
    role: tool
    range: ">=v1.0.0"
    default: v1.0.0
`,
			want: []string{"go: default 1.25", "module example.com/a: default", "module example.com/b: tool range"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseRules([]byte(tt.yaml))
			if err == nil {
				t.Fatal("ParseRules() error = nil, want an error")
			}
			for _, w := range tt.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("ParseRules() error = %q, want it to contain %q", err, w)
				}
			}
		})
	}
}

func TestCheckComponents(t *testing.T) {
	t.Parallel()

	r, err := ParseRules([]byte(validRules))
	if err != nil {
		t.Fatalf("ParseRules() error = %v", err)
	}

	all := func(string) bool { return true }
	if err := r.CheckComponents(all); err != nil {
		t.Errorf("CheckComponents(all known) error = %v, want nil", err)
	}

	onlyLib := func(id string) bool { return id == "database/lib" }
	err = r.CheckComponents(onlyLib)
	if err == nil {
		t.Fatal("CheckComponents() error = nil, want unknown component errors")
	}
	for _, w := range []string{
		`module example.com/other: unknown component "data/other"`,
		`module example.com/tool: unknown component "data/other"`,
	} {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("CheckComponents() error = %q, want it to contain %q", err, w)
		}
	}
}

// TestCheckTraits covers spec "Modules conditioned on traits" / Unknown
// trait in a condition.
func TestCheckTraits(t *testing.T) {
	t.Parallel()

	r := mustParseRules(t, traitRules)
	if err := r.CheckTraits(func(string) bool { return true }); err != nil {
		t.Errorf("CheckTraits(all known) error = %v, want nil", err)
	}
	err := r.CheckTraits(func(trait string) bool { return trait != "sql-dialect" })
	if err == nil {
		t.Fatal("CheckTraits() error = nil, want an unknown trait error")
	}
	for _, w := range []string{"module example.com/pgdriver", `"sql-dialect"`, "no component provides"} {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("CheckTraits() error = %q, want it to contain %q", err, w)
		}
	}
}
