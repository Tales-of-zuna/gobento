// Package compose combines a selection of catalog components into a Plan, the
// in-memory list of files a generated project consists of. It resolves the
// architecture's slots, checks ownership and traits, renders every template
// with a fixed data shape, and formats Go output. The catalog's base layer
// is composed into every selection; its templates include the selected
// components' contributions to its points with the contrib function. All
// composition problems are reported together.
package compose
