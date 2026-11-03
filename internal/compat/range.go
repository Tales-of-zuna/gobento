package compat

import (
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)

// Range is a set of versions described by space-separated comparators that
// must all hold, for example ">=v1.61.0 <v2.0.0". The operators are >=, >,
// <=, <, and =. The zero Range contains no comparators and matches every
// version; ParseRange never returns it.
type Range struct {
	raw         string
	comparators []comparator
}

type comparator struct {
	op      string
	version string // canonical semantic version, used for comparison
}

// operators lists the comparison operators, longest first so that ">="
// is matched before ">".
var operators = []string{">=", "<=", ">", "<", "="}

// ParseRange parses a range over module versions. Each operand must be a
// canonical semantic version such as v1.61.0.
func ParseRange(s string) (Range, error) {
	return parseRange(s, func(v string) (string, error) {
		if err := CheckModuleVersion(v); err != nil {
			return "", err
		}
		return v, nil
	})
}

// ParseGoRange parses a range over Go versions, such as ">=1.26 <1.28".
// Operands are Go release versions and are normalized with NormalizeGo.
func ParseGoRange(s string) (Range, error) {
	return parseRange(s, NormalizeGo)
}

func parseRange(s string, normalize func(string) (string, error)) (Range, error) {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return Range{}, fmt.Errorf("%w %q: no comparators", ErrInvalidRange, s)
	}
	r := Range{raw: strings.Join(fields, " ")}
	for _, f := range fields {
		op := operatorOf(f)
		if op == "" {
			return Range{}, fmt.Errorf("%w %q: comparator %q must start with one of >=, >, <=, <, =", ErrInvalidRange, s, f)
		}
		v, err := normalize(f[len(op):])
		if err != nil {
			return Range{}, fmt.Errorf("%w %q: %w", ErrInvalidRange, s, err)
		}
		r.comparators = append(r.comparators, comparator{op: op, version: v})
	}
	return r, nil
}

func operatorOf(field string) string {
	for _, op := range operators {
		if strings.HasPrefix(field, op) {
			return op
		}
	}
	return ""
}

// Contains reports whether the canonical semantic version v satisfies every
// comparator of r. Go versions must be normalized with NormalizeGo first.
func (r Range) Contains(v string) bool {
	for _, c := range r.comparators {
		cmp := semver.Compare(v, c.version)
		var ok bool
		switch c.op {
		case ">=":
			ok = cmp >= 0
		case ">":
			ok = cmp > 0
		case "<=":
			ok = cmp <= 0
		case "<":
			ok = cmp < 0
		case "=":
			ok = cmp == 0
		}
		if !ok {
			return false
		}
	}
	return true
}

// Exact returns the single version r allows when r consists of exactly one
// "=" comparator, as required for tool modules.
func (r Range) Exact() (string, bool) {
	if len(r.comparators) == 1 && r.comparators[0].op == "=" {
		return r.comparators[0].version, true
	}
	return "", false
}

// String returns the range as written, with runs of spaces collapsed.
func (r Range) String() string {
	return r.raw
}
