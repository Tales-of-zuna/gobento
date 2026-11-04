package compose

import (
	"github.com/Tales-of-zuna/gobento/internal/catalog"
	"github.com/Tales-of-zuna/gobento/internal/compat"
)

// Supported returns the selections of cat that are supported: one component
// per required kind and at most one per optional kind, all required traits
// met, default versions that resolve, and a
// composition without errors. The result is derived from the catalog alone,
// in the order of catalog.Candidates.
func Supported(cat *catalog.Catalog) []catalog.Selection {
	var out []catalog.Selection
	for _, sel := range cat.Candidates() {
		if Check(cat, sel) == nil {
			out = append(out, sel)
		}
	}
	return out
}

// Check reports why sel is not supported by cat, or nil if it is. It
// resolves the default versions and composes sel for a probe project.
func Check(cat *catalog.Catalog, sel catalog.Selection) error {
	var ids []string
	for _, c := range sel.Components() {
		ids = append(ids, c.ID())
	}
	res, err := cat.Rules().Resolve(compat.Request{Components: ids, Traits: sel.Traits()})
	if err != nil {
		return err
	}
	goVersion, err := compat.GoDirective(res.Go)
	if err != nil {
		return err
	}
	// Composition does not depend on the project beyond names that are
	// valid for any project, so a fixed, simple one is representative.
	project := Project{Name: "app", Module: "example.com/app", GoVersion: goVersion}
	_, err = Compose(Input{Project: project, Selection: sel, Versions: res.Versions(), Assets: res.AssetVersions(), Binaries: res.BinaryVersions(), Base: cat.Base(), Tooling: res.Tooling})
	return err
}
