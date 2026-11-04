# Design

## Context

The `handlers` slot contract (owned by the architecture, see
`components/architecture/mvc/files/entry/main.go.tmpl`) is:

```go
func NewRouter(items models.ItemStore) http.Handler
```

`main.go` wraps the handler in its own `http.Server` (timeouts, graceful
shutdown). `*gin.Engine` and `chi.Router` both implement `http.Handler`, so
neither framework needs anything from the architecture beyond what `nethttp`
uses. The existing `router_test.go.tmpl` of `nethttp` already exercises the
API only through `NewRouter` and `httptest`, so it is framework-independent.

## Goals / Non-Goals

**Goals:** two new framework components with idiomatic routing and identical
API behavior; an enforced shared API test; golden coverage per component;
versions as data.

**Non-Goals:** framework-specific middleware stacks (logging, CORS, request
IDs), a default framework when the flag is omitted, other frameworks (echo,
fiber), changes to the slot contract.

## Decisions

### D1. Same API, idiomatic routing, shared request handling style

- **chi:** `chi.NewRouter()` with `r.Get/r.Post`. Handlers are plain
  `http.HandlerFunc`s, so `items.go` is the same code as in `nethttp`
  (JSON helpers, `http.MaxBytesReader`, `DisallowUnknownFields`, trailing-data
  check). chi answers unknown methods on a known route with `405` by default.
- **gin:** `gin.New()` with `HandleMethodNotAllowed = true` (gin answers `404`
  otherwise). Handlers take `*gin.Context` and respond with `c.JSON`. Bodies
  are decoded with `encoding/json` over `http.MaxBytesReader(c.Writer,
  c.Request.Body, …)` instead of `c.ShouldBindJSON`: gin's binding cannot
  reject trailing data, has no per-request size limit, and only rejects
  unknown fields through a process-global switch
  (`binding.EnableDecoderDisallowUnknownFields`), which a generated library
  package must not flip.
- **No extra middleware** in any framework: `nethttp` has none, and
  `http.Server` already recovers handler panics per connection. Adding
  `gin.Recovery()` or `middleware.Recoverer` only to some frameworks would make
  their behavior differ. Users add middleware in their own project.
- **gin mode:** gin starts in debug mode and logs route registrations unless
  `GIN_MODE=release` is set. gobento does not call the process-global
  `gin.SetMode` from a library package; the gin notes document `GIN_MODE`.

*Alternative rejected:* sharing `items.go` between `nethttp` and `chi` through
a new "shared files" mechanism. It would be the only cross-component file
sharing in the catalog, for one file; the shared API test (D2) already catches
behavioral drift.

### D2. One API test, enforced identical

`files/handlers/router_test.go.tmpl` is byte-identical in every framework
component. A catalog test (`internal/catalog/shipped_test.go`) compares the
copies of all shipped framework components and names the first pair that
differs. This test is the executable form of the new requirement, and each
generated project runs it in CI (integration test, all OSes).

The Content-Type check becomes `mime.ParseMediaType(...) == "application/json"`
because gin sends `application/json; charset=utf-8`; both are valid JSON
responses.

### D3. Versions

Checked on proxy.golang.org and vuln.go.dev on 2026-11-04:

| Module | Range | Default | Reason for floor |
|---|---|---|---|
| github.com/gin-gonic/gin | `>=v1.11.0 <v2.0.0` | v1.12.0 | newest minor before default; all gin advisories are fixed in v1.9.1 |
| github.com/go-chi/chi/v5 | `>=v5.3.0 <v6.0.0` | v5.5.0 | v5.3.0 fixes GO-2026-5774/5775/5777 (and the older RedirectSlashes advisories) |

Both go.mod files require Go ≤ 1.25, below gobento's Go floor (1.26), so no
Go constraint is needed. The integration test proves floors and defaults
build and pass on every OS.

### D4. Golden coverage per component

The reference golden (`testdata/golden/`, full project for
mvc/nethttp/sqlite/sqlc) stays. A new `TestGoldenComponents` composes, for
each component of the catalog, the first supported selection containing it
and compares only the files that component contributes (`File.Component`)
with `testdata/golden-components/<kind>/<name>/`. This scales with the number
of components, not combinations, and relies on the existing invariant that a
component's files do not depend on which other components are selected
(spec: *Swapping one component leaves others untouched*).

### D5. CLI behavior

No code change: with three supported frameworks, the existing rule makes
`--framework` required in non-interactive use. CI's `sqlc-fresh` job and all
docs examples pass `--framework nethttp` explicitly.

## Risks / Trade-offs

- **gin's dependency tree is large** (validator, quic-go via its HTTP/3
  support, etc.) → generation time and go.sum size grow for gin projects;
  acceptable, it is what gin users get with `go get`.
- **Duplicated `items.go` between nethttp and chi** → drift is caught by D2;
  revisit if a third std-handler framework appears.
- **Golden directories grow per component** → small, reviewed diffs are the
  intent.

## Migration Plan

Pre-1.0. Scripts calling `gobento new` without `--framework` must add it; the
error message lists the values. Release notes (release-please) carry the
`feat!`/BREAKING note from the commit.
