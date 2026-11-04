// Package views serves the web pages: HTML rendered on the server
// from the templ components in pages.templ, made interactive with htmx, and
// the static files they load, embedded in the binary.
package views

//go:generate go run github.com/a-h/templ/cmd/templ@v0.3.1100 generate

import (
	"bytes"
	"embed"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/a-h/templ"

	"example.com/shop/internal/models"
)

// static holds the stylesheet and htmx 2.0.11, served under
// /static/ so that the pages need no network access beyond this server.
//
//go:embed static
var static embed.FS

// title is the title of every page. It is Go code rather than text in
// pages.templ, so that a CSS framework scanning the pages for class names
// does not take the project name for one.
const title = "shop"

// maxFormBytes limits the size of form submissions.
const maxFormBytes = 1 << 20

// NewHandler returns the HTTP handler for the web pages:
//
//	GET  /           lists every item, with a form that adds one
//	POST /items      adds an item from the form field name
//	GET  /static/... serves the stylesheet and htmx
//
// GET routes also answer HEAD. Unsafe requests from other origins, such as
// a form on another site that posts here, are rejected with 403.
func NewHandler(itemStore models.ItemStore) http.Handler {
	p := &pages{items: itemStore}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", p.home)
	mux.HandleFunc("POST /items", p.create)
	mux.Handle("GET /static/", http.FileServerFS(static))
	return http.NewCrossOriginProtection().Handler(mux)
}

type pages struct {
	items models.ItemStore
}

func (p *pages) home(w http.ResponseWriter, r *http.Request) {
	p.page(w, r, http.StatusOK, "", "")
}

// page renders the home page with the given status and form state.
func (p *pages) page(w http.ResponseWriter, r *http.Request, status int, name, problem string) {
	list, err := p.items.List(r.Context())
	if err != nil {
		serverError(w, r, err)
		return
	}
	render(w, r, status, Page(list, name, problem))
}

func (p *pages) create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxFormBytes)
	if err := r.ParseForm(); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			http.Error(w, "the form is too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.PostForm.Get("name"))
	if name == "" {
		p.reject(w, r, name, "Enter a name for the item.")
		return
	}

	item, err := p.items.Create(r.Context(), name)
	if err != nil {
		serverError(w, r, err)
		return
	}
	if !isHTMX(r) {
		// Post/Redirect/Get: reloading the page does not post the form again.
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	render(w, r, http.StatusOK, Created(item))
}

// reject answers a form that cannot be accepted with 422 and the form,
// showing problem next to the field.
func (p *pages) reject(w http.ResponseWriter, r *http.Request, name, problem string) {
	if !isHTMX(r) {
		p.page(w, r, http.StatusUnprocessableEntity, name, problem)
		return
	}
	// Replace the form instead of appending to the list.
	w.Header().Set("HX-Retarget", "#new-item")
	w.Header().Set("HX-Reswap", "outerHTML")
	render(w, r, http.StatusUnprocessableEntity, ItemForm(name, problem, false))
}

// isHTMX reports whether htmx sent r, rather than a plain browser request.
func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// render writes c as an HTML response with the given status. It renders
// into a buffer first, so that a failing component yields a clean 500.
func render(w http.ResponseWriter, r *http.Request, status int, c templ.Component) {
	var buf bytes.Buffer
	if err := c.Render(r.Context(), &buf); err != nil {
		serverError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := buf.WriteTo(w); err != nil {
		slog.Error("write response", "error", err)
	}
}

// serverError logs err and responds with 500 without exposing details.
func serverError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("request failed", "method", r.Method, "path", r.URL.Path, "error", err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
