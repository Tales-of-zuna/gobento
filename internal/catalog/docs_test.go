package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// TestDocsExampleManifest loads the example manifest of docs/components.md,
// the YAML block after the example-manifest marker, so the documented
// schema cannot drift from the loader.
func TestDocsExampleManifest(t *testing.T) {
	t.Parallel()

	doc, err := os.ReadFile(filepath.Join("..", "..", "docs", "components.md"))
	if err != nil {
		t.Fatal(err)
	}
	_, after, ok := strings.Cut(string(doc), "<!-- example-manifest -->\n```yaml\n")
	if !ok {
		t.Fatal("docs/components.md has no example-manifest YAML block")
	}
	manifest, _, ok := strings.Cut(after, "```")
	if !ok {
		t.Fatal("the example manifest block is not closed")
	}

	cat, err := Load(fstest.MapFS{
		"versions.yaml":                   {Data: []byte("go: {range: \">=1.26\", default: \"1.27\"}\nmodules: {}\n")},
		"architecture/mvc/component.yaml": {Data: []byte(manifest)},
	})
	if err != nil {
		t.Fatalf("Load(example manifest) error = %v", err)
	}
	arch, ok := cat.Component(KindArchitecture, "mvc")
	if !ok || len(arch.Slots) != 5 {
		t.Errorf("example manifest = %+v, want an architecture with 5 slots", arch)
	}
}
