// Package compose combines a selection of catalog components into a Plan, the
// in-memory list of files a generated project consists of. It resolves the
// architecture's slots, checks ownership and traits, renders every template
// with a fixed data shape, and formats Go output. All composition problems
// are reported together.
package compose
