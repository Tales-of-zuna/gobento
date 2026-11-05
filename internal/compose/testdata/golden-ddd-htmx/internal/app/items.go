// Package app is the application layer: it runs the item
// use cases on the domain, package domain, the only package
// of the project it imports. The HTTP adapter reaches it through the domain's
// ItemStore port.
package app

import (
	"context"
	"fmt"

	"example.com/shop/internal/domain"
)

// ItemService runs the item use cases. It implements
// domain.ItemStore, the port the HTTP adapter uses, and stores
// items through another domain.ItemStore, such as the
// persistence adapter.
type ItemService struct {
	store domain.ItemStore
}

var _ domain.ItemStore = (*ItemService)(nil)

// NewItemService returns an ItemService that stores items in store.
func NewItemService(store domain.ItemStore) *ItemService {
	return &ItemService{store: store}
}

// List returns every item, oldest first.
func (s *ItemService) List(ctx context.Context) ([]domain.Item, error) {
	return s.store.List(ctx)
}

// Create checks the domain rules for name and stores a new item with it.
func (s *ItemService) Create(ctx context.Context, name string) (domain.Item, error) {
	if err := domain.ValidateName(name); err != nil {
		return domain.Item{}, fmt.Errorf("create item: %w", err)
	}
	return s.store.Create(ctx, name)
}
