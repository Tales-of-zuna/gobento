package compat

import (
	"errors"
	"maps"
	"strings"
	"testing"
)

func mustParseRules(t *testing.T, data string) *Rules {
	t.Helper()
	r, err := ParseRules([]byte(data))
	if err != nil {
		t.Fatalf("ParseRules() error = %v", err)
	}
	return r
}

func TestResolve(t *testing.T) {
	t.Parallel()

	rules := mustParseRules(t, validRules)
	both := []string{"database/lib", "data/other"}

	tests := []struct {
		name         string
		req          Request
		wantGo       string
		wantVersions map[string]string
	}{
		{
			// Spec: Default versions / No version flags.
			name:   "defaults",
			req:    Request{Components: both},
			wantGo: "1.27",
			wantVersions: map[string]string{
				"example.com/lib":   "v1.3.0",
				"example.com/other": "v0.5.0",
				"example.com/tool":  "v1.31.1",
			},
		},
		{
			name:         "only modules of selected components",
			req:          Request{Components: []string{"database/lib"}},
			wantGo:       "1.27",
			wantVersions: map[string]string{"example.com/lib": "v1.3.0"},
		},
		{
			// Spec: User version overrides / Override within range.
			name: "override within range",
			req: Request{
				Components: both,
				Go:         "1.26",
				Pins:       map[string]string{"example.com/lib": "v1.2.5"},
			},
			wantGo: "1.26",
			wantVersions: map[string]string{
				"example.com/lib":   "v1.2.5",
				"example.com/other": "v0.5.0",
				"example.com/tool":  "v1.31.1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := rules.Resolve(tt.req)
			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}
			if got.Go != tt.wantGo {
				t.Errorf("Go = %q, want %q", got.Go, tt.wantGo)
			}
			if v := got.Versions(); !maps.Equal(v, tt.wantVersions) {
				t.Errorf("Versions() = %v, want %v", v, tt.wantVersions)
			}
			for i := 1; i < len(got.Modules); i++ {
				if got.Modules[i-1].Path >= got.Modules[i].Path {
					t.Errorf("Modules not sorted by path: %v", got.Modules)
				}
			}
		})
	}
}

func TestResolveErrors(t *testing.T) {
	t.Parallel()

	rules := mustParseRules(t, validRules)
	both := []string{"database/lib", "data/other"}

	tests := []struct {
		name    string
		req     Request
		wantIs  []error
		wantMsg []string
	}{
		{
			// Spec: Override outside range.
			name:    "pin outside range",
			req:     Request{Components: both, Pins: map[string]string{"example.com/lib": "v2.0.0"}},
			wantIs:  []error{ErrOutOfRange},
			wantMsg: []string{"example.com/lib v2.0.0", `">=v1.2.0 <v2.0.0"`},
		},
		{
			name:    "go outside range",
			req:     Request{Components: both, Go: "1.28"},
			wantIs:  []error{ErrOutOfRange},
			wantMsg: []string{"go 1.28", `">=1.26 <1.28"`},
		},
		{
			name:    "malformed go version",
			req:     Request{Components: both, Go: "go1.27"},
			wantIs:  []error{ErrInvalidVersion},
			wantMsg: []string{"go1.27"},
		},
		{
			name:    "malformed pin version",
			req:     Request{Components: both, Pins: map[string]string{"example.com/lib": "1.2.5"}},
			wantIs:  []error{ErrInvalidVersion},
			wantMsg: []string{"pin example.com/lib"},
		},
		{
			// Spec: Override for a module not in the selection.
			name:    "pin module of an unselected component",
			req:     Request{Components: []string{"data/other"}, Pins: map[string]string{"example.com/lib": "v1.2.5"}},
			wantIs:  []error{ErrNotSelected},
			wantMsg: []string{"example.com/lib"},
		},
		{
			name:    "pin unknown module",
			req:     Request{Components: both, Pins: map[string]string{"example.com/unknown": "v1.0.0"}},
			wantIs:  []error{ErrNotSelected},
			wantMsg: []string{"example.com/unknown"},
		},
		{
			// Spec: Override of a tool version.
			name:    "pin tool",
			req:     Request{Components: both, Pins: map[string]string{"example.com/tool": "v1.31.1"}},
			wantIs:  []error{ErrNotPinnable},
			wantMsg: []string{"example.com/tool", "fixed per gobento release"},
		},
		{
			// Spec: Reject incompatible combinations / Go version too old for a module.
			name:    "go too old for module",
			req:     Request{Components: both, Go: "1.26"},
			wantIs:  []error{ErrConstraint},
			wantMsg: []string{"example.com/lib v1.3.0 requires go >=1.27", "go 1.26 is selected"},
		},
		{
			name: "module too old for module",
			req: Request{Components: both, Pins: map[string]string{
				"example.com/lib":   "v1.2.0",
				"example.com/other": "v0.4.0",
			}},
			wantIs:  []error{ErrConstraint},
			wantMsg: []string{"example.com/lib v1.2.0 requires example.com/other >=v0.5.0", "example.com/other v0.4.0 is selected"},
		},
		{
			name: "every problem is reported",
			req: Request{Components: both, Go: "1.28", Pins: map[string]string{
				"example.com/tool":    "v1.0.0",
				"example.com/unknown": "v1.0.0",
				"example.com/lib":     "v9.0.0",
			}},
			wantIs:  []error{ErrOutOfRange, ErrNotPinnable, ErrNotSelected},
			wantMsg: []string{"go 1.28", "example.com/tool", "example.com/unknown", "example.com/lib v9.0.0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := rules.Resolve(tt.req)
			if err == nil {
				t.Fatal("Resolve() error = nil, want an error")
			}
			for _, target := range tt.wantIs {
				if !errors.Is(err, target) {
					t.Errorf("Resolve() error = %q, want errors.Is(%v)", err, target)
				}
			}
			for _, w := range tt.wantMsg {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("Resolve() error = %q, want it to contain %q", err, w)
				}
			}
		})
	}
}

