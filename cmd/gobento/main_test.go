package main

import (
	"runtime/debug"
	"testing"
)

func TestBuildVersion(t *testing.T) {
	info := func(v string) func() (*debug.BuildInfo, bool) {
		return func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{Main: debug.Module{Path: "github.com/Tales-of-zuna/gobento", Version: v}}, true
		}
	}
	noInfo := func() (*debug.BuildInfo, bool) { return nil, false }

	tests := []struct {
		name    string
		ldflags string
		read    func() (*debug.BuildInfo, bool)
		want    string
	}{
		{"ldflags version wins", "0.1.0", info("v9.9.9"), "0.1.0"},
		{"go install at a tag", "dev", info("v0.1.0"), "0.1.0"},
		{"pseudo-version from a checkout", "dev", info("v0.1.1-0.20261103120000-abcdef123456+dirty"), "0.1.1-0.20261103120000-abcdef123456+dirty"},
		{"devel build", "dev", info("(devel)"), "dev"},
		{"empty module version", "dev", info(""), "dev"},
		{"no build info", "dev", noInfo, "dev"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildVersion(tt.ldflags, tt.read); got != tt.want {
				t.Errorf("buildVersion(%q) = %q, want %q", tt.ldflags, got, tt.want)
			}
		})
	}
}
