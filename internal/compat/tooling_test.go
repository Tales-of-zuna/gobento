package compat

import (
	"maps"
	"strings"
	"testing"
)

const validTooling = `
tooling:
  images:
    go: { image: golang, tag: "{go}" }
    runtime:
      image: gcr.io/distroless/static-debian13
      tag: nonroot
      digest: sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329
    db: { image: localhost:5000/team/postgres, tag: 18-alpine }
  actions:
    actions/checkout: { version: v7.0.1, sha: 3d3c42e5aac5ba805825da76410c181273ba90b1 }
    github/codeql-action/init: { version: v4.1.0, sha: 0123456789abcdef0123456789abcdef01234567 }
  tools:
    golangci-lint: v2.15.0
`

func TestParseRulesTooling(t *testing.T) {
	t.Parallel()

	r, err := ParseRules([]byte(validRules + validTooling))
	if err != nil {
		t.Fatalf("ParseRules() error = %v", err)
	}
	res, err := r.Resolve(Request{Components: []string{"data/other"}, Go: "1.26"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	want := ToolingVersions{
		Images: map[string]string{
			// Spec scenario "Image tag follows the Go version".
			"go":      "golang:1.26",
			"runtime": "gcr.io/distroless/static-debian13:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329",
			"db":      "localhost:5000/team/postgres:18-alpine",
		},
		Actions: map[string]string{
			"actions/checkout":          "actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1",
			"github/codeql-action/init": "github/codeql-action/init@0123456789abcdef0123456789abcdef01234567 # v4.1.0",
		},
		Tools: map[string]string{"golangci-lint": "v2.15.0"},
	}
	if !maps.Equal(res.Tooling.Images, want.Images) {
		t.Errorf("Images = %v, want %v", res.Tooling.Images, want.Images)
	}
	if !maps.Equal(res.Tooling.Actions, want.Actions) {
		t.Errorf("Actions = %v, want %v", res.Tooling.Actions, want.Actions)
	}
	if !maps.Equal(res.Tooling.Tools, want.Tools) {
		t.Errorf("Tools = %v, want %v", res.Tooling.Tools, want.Tools)
	}

	// The default Go version is used without --go.
	res, err = r.Resolve(Request{Components: []string{"database/lib", "data/other"}})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got := res.Tooling.Images["go"]; got != "golang:1.27" {
		t.Errorf(`Images["go"] = %q, want golang:1.27`, got)
	}
}

func TestParseRulesWithoutTooling(t *testing.T) {
	t.Parallel()

	r, err := ParseRules([]byte(validRules))
	if err != nil {
		t.Fatalf("ParseRules() error = %v", err)
	}
	res, err := r.Resolve(Request{Components: []string{"database/lib", "data/other"}})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(res.Tooling.Images)+len(res.Tooling.Actions)+len(res.Tooling.Tools) != 0 || res.Tooling.Images == nil {
		t.Errorf("Tooling = %+v, want empty, non-nil maps", res.Tooling)
	}
}

func TestParseRulesToolingErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		tooling string
		want    []string
	}{
		{
			name:    "malformed action SHA",
			tooling: "  actions:\n    actions/checkout: { version: v7.0.1, sha: 3d3c42e }\n",
			want:    []string{"action actions/checkout", "40-character commit SHA"},
		},
		{
			name:    "uppercase SHA",
			tooling: "  actions:\n    actions/checkout: { version: v7.0.1, sha: 3D3C42E5AAC5BA805825DA76410C181273BA90B1 }\n",
			want:    []string{"action actions/checkout", "sha"},
		},
		{
			name:    "action version",
			tooling: "  actions:\n    actions/checkout: { version: latest, sha: 3d3c42e5aac5ba805825da76410c181273ba90b1 }\n",
			want:    []string{"action actions/checkout", `version "latest"`},
		},
		{
			name:    "action reference",
			tooling: "  actions:\n    checkout: { version: v7.0.1, sha: 3d3c42e5aac5ba805825da76410c181273ba90b1 }\n",
			want:    []string{"action checkout", "owner/repo"},
		},
		{
			name:    "digest",
			tooling: "  images:\n    runtime: { image: gcr.io/distroless/static, tag: nonroot, digest: sha256:abc }\n",
			want:    []string{"image runtime", "digest"},
		},
		{
			name:    "unknown placeholder",
			tooling: "  images:\n    go: { image: golang, tag: \"{golang}\" }\n",
			want:    []string{"image go", "tag", "the placeholders are {go} and {node}"},
		},
		{
			name:    "empty tag",
			tooling: "  images:\n    go: { image: golang }\n",
			want:    []string{"image go", "tag"},
		},
		{
			name:    "tag in the image name",
			tooling: "  images:\n    go: { image: \"golang:1.27\", tag: \"{go}\" }\n",
			want:    []string{"image go", "without tag or digest"},
		},
		{
			name:    "image key",
			tooling: "  images:\n    go-build: { image: golang, tag: \"{go}\" }\n",
			want:    []string{"image go-build", "key must match"},
		},
		{
			name:    "tool version",
			tooling: "  tools:\n    golangci-lint: v2.15\n",
			want:    []string{"tool golangci-lint", "exact version"},
		},
		{
			name:    "unknown field",
			tooling: "  images:\n    go: { image: golang, tag: \"{go}\", platform: linux }\n",
			want:    []string{"platform"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseRules([]byte(validRules + "tooling:\n" + tt.tooling))
			if err == nil {
				t.Fatal("ParseRules() error = nil")
			}
			for _, w := range tt.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q does not contain %q", err, w)
				}
			}
		})
	}
}
