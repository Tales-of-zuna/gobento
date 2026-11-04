// Package controllers serves the HTTP API with the Gin web
// framework.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"example.com/shop/internal/models"
)

// NewRouter returns the HTTP handler for the API:
//
//	GET  /health  reports that the server is up
//	GET  /items   lists all items
//	POST /items   creates an item from {"name": "..."}
//
// GET routes also answer HEAD. The engine has no middleware: http.Server
// already recovers from panics in handlers, and request logging is left to
// you. In debug mode, the default unless GIN_MODE=release is set, Gin prints
// the registered routes at startup.
func NewRouter(items models.ItemStore) http.Handler {
	h := &itemHandler{items: items}
	r := gin.New()
	// Answer 405, not 404, for a known path with an unsupported method, and
	// 404 instead of a redirect for paths that differ only in a trailing
	// slash or in case, like the other frameworks.
	r.HandleMethodNotAllowed = true
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false
	r.GET("/health", health)
	r.HEAD("/health", health)
	r.GET("/items", h.list)
	r.HEAD("/items", h.list)
	r.POST("/items", h.create)
	return r
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
