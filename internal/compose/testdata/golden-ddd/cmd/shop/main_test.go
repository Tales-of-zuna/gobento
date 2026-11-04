package main

import (
	"context"
	"mime"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/shop/internal/app"
	"example.com/shop/internal/domain"
)

// memStore is an in-memory domain.ItemStore that stands in for
// the persistence adapter, so the routing test needs no database.
type memStore struct {
	items []domain.Item
}

func (s *memStore) List(context.Context) ([]domain.Item, error) {
	return s.items, nil
}

func (s *memStore) Create(_ context.Context, name string) (domain.Item, error) {
	item := domain.Item{ID: int64(len(s.items) + 1), Name: name, CreatedAt: time.Now()}
	s.items = append(s.items, item)
	return item, nil
}

// TestHandlerRoutes checks which handler serves which path, with the
// application service wired in as in main: the API at
// the root.
func TestHandlerRoutes(t *testing.T) {
	tests := []struct {
		target    string
		mediaType string
	}{
		{"/health", "application/json"},
		{"/items", "application/json"},
	}
	h := newHandler(app.NewItemService(&memStore{}))
	for _, tt := range tests {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.target, nil))
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", tt.target, rec.Code, http.StatusOK)
			continue
		}
		if mt, _, err := mime.ParseMediaType(rec.Header().Get("Content-Type")); err != nil || mt != tt.mediaType {
			t.Errorf("GET %s Content-Type = %q, want %s", tt.target, rec.Header().Get("Content-Type"), tt.mediaType)
		}
	}
}
