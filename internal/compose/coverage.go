package compose

import (
	"slices"

	"github.com/Tales-of-zuna/gobento/internal/catalog"
)

// Coverage returns the selections of supported that CI generates and tests:
// every selection without a component of an optional kind, and a subset of
// the others in which every pair of components that appears together in
// one of them appears together at least once (a pairwise covering set).
// The result keeps the order of supported and depends on nothing else.
//
// Required kinds are covered exhaustively, so every backend combination is
// proven. An optional component, such as a frontend, meets the rest of the
// project only through the architecture's contracts, so pairs bound the
// combinations worth building, and their number grows with the largest pair
// product instead of with the product of all kinds.
func Coverage(supported []catalog.Selection) []catalog.Selection {
	var rest []catalog.Selection
	for _, sel := range supported {
		if hasOptional(sel) {
			rest = append(rest, sel)
		}
	}
	chosen := make(map[int]bool)
	uncovered := make(map[pair]bool)
	for _, sel := range rest {
		for _, p := range pairs(sel) {
			uncovered[p] = true
		}
	}
	// Greedy: take the first selection that covers the most uncovered
	// pairs until none is left. Ties go to the earlier selection, so the
	// result is deterministic.
	for len(uncovered) > 0 {
		best, bestCount := -1, 0
		for i, sel := range rest {
			if chosen[i] {
				continue
			}
			n := 0
			for _, p := range pairs(sel) {
				if uncovered[p] {
					n++
				}
			}
			if n > bestCount {
				best, bestCount = i, n
			}
		}
		chosen[best] = true
		for _, p := range pairs(rest[best]) {
			delete(uncovered, p)
		}
	}

	var out []catalog.Selection
	for _, sel := range supported {
		if !hasOptional(sel) {
			out = append(out, sel)
			continue
		}
		i := slices.IndexFunc(rest, func(s catalog.Selection) bool { return s.String() == sel.String() })
		if chosen[i] {
			out = append(out, sel)
		}
	}
	return out
}

// pair is two components of one selection, by ID, in kind order.
type pair struct{ a, b string }

// pairs returns every pair of components in sel.
func pairs(sel catalog.Selection) []pair {
	comps := sel.Components()
	var out []pair
	for i := range comps {
		for j := i + 1; j < len(comps); j++ {
			out = append(out, pair{comps[i].ID(), comps[j].ID()})
		}
	}
	return out
}

// hasOptional reports whether sel has a component of an optional kind.
func hasOptional(sel catalog.Selection) bool {
	return slices.ContainsFunc(sel.Components(), func(c *catalog.Component) bool { return c.Kind.Optional() })
}
