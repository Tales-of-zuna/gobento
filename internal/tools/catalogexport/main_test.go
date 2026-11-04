package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// websiteCatalog is the committed export the website reads.
const websiteCatalog = "../../../website/src/generated/catalog.json"

// TestCommittedCatalogIsCurrent fails when the catalog or the exporter
// changed without regenerating the website's copy. Fix it with
// `make catalog`.
func TestCommittedCatalogIsCurrent(t *testing.T) {
	t.Parallel()

	var got bytes.Buffer
	if err := run(nil, &got); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	want, err := os.ReadFile(websiteCatalog)
	if err != nil {
		t.Fatalf("read the committed catalog: %v", err)
	}
	if !bytes.Equal(got.Bytes(), want) {
		t.Errorf("%s is out of date; run `make catalog` and commit the result", websiteCatalog)
	}
}

func TestRunWritesFile(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	if err := run(nil, &stdout); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	name := filepath.Join(t.TempDir(), "catalog.json")
	if err := os.WriteFile(name, []byte("stale"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-o", name}, nil); err != nil {
		t.Fatalf("run(-o) error = %v", err)
	}
	got, err := os.ReadFile(filepath.Clean(name))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, stdout.Bytes()) {
		t.Error("the file written with -o differs from the stdout output")
	}
	if entries, _ := os.ReadDir(filepath.Dir(name)); len(entries) != 1 {
		t.Errorf("directory holds %d entries, want only catalog.json", len(entries))
	}

	if err := run([]string{"extra"}, nil); err == nil {
		t.Error("run(extra) error = nil, want an error")
	}
}
