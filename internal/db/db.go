package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// NewDB opens the SQLite database at path and verifies it is reachable.
//
// The DSN enables per-connection pragmas that matter for correctness on a
// single-user CLI: foreign_keys enforces ON DELETE CASCADE, WAL improves
// read/write concurrency, and busy_timeout avoids spurious "database is
// locked" errors. Writes are serialized via a single open connection.
func NewDB(ctx context.Context, path string) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)",
		path,
	)

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// SQLite allows a single writer; serialize access to keep transactions and
	// writes from contending for a connection.
	sqlDB.SetMaxOpenConns(1)

	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("unable to reach database: %w", err)
	}

	return sqlDB, nil
}
