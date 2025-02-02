package db

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	_ "embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

func NewSQLite(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return db, fmt.Errorf("sql.NewSQLite failed: %w", err)
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

//go:embed migrations/*.sql
var migrations embed.FS

func NewSQLiteWithMigrations(path string) (*sql.DB, error) {
	db, err := NewSQLite(path)
	if err != nil {
		return nil, err
	}

	fsSource, err := iofs.New(migrations, "migrations")
	if err != nil {
		return nil, fmt.Errorf("NewSQLiteWithMigrations creating FS source failed. %w", err)
	}

	dbDriver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return nil, fmt.Errorf("NewSQLiteWithMigrations creating DB driver failed. %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", fsSource, "", dbDriver)
	if err != nil {
		return nil, fmt.Errorf("NewSQLiteWithMigrations creating migrate instance failed. %w", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return db, nil
		}
		return nil, fmt.Errorf("NewSQLiteWithMigrations migration failed. %w", err)
	}

	return db, nil
}
