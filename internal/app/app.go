package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shajith-dev/taskmem/internal/config"
	"github.com/shajith-dev/taskmem/internal/db"
	"github.com/shajith-dev/taskmem/internal/service"
)

type App struct {
	Pool     *pgxpool.Pool
	Tasks    *service.TaskService
	Files    *service.FileService
}

func New(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	taskRepo      := db.NewTaskRepo(pool)
	taskGraphRepo := db.NewTaskGraphRepo(pool)
	fileRepo      := db.NewFileRepo(pool)

	return &App{
		Pool:  pool,
		Tasks: service.NewTaskService(taskRepo, taskGraphRepo),
		Files: service.NewFileService(fileRepo),
	}, nil
}

func (a *App) Close() {
	a.Pool.Close()
}
