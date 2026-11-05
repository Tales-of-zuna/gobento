package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"example.com/shop/internal/models"
)

// maxBodyBytes limits the size of request bodies.
const maxBodyBytes = 1 << 20

type itemHandler struct {
	items models.ItemStore
}

func (h *itemHandler) list(c *gin.Context) {
	// Local names avoid "items": it may be the name of the contract package.
	list, err := h.items.List(c.Request.Context())
	if err != nil {
		serverError(c, err)
		return
	}
	if list == nil {
		list = []models.Item{}
	}
	c.JSON(http.StatusOK, list)
}

type createItemRequest struct {
	Name string `json:"name"`
}

func (h *itemHandler) create(c *gin.Context) {
	// The body is decoded with encoding/json rather than c.ShouldBindJSON,
	// which cannot limit the body size per request, reject data after the
	// JSON object, or reject unknown fields without a process-wide setting.
	var req createItemRequest
	dec := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(c, http.StatusRequestEntityTooLarge, "request body is too large")
			return
		}
		writeError(c, http.StatusBadRequest, `invalid JSON body; send {"name": "..."}`)
		return
	}
	// The body must hold exactly one JSON value.
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(c, http.StatusBadRequest, "invalid JSON body; unexpected data after the JSON object")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(c, http.StatusBadRequest, "name is required")
		return
	}

	item, err := h.items.Create(c.Request.Context(), req.Name)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

// writeError writes a JSON error response.
func writeError(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}

// serverError logs err and responds with 500 without exposing details.
func serverError(c *gin.Context, err error) {
	slog.Error("request failed", "method", c.Request.Method, "path", c.Request.URL.Path, "error", err)
	writeError(c, http.StatusInternalServerError, "internal server error")
}
