package service

import (
	"context"
	"fmt"

	"github.com/shajith-dev/taskmem/internal/db"
	"github.com/shajith-dev/taskmem/internal/models"
)

type FileService struct {
	files *db.FileRepo
}

func NewFileService(files *db.FileRepo) *FileService {
	return &FileService{files: files}
}

// AttachPath upserts the file by path and attaches it to the given task.
func (s *FileService) AttachPath(ctx context.Context, taskID int64, filePath string) (*models.File, error) {
	f, err := s.files.Upsert(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("upsert file: %w", err)
	}

	if err := s.files.AttachToTask(ctx, taskID, f.ID); err != nil {
		return nil, fmt.Errorf("attach file to task: %w", err)
	}

	return f, nil
}

func (s *FileService) BulkAttachPaths(ctx context.Context, taskID int64, filePaths []string) ([]*models.File, error) {
	files, err := s.files.BulkAttachToTask(ctx, taskID, filePaths)
	if err != nil {
		return nil, fmt.Errorf("bulk attach files: %w", err)
	}
	return files, nil
}

func (s *FileService) DetachPath(ctx context.Context, taskID int64, filePath string) error {
	f, err := s.files.Upsert(ctx, filePath)
	if err != nil {
		return fmt.Errorf("lookup file: %w", err)
	}

	return s.files.DetachFromTask(ctx, taskID, f.ID)
}

func (s *FileService) ListByTask(ctx context.Context, taskID int64) ([]*models.File, error) {
	files, err := s.files.ListByTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("list files by task: %w", err)
	}
	return files, nil
}
