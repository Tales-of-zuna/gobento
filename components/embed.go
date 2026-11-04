// Package components holds gobento's component catalog as data: one directory
// per component (components/<kind>/<name>/) with a component.yaml manifest and
// .tmpl template files, the base layer (base/) that every generated project
// gets, and versions.yaml, the version-compatibility rules.
//
// No file here other than this one is Go code, so templates are never built,
// vetted, or linted as part of gobento. Package internal/catalog loads and
// validates the files.
package components

import (
	"embed"
	"io/fs"
)

// files holds the catalog. The all: prefix includes files whose names start
// with "." or "_", so that the catalog loader can reject them instead of
// having them silently left out of the binary.
//
//go:embed versions.yaml all:base all:architecture all:framework all:database all:data
var files embed.FS

// FS returns the embedded catalog as a read-only file system rooted at the
// components directory.
func FS() fs.FS {
	return files
}
