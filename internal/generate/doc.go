// Package generate turns a composed plan into a project on disk. It runs the
// Go toolchain preflight, writes the plan into a temporary directory, resolves
// dependencies with the go command, verifies the resulting go.mod, and places
// the result atomically, cleaning up on any error or interrupt.
package generate
