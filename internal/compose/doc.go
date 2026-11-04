// Package compose combines a selection of catalog components into a Plan, the
// in-memory list of files a generated project consists of. It resolves the
// architecture's slots, checks ownership and traits, renders every template
// with a fixed data shape, and formats Go output. The catalog's base layer
// is composed into every selection; its templates, and those of components
// whose kind declares points (kind.yaml), include the selected components'
// contributions to their points with the contrib function. A template that
// renders to whitespace produces no file. All composition problems are
// reported together.
package compose
