package compat

import (
	"errors"
	"strings"
	"testing"
)

// npmRules is validRules with a Node.js rule, a scaffolder, a development
// dependency, and an image tagged with the Node.js release line.
const npmRules = validRules + `
node:
  range: ">=v22.12.0"
  default: "26"
npm:
  create-thing:
    components: [data/other]
    role: scaffolder
    version: 9.3.3
  "@scope/plugin":
    components: [data/other]
    role: dev-dependency
    version: 4.1.0
`

func TestParseRulesNPM(t *testing.T) {
	t.Parallel()

	r, err := ParseRules([]byte(npmRules))
	if err != nil {
		t.Fatalf("ParseRules() error = %v", err)
	}
	if r.Node.Default != "26" || r.Node.Range.String() != ">=v22.12.0" {
		t.Errorf("Node = %+v", r.Node)
	}
	if len(r.NPM) != 2 || r.NPM[0].Name != "@scope/plugin" || r.NPM[1].Name != "create-thing" {
		t.Fatalf("NPM = %+v, want @scope/plugin and create-thing in name order", r.NPM)
	}
	p, ok := r.NPMPackage("create-thing")
	if !ok || p.Role != NPMScaffolder || p.Version != "9.3.3" || p.CreateName() != "thing" {
		t.Errorf("NPMPackage(create-thing) = %+v, %v", p, ok)
	}
	for v, want := range map[string]bool{"v24.22.0": true, "22.12.0": true, "v22.11.9": false, "v20.19.0": false} {
		got, err := r.Node.CheckNodeVersion(v + "\n")
		if err != nil || got != want {
			t.Errorf("CheckNodeVersion(%s) = %v, %v; want %v", v, got, err, want)
		}
	}
	if _, err := r.Node.CheckNodeVersion("nightly"); err == nil {
		t.Error("CheckNodeVersion(nightly) error = nil")
	}
}

func TestResolveNPM(t *testing.T) {
	t.Parallel()

	r, err := ParseRules([]byte(npmRules))
	if err != nil {
		t.Fatal(err)
	}
	res, err := r.Resolve(Request{Components: []string{"data/other"}})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got := res.NPMVersions(); len(got) != 2 || got["create-thing"] != "9.3.3" {
		t.Errorf("NPMVersions() = %v", got)
	}
	res, err = r.Resolve(Request{Components: []string{"database/lib"}})
	if err != nil || len(res.NPM) != 0 {
		t.Errorf("Resolve() without the component: NPM = %v, error = %v", res.NPM, err)
	}

	// Spec scenario "Pin of an npm package".
	_, err = r.Resolve(Request{Components: []string{"data/other"}, Pins: map[string]string{"create-thing": "9.0.0"}})
	if !errors.Is(err, ErrNotPinnable) || !strings.Contains(err.Error(), "npm package versions are fixed") {
		t.Errorf("Resolve() pin error = %v, want ErrNotPinnable", err)
	}
}

func TestNodeImagePlaceholder(t *testing.T) {
	t.Parallel()

	img := Image{Name: "node", Tag: "{node}-alpine"}
	if got := img.Ref("1.27", "26"); got != "node:26-alpine" {
		t.Errorf("Ref() = %q", got)
	}
	_, err := ParseRules([]byte(validRules + "tooling:\n  images:\n    node: {image: node, tag: \"{node}-alpine\"}\n"))
	if err == nil || !strings.Contains(err.Error(), "no node section") {
		t.Errorf("ParseRules() error = %v, want the missing node section", err)
	}
	if _, err := ParseRules([]byte(npmRules + "tooling:\n  images:\n    node: {image: node, tag: \"{node}-alpine\"}\n")); err != nil {
		t.Errorf("ParseRules() with a node section error = %v", err)
	}
}

// TestParseRulesNPMErrors covers the spec scenarios "Scaffolder name" and
// "Default Node.js line inside the range", and the other npm checks.
func TestParseRulesNPMErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, rules, want string
	}{
		{"scaffolder name", validRules + "node: {range: \">=v22.12.0\", default: \"26\"}\nnpm:\n  vite: {components: [data/other], role: scaffolder, version: 9.3.3}\n", `starts with "create-"`},
		{"no components", validRules + "node: {range: \">=v22.12.0\", default: \"26\"}\nnpm:\n  create-x: {role: scaffolder, version: 9.3.3}\n", "no components listed"},
		{"role", validRules + "node: {range: \">=v22.12.0\", default: \"26\"}\nnpm:\n  create-x: {components: [data/other], role: tool, version: 9.3.3}\n", `role "tool"`},
		{"version", validRules + "node: {range: \">=v22.12.0\", default: \"26\"}\nnpm:\n  create-x: {components: [data/other], role: scaffolder, version: ^9.3.3}\n", "exact version"},
		{"node required", validRules + "npm:\n  create-x: {components: [data/other], role: scaffolder, version: 9.3.3}\n", "Node.js range and default are required"},
		{"default outside range", validRules + "node: {range: \">=v22.12.0\", default: \"20\"}\n", "default 20 is outside its range"},
		{"default form", validRules + "node: {range: \">=v22.12.0\", default: \"v26\"}\n", "release line such as 26"},
		{"range", validRules + "node: {range: \">=22.12\", default: \"26\"}\n", "node: range"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseRules([]byte(tt.rules))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("ParseRules() error = %v, want it to contain %q", err, tt.want)
			}
		})
	}
}
