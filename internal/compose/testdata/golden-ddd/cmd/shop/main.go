// Command shop serves the shop HTTP API.
//
// Configuration comes from the environment:
//
//	PORT          port to listen on (default 8080)
//	DATABASE_URL  database connection string (default database.DefaultURL)
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/shop/internal/adapters/database"
	"example.com/shop/internal/adapters/httpapi"
	"example.com/shop/internal/adapters/persistence"
	"example.com/shop/internal/app"
	"example.com/shop/internal/domain"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

// run wires the application together and serves HTTP until the process is
// interrupted, then shuts down gracefully. It is the only place that knows
// every layer: the HTTP adapter gets the application service, which stores
// items through the persistence adapter.
func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Open(ctx, getenv("DATABASE_URL", database.DefaultURL))
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("close database", "error", err)
		}
	}()

	srv := &http.Server{
		Addr:              ":" + getenv("PORT", "8080"),
		Handler:           newHandler(app.NewItemService(persistence.New(db))),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errc := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", srv.Addr)
		errc <- srv.ListenAndServe()
	}()

	select {
	case err := <-errc:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shut down: %w", err)
	}
	// Wait for the server goroutine, which returns ErrServerClosed.
	if err := <-errc; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// newHandler returns the application's HTTP handler: the API router, which
// reaches service, the application service, only through the
// domain.ItemStore port.
func newHandler(service domain.ItemStore) http.Handler {
	return httpapi.NewRouter(service)
}

// getenv returns the value of the environment variable key, or fallback if
// it is unset or empty.
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
