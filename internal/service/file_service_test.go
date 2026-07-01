package service_test

import (
	"context"
	"testing"

	"github.com/shajith-dev/taskmem/internal/db"
	"github.com/shajith-dev/taskmem/internal/service"
	"github.com/shajith-dev/taskmem/internal/testutil"
)

func TestFileServiceAttachAndList(t *testing.T) {
	sqlDB := testutil.NewDB(t)
	ctx := context.Background()
	tasks := service.NewTaskService(db.NewTaskRepo(sqlDB), db.NewTaskGraphRepo(sqlDB))
	fileSvc := service.NewFileService(db.NewFileRepo(sqlDB))

	task, err := tasks.Create(ctx, "root", "inherit", nil, false)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if _, err := fileSvc.AttachPath(ctx, task.ID, "src/a.go"); err != nil {
		t.Fatalf("attach: %v", err)
	}
	list, err := fileSvc.ListByTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].FilePath != "src/a.go" {
		t.Errorf("list = %+v, want one file", list)
	}

	if err := fileSvc.DetachPath(ctx, task.ID, "src/a.go"); err != nil {
		t.Fatalf("detach: %v", err)
	}
	list, err = fileSvc.ListByTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("list after detach: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("list = %+v, want empty", list)
	}
}

func TestFileServiceBulkAttach(t *testing.T) {
	sqlDB := testutil.NewDB(t)
	ctx := context.Background()
	tasks := service.NewTaskService(db.NewTaskRepo(sqlDB), db.NewTaskGraphRepo(sqlDB))
	fileSvc := service.NewFileService(db.NewFileRepo(sqlDB))

	task, err := tasks.Create(ctx, "root", "inherit", nil, false)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if _, err := fileSvc.BulkAttachPaths(ctx, task.ID, []string{"a.go", "b.go"}); err != nil {
		t.Fatalf("bulk attach: %v", err)
	}
	list, err := fileSvc.ListByTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("files = %d, want 2", len(list))
	}
}
