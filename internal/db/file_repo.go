package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shajith-dev/taskmem/internal/models"
)

type FileRepo struct {
	pool *pgxpool.Pool
}

func NewFileRepo(pool *pgxpool.Pool) *FileRepo {
	return &FileRepo{pool: pool}
}

// Upsert inserts a file or returns the existing one if the path already exists.
func (r *FileRepo) Upsert(ctx context.Context, filePath string) (*models.File, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO files (file_path) VALUES ($1)
		ON CONFLICT (file_path) DO UPDATE SET file_path = EXCLUDED.file_path
		RETURNING id, file_path
	`, filePath)

	f := &models.File{}
	if err := row.Scan(&f.ID, &f.FilePath); err != nil {
		return nil, fmt.Errorf("upsert file: %w", err)
	}
	return f, nil
}

func (r *FileRepo) GetByID(ctx context.Context, id int64) (*models.File, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, file_path FROM files WHERE id = $1`, id)

	f := &models.File{}
	if err := row.Scan(&f.ID, &f.FilePath); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get file by id: %w", err)
	}
	return f, nil
}

func (r *FileRepo) AttachToTask(ctx context.Context, taskID, fileID int64) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO task_files (task_id, file_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, taskID, fileID)
	if err != nil {
		return fmt.Errorf("attach file to task: %w", err)
	}
	return nil
}

// BulkAttachToTask upserts all file paths and links them to the task in two round-trips,
// wrapped in a single transaction so both operations succeed or roll back together.
func (r *FileRepo) BulkAttachToTask(ctx context.Context, taskID int64, filePaths []string) ([]*models.File, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin bulk attach tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Round-trip 1: upsert all file paths in a batch.
	upsertBatch := &pgx.Batch{}
	for _, fp := range filePaths {
		upsertBatch.Queue(`
			INSERT INTO files (file_path) VALUES ($1)
			ON CONFLICT (file_path) DO UPDATE SET file_path = EXCLUDED.file_path
			RETURNING id, file_path
		`, fp)
	}

	br := tx.SendBatch(ctx, upsertBatch)
	files := make([]*models.File, len(filePaths))
	for i := range filePaths {
		f := &models.File{}
		if err := br.QueryRow().Scan(&f.ID, &f.FilePath); err != nil {
			br.Close()
			return nil, fmt.Errorf("upsert file batch: %w", err)
		}
		files[i] = f
	}
	if err := br.Close(); err != nil {
		return nil, fmt.Errorf("close upsert file batch: %w", err)
	}

	// Round-trip 2: attach all file IDs to the task in a batch.
	attachBatch := &pgx.Batch{}
	for _, f := range files {
		attachBatch.Queue(`
			INSERT INTO task_files (task_id, file_id) VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, taskID, f.ID)
	}

	abr := tx.SendBatch(ctx, attachBatch)
	for range files {
		if _, err := abr.Exec(); err != nil {
			abr.Close()
			return nil, fmt.Errorf("attach file batch: %w", err)
		}
	}
	if err := abr.Close(); err != nil {
		return nil, fmt.Errorf("close attach file batch: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit bulk attach: %w", err)
	}

	return files, nil
}

func (r *FileRepo) DetachFromTask(ctx context.Context, taskID, fileID int64) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM task_files WHERE task_id = $1 AND file_id = $2
	`, taskID, fileID)
	if err != nil {
		return fmt.Errorf("detach file from task: %w", err)
	}
	return nil
}

func (r *FileRepo) ListByTask(ctx context.Context, taskID int64) ([]*models.File, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT f.id, f.file_path
		FROM files f
		JOIN task_files tf ON tf.file_id = f.id
		WHERE tf.task_id = $1
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
