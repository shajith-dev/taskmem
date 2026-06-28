package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shajith-dev/taskmem/internal/models"
)

type TaskRepo struct {
	pool *pgxpool.Pool
}

func NewTaskRepo(pool *pgxpool.Pool) *TaskRepo {
	return &TaskRepo{pool: pool}
}

func (r *TaskRepo) Create(ctx context.Context, t *models.Task) (*models.Task, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO tasks (parent, status, description, scratchpad, model, use_subagent)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, parent, status, description, scratchpad, model, use_subagent, created_at, updated_at
	`, t.Parent, t.Status, t.Description, t.Scratchpad, t.Model, t.UseSubagent)

	return scanTask(row)
}

func (r *TaskRepo) GetByID(ctx context.Context, id int64) (*models.Task, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, parent, status, description, scratchpad, model, use_subagent, created_at, updated_at
		FROM tasks WHERE id = $1
	`, id)

	t, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return t, nil
}

func (r *TaskRepo) ListByParent(ctx context.Context, parentID *int64) ([]*models.Task, error) {
	var rows pgx.Rows
	var err error

	if parentID == nil {
		rows, err = r.pool.Query(ctx, `
			SELECT id, parent, status, description, scratchpad, model, use_subagent, created_at, updated_at
			FROM tasks WHERE parent IS NULL ORDER BY id
		`)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, parent, status, description, scratchpad, model, use_subagent, created_at, updated_at
			FROM tasks WHERE parent = $1 ORDER BY id
		`, *parentID)
	}
	if err != nil {
		return nil, fmt.Errorf("list tasks by parent: %w", err)
	}
	defer rows.Close()

	return collectTasks(rows)
}

func (r *TaskRepo) UpdateStatus(ctx context.Context, id int64, status models.TaskStatus) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE tasks SET status = $1 WHERE id = $2
	`, status, id)
	if err != nil {
		return fmt.Errorf("update task status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *TaskRepo) Update(ctx context.Context, t *models.Task) (*models.Task, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE tasks
		SET parent = $1, status = $2, description = $3, scratchpad = $4, model = $5, use_subagent = $6
		WHERE id = $7
		RETURNING id, parent, status, description, scratchpad, model, use_subagent, created_at, updated_at
	`, t.Parent, t.Status, t.Description, t.Scratchpad, t.Model, t.UseSubagent, t.ID)

	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return updated, nil
}

func (r *TaskRepo) BulkCreate(ctx context.Context, tasks []*models.Task) ([]*models.Task, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin bulk create tx: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	for _, t := range tasks {
		batch.Queue(`
			INSERT INTO tasks (parent, status, description, scratchpad, model, use_subagent)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, parent, status, description, scratchpad, model, use_subagent, created_at, updated_at
		`, t.Parent, t.Status, t.Description, t.Scratchpad, t.Model, t.UseSubagent)
	}

	br := tx.SendBatch(ctx, batch)

	var created []*models.Task
	for range tasks {
		row := br.QueryRow()
		t, err := scanTask(row)
		if err != nil {
			br.Close()
			return nil, fmt.Errorf("bulk create scan: %w", err)
		}
		created = append(created, t)
	}

	if err := br.Close(); err != nil {
		return nil, fmt.Errorf("close bulk create batch: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit bulk create: %w", err)
	}

	return created, nil
}

func (r *TaskRepo) UpdateScratchpad(ctx context.Context, id int64, scratchpad *string) error {
	tag, err := r.pool.Exec(ctx, `UPDATE tasks SET scratchpad = $1 WHERE id = $2`, scratchpad, id)
	if err != nil {
		return fmt.Errorf("update task scratchpad: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *TaskRepo) Delete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// scanTask scans a single row into a Task.
type scanner interface {
	Scan(dest ...any) error
}

func scanTask(row scanner) (*models.Task, error) {
	t := &models.Task{}
	err := row.Scan(
		&t.ID, &t.Parent, &t.Status, &t.Description,
		&t.Scratchpad, &t.Model, &t.UseSubagent,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan task: %w", err)
	}
	return t, nil
}

func collectTasks(rows pgx.Rows) ([]*models.Task, error) {
	var tasks []*models.Task
	for rows.Next() {
		t := &models.Task{}
		err := rows.Scan(
			&t.ID, &t.Parent, &t.Status, &t.Description,
			&t.Scratchpad, &t.Model, &t.UseSubagent,
			&t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan task row: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
