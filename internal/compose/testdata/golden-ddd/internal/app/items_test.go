package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"example.com/shop/internal/domain"
)

// memStore is an in-memory domain.ItemStore for tests.
type memStore struct {
	items []domain.Item
}

func (s *memStore) List(context.Context) ([]domain.Item, error) {
	return s.items, nil
}

func (s *memStore) Create(_ context.Context, name string) (domain.Item, error) {
	it := domain.Item{ID: int64(len(s.items) + 1), Name: name, CreatedAt: time.Now()}
	s.items = append(s.items, it)
	return it, nil
}

func TestItemService(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	s := NewItemService(&memStore{})
	created, err := s.Create(ctx, "first")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == 0 || created.Name != "first" {
		t.Errorf("Create() = %+v, want an ID and the name", created)
	}
	list, err := s.List(ctx)
	if err != nil || len(list) != 1 || list[0] != created {
		t.Errorf("List() = %+v, %v; want the created item", list, err)
	}
}

func TestItemServiceRejectsInvalidNames(t *testing.T) {
	t.Parallel()

	store := &memStore{}
	s := NewItemService(store)
	for _, name := range []string{"", " padded "} {
		if _, err := s.Create(t.Context(), name); !errors.Is(err, domain.ErrInvalidName) {
			t.Errorf("Create(%q) error = %v, want domain.ErrInvalidName", name, err)
		}
	}
	if len(store.items) != 0 {
		t.Errorf("store holds %+v, want nothing stored for invalid names", store.items)
	}
}
