package catalog

import (
	"maps"
	"strings"
	"testing"
)

// TestCandidatesExcludesTraitMismatch covers the spec scenario "Trait
// mismatch is not supported": data/pgonly requires the postgres dialect,
// which no database in the fixture catalog provides.
func TestCandidatesExcludesTraitMismatch(t *testing.T) {
	t.Parallel()

	cat := loadFixture(t, "catalog")
	var got []string
	for _, s := range cat.Candidates() {
		got = append(got, s.String())
	}
	if want := "layered/web/lite/query"; strings.Join(got, ",") != want {
		t.Errorf("Candidates() = %v, want [%s]", got, want)
	}
}

func TestCandidatesCrossProduct(t *testing.T) {
	t.Parallel()

	comp := func(kind Kind, name string, provides map[string]string) *Component {
		return &Component{Kind: kind, Name: name, Provides: provides}
	}
	cat := &Catalog{components: []*Component{
		comp(KindArchitecture, "a1", nil),
		comp(KindArchitecture, "a2", nil),
		comp(KindFramework, "f1", nil),
		comp(KindDatabase, "d1", nil),
		comp(KindDatabase, "d2", nil),
		comp(KindData, "x1", nil),
	}}
	var got []string
	for _, s := range cat.Candidates() {
		got = append(got, s.String())
	}
	want := "a1/f1/d1/x1,a1/f1/d2/x1,a2/f1/d1/x1,a2/f1/d2/x1"
	if strings.Join(got, ",") != want {
		t.Errorf("Candidates() = %v, want %s", got, want)
	}

	cat.components = cat.components[:5] // no data component
	if s := cat.Candidates(); len(s) != 0 {
		t.Errorf("Candidates() without a data component = %v, want none", s)
	}
}

// TestCandidatesOptionalKind covers the spec scenario "Optional kind left
// out": every combination of the required kinds is a candidate without a
// frontend and with each frontend, "none" first.
func TestCandidatesOptionalKind(t *testing.T) {
	t.Parallel()

	comp := func(kind Kind, name string) *Component { return &Component{Kind: kind, Name: name} }
	cat := &Catalog{components: []*Component{
		comp(KindArchitecture, "a1"),
		comp(KindFramework, "f1"),
		comp(KindDatabase, "d1"),
		comp(KindDatabase, "d2"),
		comp(KindData, "x1"),
		comp(KindFrontend, "u1"),
		comp(KindFrontend, "u2"),
	}}
	var got []string
	for _, s := range cat.Candidates() {
		got = append(got, s.String())
	}
	want := "a1/f1/d1/x1,a1/f1/d1/x1/u1,a1/f1/d1/x1/u2,a1/f1/d2/x1,a1/f1/d2/x1/u1,a1/f1/d2/x1/u2"
	if strings.Join(got, ",") != want {
		t.Errorf("Candidates() = %v, want %s", got, want)
	}
}

func TestSelectionOptionalKind(t *testing.T) {
	t.Parallel()

	for _, k := range []Kind{KindFrontend, KindCSS} {
		if !k.Optional() {
			t.Errorf("%s.Optional() = false, want true", k)
		}
	}
	for _, k := range []Kind{KindArchitecture, KindFramework, KindDatabase, KindData} {
		if k.Optional() {
			t.Errorf("%s.Optional() = true, want false", k)
		}
	}

	ui := &Component{Kind: KindFrontend, Name: "htmx"}
	sel := Selection{}.With(&Component{Kind: KindArchitecture, Name: "mvc"})
	if got := sel.String(); got != "mvc/-/-/-" {
		t.Errorf("String() = %q, want mvc/-/-/- (an optional kind is left out)", got)
	}
	if sel.Component(KindFrontend) != nil {
		t.Error("Component(frontend) != nil without a frontend")
	}
	sel = sel.With(ui)
	if got := sel.String(); got != "mvc/-/-/-/htmx" {
		t.Errorf("String() = %q, want mvc/-/-/-/htmx", got)
	}
	if sel.Component(KindFrontend) != ui || sel.Frontend != ui {
		t.Error("With(frontend) did not set the frontend")
	}
	if n := len(sel.Components()); n != 2 {
		t.Errorf("Components() has %d entries, want 2", n)
	}
	css := &Component{Kind: KindCSS, Name: "tailwind"}
	sel = sel.With(css)
	if got := sel.String(); got != "mvc/-/-/-/htmx/tailwind" {
		t.Errorf("String() = %q, want mvc/-/-/-/htmx/tailwind", got)
	}
	if sel.Component(KindCSS) != css || sel.CSS != css || sel.Components()[2] != css {
		t.Error("With(css) did not set the CSS framework last in kind order")
	}
}

func TestSelectionCheckTraits(t *testing.T) {
	t.Parallel()

	sqlite := &Component{Kind: KindDatabase, Name: "sqlite", Provides: map[string]string{
		"sql-dialect": "sqlite",
		"conn":        "database/sql",
	}}

	tests := []struct {
		name    string
		sel     Selection
		wantErr []string
	}{
		{
			name: "met",
			sel: Selection{Database: sqlite, Data: &Component{Kind: KindData, Name: "sqlc", Requires: map[string][]string{
				"sql-dialect": {"sqlite", "postgres"},
				"conn":        {"database/sql"},
			}}},
		},
		{
			name: "value not accepted",
			sel: Selection{Database: sqlite, Data: &Component{Kind: KindData, Name: "pg", Requires: map[string][]string{
				"sql-dialect": {"postgres"},
			}}},
			wantErr: []string{"trait sql-dialect: data/pg requires one of [postgres], but database/sqlite provides \"sqlite\""},
		},
		{
			name: "not provided",
			sel: Selection{Database: sqlite, Data: &Component{Kind: KindData, Name: "kv", Requires: map[string][]string{
				"kv-store": {"redis"},
			}}},
			wantErr: []string{"trait kv-store: data/kv requires one of [redis], but no selected component provides it"},
		},
		{
			name: "conflicting providers",
			sel: Selection{
				Framework: &Component{Kind: KindFramework, Name: "f", Provides: map[string]string{"conn": "pgx"}},
				Database:  sqlite,
			},
			wantErr: []string{`trait conn: framework/f provides "pgx", but database/sqlite provides "database/sql"`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.sel.CheckTraits()
			if len(tt.wantErr) == 0 {
				if err != nil {
					t.Fatalf("CheckTraits() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatal("CheckTraits() error = nil, want an error")
			}
			for _, w := range tt.wantErr {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("CheckTraits() error = %q\nwant it to contain %q", err, w)
				}
			}
		})
	}
}

func TestSelectionHelpers(t *testing.T) {
	t.Parallel()

	s := Selection{
		Architecture: &Component{Kind: KindArchitecture, Name: "mvc"},
		Database:     &Component{Kind: KindDatabase, Name: "sqlite", Provides: map[string]string{"conn": "database/sql"}},
	}
	if got := s.String(); got != "mvc/-/sqlite/-" {
		t.Errorf("String() = %q, want mvc/-/sqlite/-", got)
	}
	if got := s.Components(); len(got) != 2 || got[0].Name != "mvc" || got[1].Name != "sqlite" {
		t.Errorf("Components() = %v", got)
	}
	if got := s.Traits(); !maps.Equal(got, map[string]string{"conn": "database/sql"}) {
		t.Errorf("Traits() = %v", got)
	}
}
