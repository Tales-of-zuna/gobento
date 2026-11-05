// Package controllers serves the HTTP API with the chi
// router.
package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"example.com/shop/internal/models"
)

// NewRouter returns the HTTP handler for the API:
//
//	GET  /health  reports that the server is up
//	GET  /items   lists all items
//	POST /items   creates an item from {"name": "..."}
//
// GET routes also answer HEAD.
func NewRouter(items models.ItemStore) http.Handler {
	h := &itemHandler{items: items}
	r := chi.NewRouter()
	r.Get("/health", health)
	r.Head("/health", health)
	r.Get("/items", h.list)
	r.Head("/items", h.list)
	r.Post("/items", h.create)
	return r
}

func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
