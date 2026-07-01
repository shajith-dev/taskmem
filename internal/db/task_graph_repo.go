package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/shajith-dev/taskmem/internal/models"
)

type TaskGraphRepo struct {
	db *sql.DB
}

func NewTaskGraphRepo(db *sql.DB) *TaskGraphRepo {
	return &TaskGraphRepo{db: db}
}

func (r *TaskGraphRepo) AddDependency(ctx context.Context, taskID, dependsOn int64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO task_graph (task_id, depends_on) VALUES (?, ?)
		ON CONFLICT DO NOTHING
	`, taskID, dependsOn)
	if err != nil {
		return fmt.Errorf("add dependency: %w", err)
	}
	return nil
}

func (r *TaskGraphRepo) RemoveDependency(ctx context.Context, taskID, dependsOn int64) error {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM task_graph WHERE task_id = ? AND depends_on = ?
	`, taskID, dependsOn)
	if err != nil {
		return fmt.Errorf("remove dependency: %w", err)
	}
	return checkAffected(res, "remove dependency")
}

// GetDependencies returns all tasks that taskID depends on.
func (r *TaskGraphRepo) GetDependencies(ctx context.Context, taskID int64) ([]*models.TaskGraph, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT task_id, depends_on FROM task_graph WHERE task_id = ?
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("get dependencies: %w", err)
	}
	defer rows.Close()

	return collectTaskGraph(rows)
}

// GetDependents returns all tasks that depend on taskID.
func (r *TaskGraphRepo) GetDependents(ctx context.Context, taskID int64) ([]*models.TaskGraph, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT task_id, depends_on FROM task_graph WHERE depends_on = ?
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("get dependents: %w", err)
	}
	defer rows.Close()

	return collectTaskGraph(rows)
}

func collectTaskGraph(rows *sql.Rows) ([]*models.TaskGraph, error) {
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
