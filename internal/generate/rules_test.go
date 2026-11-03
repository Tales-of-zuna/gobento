package generate

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"golang.org/x/mod/module"
)

func TestReservedNames(t *testing.T) {
	t.Parallel()

	names := ReservedNames()
	if len(names) != 22 {
		t.Errorf("len(ReservedNames()) = %d, want 22", len(names))
	}
	for _, want := range []string{"con", "prn", "aux", "nul", "com1", "com9", "lpt1", "lpt9"} {
		if !slices.Contains(names, want) {
			t.Errorf("ReservedNames() lacks %q", want)
		}
	}
	for _, name := range names {
		// A reserved name passes the pattern; the separate reserved-name
		// check is what rejects it.
		if !namePattern.MatchString(name) {
			t.Errorf("reserved name %q does not match NamePattern", name)
		}
	}
	for _, ok := range []string{"com", "com0", "com10", "lpt0", "console", "aux1"} {
		if isWindowsReserved(ok) {
			t.Errorf("isWindowsReserved(%q) = true, want false", ok)
		}
	}
}

// modulePathPattern is the compiled ModulePathPattern.
var modulePathPattern = regexp.MustCompile(ModulePathPattern)

// checkModulePathPattern reports a disagreement between ModulePathPattern
// and module.CheckImportPath that the pattern's documentation does not
// allow: the pattern must accept every valid path, and may accept an invalid
// one only when it is invalid because of Windows naming rules.
func checkModulePathPattern(t *testing.T, path string) {
	t.Helper()
	err := module.CheckImportPath(path)
	matched := modulePathPattern.MatchString(path)
	switch {
	case err == nil && !matched:
		t.Errorf("ModulePathPattern rejects %q, which CheckImportPath accepts", path)
	case err != nil && matched && !windowsOnly(err):
		t.Errorf("ModulePathPattern accepts %q, which CheckImportPath rejects: %v", path, err)
	}
}

// modulePathSeeds covers each rule of CheckImportPath.
var modulePathSeeds = []string{
	"shop", "example.com/shop", "github.com/Tales-of-zuna/gobento", "a", "a/b",
	"gopkg.in/yaml.v3", "example.com/a_b~c+d-e", ".hidden/x", "x/.y", "my-app",
	"", "/", "a/", "/a", "a//b", "-a", "a/-b", "a.", "a/b.", ".", "..", "a/./b",
	"a/../b", "My App", "Not A Path", "a@b", "a b", "a\\b", "é", "a:b",
	"con", "example.com/con", "example.com/con.go", "abc~1", "x/abc~12",
	"abc~", "abc~x", "~", "+", "_", "a.b.c",
}

func TestModulePathPattern(t *testing.T) {
	t.Parallel()

	for _, path := range modulePathSeeds {
		checkModulePathPattern(t, path)
	}
	// The shared pattern must also be valid JavaScript; RE2 and
	// JavaScript agree on this subset (no lookaround, no flags).
	if strings.ContainsAny(ModulePathPattern+NamePattern, "\\") {
		t.Errorf("patterns use escapes, which may differ between Go and JavaScript")
	}
}

func FuzzModulePathPattern(f *testing.F) {
	for _, path := range modulePathSeeds {
		f.Add(path)
	}
	f.Fuzz(checkModulePathPattern)
}

// windowsOnly reports whether err rejects a path only because of Windows
// naming rules: a reserved device name or a short name like abc~1.
func windowsOnly(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "on Windows") || strings.Contains(msg, "trailing tilde and digits")
}