func TestResolveConstraintOnUnselectedModuleDoesNotApply(t *testing.T) {
	t.Parallel()

	rules := mustParseRules(t, validRules)
	// example.com/lib v1.2.0 requires example.com/other >=v0.5.0, but
	// data/other is not selected, so the requirement does not apply.
	got, err := rules.Resolve(Request{
		Components: []string{"database/lib"},
		Pins:       map[string]string{"example.com/lib": "v1.2.0"},
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if v := got.Versions()["example.com/lib"]; v != "v1.2.0" {
		t.Errorf("example.com/lib = %s, want v1.2.0", v)
	}
}

// traitRules has one data component whose drivers depend on the dialect.
const traitRules = `
go: { range: ">=1.26 <1.28", default: "1.27" }
modules:
  example.com/orm:
    components: [data/orm]
    role: library
    range: ">=v1.0.0 <v2.0.0"
    default: v1.1.0
  example.com/litedriver:
    components: [data/orm]
    traits: { sql-dialect: [sqlite] }
    role: library
    range: ">=v1.0.0 <v2.0.0"
    default: v1.0.0
  example.com/pgdriver:
    components: [data/orm]
    traits: { sql-dialect: [postgresql] }
    role: library
    range: ">=v1.0.0 <v2.0.0"
    default: v1.2.0
    constraints:
      - when: ">=v1.2.0"
        require: { example.com/orm: ">=v1.1.0" }
`

// TestResolveTraitConditions covers spec "Modules conditioned on traits".
func TestResolveTraitConditions(t *testing.T) {
	t.Parallel()

	rules := mustParseRules(t, traitRules)
	orm := []string{"data/orm"}
	tests := []struct {
		name   string
		traits map[string]string
		want   map[string]string
	}{
		{
			// Spec: Dialect-specific driver left out.
			name:   "sqlite",
			traits: map[string]string{"sql-dialect": "sqlite", "conn": "database/sql"},
			want:   map[string]string{"example.com/orm": "v1.1.0", "example.com/litedriver": "v1.0.0"},
		},
		{
			name:   "postgresql",
			traits: map[string]string{"sql-dialect": "postgresql"},
			want:   map[string]string{"example.com/orm": "v1.1.0", "example.com/pgdriver": "v1.2.0"},
		},
		{
			name:   "trait not provided",
			traits: map[string]string{},
			want:   map[string]string{"example.com/orm": "v1.1.0"},
		},
		{
			name:   "traits not known yet",
			traits: nil,
			want: map[string]string{
				"example.com/orm": "v1.1.0", "example.com/litedriver": "v1.0.0", "example.com/pgdriver": "v1.2.0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := rules.Resolve(Request{Components: orm, Traits: tt.traits})
			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}
			if v := got.Versions(); !maps.Equal(v, tt.want) {
				t.Errorf("Versions() = %v, want %v", v, tt.want)
			}
		})
	}

	t.Run("pin of a module whose condition is not met", func(t *testing.T) {
		t.Parallel()

		// Spec: Pin for a module whose trait condition is not met.
		_, err := rules.Resolve(Request{
			Components: orm,
			Traits:     map[string]string{"sql-dialect": "sqlite"},
			Pins:       map[string]string{"example.com/pgdriver": "v1.0.0"},
		})
		if !errors.Is(err, ErrNotSelected) || !strings.Contains(err.Error(), "example.com/pgdriver") {
			t.Errorf("Resolve() error = %v, want ErrNotSelected naming example.com/pgdriver", err)
		}
	})

	t.Run("constraints of unused modules do not apply", func(t *testing.T) {
		t.Parallel()

		_, err := rules.Resolve(Request{
			Components: orm,
			Traits:     map[string]string{"sql-dialect": "sqlite"},
			Pins:       map[string]string{"example.com/orm": "v1.0.0"},
		})
		if err != nil {
			t.Errorf("Resolve() error = %v, want nil", err)
		}
	})
}
