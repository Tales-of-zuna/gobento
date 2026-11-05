# Design

## Context

An architecture declares slots (one directory = one Go package each) and owns
the entry point and the contract package (`Item`, `ItemStore`). Framework,
database, and data-access templates only use the slot names `models`,
`handlers`, `database`, and `store`: handlers provide
`NewRouter(models.ItemStore) http.Handler`, stores provide
`New(db) models.ItemStore`, and the shared API test in every framework checks
the HTTP behavior. `mvc` and `vertical-slice` wire
`NewRouter(store.New(db))` directly, so they have no application layer. See
proposal.md for the motivation.

## Goals / Non-Goals

**Goals:**

- A DDD / hexagonal layout a Go team would recognize and keep: a domain
  package, an application service, and adapters, with dependencies pointing
  inward and only `main` knowing every package.
- A real application layer (use cases plus domain rules) between the HTTP
  adapter and the store, without changing any framework, database, or data
  component, the `ItemStore` contract, or the core.

**Non-Goals:**

- Domain events, CQRS, unit of work, or value-object types for every field:
  they would be ceremony for a one-entity example.
- Changing the handler-facing contract (for example a separate use-case
  interface): it is shared by three frameworks and their byte-identical API
  test.
- Several aggregates or a user-chosen aggregate name.

## Decisions

### D1. Name `ddd`, title "Domain-driven design"

`ddd` is what the roadmap, users, and searches call the style, and it is short
to type in `--arch ddd`. The title is 20 characters, so the home belt shows its
initials, `DDD`, like `MVC` and `VS`. Alternatives: `hexagonal` (narrower: it
names the ports-and-adapters part only) and `clean` (a different, more layered
pattern).

### D2. Layout

| Slot       | Directory                       | Package       | Filled by    |
| ---------- | ------------------------------- | ------------- | ------------ |
| `entry`    | `cmd/{{.Project.Name}}`         | `main`        | architecture |
| `models`   | `internal/domain`               | `domain`      | architecture |
| `app`      | `internal/app`                  | `app`         | architecture |
| `handlers` | `internal/adapters/httpapi`     | `httpapi`     | framework    |
| `store`    | `internal/adapters/persistence` | `persistence` | data         |
| `database` | `internal/adapters/database`    | `database`    | database     |

- Three layers, as in ThreeDotsLabs' "Wild Workouts" Go DDD example
  (`domain`, `app`, `adapters`) and the usual ports-and-adapters shape: the
  domain is the center, the application layer runs use cases on it, and every
  piece of I/O is an adapter. Evans' four-layer names were rejected for Go:
  `interfaces` reads as Go interfaces, and `infrastructure`/`application` are
  long package names that Go code would alias anyway.
- One `domain` package instead of `internal/domain/item`: with one aggregate,
  a package per aggregate is premature. The README says to split into
  `internal/domain/<aggregate>` when the domain grows. A package named `item`
  would also clash with the GORM store's unexported `item` row type in the
  same file that imports it.
- `httpapi`, not `http`, so it does not collide with `net/http` in every
  handler file (same reason as `itemshttp` in `vertical-slice`).
- `persistence` holds the repositories of every aggregate, as is common in Go
  services; sqlc's `sqlcgen` subpackage stays below it.
- `database` sits with the adapters: it is the shared connection and schema
  the persistence adapter runs on.
- Directory names equal package names, so no slot needs `package:` except the
  entry point.

### D3. The application service implements the port the handlers use

Handlers take a `models.ItemStore`, and the contract package is the domain.
Instead of changing that contract, the architecture adds an
architecture-owned slot `app` with `ItemService`:

```go
func NewItemService(store domain.ItemStore) *ItemService
func (s *ItemService) List(ctx) ([]domain.Item, error)
func (s *ItemService) Create(ctx, name string) (domain.Item, error) // checks domain.ValidateName
```

`ItemService` implements `domain.ItemStore`, so `main` wires
`httpapi.NewRouter(app.NewItemService(persistence.New(db)))`. The handlers
depend on the port only, the service is the only caller of the persistence
adapter, and the frameworks are untouched. In this example the use-case port
and the repository port have the same methods, so one interface serves both;
the generated README says to give the handlers their own interface once the
use cases go beyond storing and listing.

- Why an extra slot, not a file in `models` or the entry: a slot is the only
  way an architecture places its own package, and slot names other components
  do not use are invisible to them. The core already allows any number of
  architecture-filled slots; no core change.
- Alternative: put the service in `domain`. Rejected: use cases orchestrate
  ports and belong outside the domain.
- Alternative: change `NewRouter` to take a service interface in all three
  frameworks. Rejected: no behavior gain, and it rewrites every framework and
  its golden files.

### D4. Domain rule

`domain.ValidateName` returns an error wrapping `domain.ErrInvalidName` for an
empty name or one with leading or trailing white space. The handlers already
trim and reject empty names with 400, so the HTTP behavior and the shared API
test are unchanged; the rule makes the invariant hold for any caller of the
service, and the domain and app packages get their own unit tests (plain Go,
no database). `Item` keeps its JSON tags because the shipped handlers encode
it directly; the README notes that response types go in `httpapi` when the API
and the domain diverge.

### D5. Tests and goldens

- `testdata/golden-ddd` holds `ddd/nethttp/sqlite/sqlc`.
- `TestGoldenComponents` picks, for a non-architecture component, the
  reference selection (`mvc/nethttp/sqlite/sqlc`) with that component
  swapped in when it is supported, and otherwise the first supported
  selection that contains it. Candidates are sorted by name, so `ddd` now
  sorts first; without this, adding an architecture would rewrite the goldens
  of every framework, database, and data component.
- `TestShippedFrameworksAreLayoutIndependent` compares every non-`mvc`
  architecture with `mvc`, so it covers `ddd`.
- The integration test also checks, when the plan has an `app` slot, that the
  `models` package imports no other package of the project.

## Risks / Trade-offs

- [The handlers' port is named `ItemStore` although it is served by a
  service] → Inherent to keeping the shared contract; the README and the
  `domain` doc comment explain who implements it on each side.
- [A third copy of `main.go`] → Covered by goldens and the integration test;
  `ddd`'s differs (it wires the service), which supports keeping the copies
  separate (vertical-slice D4).
- [`ddd` sorts first in candidate lists, interactive questions, and the
  configurator] → Order is alphabetical by design; the configurator has no
  default architecture, and docs always pass `--arch`.

## Migration Plan

Nothing changes for existing architectures. Scripts that parse the `--arch`
error now see `ddd` in the list of values.
