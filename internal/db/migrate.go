package db

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
)

// MigrateDB applies all up migrations from fsys against an already-open
// database. dir is the path to the migration files within fsys (e.g.
// "migrations").
func MigrateDB(ctx context.Context, sqlDB *sql.DB, fsys fs.FS, dir string) error {
	// Migrations run automatically on startup; keep goose quiet so it doesn't
	// pollute the CLI's (machine-readable) output.
	goose.SetLogger(goose.NopLogger())
	goose.SetBaseFS(fsys)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.UpContext(ctx, sqlDB, dir); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}

// Migrate opens the database at path, applies all up migrations, and closes it.
// Used by the standalone `migrate` command; normal commands migrate on startup.
func Migrate(ctx context.Context, path string, fsys fs.FS, dir string) error {
	sqlDB, err := NewDB(ctx, path)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	return MigrateDB(ctx, sqlDB, fsys, dir)
}
