package compat

import (
	"strings"
	"testing"
)

// conditionalRules is validRules with a binary and an npm package that
// data/other uses only for one value of the trait build each.
const conditionalRules = validRules + `
binaries:
  tool:
    components: [data/other]
    traits: {build: [go]}
    version: 4.3.4
    platforms:
      linux/amd64: { url: "https://releases.example.com/v4.3.4/tool-linux-x64", sha256: ` + "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" + ` }
node:
  range: ">=v22.12.0"
  default: "26"
npm:
  "@scope/plugin":
    components: [data/other]
    traits: {build: [vite]}
    role: dev-dependency
    version: 4.3.4
`

// TestResolveBinaryAndNPMTraitConditions covers the spec scenario "Tailwind's npm
// packages from the rules" as far as resolution goes: a binary or an npm
// package with a trait condition is used only when the selection's traits
// meet it, and always when the traits are not known yet.
func TestResolveBinaryAndNPMTraitConditions(t *testing.T) {
	t.Parallel()

	r, err := ParseRules([]byte(conditionalRules))
	if err != nil {
		t.Fatalf("ParseRules() error = %v", err)
	}
	for _, tc := range []struct {
		traits      map[string]string
		binary, pkg bool
	}{
		{traits: map[string]string{"build": "go"}, binary: true},
		{traits: map[string]string{"build": "vite"}, pkg: true},
		{traits: map[string]string{}},
		{traits: nil, binary: true, pkg: true},
	} {
		res, err := r.Resolve(Request{Components: []string{"data/other"}, Traits: tc.traits})
		if err != nil {
			t.Fatalf("Resolve(%v) error = %v", tc.traits, err)
		}
		if got := len(res.Binaries) == 1; got != tc.binary {
			t.Errorf("Resolve(%v) binaries = %v, want the binary: %v", tc.traits, res.Binaries, tc.binary)
		}
		if got := len(res.NPM) == 1; got != tc.pkg {
			t.Errorf("Resolve(%v) npm = %v, want the package: %v", tc.traits, res.NPM, tc.pkg)
		}
	}
	err = r.CheckTraits(func(trait string) bool { return trait != "build" })
	for _, want := range []string{`binary tool: trait condition on "build"`, `npm @scope/plugin: trait condition on "build"`} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("CheckTraits() error = %v, want %q", err, want)
		}
	}
}

// TestParseRulesTraitConditionErrors checks that binaries and npm packages
// reject malformed trait conditions like modules do.
func TestParseRulesTraitConditionErrors(t *testing.T) {
	t.Parallel()

	bad := strings.ReplaceAll(conditionalRules, "traits: {build: [go]}", "traits: {build: []}")
	bad = strings.ReplaceAll(bad, "traits: {build: [vite]}", `traits: {build: [""]}`)
	_, err := ParseRules([]byte(bad))
	for _, want := range []string{"binary tool: traits: build lists no values", "npm @scope/plugin: traits: build lists an empty value"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("ParseRules() error = %v, want %q", err, want)
		}
	}
}
