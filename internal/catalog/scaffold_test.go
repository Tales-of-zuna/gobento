package catalog

import (
	"strings"
	"testing"
)

// scaffoldRules is a rules file with one scaffolder for frontend/spa.
const scaffoldRules = "go: {range: \">=1.26\", default: \"1.27\"}\nmodules: {}\n" +
	"node: {range: \">=v22.12.0\", default: \"26\"}\n" +
	"npm:\n  create-thing: {components: [frontend/spa], role: scaffolder, version: 1.2.3}\n" +
	"  helper: {components: [frontend/spa], role: dev-dependency, version: 1.0.0}\n"

func TestLoadScaffold(t *testing.T) {
	t.Parallel()

	cat, err := Load(baseCatalog(map[string]string{
		"versions.yaml": scaffoldRules,
		"frontend/spa/component.yaml": "name: spa\ntitle: SPA\n" +
			"scaffold: {with: create-thing, template: spa-ts, dir: frontend, remove: [src/demo.css]}\n",
		"frontend/spa/files/root/frontend/vite.config.ts.tmpl": "export default {}\n",
	}))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	c, _ := cat.Component(KindFrontend, "spa")
	if sc := c.Scaffold; sc == nil || sc.With != "create-thing" || sc.Template != "spa-ts" || sc.Dir != "frontend" ||
		strings.Join(sc.Remove, ",") != "src/demo.css" {
		t.Errorf("Scaffold = %+v", c.Scaffold)
	}
}

// TestLoadScaffoldErrors covers the spec scenario "Unknown scaffolder" and
// the other scaffold checks.
func TestLoadScaffoldErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, scaffold, want string
	}{
		{"unknown", "{with: create-other, template: spa-ts, dir: frontend}", `"create-other" is not a scaffolder`},
		{"not a scaffolder", "{with: helper, template: spa-ts, dir: frontend}", `"helper" is not a scaffolder`},
		{"dir", "{with: create-thing, template: spa-ts, dir: ../web}", "scaffold: dir"},
		{"template", "{with: create-thing, template: Spa, dir: frontend}", "scaffold: template"},
		{"remove", "{with: create-thing, template: spa-ts, dir: frontend, remove: [/etc]}", "scaffold: remove"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := Load(baseCatalog(map[string]string{
				"versions.yaml":               scaffoldRules,
				"frontend/spa/component.yaml": "name: spa\ntitle: SPA\nscaffold: " + tt.scaffold + "\n",
			}))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Load() error = %v, want it to contain %q", err, tt.want)
			}
		})
	}

	_, err := Load(baseCatalog(map[string]string{
		"versions.yaml": strings.Replace(scaffoldRules, "[frontend/spa], role: scaffolder", "[frontend/spa2], role: scaffolder", 1),
		"frontend/spa/component.yaml": "name: spa\ntitle: SPA\n" +
			"scaffold: {with: create-thing, template: spa-ts, dir: frontend}\n",
		"frontend/spa2/component.yaml": "name: spa2\ntitle: SPA 2\n",
	}))
	if err == nil || !strings.Contains(err.Error(), "does not list frontend/spa") {
		t.Errorf("Load() error = %v, want the unlisted component", err)
	}
}
