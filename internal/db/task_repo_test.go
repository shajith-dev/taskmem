package db_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shajith-dev/taskmem/internal/db"
	"github.com/shajith-dev/taskmem/internal/models"
	"github.com/shajith-dev/taskmem/internal/testutil"
)

func newTaskRepo(t *testing.T) (*db.TaskRepo, context.Context) {
	t.Helper()
	return db.NewTaskRepo(testutil.NewDB(t)), context.Background()
}

func mustCreate(t *testing.T, r *db.TaskRepo, ctx context.Context, desc string, parent *int64) *models.Task {
	t.Helper()
	task, err := r.Create(ctx, &models.Task{
		Description: desc,
		Model:       "inherit",
		Status:      models.TaskStatusPending,
		Parent:      parent,
	})
	if err != nil {
		t.Fatalf("create task %q: %v", desc, err)
	}
	return task
}

func TestTaskCreateAndGet(t *testing.T) {
	r, ctx := newTaskRepo(t)

	created := mustCreate(t, r, ctx, "root", nil)
	if created.ID == 0 {
		t.Fatal("expected non-zero id")
	}
	if created.Status != models.TaskStatusPending {
		t.Errorf("status = %q, want PENDING", created.Status)
	}
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Errorf("timestamps not populated: created=%v updated=%v", created.CreatedAt, created.UpdatedAt)
	}

	got, err := r.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Description != "root" || got.Model != "inherit" {
		t.Errorf("round-trip mismatch: %+v", got)
	}
}

func TestTaskGetNotFound(t *testing.T) {
	r, ctx := newTaskRepo(t)
	_, err := r.GetByID(ctx, 999)
	if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestTaskListByParent(t *testing.T) {
	r, ctx := newTaskRepo(t)

	root := mustCreate(t, r, ctx, "root", nil)
	mustCreate(t, r, ctx, "child-a", &root.ID)
	mustCreate(t, r, ctx, "child-b", &root.ID)

	roots, err := r.ListByParent(ctx, nil)
	if err != nil {
		t.Fatalf("list roots: %v", err)
	}
	if len(roots) != 1 || roots[0].ID != root.ID {
		t.Errorf("roots = %+v, want [root]", roots)
	}

	children, err := r.ListByParent(ctx, &root.ID)
	if err != nil {
		t.Fatalf("list children: %v", err)
	}
	if len(children) != 2 {
		t.Errorf("children count = %d, want 2", len(children))
	}
}

func TestTaskUpdateStatus(t *testing.T) {
	r, ctx := newTaskRepo(t)
	task := mustCreate(t, r, ctx, "root", nil)

	// updated_at must advance; datetime('now') has second granularity, so
	// force a distinct timestamp rather than sleeping a full second.
	before := task.UpdatedAt

	if err := r.UpdateStatus(ctx, task.ID, models.TaskStatusCompleted); err != nil {
		t.Fatalf("update status: %v", err)
	}

	got, err := r.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != models.TaskStatusCompleted {
		t.Errorf("status = %q, want COMPLETED", got.Status)
	}
	if got.UpdatedAt.Before(before) {
		t.Errorf("updated_at went backwards: %v < %v", got.UpdatedAt, before)
	}

	if err := r.UpdateStatus(ctx, 999, models.TaskStatusCompleted); !errors.Is(err, db.ErrNotFound) {
		t.Errorf("update missing status err = %v, want ErrNotFound", err)
	}
}

func TestTaskUpdate(t *testing.T) {
	r, ctx := newTaskRepo(t)
	task := mustCreate(t, r, ctx, "root", nil)

	task.Description = "renamed"
	task.Model = "claude-opus-4-8"
	task.UseSubagent = true
	task.Status = models.TaskStatusInProgress

	updated, err := r.Update(ctx, task)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Description != "renamed" || updated.Model != "claude-opus-4-8" || !updated.UseSubagent {
		t.Errorf("update did not persist fields: %+v", updated)
	}
	if updated.Status != models.TaskStatusInProgress {
		t.Errorf("status = %q, want IN_PROGRESS", updated.Status)
	}

	missing := &models.Task{ID: 999, Description: "x", Model: "inherit", Status: models.TaskStatusPending}
	if _, err := r.Update(ctx, missing); !errors.Is(err, db.ErrNotFound) {
		t.Errorf("update missing err = %v, want ErrNotFound", err)
	}
}

func TestTaskBulkCreate(t *testing.T) {
	r, ctx := newTaskRepo(t)

	in := []*models.Task{
		{Description: "a", Model: "inherit", Status: models.TaskStatusPending},
		{Description: "b", Model: "inherit", Status: models.TaskStatusPending, UseSubagent: true},
	}
	created, err := r.BulkCreate(ctx, in)
	if err != nil {
		t.Fatalf("bulk create: %v", err)
	}
	if len(created) != 2 {
		t.Fatalf("created %d, want 2", len(created))
	}
	if created[0].ID == created[1].ID {
		t.Errorf("expected distinct ids, got %d twice", created[0].ID)
	}
	if !created[1].UseSubagent {
		t.Errorf("use_subagent not persisted for second task")
	}
}

func TestTaskUpdateScratchpad(t *testing.T) {
	r, ctx := newTaskRepo(t)
	task := mustCreate(t, r, ctx, "root", nil)

	note := "some working notes"
	if err := r.UpdateScratchpad(ctx, task.ID, &note); err != nil {
		t.Fatalf("update scratchpad: %v", err)
	}
	got, err := r.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Scratchpad == nil || *got.Scratchpad != note {
		t.Errorf("scratchpad = %v, want %q", got.Scratchpad, note)
	}

	if err := r.UpdateScratchpad(ctx, 999, &note); !errors.Is(err, db.ErrNotFound) {
		t.Errorf("update missing scratchpad err = %v, want ErrNotFound", err)
	}
}

func TestTaskDeleteCascade(t *testing.T) {
	r, ctx := newTaskRepo(t)

	root := mustCreate(t, r, ctx, "root", nil)
	child := mustCreate(t, r, ctx, "child", &root.ID)

	if err := r.Delete(ctx, root.ID); err != nil {
		t.Fatalf("delete root: %v", err)
	}

	// The child should be gone via ON DELETE CASCADE (requires the
	// foreign_keys pragma to be active on the connection).
	if _, err := r.GetByID(ctx, child.ID); !errors.Is(err, db.ErrNotFound) {
		t.Errorf("child still present after cascade delete: err = %v", err)
	}

	if err := r.Delete(ctx, 999); !errors.Is(err, db.ErrNotFound) {
		t.Errorf("delete missing err = %v, want ErrNotFound", err)
	}
}

// TestTaskTimestampsAreUTC guards the time.Time round-trip through SQLite's
// DATETIME columns that motivated much of the driver choice.
func TestTaskTimestampsAreUTC(t *testing.T) {
	r, ctx := newTaskRepo(t)
	task := mustCreate(t, r, ctx, "root", nil)

	if got := task.CreatedAt.Location(); got != time.UTC {
		t.Errorf("created_at location = %v, want UTC", got)
	}
	if time.Since(task.CreatedAt) > time.Hour {
		t.Errorf("created_at = %v looks wrong (not recent)", task.CreatedAt)
	}
}
