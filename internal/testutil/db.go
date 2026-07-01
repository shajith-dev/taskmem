// Package testutil provides shared helpers for tests that need a real,
// migrated SQLite database.
package testutil

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/shajith-dev/taskmem/internal/db"
)

// NewDB returns a freshly migrated SQLite database backed by a temp file that
// is removed automatically when the test finishes. Each call gets an isolated
// database, so tests never share state.
func NewDB(t *testing.T) *sql.DB {
	t.Helper()

	path := filepath.Join(t.TempDir(), "test.db")

	sqlDB, err := db.NewDB(context.Background(), path)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	// Mirror production: migrations are rooted at the repo and applied from the
	// "migrations" directory.
	if err := db.MigrateDB(context.Background(), sqlDB, os.DirFS(repoRoot()), "migrations"); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	return sqlDB
}

// repoRoot locates the repository root relative to this source file so tests
// work regardless of the current working directory.
func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	// this file lives at <root>/internal/testutil/db.go
	return filepath.Join(filepath.Dir(file), "..", "..")
}
