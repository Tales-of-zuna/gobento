package compat

import (
	"errors"
	"strings"
	"testing"
)

// assetRules is validRules with one vendored asset.
const assetRules = validRules + `
assets:
  lib.js:
    components: [data/other]
    version: 2.0.11
    source: https://registry.example.com/lib.js/-/lib.js-2.0.11.tgz
    path: package/dist/lib.min.js
    sha256: d6fdc75f204e6bdefa99b69bf1e6d4ac69b8a364f77929f45c13476b4000f717
`

func TestParseRulesAssets(t *testing.T) {
	t.Parallel()

	r, err := ParseRules([]byte(assetRules))
	if err != nil {
		t.Fatalf("ParseRules() error = %v", err)
	}
	a, ok := r.Asset("lib.js")
	if !ok {
		t.Fatal(`Asset("lib.js") not found`)
	}
	want := Asset{
		Name:       "lib.js",
		Components: []string{"data/other"},
		Version:    "2.0.11",
		Source:     "https://registry.example.com/lib.js/-/lib.js-2.0.11.tgz",
		Path:       "package/dist/lib.min.js",
		SHA256:     "d6fdc75f204e6bdefa99b69bf1e6d4ac69b8a364f77929f45c13476b4000f717",
	}
	if a.Name != want.Name || a.Version != want.Version || a.Source != want.Source || a.Path != want.Path ||
		a.SHA256 != want.SHA256 || strings.Join(a.Components, ",") != "data/other" {
		t.Errorf("Asset() = %+v, want %+v", a, want)
	}
	if _, ok := r.Asset("missing"); ok {
		t.Error(`Asset("missing") found, want not found`)
	}
}

// TestParseRulesAssetErrors covers the spec scenario "Malformed checksum" and
// the other asset checks.
func TestParseRulesAssetErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		asset string
		want  string
	}{
		{"malformed checksum", "{components: [data/other], version: 1.0.0, source: 'https://x.example/a.tgz', path: a.js, sha256: abc}", "asset a.js: sha256"},
		{"uppercase checksum", "{components: [data/other], version: 1.0.0, source: 'https://x.example/a.tgz', path: a.js, sha256: " + strings.Repeat("A", 64) + "}", "asset a.js: sha256"},
		{"no components", "{version: 1.0.0, source: 'https://x.example/a.tgz', path: a.js, sha256: " + strings.Repeat("a", 64) + "}", "asset a.js: no components listed"},
		{"bad version", "{components: [data/other], version: latest, source: 'https://x.example/a.tgz', path: a.js, sha256: " + strings.Repeat("a", 64) + "}", `asset a.js: version "latest"`},
		{"http source", "{components: [data/other], version: 1.0.0, source: 'http://x.example/a.tgz', path: a.js, sha256: " + strings.Repeat("a", 64) + "}", "asset a.js: source"},
		{"escaping path", "{components: [data/other], version: 1.0.0, source: 'https://x.example/a.tgz', path: ../a.js, sha256: " + strings.Repeat("a", 64) + "}", "asset a.js: path"},
		{"unknown field", "{components: [data/other], version: 1.0.0, source: 'https://x.example/a.tgz', path: a.js, sha256: " + strings.Repeat("a", 64) + ", license: MIT}", "field license not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseRules([]byte(validRules + "assets:\n  a.js: " + tt.asset + "\n"))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("ParseRules() error = %v, want it to contain %q", err, tt.want)
			}
		})
	}
}

func TestParseRulesAssetNamedLikeModule(t *testing.T) {
	t.Parallel()

	_, err := ParseRules([]byte(validRules + "assets:\n  example.com/lib: {components: [data/other], version: 1.0.0, source: 'https://x.example/a.tgz', path: a.js, sha256: " + strings.Repeat("a", 64) + "}\n"))
	if err == nil || !strings.Contains(err.Error(), "asset example.com/lib: the name is also a module") {
		t.Errorf("ParseRules() error = %v, want a name clash", err)
	}
}

func TestCheckComponentsAssets(t *testing.T) {
	t.Parallel()

	r, err := ParseRules([]byte(assetRules))
	if err != nil {
		t.Fatal(err)
	}
	err = r.CheckComponents(func(id string) bool { return id != "data/other" })
	if err == nil || !strings.Contains(err.Error(), `asset lib.js: unknown component "data/other"`) {
		t.Errorf("CheckComponents() error = %v, want the asset named", err)
	}
}

func TestResolveAssets(t *testing.T) {
	t.Parallel()

	r, err := ParseRules([]byte(assetRules))
	if err != nil {
		t.Fatal(err)
	}
	res, err := r.Resolve(Request{Components: []string{"database/lib", "data/other"}})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got := res.AssetVersions(); len(got) != 1 || got["lib.js"] != "2.0.11" {
		t.Errorf("AssetVersions() = %v, want lib.js 2.0.11", got)
	}

	res, err = r.Resolve(Request{Components: []string{"database/lib"}})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(res.Assets) != 0 {
		t.Errorf("Assets = %v for a selection without data/other, want none", res.Assets)
	}
}

// TestResolvePinAsset covers the spec scenario "Override of a vendored
// asset".
func TestResolvePinAsset(t *testing.T) {
	t.Parallel()

	r, err := ParseRules([]byte(assetRules))
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.Resolve(Request{Components: []string{"database/lib", "data/other"}, Pins: map[string]string{"lib.js": "2.0.10"}})
	if !errors.Is(err, ErrNotPinnable) {
		t.Fatalf("Resolve() error = %v, want ErrNotPinnable", err)
	}
	for _, want := range []string{"lib.js", "vendored asset", "2.0.11"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Resolve() error = %q, want it to contain %q", err, want)
		}
	}
}
