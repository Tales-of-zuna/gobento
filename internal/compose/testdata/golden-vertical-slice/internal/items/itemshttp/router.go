// Package itemshttp serves the HTTP API with the standard
// library's net/http.
package itemshttp

import (
	"net/http"

	"example.com/shop/internal/items"
)

// NewRouter returns the HTTP handler for the API:
//
//	GET  /health  reports that the server is up
//	GET  /items   lists all items
//	POST /items   creates an item from {"name": "..."}
//
// GET routes also answer HEAD.
func NewRouter(items items.ItemStore) http.Handler {
	h := &itemHandler{items: items}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /items", h.list)
	mux.HandleFunc("POST /items", h.create)
	return mux
}

func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
