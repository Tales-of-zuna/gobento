// Package items is the items feature: the Item type and the
// ItemStore interface that connects the feature's packages.
//
// The feature's HTTP handlers (itemshttp) and its data access
// (itemsstore) live in subpackages that depend on this package
// and never on each other, so the handlers work with any store.
package items

import (
	"context"
	"time"
)

// Item is an example resource.
type Item struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// ItemStore stores items.
type ItemStore interface {
	// List returns every item, oldest first.
	List(ctx context.Context) ([]Item, error)
	// Create stores a new item with the given name and returns it.
	Create(ctx context.Context, name string) (Item, error)
}
