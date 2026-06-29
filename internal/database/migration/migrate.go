package migration

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
	_ "github.com/go-sql-driver/mysql"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// Run applies all pending database migrations.
func Run(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("mysql"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}

// Down rolls back the last migration.
func Down(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.Down(db, "migrations"); err != nil {
		return fmt.Errorf("rollback migration: %w", err)
	}
	return nil
}

// Status reports the status of all migrations.
func Status(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	return goose.Status(db, "migrations")
}
