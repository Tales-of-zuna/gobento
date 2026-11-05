package compat

import (
	"errors"
	"strings"
	"testing"
)

// binaryRules is validRules with one platform binary.
var binaryRules = validRules + `
binaries:
  tool:
    components: [data/other]
    version: 4.3.4
    checksums: https://releases.example.com/v4.3.4/sha256sums.txt
    platforms:
      windows/amd64: { url: "https://releases.example.com/v4.3.4/tool-windows-x64.exe", sha256: ` + strings.Repeat("b", 64) + ` }
      linux/amd64: { url: "https://releases.example.com/v4.3.4/tool-linux-x64", sha256: ` + strings.Repeat("a", 64) + ` }
`

func TestParseRulesBinaries(t *testing.T) {
	t.Parallel()

	r, err := ParseRules([]byte(binaryRules))
	if err != nil {
		t.Fatalf("ParseRules() error = %v", err)
	}
	b, ok := r.Binary("tool")
	if !ok {
		t.Fatal(`Binary("tool") not found`)
	}
	if b.Version != "4.3.4" || b.Checksums != "https://releases.example.com/v4.3.4/sha256sums.txt" || strings.Join(b.Components, ",") != "data/other" {
		t.Errorf("Binary() = %+v", b)
	}
	if len(b.Platforms) != 2 {
		t.Fatalf("Platforms = %+v, want 2", b.Platforms)
	}
	linux := b.Platforms[0]
	if linux.Key() != "linux/amd64" || linux.OS != "linux" || linux.Arch != "amd64" || linux.SHA256 != strings.Repeat("a", 64) {
		t.Errorf("Platforms[0] = %+v, want linux/amd64 first", linux)
	}
	if got := b.Platforms[1].FileName(); got != "tool-windows-x64.exe" {
		t.Errorf("FileName() = %q", got)
	}
	if _, ok := r.Binary("missing"); ok {
		t.Error(`Binary("missing") found`)
	}
}

// TestParseRulesBinaryErrors covers the spec scenarios "URL of another
// version" and "Malformed platform", and the other binary checks.
func TestParseRulesBinaryErrors(t *testing.T) {
	t.Parallel()

	sum := strings.Repeat("a", 64)
	tests := []struct {
		name   string
		binary string
		want   string
	}{
		{"url of another version", "{components: [data/other], version: 4.3.4, platforms: {linux/amd64: {url: 'https://x.example/v4.3.3/tool', sha256: " + sum + "}}}", `binary tool: platform linux/amd64: url "https://x.example/v4.3.3/tool" does not contain the version 4.3.4`},
		{"malformed platform", "{components: [data/other], version: 4.3.4, platforms: {linux-amd64: {url: 'https://x.example/4.3.4/tool', sha256: " + sum + "}}}", `binary tool: platform "linux-amd64" must be GOOS/GOARCH`},
		{"malformed checksum", "{components: [data/other], version: 4.3.4, platforms: {linux/amd64: {url: 'https://x.example/4.3.4/tool', sha256: abc}}}", "binary tool: platform linux/amd64: sha256"},
		{"http url", "{components: [data/other], version: 4.3.4, platforms: {linux/amd64: {url: 'http://x.example/4.3.4/tool', sha256: " + sum + "}}}", "binary tool: platform linux/amd64: url"},
		{"http checksums", "{components: [data/other], version: 4.3.4, checksums: 'http://x.example/sums', platforms: {linux/amd64: {url: 'https://x.example/4.3.4/tool', sha256: " + sum + "}}}", "binary tool: checksums"},
		{"no platforms", "{components: [data/other], version: 4.3.4}", "binary tool: no platforms listed"},
		{"no components", "{version: 4.3.4, platforms: {linux/amd64: {url: 'https://x.example/4.3.4/tool', sha256: " + sum + "}}}", "binary tool: no components listed"},
		{"bad version", "{components: [data/other], version: latest, platforms: {linux/amd64: {url: 'https://x.example/latest/tool', sha256: " + sum + "}}}", `binary tool: version "latest"`},
		{"unknown field", "{components: [data/other], version: 4.3.4, license: MIT, platforms: {linux/amd64: {url: 'https://x.example/4.3.4/tool', sha256: " + sum + "}}}", "field license not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseRules([]byte(validRules + "binaries:\n  tool: " + tt.binary + "\n"))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("ParseRules() error = %v, want it to contain %q", err, tt.want)
			}
		})
	}
}

func TestParseRulesBinaryNamedLikeModule(t *testing.T) {
	t.Parallel()

	_, err := ParseRules([]byte(validRules + "binaries:\n  example.com/lib: {components: [data/other], version: 1.0.0, platforms: {linux/amd64: {url: 'https://x.example/1.0.0/t', sha256: " + strings.Repeat("a", 64) + "}}}\n"))
	if err == nil || !strings.Contains(err.Error(), "binary example.com/lib: the name is also a module or an asset") {
		t.Errorf("ParseRules() error = %v, want a name clash", err)
	}
}

func TestCheckComponentsBinaries(t *testing.T) {
	t.Parallel()

	r, err := ParseRules([]byte(binaryRules))
	if err != nil {
		t.Fatal(err)
	}
	err = r.CheckComponents(func(id string) bool { return id != "data/other" })
	if err == nil || !strings.Contains(err.Error(), `binary tool: unknown component "data/other"`) {
		t.Errorf("CheckComponents() error = %v, want the binary named", err)
	}
}

// TestResolveBinaries covers the spec scenario "Binary visible to
// templates" as far as resolution goes: a selection gets the binaries of
// its components, with every platform.
func TestResolveBinaries(t *testing.T) {
	t.Parallel()

	r, err := ParseRules([]byte(binaryRules))
	if err != nil {
		t.Fatal(err)
	}
	res, err := r.Resolve(Request{Components: []string{"database/lib", "data/other"}})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	got := res.BinaryVersions()
	if len(got) != 1 || got["tool"].Version != "4.3.4" || len(got["tool"].Platforms) != 2 {
		t.Errorf("BinaryVersions() = %+v, want tool 4.3.4 with two platforms", got)
	}

	res, err = r.Resolve(Request{Components: []string{"database/lib"}})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(res.Binaries) != 0 {
		t.Errorf("Binaries = %v for a selection without data/other, want none", res.Binaries)
	}
}

// TestResolvePinBinary covers the spec scenario "Pin for a binary".
func TestResolvePinBinary(t *testing.T) {
	t.Parallel()

	r, err := ParseRules([]byte(binaryRules))
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.Resolve(Request{Components: []string{"database/lib", "data/other"}, Pins: map[string]string{"tool": "4.0.0"}})
	if !errors.Is(err, ErrNotPinnable) {
		t.Fatalf("Resolve() error = %v, want ErrNotPinnable", err)
	}
	for _, want := range []string{"tool", "binary versions are fixed", "4.3.4"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Resolve() error = %q, want it to contain %q", err, want)
		}
	}
}
