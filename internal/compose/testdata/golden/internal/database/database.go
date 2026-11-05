// Package database opens the SQLite database and applies
// its schema. It uses modernc.org/sqlite, a pure-Go driver, so the
// application builds without cgo.
package database

import (
	"context"
	"database/sql"
	_ "embed" // for the schema
	"errors"
	"fmt"

	_ "modernc.org/sqlite" // registers the "sqlite" driver
)

// DefaultURL is the database the application uses when DATABASE_URL is
// unset: the file shop.db in the working directory.
const DefaultURL = "shop.db"

//go:embed schema.sql
var schema string

// Open opens the SQLite database at dsn, a file path or a modernc.org/sqlite
// connection string, and creates any missing tables.
func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", dsn, err)
	}
	// SQLite allows one writer at a time; a single connection avoids
	// "database is locked" errors and keeps in-memory databases shared.
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return nil, errors.Join(fmt.Errorf("apply schema: %w", err), db.Close())
	}
	return db, nil
}
