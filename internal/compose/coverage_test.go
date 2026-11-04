package compose

import (
	"slices"
	"strings"
	"testing"

	"github.com/Tales-of-zuna/gobento/internal/catalog"
)

// TestCoverage covers the spec scenarios "Backend combinations are
// exhaustive" and "Frontend combinations are covered pairwise" for the
// shipped catalog: Coverage keeps every supported selection without a
// frontend, and its selections with a frontend contain every pair of
// components that some supported selection with a frontend contains.
func TestCoverage(t *testing.T) {
	t.Parallel()

	supported := Supported(shippedCatalog(t))
	covered := Coverage(supported)
	checkCoverage(t, supported, covered)
	if len(covered) >= len(supported) {
		t.Errorf("Coverage() keeps %d of %d selections, want fewer", len(covered), len(supported))
	}
	t.Logf("CI covers %d of %d supported selections", len(covered), len(supported))
	if again := Coverage(supported); !slices.EqualFunc(again, covered, sameSelection) {
		t.Error("Coverage() is not deterministic")
	}
}

func TestCoverageWithoutOptionalKinds(t *testing.T) {
	t.Parallel()

	cat := loadCatalog(t, baseFixture())
	supported := Supported(cat)
	if got := Coverage(supported); !slices.EqualFunc(got, supported, sameSelection) {
		t.Errorf("Coverage() = %v, want every selection %v", got, supported)
	}
}

func TestCoverageFixture(t *testing.T) {
	t.Parallel()

	files := webFixture()
	files["framework/api/component.yaml"] = "name: api\ntitle: API\n"
	files["framework/api/files/handlers/router.go.tmpl"] = files["framework/web/files/handlers/router.go.tmpl"]
	files["frontend/spa/component.yaml"] = "name: spa\ntitle: SPA\n"
	files["frontend/spa/files/web/views.go.tmpl"] = files["frontend/pages/files/web/views.go.tmpl"]
	supported := Supported(loadCatalog(t, files))
	checkCoverage(t, supported, Coverage(supported))
}

// checkCoverage reports selections without an optional component that
// covered lacks, and pairs of components of supported selections with one
// that no selection of covered with one contains.
func checkCoverage(t *testing.T, supported, covered []catalog.Selection) {
	t.Helper()
	want := make(map[pair]bool)
	got := make(map[pair]bool)
	for _, sel := range supported {
		if !hasOptional(sel) {
			if !slices.ContainsFunc(covered, func(s catalog.Selection) bool { return sameSelection(s, sel) }) {
				t.Errorf("Coverage() leaves out %s, which has no optional component", sel)
			}
			continue
		}
		for _, p := range pairs(sel) {
			want[p] = true
		}
	}
	for _, sel := range covered {
		if !slices.ContainsFunc(supported, func(s catalog.Selection) bool { return sameSelection(s, sel) }) {
			t.Errorf("Coverage() returns %s, which is not supported", sel)
		}
		if hasOptional(sel) {
			for _, p := range pairs(sel) {
				got[p] = true
			}
		}
	}
	var missing []string
	for p := range want {
		if !got[p] {
			missing = append(missing, p.a+"+"+p.b)
		}
	}
	if len(missing) > 0 {
		slices.Sort(missing)
		t.Errorf("Coverage() misses the pairs %s", strings.Join(missing, ", "))
	}
}

func sameSelection(a, b catalog.Selection) bool {
	return a.String() == b.String()
}
