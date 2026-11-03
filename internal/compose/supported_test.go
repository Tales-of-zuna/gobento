package compose

import (
	"strings"
	"testing"
)

// TestSupportedRequiresComposition covers spec "Supported combinations": a
// selection is supported only when its traits are met and it composes.
func TestSupportedRequiresComposition(t *testing.T) {
	t.Parallel()

	files := withFiles(map[string]string{
		// data/broken meets every trait but fills no store, a required slot.
		"data/broken/component.yaml":             "name: broken\ntitle: Broken\n",
		"data/broken/files/root/broken.txt.tmpl": "x\n",
		// data/pg requires a dialect no database provides.
		"data/pg/component.yaml":            "name: pg\ntitle: PG\nrequires:\n  sql-dialect: [postgres]\n",
		"data/pg/files/store/store.go.tmpl": "package {{.Slots.store.Package}}\n",
	})
	cat := loadCatalog(t, files)

	var candidates, supported []string
	for _, s := range cat.Candidates() {
		candidates = append(candidates, s.String())
	}
	for _, s := range Supported(cat) {
		supported = append(supported, s.String())
	}

	if strings.Contains(strings.Join(candidates, ","), "/pg") {
		t.Errorf("Candidates() = %v, want no selection with data/pg (trait mismatch)", candidates)
	}
	if !strings.Contains(strings.Join(candidates, ","), "layered/web/lite/broken") {
		t.Errorf("Candidates() = %v, want layered/web/lite/broken (traits are met)", candidates)
	}
	want := "hexagonal/web/lite/orm,hexagonal/web/lite/query,layered/web/lite/orm,layered/web/lite/query"
	if got := strings.Join(supported, ","); got != want {
		t.Errorf("Supported() = %s, want %s", got, want)
	}

	sel := pick(t, cat, "layered", "web", "lite", "broken")
	if err := Check(cat, sel); err == nil || !strings.Contains(err.Error(), "slot store") {
		t.Errorf("Check(broken) error = %v, want an unfilled store slot", err)
	}
}
