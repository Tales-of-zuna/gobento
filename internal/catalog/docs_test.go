package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// docsPage is the contributor guide for components on the website.
const docsPage = "../../website/content/docs/contributing/adding-a-component.mdx"

// TestDocsExampleManifest loads the example manifest of the contributor
// guide, the YAML block after the example-manifest marker, so the documented
// schema cannot drift from the loader.
func TestDocsExampleManifest(t *testing.T) {
	t.Parallel()

	doc, err := os.ReadFile(filepath.FromSlash(docsPage))
	if err != nil {
		t.Fatal(err)
	}
	_, after, ok := strings.Cut(string(doc), "{/* example-manifest */}\n\n```yaml")
	if !ok {
		t.Fatalf("%s has no example-manifest YAML block", docsPage)
	}
	// Skip the rest of the fence line, which may carry a title.
	_, after, _ = strings.Cut(after, "\n")
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
