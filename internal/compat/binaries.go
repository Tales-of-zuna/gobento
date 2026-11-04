package compat

import (
	"fmt"
	"maps"
	"net/url"
	"regexp"
	"slices"
	"strings"
)

// Binary is a program that generated projects download on demand, one
// file per platform, such as the standalone Tailwind CSS CLI. Like an
// asset's, its version is fixed per gobento release and cannot be pinned.
type Binary struct {
	// Name identifies the binary, for example tailwindcss.
	Name string
	// Components lists the components that use the binary, as
	// "<kind>/<name>" identifiers.
	Components []string
	// Traits optionally restricts the binary to selections whose traits
	// have one of the listed values, as for modules.
	Traits map[string][]string
	// Version is the upstream release, for example 4.3.4.
	Version string
	// Checksums is the https URL of the release's checksum file, which
	// lists "<sha256>  <file name>" lines, or "" if it publishes none.
	Checksums string
	// Platforms lists the downloads, sorted by OS and then architecture.
	Platforms []Platform
}

// Platform is the download of a binary for one GOOS/GOARCH platform.
type Platform struct {
	// OS is the GOOS value, for example linux.
	OS string
	// Arch is the GOARCH value, for example amd64.
	Arch string
	// URL is the https URL of the file, which contains the version.
	URL string
	// SHA256 is the file's SHA-256 checksum in lowercase hex.
	SHA256 string
}

// Key returns the platform as "<os>/<arch>", as in the rules file.
func (p Platform) Key() string {
	return p.OS + "/" + p.Arch
}

// FileName returns the last element of the URL's path, the name of the
// file in the release.
func (p Platform) FileName() string {
	u, err := url.Parse(p.URL)
	if err != nil {
		return ""
	}
	return u.Path[strings.LastIndex(u.Path, "/")+1:]
}

type binaryFile struct {
	Components []string                `yaml:"components"`
	Traits     map[string][]string     `yaml:"traits"`
	Version    string                  `yaml:"version"`
	Checksums  string                  `yaml:"checksums"`
	Platforms  map[string]platformFile `yaml:"platforms"`
}

type platformFile struct {
	URL    string `yaml:"url"`
	SHA256 string `yaml:"sha256"`
}

// platformPattern matches a GOOS/GOARCH pair such as linux/amd64.
var platformPattern = regexp.MustCompile(`^[a-z][a-z0-9]*/[a-z][a-z0-9]*$`)

// parseBinary converts and validates one binary entry.
func parseBinary(name string, bf binaryFile) (Binary, []error) {
	var errs []error
	fail := func(format string, args ...any) {
		errs = append(errs, fmt.Errorf("binary %s: "+format, append([]any{name}, args...)...))
	}
	if strings.TrimSpace(name) == "" {
		fail("empty name")
	}
	if len(bf.Components) == 0 {
		fail("no components listed")
	}
	checkTraitConditions(bf.Traits, fail)
	if !assetVersionPattern.MatchString(bf.Version) {
		fail("version %q must be a release such as 4.3.4", bf.Version)
	}
	if bf.Checksums != "" && !isHTTPS(bf.Checksums) {
		fail("checksums %q must be an https URL of the release's checksum file", bf.Checksums)
	}
	if len(bf.Platforms) == 0 {
		fail("no platforms listed")
	}
	b := Binary{Name: name, Components: bf.Components, Traits: bf.Traits, Version: bf.Version, Checksums: bf.Checksums}
	for _, key := range slices.Sorted(maps.Keys(bf.Platforms)) {
		pf := bf.Platforms[key]
		if !platformPattern.MatchString(key) {
			fail("platform %q must be GOOS/GOARCH, for example linux/amd64", key)
			continue
		}
		switch {
		case !isHTTPS(pf.URL):
			fail("platform %s: url %q must be an https URL", key, pf.URL)
		case bf.Version != "" && !strings.Contains(pf.URL, bf.Version):
			fail("platform %s: url %q does not contain the version %s", key, pf.URL, bf.Version)
		}
		if !sha256Pattern.MatchString(pf.SHA256) {
			fail("platform %s: sha256 %q must be 64 lowercase hexadecimal digits", key, pf.SHA256)
		}
		goos, goarch, _ := strings.Cut(key, "/")
		b.Platforms = append(b.Platforms, Platform{OS: goos, Arch: goarch, URL: pf.URL, SHA256: pf.SHA256})
	}
	return b, errs
}

// isHTTPS reports whether s is an absolute https URL with a host.
func isHTTPS(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme == "https" && u.Host != ""
}

// Binary returns the binary with the given name.
func (r *Rules) Binary(name string) (Binary, bool) {
	i := slices.IndexFunc(r.Binaries, func(b Binary) bool { return b.Name == name })
	if i < 0 {
		return Binary{}, false
	}
	return r.Binaries[i], true
}

// BinaryVersions returns the binaries of the selection keyed by name.
func (r Resolved) BinaryVersions() map[string]Binary {
	out := make(map[string]Binary, len(r.Binaries))
	for _, b := range r.Binaries {
		out[b.Name] = b
	}
	return out
}
