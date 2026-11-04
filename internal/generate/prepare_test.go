package generate

import (
	"errors"
	"strings"
	"testing"

	"github.com/Tales-of-zuna/gobento/components"
	"github.com/Tales-of-zuna/gobento/internal/catalog"
)

func shippedCatalog(t *testing.T) *catalog.Catalog {
	t.Helper()
	cat, err := catalog.Load(components.FS())
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}
	return cat
}

func TestPrepare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		spec       Spec
		wantModule string
		wantGo     string
	}{
		{
			// Spec: Fully specified by flags.
			name: "fully specified",
			spec: Spec{Name: "shop", Module: "example.com/shop", Components: map[catalog.Kind]string{
				catalog.KindArchitecture: "mvc", catalog.KindFramework: "nethttp",
				catalog.KindDatabase: "sqlite", catalog.KindData: "sqlc",
			}},
			wantModule: "example.com/shop",
			wantGo:     "1.27",
		},
		{
			// Spec: Module path defaults to the project name, and
			// Automatic selection of the only option.
			name:       "defaults",
			spec:       Spec{Name: "shop", Components: map[catalog.Kind]string{catalog.KindFramework: "nethttp"}},
			wantModule: "shop",
			wantGo:     "1.27",
		},
		{
			name:       "version overrides",
			spec:       Spec{Name: "my-app", Components: map[catalog.Kind]string{catalog.KindFramework: "nethttp"}, Go: "1.26", Pins: map[string]string{"modernc.org/sqlite": "v1.61.0"}},
			wantModule: "my-app",
			wantGo:     "1.26",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p, err := Prepare(shippedCatalog(t), tt.spec)
			if err != nil {
				t.Fatalf("Prepare() error = %v", err)
			}
			if p.Module != tt.wantModule || p.Versions.Go != tt.wantGo {
				t.Errorf("Prepare() module %q go %q, want %q and %q", p.Module, p.Versions.Go, tt.wantModule, tt.wantGo)
			}
			if got := p.Selection.String(); got != "mvc/nethttp/sqlite/sqlc" {
				t.Errorf("Selection = %s", got)
			}
			if _, ok := p.Plan.File("cmd/" + tt.spec.Name + "/main.go"); !ok {
				t.Errorf("plan has no cmd/%s/main.go: %v", tt.spec.Name, p.Plan.Paths())
			}
			req := p.Request("/tmp/x")
			if req.Module != p.Module || req.GoVersion != tt.wantGo || len(req.Modules) == 0 {
				t.Errorf("Request() = %+v", req)
			}
		})
	}
}

func TestPrepareErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		spec    Spec
		wantIs  []error
		wantMsg []string
	}{
		{
			// Spec: Invalid project name.
			name:    "invalid name",
			spec:    Spec{Name: "My App"},
			wantIs:  []error{ErrInvalidName},
			wantMsg: []string{"lowercase letters, digits, - and _"},
		},
		{
			// Spec: Invalid module path.
			name:    "invalid module",
			spec:    Spec{Name: "shop", Module: "Not A Path"},
			wantIs:  []error{ErrInvalidModule},
			wantMsg: []string{`"Not A Path"`, "example.com/shop"},
		},
		{
			// Spec: Several problems.
			name: "several problems",
			spec: Spec{Name: "shop", Module: "Not A Path", Components: map[catalog.Kind]string{
				catalog.KindFramework: "express",
			}},
			wantIs:  []error{ErrInvalidModule},
			wantMsg: []string{"Not A Path", `unknown framework "express"; supported: chi, gin, nethttp`},
		},
		{
			// Spec: Several options require a choice.
			name:    "several options",
			spec:    Spec{Name: "shop"},
			wantMsg: []string{"choose a framework; supported: chi, gin, nethttp"},
		},
		{
			name:    "version outside range",
			spec:    Spec{Name: "shop", Pins: map[string]string{"modernc.org/sqlite": "v1.0.0"}},
			wantMsg: []string{"modernc.org/sqlite v1.0.0", ">=v1.61.0 <v2.0.0"},
		},
		{
			name:    "name and version problems together",
			spec:    Spec{Name: "Shop", Go: "1.20"},
			wantIs:  []error{ErrInvalidName},
			wantMsg: []string{"go 1.20"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := Prepare(shippedCatalog(t), tt.spec)
			if err == nil {
				t.Fatal("Prepare() error = nil, want an error")
			}
			for _, target := range tt.wantIs {
				if !errors.Is(err, target) {
					t.Errorf("Prepare() error = %q, want errors.Is(%v)", err, target)
				}
			}
			for _, w := range tt.wantMsg {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("Prepare() error = %q\nwant it to contain %q", err, w)
				}
			}
		})
	}
}

func TestSelectionErrorAs(t *testing.T) {
	t.Parallel()

	_, err := Prepare(shippedCatalog(t), Spec{Name: "shop", Components: map[catalog.Kind]string{catalog.KindFramework: "nethttp", catalog.KindDatabase: "postgres"}})
	var se *SelectionError
	if !errors.As(err, &se) {
		t.Fatalf("Prepare() error = %v, want a *SelectionError", err)
	}
	if se.Kind != catalog.KindDatabase || se.Value != "postgres" || len(se.Supported) != 1 || se.Supported[0] != "sqlite" {
		t.Errorf("SelectionError = %+v", se)
	}
}

func TestPrepareReportsVersionsWithSelectionProblems(t *testing.T) {
	t.Parallel()

	_, err := Prepare(shippedCatalog(t), Spec{
		Name:       "shop",
		Components: map[catalog.Kind]string{catalog.KindFramework: "express", catalog.KindDatabase: "postgres"},
		Go:         "1.20",
		Pins:       map[string]string{"modernc.org/sqlite": "v1.0.0"},
	})
	for _, w := range []string{`unknown framework "express"`, `unknown database "postgres"`, "go 1.20", "modernc.org/sqlite v1.0.0"} {
		if err == nil || !strings.Contains(err.Error(), w) {
			t.Errorf("Prepare() error = %v\nwant it to contain %q", err, w)
		}
	}
}

func TestPrepareRejectsWindowsReservedNames(t *testing.T) {
	t.Parallel()

	cat := shippedCatalog(t)
	for _, name := range []string{"con", "prn", "aux", "nul", "com1", "com9", "lpt1", "lpt9"} {
		_, err := Prepare(cat, Spec{Name: name})
		if !errors.Is(err, ErrInvalidName) || !strings.Contains(err.Error(), "reserved on Windows") {
			t.Errorf("Prepare(%q) error = %v, want a reserved-name error", name, err)
		}
	}
	for _, name := range []string{"console", "com0", "com10", "lpt", "auxiliary", "nul-app"} {
		if _, err := Prepare(cat, Spec{Name: name, Components: map[catalog.Kind]string{catalog.KindFramework: "nethttp"}}); err != nil {
			t.Errorf("Prepare(%q) error = %v, want nil", name, err)
		}
	}
}
