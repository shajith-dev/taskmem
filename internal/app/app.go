package app

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/shajith-dev/taskmem/internal/config"
	"github.com/shajith-dev/taskmem/internal/db"
	"github.com/shajith-dev/taskmem/internal/service"
)

type App struct {
	DB    *sql.DB
	Tasks *service.TaskService
	Files *service.FileService
}

// New loads config, opens the SQLite database (creating its parent directory if
// needed), applies any pending migrations, and wires up the services. Callers
// never have to run migrations or configure a database manually.
func New(ctx context.Context, migrationsFS fs.FS) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(cfg.DatabaseURL), 0o755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	sqlDB, err := db.NewDB(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.MigrateDB(ctx, sqlDB, migrationsFS, "migrations"); err != nil {
		sqlDB.Close()
		return nil, err
	}

	taskRepo := db.NewTaskRepo(sqlDB)
	taskGraphRepo := db.NewTaskGraphRepo(sqlDB)
	fileRepo := db.NewFileRepo(sqlDB)

	return &App{
		DB:    sqlDB,
		Tasks: service.NewTaskService(taskRepo, taskGraphRepo),
		Files: service.NewFileService(fileRepo),
	}, nil
}

func (a *App) Close() {
	a.DB.Close()
}
