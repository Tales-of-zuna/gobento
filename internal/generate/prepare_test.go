package generate

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/Tales-of-zuna/gobento/components"
	"github.com/Tales-of-zuna/gobento/internal/catalog"
	"github.com/Tales-of-zuna/gobento/internal/compat"
)

func shippedCatalog(t *testing.T) *catalog.Catalog {
	t.Helper()
	cat, err := catalog.Load(components.FS())
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}
	return cat
}

// reference names every component of the reference selection
// mvc/nethttp/sqlite/sqlc.
var reference = map[catalog.Kind]string{
	catalog.KindArchitecture: "mvc", catalog.KindFramework: "nethttp", catalog.KindDatabase: "sqlite", catalog.KindData: "sqlc",
}

func TestPrepare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		spec       Spec
		wantModule string
		wantGo     string
		wantSel    string
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
			wantSel:    "mvc/nethttp/sqlite/sqlc",
		},
		{
			// Spec: Module path defaults to the project name, and
			// Single option is chosen automatically.
			name:       "defaults",
			spec:       Spec{Name: "shop", Components: reference},
			wantModule: "shop",
			wantGo:     "1.27",
			wantSel:    "mvc/nethttp/sqlite/sqlc",
		},
		{
			name:       "version overrides",
			spec:       Spec{Name: "my-app", Components: reference, Go: "1.26", Pins: map[string]string{"modernc.org/sqlite": "v1.61.0"}},
			wantModule: "my-app",
			wantGo:     "1.26",
			wantSel:    "mvc/nethttp/sqlite/sqlc",
		},
		{
			// Spec: GORM data access.
			name: "gorm",
			spec: Spec{Name: "shop", Components: map[catalog.Kind]string{catalog.KindArchitecture: "mvc", catalog.KindFramework: "nethttp", catalog.KindDatabase: "sqlite", catalog.KindData: "gorm"},
				Pins: map[string]string{"gorm.io/gorm": "v1.31.2", "github.com/libtnb/sqlite": "v1.1.0"}},
			wantModule: "shop",
			wantGo:     "1.27",
			wantSel:    "mvc/nethttp/sqlite/gorm",
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
			if got := p.Selection.String(); got != tt.wantSel {
				t.Errorf("Selection = %s, want %s", got, tt.wantSel)
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
			// Spec: Several options require a choice, for the shipped
			// data-access components.
			name:    "data access not chosen",
			spec:    Spec{Name: "shop"},
			wantMsg: []string{"choose a data-access library; supported: gorm, sqlc"},
		},
		{
			// Spec: Pin for a module whose trait condition is not met.
			name: "pin of the postgres driver with sqlite",
			spec: Spec{Name: "shop", Components: map[catalog.Kind]string{catalog.KindArchitecture: "mvc", catalog.KindFramework: "nethttp", catalog.KindDatabase: "sqlite", catalog.KindData: "gorm"},
				Pins: map[string]string{"gorm.io/driver/postgres": "v1.6.3"}},
			wantIs:  []error{compat.ErrNotSelected},
			wantMsg: []string{"cannot pin gorm.io/driver/postgres"},
		},
		{
			name: "gorm floor below a dialect constraint",
			spec: Spec{Name: "shop", Components: map[catalog.Kind]string{catalog.KindArchitecture: "mvc", catalog.KindFramework: "nethttp", catalog.KindDatabase: "sqlite", catalog.KindData: "gorm"},
				Pins: map[string]string{"gorm.io/gorm": "v1.31.1"}},
			wantIs:  []error{compat.ErrConstraint},
			wantMsg: []string{"github.com/libtnb/sqlite v1.2.2 requires gorm.io/gorm >=v1.31.2"},
		},
		{
			name: "version outside range",
			spec: Spec{
				Name: "shop", Components: reference,
				Pins: map[string]string{"modernc.org/sqlite": "v1.0.0"},
			},
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

	_, err := Prepare(shippedCatalog(t), Spec{Name: "shop", Components: map[catalog.Kind]string{catalog.KindArchitecture: "mvc", catalog.KindFramework: "nethttp", catalog.KindDatabase: "mysql"}})
	var se *SelectionError
	if !errors.As(err, &se) {
		t.Fatalf("Prepare() error = %v, want a *SelectionError", err)
	}
	if se.Kind != catalog.KindDatabase || se.Value != "mysql" || !slices.Equal(se.Supported, []string{"postgres", "sqlite"}) {
		t.Errorf("SelectionError = %+v", se)
	}
}

func TestPrepareReportsVersionsWithSelectionProblems(t *testing.T) {
	t.Parallel()

	_, err := Prepare(shippedCatalog(t), Spec{
		Name:       "shop",
		Components: map[catalog.Kind]string{catalog.KindFramework: "express", catalog.KindDatabase: "mysql"},
		Go:         "1.20",
		Pins:       map[string]string{"modernc.org/sqlite": "v1.0.0"},
	})
	for _, w := range []string{`unknown framework "express"`, `unknown database "mysql"`, "go 1.20", "modernc.org/sqlite v1.0.0"} {
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
		if _, err := Prepare(cat, Spec{Name: name, Components: reference}); err != nil {
			t.Errorf("Prepare(%q) error = %v, want nil", name, err)
		}
	}
}

func TestCheckName(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"shop":   "",
		"my-app": "",
		"a_1":    "",
		"My App": NameRule,
		"":       NameRule,
		"1shop":  NameRule,
		"con":    ReservedNameRule,
		"lpt9":   ReservedNameRule,
	}
	for name, want := range tests {
		err := CheckName(name)
		if want == "" {
			if err != nil {
				t.Errorf("CheckName(%q) = %v, want nil", name, err)
			}
			continue
		}
		if !errors.Is(err, ErrInvalidName) || !strings.Contains(err.Error(), want) {
			t.Errorf("CheckName(%q) = %v, want ErrInvalidName explaining %q", name, err, want)
		}
	}
}

func TestCheckModule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path, name string
		want       string
	}{
		{path: "example.com/shop", name: "shop"},
		{path: "shop", name: "shop"},
		{path: "Not A Path", name: "shop", want: "use a path such as example.com/shop"},
		{path: "con", name: "con", want: "use a path such as example.com/app"},
		{path: "", name: "My App", want: "use a path such as example.com/app"},
	}
	for _, tt := range tests {
		err := CheckModule(tt.path, tt.name)
		if tt.want == "" {
			if err != nil {
				t.Errorf("CheckModule(%q, %q) = %v, want nil", tt.path, tt.name, err)
			}
			continue
		}
		if !errors.Is(err, ErrInvalidModule) || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("CheckModule(%q, %q) = %v, want ErrInvalidModule containing %q", tt.path, tt.name, err, tt.want)
		}
	}
}

// TestPrepareRejectsToolingPins covers the spec scenario "Pin for a tooling
// entry": tooling versions are not modules and cannot be pinned.
func TestPrepareRejectsToolingPins(t *testing.T) {
	t.Parallel()

	cat := shippedCatalog(t)
	_, err := Prepare(cat, Spec{
		Name: "shop",
		Components: map[catalog.Kind]string{
			catalog.KindArchitecture: "mvc", catalog.KindFramework: "nethttp",
			catalog.KindDatabase: "sqlite", catalog.KindData: "sqlc",
		},
		Pins: map[string]string{"golangci-lint": "v2.15.0"},
	})
	if !errors.Is(err, compat.ErrNotSelected) {
		t.Fatalf("Prepare() error = %v, want %v", err, compat.ErrNotSelected)
	}
}
