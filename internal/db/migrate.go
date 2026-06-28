package db

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// Migrate applies all up migrations from fsys. dir is the path to the migration
// files within fsys (e.g. "migrations").
func Migrate(ctx context.Context, databaseURL string, fsys fs.FS, dir string) error {
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open sql db: %w", err)
	}
	defer sqlDB.Close()

	goose.SetBaseFS(fsys)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.UpContext(ctx, sqlDB, dir); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
