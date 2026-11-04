// Package store implements the models
// interfaces with GORM. It uses the connection and the tables of package
// database: the schema is defined only in
// internal/database/schema.sql, and this package
// never creates or migrates tables.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"example.com/shop/internal/models"
)

// item maps a row of the items table. The database assigns id and
// created_at (see the schema), so GORM never writes created_at and reads
// both back after an insert.
type item struct {
	ID        int64     `gorm:"primaryKey"`
	Name      string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"->;default:CURRENT_TIMESTAMP"`
}

// TableName returns the table in the schema that item maps to.
func (item) TableName() string { return "items" }

// itemStore is a models.ItemStore backed by GORM.
type itemStore struct {
	db *gorm.DB
	// err is the error from setting up GORM, returned by every method.
	err error
}

// New returns a models.ItemStore that stores items in db,
// an open connection whose schema is already applied. GORM uses db as is
// and opens no connection of its own. If GORM cannot be set up on db, every
// method of the returned store reports that error.
func New(db *sql.DB) models.ItemStore {
	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{
		// The database package has already connected and applied the schema.
		DisableAutomaticPing: true,
		// Errors are returned to the callers, which log them.
		Logger: logger.Discard,
	})
	if err != nil {
		return &itemStore{err: fmt.Errorf("set up gorm: %w", err)}
	}
	return &itemStore{db: gdb}
}

func (s *itemStore) List(ctx context.Context) ([]models.Item, error) {
	if s.err != nil {
		return nil, s.err
	}
	var rows []item
	if err := s.db.WithContext(ctx).Order("id").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	items := make([]models.Item, len(rows))
	for i, row := range rows {
		items[i] = toModel(row)
	}
	return items, nil
}

func (s *itemStore) Create(ctx context.Context, name string) (models.Item, error) {
	if s.err != nil {
		return models.Item{}, s.err
	}
	row := item{Name: name}
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		return models.Item{}, fmt.Errorf("create item: %w", err)
	}
	return toModel(row), nil
}

func toModel(row item) models.Item {
	return models.Item{ID: row.ID, Name: row.Name, CreatedAt: row.CreatedAt}
}
