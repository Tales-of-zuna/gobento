package catalog

import (
	"maps"
	"strings"
	"testing"
)

// TestSupportedExcludesTraitMismatch covers the spec scenario "Trait
// mismatch is not supported": data/pgonly requires the postgres dialect,
// which no database in the fixture catalog provides.
func TestSupportedExcludesTraitMismatch(t *testing.T) {
	t.Parallel()

	cat := loadFixture(t, "catalog")
	var got []string
	for _, s := range cat.Supported() {
		got = append(got, s.String())
	}
	if want := "layered/web/lite/query"; strings.Join(got, ",") != want {
		t.Errorf("Supported() = %v, want [%s]", got, want)
	}
}

func TestSupportedCrossProduct(t *testing.T) {
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
	for _, s := range cat.Supported() {
		got = append(got, s.String())
	}
	want := "a1/f1/d1/x1,a1/f1/d2/x1,a2/f1/d1/x1,a2/f1/d2/x1"
	if strings.Join(got, ",") != want {
		t.Errorf("Supported() = %v, want %s", got, want)
	}

	cat.components = cat.components[:5] // no data component
	if s := cat.Supported(); len(s) != 0 {
		t.Errorf("Supported() without a data component = %v, want none", s)
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
