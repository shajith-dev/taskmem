package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shajith-dev/taskmem/internal/models"
)

type TaskGraphRepo struct {
	pool *pgxpool.Pool
}

func NewTaskGraphRepo(pool *pgxpool.Pool) *TaskGraphRepo {
	return &TaskGraphRepo{pool: pool}
}

func (r *TaskGraphRepo) AddDependency(ctx context.Context, taskID, dependsOn int64) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO task_graph (task_id, depends_on) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, taskID, dependsOn)
	if err != nil {
		return fmt.Errorf("add dependency: %w", err)
	}
	return nil
}

func (r *TaskGraphRepo) RemoveDependency(ctx context.Context, taskID, dependsOn int64) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM task_graph WHERE task_id = $1 AND depends_on = $2
	`, taskID, dependsOn)
	if err != nil {
		return fmt.Errorf("remove dependency: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// GetDependencies returns all tasks that taskID depends on.
func (r *TaskGraphRepo) GetDependencies(ctx context.Context, taskID int64) ([]*models.TaskGraph, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT task_id, depends_on FROM task_graph WHERE task_id = $1
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("get dependencies: %w", err)
	}
	defer rows.Close()

	return collectTaskGraph(rows)
}

// GetDependents returns all tasks that depend on taskID.
func (r *TaskGraphRepo) GetDependents(ctx context.Context, taskID int64) ([]*models.TaskGraph, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT task_id, depends_on FROM task_graph WHERE depends_on = $1
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("get dependents: %w", err)
	}
	defer rows.Close()

	return collectTaskGraph(rows)
}

func collectTaskGraph(rows pgx.Rows) ([]*models.TaskGraph, error) {
	var edges []*models.TaskGraph
	for rows.Next() {
		e := &models.TaskGraph{}
		if err := rows.Scan(&e.TaskID, &e.DependsOn); err != nil {
			return nil, fmt.Errorf("scan task_graph row: %w", err)
		}
		edges = append(edges, e)
	}
	return edges, rows.Err()
}
