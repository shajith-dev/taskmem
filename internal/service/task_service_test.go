package service_test

import (
	"context"
	"strings"
	"testing"

	"github.com/shajith-dev/taskmem/internal/db"
	"github.com/shajith-dev/taskmem/internal/models"
	"github.com/shajith-dev/taskmem/internal/service"
	"github.com/shajith-dev/taskmem/internal/testutil"
)

func newTaskService(t *testing.T) (*service.TaskService, context.Context) {
	t.Helper()
	sqlDB := testutil.NewDB(t)
	svc := service.NewTaskService(db.NewTaskRepo(sqlDB), db.NewTaskGraphRepo(sqlDB))
	return svc, context.Background()
}

func TestServiceCreateDefaultsModel(t *testing.T) {
	svc, ctx := newTaskService(t)

	// Empty model should default to "inherit".
	task, err := svc.Create(ctx, "do a thing", "", nil, false)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if task.Model != "inherit" {
		t.Errorf("model = %q, want inherit", task.Model)
	}
}

func TestServiceUpdateStatusValidation(t *testing.T) {
	svc, ctx := newTaskService(t)
	task, err := svc.Create(ctx, "t", "inherit", nil, false)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := svc.UpdateStatus(ctx, task.ID, models.TaskStatus("BOGUS")); err == nil {
		t.Error("expected invalid status to be rejected")
	}
	if err := svc.UpdateStatus(ctx, task.ID, models.TaskStatusCompleted); err != nil {
		t.Errorf("valid status rejected: %v", err)
	}
}

func TestServiceScratchpadAppend(t *testing.T) {
	svc, ctx := newTaskService(t)
	task, err := svc.Create(ctx, "t", "inherit", nil, false)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err := svc.SetScratchpad(ctx, task.ID, "line one"); err != nil {
		t.Fatalf("set scratchpad: %v", err)
	}
	got, err := svc.AppendScratchpad(ctx, task.ID, "line two")
	if err != nil {
		t.Fatalf("append scratchpad: %v", err)
	}
	if got.Scratchpad == nil {
		t.Fatal("scratchpad is nil after append")
	}
	if want := "line one\nline two"; *got.Scratchpad != want {
		t.Errorf("scratchpad = %q, want %q", *got.Scratchpad, want)
	}

	// Appending to an empty scratchpad should not prepend a newline.
	task2, err := svc.Create(ctx, "t2", "inherit", nil, false)
	if err != nil {
		t.Fatalf("create t2: %v", err)
	}
	got2, err := svc.AppendScratchpad(ctx, task2.ID, "first")
	if err != nil {
		t.Fatalf("append to empty: %v", err)
	}
	if got2.Scratchpad == nil || strings.HasPrefix(*got2.Scratchpad, "\n") {
		t.Errorf("append to empty scratchpad = %q, want no leading newline", *got2.Scratchpad)
	}
}

func TestServiceCircularDependencyRejected(t *testing.T) {
	svc, ctx := newTaskService(t)

	a, err := svc.Create(ctx, "a", "inherit", nil, false)
	if err != nil {
		t.Fatalf("create a: %v", err)
	}
	b, err := svc.Create(ctx, "b", "inherit", nil, false)
	if err != nil {
		t.Fatalf("create b: %v", err)
	}

	// a depends on b is fine.
	if err := svc.AddDependency(ctx, a.ID, b.ID); err != nil {
		t.Fatalf("add a->b: %v", err)
	}
	// b depends on a would create a direct cycle and must be rejected.
	if err := svc.AddDependency(ctx, b.ID, a.ID); err == nil {
		t.Error("expected direct circular dependency to be rejected")
	}
}

func TestServiceUpdatePreservesAndChanges(t *testing.T) {
	svc, ctx := newTaskService(t)
	task, err := svc.Create(ctx, "orig", "inherit", nil, false)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	task.Description = "changed"
	task.UseSubagent = true
	updated, err := svc.Update(ctx, task)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Description != "changed" || !updated.UseSubagent {
		t.Errorf("update result = %+v", updated)
	}

	// Confirm it persisted.
	got, err := svc.Get(ctx, task.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Description != "changed" || !got.UseSubagent {
		t.Errorf("persisted = %+v", got)
	}
}
