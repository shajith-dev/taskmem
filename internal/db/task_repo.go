package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/shajith-dev/taskmem/internal/models"
)

type TaskRepo struct {
	db *sql.DB
}

func NewTaskRepo(db *sql.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

func (r *TaskRepo) Create(ctx context.Context, t *models.Task) (*models.Task, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO tasks (parent, status, description, scratchpad, model, use_subagent)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id, parent, status, description, scratchpad, model, use_subagent, created_at, updated_at
	`, t.Parent, t.Status, t.Description, t.Scratchpad, t.Model, t.UseSubagent)

	return scanTask(row)
}

func (r *TaskRepo) GetByID(ctx context.Context, id int64) (*models.Task, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, parent, status, description, scratchpad, model, use_subagent, created_at, updated_at
		FROM tasks WHERE id = ?
	`, id)

	t, err := scanTask(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return t, nil
}

func (r *TaskRepo) ListByParent(ctx context.Context, parentID *int64) ([]*models.Task, error) {
	var rows *sql.Rows
	var err error

	if parentID == nil {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, parent, status, description, scratchpad, model, use_subagent, created_at, updated_at
			FROM tasks WHERE parent IS NULL ORDER BY id
		`)
	} else {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, parent, status, description, scratchpad, model, use_subagent, created_at, updated_at
			FROM tasks WHERE parent = ? ORDER BY id
		`, *parentID)
	}
	if err != nil {
		return nil, fmt.Errorf("list tasks by parent: %w", err)
	}
	defer rows.Close()

	return collectTasks(rows)
}

func (r *TaskRepo) UpdateStatus(ctx context.Context, id int64, status models.TaskStatus) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE tasks SET status = ?, updated_at = datetime('now') WHERE id = ?
	`, status, id)
	if err != nil {
		return fmt.Errorf("update task status: %w", err)
	}
	return checkAffected(res, "update task status")
}

func (r *TaskRepo) Update(ctx context.Context, t *models.Task) (*models.Task, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE tasks
		SET parent = ?, status = ?, description = ?, scratchpad = ?, model = ?, use_subagent = ?, updated_at = datetime('now')
		WHERE id = ?
		RETURNING id, parent, status, description, scratchpad, model, use_subagent, created_at, updated_at
	`, t.Parent, t.Status, t.Description, t.Scratchpad, t.Model, t.UseSubagent, t.ID)

	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return updated, nil
}

func (r *TaskRepo) BulkCreate(ctx context.Context, tasks []*models.Task) ([]*models.Task, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin bulk create tx: %w", err)
	}
	defer tx.Rollback()

	var created []*models.Task
	for _, t := range tasks {
		row := tx.QueryRowContext(ctx, `
			INSERT INTO tasks (parent, status, description, scratchpad, model, use_subagent)
			VALUES (?, ?, ?, ?, ?, ?)
			RETURNING id, parent, status, description, scratchpad, model, use_subagent, created_at, updated_at
		`, t.Parent, t.Status, t.Description, t.Scratchpad, t.Model, t.UseSubagent)

		ct, err := scanTask(row)
		if err != nil {
			return nil, fmt.Errorf("bulk create scan: %w", err)
		}
		created = append(created, ct)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit bulk create: %w", err)
	}

	return created, nil
}

func (r *TaskRepo) UpdateScratchpad(ctx context.Context, id int64, scratchpad *string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE tasks SET scratchpad = ?, updated_at = datetime('now') WHERE id = ?`, scratchpad, id)
	if err != nil {
		return fmt.Errorf("update task scratchpad: %w", err)
	}
	return checkAffected(res, "update task scratchpad")
}

func (r *TaskRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return checkAffected(res, "delete task")
}

// checkAffected returns ErrNotFound when a write matched no rows.
func checkAffected(res sql.Result, op string) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if n == 0 {
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

func collectTasks(rows *sql.Rows) ([]*models.Task, error) {
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
