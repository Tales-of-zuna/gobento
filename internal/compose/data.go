package compose

// Data is the value every template is executed with. Its shape is part of
// the contributor API for writing components and is documented in
// website/content/docs/contributing/adding-a-component.mdx; change it only
// together with that page and the golden test.
type Data struct {
	// Project describes the generated project.
	Project Project
	// Slots maps each slot the architecture declares to its location, for
	// example {{.Slots.handlers.Import}}.
	Slots map[string]SlotData
	// Traits maps trait names to the values the selection provides. Use
	// index for names with dashes: {{index .Traits "sql-dialect"}}.
	Traits map[string]string
	// Versions maps module paths to their chosen versions, for example
	// {{index .Versions "modernc.org/sqlite"}}.
	Versions map[string]string
	// Assets maps the names of the vendored assets of the selected
	// components to their versions, for example
	// {{index .Assets "htmx.org"}}.
	Assets map[string]string
	// APIPrefix is the path prefix the API router is served under: "" when
	// no frontend is selected, so the API is at the root, and "/api" when
	// one is, so the frontend serves every other path. Write API paths as
	// {{.APIPrefix}}/items.
	APIPrefix string
	// Images maps the image keys of the rules file's tooling section to
	// full references (name:tag, then @digest when pinned), for example
	// {{.Images.runtime}}.
	Images map[string]string
	// Actions maps GitHub Action references to complete values for uses:,
	// the commit SHA followed by a version comment, for example
	// {{index .Actions "actions/checkout"}}.
	Actions map[string]string
	// Tools maps tool names to their exact versions, for example
	// {{index .Tools "golangci-lint"}}.
	Tools map[string]string
}

// FrontendAPIPrefix is the value of Data.APIPrefix when a frontend is
// selected.
const FrontendAPIPrefix = "/api"

// Project describes the generated project.
type Project struct {
	// Name is the project name, for example shop.
	Name string
	// Module is the Go module path, for example example.com/shop.
	Module string
	// GoVersion is the chosen Go version as written in go.mod, for
	// example 1.27.
	GoVersion string
}

// SlotData locates one slot in the generated project.
type SlotData struct {
	// Dir is the slot directory relative to the project root, with forward
	// slashes, for example internal/controllers.
	Dir string
	// Package is the slot's Go package name, for example controllers.
	Package string
	// Import is the slot's import path, for example
	// example.com/shop/internal/controllers.
	Import string
	// Filled reports whether a selected component has templates for the
	// slot. An optional slot, such as web, may be unfilled; test it with
	// {{if .Slots.web.Filled}}.
	Filled bool
}
