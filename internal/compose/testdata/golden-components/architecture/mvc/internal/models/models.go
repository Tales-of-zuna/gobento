// Package models defines the application's data types and
// the interfaces that connect its packages. Controllers depend only on these
// interfaces, never on a particular database or data-access library.
package models

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
