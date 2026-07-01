package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/shajith-dev/taskmem/internal/models"
)

type FileRepo struct {
	db *sql.DB
}

func NewFileRepo(db *sql.DB) *FileRepo {
	return &FileRepo{db: db}
}

// Upsert inserts a file or returns the existing one if the path already exists.
func (r *FileRepo) Upsert(ctx context.Context, filePath string) (*models.File, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO files (file_path) VALUES (?)
		ON CONFLICT (file_path) DO UPDATE SET file_path = excluded.file_path
		RETURNING id, file_path
	`, filePath)

	f := &models.File{}
	if err := row.Scan(&f.ID, &f.FilePath); err != nil {
		return nil, fmt.Errorf("upsert file: %w", err)
	}
	return f, nil
}

func (r *FileRepo) GetByID(ctx context.Context, id int64) (*models.File, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, file_path FROM files WHERE id = ?`, id)

	f := &models.File{}
	if err := row.Scan(&f.ID, &f.FilePath); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get file by id: %w", err)
	}
	return f, nil
}

func (r *FileRepo) AttachToTask(ctx context.Context, taskID, fileID int64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO task_files (task_id, file_id) VALUES (?, ?)
		ON CONFLICT DO NOTHING
	`, taskID, fileID)
	if err != nil {
		return fmt.Errorf("attach file to task: %w", err)
	}
	return nil
}

// BulkAttachToTask upserts all file paths and links them to the task, wrapped in
// a single transaction so both operations succeed or roll back together.
func (r *FileRepo) BulkAttachToTask(ctx context.Context, taskID int64, filePaths []string) ([]*models.File, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin bulk attach tx: %w", err)
	}
	defer tx.Rollback()

	files := make([]*models.File, len(filePaths))
	for i, fp := range filePaths {
		row := tx.QueryRowContext(ctx, `
			INSERT INTO files (file_path) VALUES (?)
			ON CONFLICT (file_path) DO UPDATE SET file_path = excluded.file_path
			RETURNING id, file_path
		`, fp)

		f := &models.File{}
		if err := row.Scan(&f.ID, &f.FilePath); err != nil {
			return nil, fmt.Errorf("upsert file: %w", err)
		}
		files[i] = f
	}

	for _, f := range files {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO task_files (task_id, file_id) VALUES (?, ?)
			ON CONFLICT DO NOTHING
		`, taskID, f.ID); err != nil {
			return nil, fmt.Errorf("attach file: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit bulk attach: %w", err)
	}

	return files, nil
}

func (r *FileRepo) DetachFromTask(ctx context.Context, taskID, fileID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM task_files WHERE task_id = ? AND file_id = ?
	`, taskID, fileID)
	if err != nil {
		return fmt.Errorf("detach file from task: %w", err)
	}
	return nil
}

func (r *FileRepo) ListByTask(ctx context.Context, taskID int64) ([]*models.File, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT f.id, f.file_path
		FROM files f
		JOIN task_files tf ON tf.file_id = f.id
		WHERE tf.task_id = ?
		ORDER BY f.file_path
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("list files by task: %w", err)
	}
	defer rows.Close()

	var files []*models.File
	for rows.Next() {
		f := &models.File{}
		if err := rows.Scan(&f.ID, &f.FilePath); err != nil {
			return nil, fmt.Errorf("scan file row: %w", err)
		}
		files = append(files, f)
	}
	return files, rows.Err()
}
