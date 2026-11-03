package compat

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)

// ErrInvalidVersion reports a version string that is not in the expected
// form: a canonical semantic version such as v1.2.3 for modules, or a Go
// release version such as 1.27 or 1.27.1 for Go.
var ErrInvalidVersion = errors.New("invalid version")

// ErrInvalidRange reports a malformed range expression.
var ErrInvalidRange = errors.New("invalid range")

// CheckModuleVersion reports whether v is a canonical semantic version with
// a leading "v" and all three components, such as v1.61.1 or v2.0.0-rc.1.
// Shorthands like v1.61 are rejected because go.mod requires full versions.
func CheckModuleVersion(v string) error {
	if !semver.IsValid(v) || semver.Canonical(v) != v {
		return fmt.Errorf("%w %q: want a full semantic version such as v1.2.3", ErrInvalidVersion, v)
	}
	return nil
}

// NormalizeGo converts a Go release version as written in go.mod or on the
// command line (1.27 or 1.27.1) into the semantic version used for
// comparisons (v1.27.0 or v1.27.1). Prereleases such as 1.27rc1 and other
// forms are rejected.
func NormalizeGo(v string) (string, error) {
	parts := strings.Split(v, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return "", fmt.Errorf("%w %q: want a Go version such as 1.27 or 1.27.1", ErrInvalidVersion, v)
	}
	for _, p := range parts {
		if !isDecimal(p) {
			return "", fmt.Errorf("%w %q: want a Go version such as 1.27 or 1.27.1", ErrInvalidVersion, v)
		}
	}
	if len(parts) == 2 {
		parts = append(parts, "0")
	}
	return "v" + strings.Join(parts, "."), nil
}

// isDecimal reports whether s is a non-empty run of ASCII digits without a
// redundant leading zero.
func isDecimal(s string) bool {
	if s == "" || (len(s) > 1 && s[0] == '0') {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// GoDirective returns the Go version as written in a generated go.mod: the
// full release version, such as 1.27.0 for 1.27. Go orders 1.27 before
// 1.27.0, so a short directive would be raised by any dependency that
// declares go 1.27.0.
func GoDirective(v string) (string, error) {
	norm, err := NormalizeGo(v)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(norm, "v"), nil
}
