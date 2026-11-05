// Package catalogjson converts gobento's component catalog, version rules,
// supported selections, and input rules into one JSON document.
//
// The document is read by the website (website/src/generated/catalog.json,
// written by internal/tools/catalogexport) and printed by `gobento list
// --json`, so both always show the same data. Its schema is
// website/lib/catalog/schema.ts; the types in this package mirror it, and
// the two change together. Fields are added, not renamed or removed.
package catalogjson
