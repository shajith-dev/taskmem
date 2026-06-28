package service

import (
	"context"
	"fmt"

	"github.com/shajith-dev/taskmem/internal/db"
	"github.com/shajith-dev/taskmem/internal/models"
)

type TaskService struct {
	tasks      *db.TaskRepo
	taskGraph  *db.TaskGraphRepo
}

func NewTaskService(tasks *db.TaskRepo, taskGraph *db.TaskGraphRepo) *TaskService {
	return &TaskService{tasks: tasks, taskGraph: taskGraph}
}

func (s *TaskService) Create(ctx context.Context, description, model string, parentID *int64, useSubagent bool) (*models.Task, error) {
	t := &models.Task{
		Description: description,
		Model:       model,
		Parent:      parentID,
		UseSubagent: useSubagent,
		Status:      models.TaskStatusPending,
	}
	if t.Model == "" {
		t.Model = "inherit"
	}

	created, err := s.tasks.Create(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	return created, nil
}

type BulkTaskInput struct {
	Description string
	Model       string
	ParentID    *int64
	UseSubagent bool
}

func (s *TaskService) BulkCreate(ctx context.Context, inputs []BulkTaskInput) ([]*models.Task, error) {
	tasks := make([]*models.Task, len(inputs))
	for i, inp := range inputs {
		model := inp.Model
		if model == "" {
			model = "inherit"
		}
		tasks[i] = &models.Task{
			Description: inp.Description,
			Model:       model,
			Parent:      inp.ParentID,
			UseSubagent: inp.UseSubagent,
			Status:      models.TaskStatusPending,
		}
	}

	created, err := s.tasks.BulkCreate(ctx, tasks)
	if err != nil {
		return nil, fmt.Errorf("bulk create tasks: %w", err)
	}
	return created, nil
}

func (s *TaskService) Get(ctx context.Context, id int64) (*models.Task, error) {
	t, err := s.tasks.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	return t, nil
}

func (s *TaskService) ListChildren(ctx context.Context, parentID *int64) ([]*models.Task, error) {
	tasks, err := s.tasks.ListByParent(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("list children: %w", err)
	}
	return tasks, nil
}

func (s *TaskService) UpdateStatus(ctx context.Context, id int64, status models.TaskStatus) error {
	if !status.Valid() {
		return fmt.Errorf("invalid status %q", status)
	}
	if err := s.tasks.UpdateStatus(ctx, id, status); err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}

func (s *TaskService) Update(ctx context.Context, t *models.Task) (*models.Task, error) {
	updated, err := s.tasks.Update(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}
	return updated, nil
}

func (s *TaskService) SetScratchpad(ctx context.Context, id int64, text string) (*models.Task, error) {
	if err := s.tasks.UpdateScratchpad(ctx, id, &text); err != nil {
		return nil, fmt.Errorf("set scratchpad: %w", err)
	}
	t, err := s.tasks.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get task after set scratchpad: %w", err)
	}
	return t, nil
}

func (s *TaskService) AppendScratchpad(ctx context.Context, id int64, text string) (*models.Task, error) {
	current, err := s.tasks.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get task for append scratchpad: %w", err)
	}
	var newVal string
	if current.Scratchpad == nil || *current.Scratchpad == "" {
		newVal = text
	} else {
		newVal = *current.Scratchpad + "\n" + text
	}
	if err := s.tasks.UpdateScratchpad(ctx, id, &newVal); err != nil {
		return nil, fmt.Errorf("append scratchpad: %w", err)
	}
	current.Scratchpad = &newVal
	return current, nil
}

func (s *TaskService) Delete(ctx context.Context, id int64) error {
	if err := s.tasks.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}

func (s *TaskService) AddDependency(ctx context.Context, taskID, dependsOn int64) error {
	// Prevent circular dependency: check if dependsOn already depends on taskID.
	deps, err := s.taskGraph.GetDependencies(ctx, dependsOn)
	if err != nil {
		return fmt.Errorf("check circular dependency: %w", err)
	}
	for _, d := range deps {
		if d.DependsOn == taskID {
			return fmt.Errorf("circular dependency: task %d already depends on task %d", dependsOn, taskID)
		}
	}

	return s.taskGraph.AddDependency(ctx, taskID, dependsOn)
}

func (s *TaskService) RemoveDependency(ctx context.Context, taskID, dependsOn int64) error {
	return s.taskGraph.RemoveDependency(ctx, taskID, dependsOn)
}

func (s *TaskService) GetDependencies(ctx context.Context, taskID int64) ([]*models.TaskGraph, error) {
	return s.taskGraph.GetDependencies(ctx, taskID)
}

func (s *TaskService) GetDependents(ctx context.Context, taskID int64) ([]*models.TaskGraph, error) {
	return s.taskGraph.GetDependents(ctx, taskID)
}
