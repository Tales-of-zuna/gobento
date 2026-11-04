package compat

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/mod/semver"
)

// GoPlaceholder in an image tag stands for the selected Go version, for
// example "{go}" becomes "1.27".
const GoPlaceholder = "{go}"

var (
	// imageKeyPattern matches image keys, which templates use as field
	// names (.Images.<key>) and therefore must be identifiers.
	imageKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	// toolKeyPattern matches tool keys, for example golangci-lint.
	toolKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	// imageNamePattern matches image names without tag or digest, with an
	// optional registry host and port, for example
	// gcr.io/distroless/static-debian13 or localhost:5000/app.
	imageNamePattern = regexp.MustCompile(`^(?:[a-z0-9.-]+(?::[0-9]+)?/)?[a-z0-9]+(?:(?:[._]|__|-+)[a-z0-9]+)*(?:/[a-z0-9]+(?:(?:[._]|__|-+)[a-z0-9]+)*)*$`)
	// imageTagPattern matches image tags, as Docker defines them.
	imageTagPattern = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$`)
	// digestPattern matches sha256 image digests.
	digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	// actionPattern matches action references: owner/repo, optionally with
	// a path to an action inside the repository.
	actionPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)*$`)
	// shaPattern matches full commit SHAs.
	shaPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
)

// Tooling lists the versions of the project tooling that are not Go module
// requirements: container images, GitHub Actions, and tools. They are fixed
// per gobento release and cannot be pinned by users.
type Tooling struct {
	// Images maps image keys, such as runtime, to images.
	Images map[string]Image
	// Actions maps action references, such as actions/checkout, to the
	// pinned release.
	Actions map[string]Action
	// Tools maps tool names, such as golangci-lint, to exact versions.
	Tools map[string]string
}

// Image is a container image.
type Image struct {
	// Name is the image name without tag or digest, for example golang.
	Name string
	// Tag is the tag. GoPlaceholder in it stands for the selected Go
	// version.
	Tag string
	// Digest optionally pins the image, for example sha256:<64 hex digits>.
	Digest string
}

// Ref returns the image reference for the Go version goVersion:
// name:tag, followed by @digest when the image has one.
func (i Image) Ref(goVersion string) string {
	ref := i.Name + ":" + strings.ReplaceAll(i.Tag, GoPlaceholder, goVersion)
	if i.Digest != "" {
		ref += "@" + i.Digest
	}
	return ref
}

// Action is a GitHub Action release, pinned by commit.
type Action struct {
	// Version is the release, for example v7.0.1.
	Version string
	// SHA is the full commit SHA the release tag points to.
	SHA string
}

// ToolingVersions are the tooling versions resolved for one generation, as
// templates see them.
type ToolingVersions struct {
	// Images maps image keys to full references, for example
	// "golang:1.27" or "gcr.io/distroless/static-debian13:nonroot@sha256:...".
	Images map[string]string
	// Actions maps action references to complete values for uses:, for
	// example "actions/checkout@<sha> # v7.0.1".
	Actions map[string]string
	// Tools maps tool names to their versions.
	Tools map[string]string
}

// Resolve returns the tooling versions for the Go version goVersion.
func (t Tooling) Resolve(goVersion string) ToolingVersions {
	out := ToolingVersions{
		Images:  make(map[string]string, len(t.Images)),
		Actions: make(map[string]string, len(t.Actions)),
		Tools:   make(map[string]string, len(t.Tools)),
	}
	for k, img := range t.Images {
		out.Images[k] = img.Ref(goVersion)
	}
	for ref, a := range t.Actions {
		out.Actions[ref] = ref + "@" + a.SHA + " # " + a.Version
	}
	maps.Copy(out.Tools, t.Tools)
	return out
}

// toolingFile mirrors the YAML layout of the rules file's tooling section.
type toolingFile struct {
	Images  map[string]imageFile  `yaml:"images"`
	Actions map[string]actionFile `yaml:"actions"`
	Tools   map[string]string     `yaml:"tools"`
}

type imageFile struct {
	Image  string `yaml:"image"`
	Tag    string `yaml:"tag"`
	Digest string `yaml:"digest"`
}

type actionFile struct {
	Version string `yaml:"version"`
	SHA     string `yaml:"sha"`
}

// parseTooling converts and validates the tooling section.
func parseTooling(tf toolingFile) (Tooling, []error) {
	var errs []error
	fail := func(format string, args ...any) {
		errs = append(errs, fmt.Errorf("tooling: "+format, args...))
	}
	t := Tooling{
		Images:  make(map[string]Image, len(tf.Images)),
		Actions: make(map[string]Action, len(tf.Actions)),
		Tools:   make(map[string]string, len(tf.Tools)),
	}

	for _, key := range slices.Sorted(maps.Keys(tf.Images)) {
		f := tf.Images[key]
		if !imageKeyPattern.MatchString(key) {
			fail("image %s: key must match %s", key, imageKeyPattern)
		}
		if !imageNamePattern.MatchString(f.Image) {
			fail("image %s: image %q is not an image name without tag or digest, such as golang or gcr.io/distroless/static-debian13", key, f.Image)
		}
		// The placeholder stands for a version such as 1.27 or 1.26.3.
		if tag := strings.ReplaceAll(f.Tag, GoPlaceholder, "1.27"); !imageTagPattern.MatchString(tag) {
			fail("image %s: tag %q is not a valid tag; %s is the only placeholder", key, f.Tag, GoPlaceholder)
		}
		if f.Digest != "" && !digestPattern.MatchString(f.Digest) {
			fail("image %s: digest %q must be sha256: followed by 64 lowercase hexadecimal digits", key, f.Digest)
		}
		t.Images[key] = Image{Name: f.Image, Tag: f.Tag, Digest: f.Digest}
	}

	for _, ref := range slices.Sorted(maps.Keys(tf.Actions)) {
		f := tf.Actions[ref]
		if !actionPattern.MatchString(ref) {
			fail("action %s: must be owner/repo, optionally followed by a path", ref)
		}
		if !semver.IsValid(f.Version) {
			fail("action %s: version %q must be a release such as v1.2.3", ref, f.Version)
		}
		if !shaPattern.MatchString(f.SHA) {
			fail("action %s: sha %q must be the full 40-character commit SHA of the release", ref, f.SHA)
		}
		t.Actions[ref] = Action(f)
	}

	for _, name := range slices.Sorted(maps.Keys(tf.Tools)) {
		v := tf.Tools[name]
		if !toolKeyPattern.MatchString(name) {
			fail("tool %s: name must match %s", name, toolKeyPattern)
		}
		if !semver.IsValid(v) || semver.Canonical(v) != v {
			fail("tool %s: version %q must be an exact version such as v1.2.3", name, v)
		}
		t.Tools[name] = v
	}
	return t, errs
}
