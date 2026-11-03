package store

import (
	"context"
	"path/filepath"
	"testing"

	"example.com/shop/internal/database"
)

// TestItemStore runs the store against a real database file.
func TestItemStore(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	// Windows cannot remove open files, so close before TempDir cleanup.
	t.Cleanup(func() { _ = db.Close() })

	s := New(db)
	items, err := s.List(ctx)
	if err != nil || len(items) != 0 {
		t.Fatalf("List() on an empty database = %v, %v; want no items", items, err)
	}

	created, err := s.Create(ctx, "first")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == 0 || created.Name != "first" || created.CreatedAt.IsZero() {
		t.Errorf("Create() = %+v, want an ID, the name, and a creation time", created)
	}

	if _, err := s.Create(ctx, "second"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	items, err = s.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 2 || items[0].Name != "first" || items[1].Name != "second" {
		t.Errorf("List() = %+v, want first and second in order", items)
	}
}
