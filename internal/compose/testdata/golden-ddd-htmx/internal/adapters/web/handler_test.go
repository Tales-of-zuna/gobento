package web

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"example.com/shop/internal/domain"
)

// fakeStore is an in-memory domain.ItemStore for tests.
type fakeStore struct {
	mu    sync.Mutex
	items []domain.Item
	err   error
}

func (s *fakeStore) List(context.Context) ([]domain.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.items, s.err
}

func (s *fakeStore) Create(_ context.Context, name string) (domain.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return domain.Item{}, s.err
	}
	item := domain.Item{ID: int64(len(s.items) + 1), Name: name, CreatedAt: time.Now()}
	s.items = append(s.items, item)
	return item, nil
}

// serve sends one request to h. A non-nil form is posted form-encoded.
func serve(t *testing.T, h http.Handler, method, target string, form url.Values, header map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var body string
	if form != nil {
		body = form.Encode()
	}
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// htmx are the headers htmx sends with its requests.
var htmx = map[string]string{"HX-Request": "true"}

func wantHTML(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if mt, _, err := mime.ParseMediaType(rec.Header().Get("Content-Type")); err != nil || mt != "text/html" {
		t.Errorf("Content-Type = %q, want text/html", rec.Header().Get("Content-Type"))
	}
}

func TestHomeListsItems(t *testing.T) {
	store := &fakeStore{items: []domain.Item{{ID: 1, Name: "first", CreatedAt: time.Now()}}}
	rec := serve(t, NewHandler(store), http.MethodGet, "/", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want %d", rec.Code, http.StatusOK)
	}
	wantHTML(t, rec)
	for _, want := range []string{"first", `<script src="/static/htmx.min.js"`, `hx-post="/items"`} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("GET / does not contain %q:\n%s", want, rec.Body)
		}
	}
}

func TestHomeEscapesNames(t *testing.T) {
	store := &fakeStore{items: []domain.Item{{ID: 1, Name: "<script>alert(1)</script>"}}}
	rec := serve(t, NewHandler(store), http.MethodGet, "/", nil, nil)
	if strings.Contains(rec.Body.String(), "<script>alert(1)") || !strings.Contains(rec.Body.String(), "&lt;script&gt;") {
		t.Errorf("GET / does not escape item names:\n%s", rec.Body)
	}
}

func TestCreateWithHTMX(t *testing.T) {
	store := &fakeStore{}
	h := NewHandler(store)
	rec := serve(t, h, http.MethodPost, "/items", url.Values{"name": {" second "}}, htmx)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /items status = %d, want %d; body %s", rec.Code, http.StatusOK, rec.Body)
	}
	wantHTML(t, rec)
	for _, want := range []string{"<li>", "second", `hx-swap-oob="true"`} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("POST /items response does not contain %q:\n%s", want, rec.Body)
		}
	}
	if len(store.items) != 1 || store.items[0].Name != "second" {
		t.Errorf("store holds %+v, want one item named second", store.items)
	}
	if rec := serve(t, h, http.MethodGet, "/", nil, nil); !strings.Contains(rec.Body.String(), "second") {
		t.Errorf("GET / after POST does not list the new item:\n%s", rec.Body)
	}
}

func TestCreateWithoutJavaScript(t *testing.T) {
	store := &fakeStore{}
	rec := serve(t, NewHandler(store), http.MethodPost, "/items", url.Values{"name": {"plain"}}, nil)
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/" {
		t.Errorf("POST /items = %d to %q, want %d to /", rec.Code, rec.Header().Get("Location"), http.StatusSeeOther)
	}
	if len(store.items) != 1 {
		t.Errorf("store holds %d items, want 1", len(store.items))
	}
}

func TestCreateRejectsBlankName(t *testing.T) {
	for _, tt := range []struct {
		name   string
		header map[string]string
	}{
		{"htmx", htmx},
		{"plain form", nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{}
			rec := serve(t, NewHandler(store), http.MethodPost, "/items", url.Values{"name": {"  "}}, tt.header)
			if rec.Code != http.StatusUnprocessableEntity {
				t.Fatalf("POST /items status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
			}
			if !strings.Contains(rec.Body.String(), `role="alert"`) {
				t.Errorf("response shows no message:\n%s", rec.Body)
			}
			if tt.header != nil && rec.Header().Get("HX-Retarget") != "#new-item" {
				t.Errorf("HX-Retarget = %q, want #new-item", rec.Header().Get("HX-Retarget"))
			}
			if len(store.items) != 0 {
				t.Errorf("store holds %+v, want nothing", store.items)
			}
		})
	}
}

func TestCreateRejectsLargeForms(t *testing.T) {
	rec := serve(t, NewHandler(&fakeStore{}), http.MethodPost, "/items", url.Values{"name": {strings.Repeat("a", maxFormBytes)}}, nil)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("POST /items status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestCrossOriginPostIsRejected(t *testing.T) {
	store := &fakeStore{}
	rec := serve(t, NewHandler(store), http.MethodPost, "/items", url.Values{"name": {"x"}}, map[string]string{"Sec-Fetch-Site": "cross-site"})
	if rec.Code != http.StatusForbidden {
		t.Errorf("cross-site POST /items status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if len(store.items) != 0 {
		t.Errorf("store holds %+v, want nothing", store.items)
	}
}

func TestStoreErrors(t *testing.T) {
	store := &fakeStore{err: errors.New("disk full")}
	h := NewHandler(store)
	for _, rec := range []*httptest.ResponseRecorder{
		serve(t, h, http.MethodGet, "/", nil, nil),
		serve(t, h, http.MethodPost, "/items", url.Values{"name": {"x"}}, htmx),
	} {
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
		if strings.Contains(rec.Body.String(), "disk full") {
			t.Errorf("response exposes an internal error: %s", rec.Body)
		}
	}
}

func TestStaticFiles(t *testing.T) {
	h := NewHandler(&fakeStore{})
	for target, want := range map[string]string{
		"/static/htmx.min.js": "htmx",
		"/static/styles.css":  "--accent",
	} {
		rec := serve(t, h, http.MethodGet, target, nil, nil)
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), want) {
			t.Errorf("GET %s = %d, want 200 with %q", target, rec.Code, want)
		}
	}
}

func TestRouting(t *testing.T) {
	tests := []struct {
		method, target string
		want           int
	}{
		{http.MethodHead, "/", http.StatusOK},
		{http.MethodGet, "/items", http.StatusMethodNotAllowed},
		{http.MethodDelete, "/", http.StatusMethodNotAllowed},
		{http.MethodGet, "/nowhere", http.StatusNotFound},
		{http.MethodGet, "/static/missing.js", http.StatusNotFound},
	}
	for _, tt := range tests {
		rec := serve(t, NewHandler(&fakeStore{}), tt.method, tt.target, nil, nil)
		if rec.Code != tt.want {
			t.Errorf("%s %s status = %d, want %d", tt.method, tt.target, rec.Code, tt.want)
		}
	}
}
