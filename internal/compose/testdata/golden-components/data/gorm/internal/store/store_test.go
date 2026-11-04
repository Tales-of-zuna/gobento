package store

import (
	"testing"

	"example.com/shop/internal/database"
	"example.com/shop/internal/database/dbtest"
)

// TestItemStore runs the store against a real, empty database.
func TestItemStore(t *testing.T) {
	ctx := t.Context()
	s := New(dbtest.Open(t))
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
	if len(items) != 2 || items[0].ID != created.ID || items[0].Name != "first" || items[1].Name != "second" {
		t.Errorf("List() = %+v, want first and second in order", items)
	}
}

// TestStoreOnClosedDatabase checks that a store on a closed connection
// returns errors instead of panicking, including when GORM cannot be set up.
func TestStoreOnClosedDatabase(t *testing.T) {
	ctx := t.Context()
	db, err := database.Open(ctx, dbtest.URL(t))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	s := New(db)
	if _, err := s.List(ctx); err == nil {
		t.Error("List() on a closed database: error = nil, want an error")
	}
	if _, err := s.Create(ctx, "x"); err == nil {
		t.Error("Create() on a closed database: error = nil, want an error")
	}
}
