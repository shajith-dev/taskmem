package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/shajith-dev/taskmem/internal/db"
	"github.com/shajith-dev/taskmem/internal/testutil"
)

func TestFileUpsertIsIdempotent(t *testing.T) {
	sqlDB := testutil.NewDB(t)
	ctx := context.Background()
	files := db.NewFileRepo(sqlDB)

	first, err := files.Upsert(ctx, "src/a.go")
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	second, err := files.Upsert(ctx, "src/a.go")
	if err != nil {
		t.Fatalf("re-upsert: %v", err)
	}
	if first.ID != second.ID {
		t.Errorf("upsert created a duplicate: %d != %d", first.ID, second.ID)
	}

	got, err := files.GetByID(ctx, first.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.FilePath != "src/a.go" {
		t.Errorf("path = %q, want src/a.go", got.FilePath)
	}

	if _, err := files.GetByID(ctx, 999); !errors.Is(err, db.ErrNotFound) {
		t.Errorf("get missing err = %v, want ErrNotFound", err)
	}
}

func TestFileAttachDetachAndList(t *testing.T) {
	sqlDB := testutil.NewDB(t)
	ctx := context.Background()
	tasks := db.NewTaskRepo(sqlDB)
	files := db.NewFileRepo(sqlDB)

	task := mustCreate(t, tasks, ctx, "root", nil)

	f, err := files.Upsert(ctx, "src/a.go")
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := files.AttachToTask(ctx, task.ID, f.ID); err != nil {
		t.Fatalf("attach: %v", err)
	}
	// Attaching twice is a no-op.
	if err := files.AttachToTask(ctx, task.ID, f.ID); err != nil {
		t.Fatalf("re-attach: %v", err)
	}

	list, err := files.ListByTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].FilePath != "src/a.go" {
		t.Errorf("list = %+v, want one file", list)
	}

	if err := files.DetachFromTask(ctx, task.ID, f.ID); err != nil {
		t.Fatalf("detach: %v", err)
	}
	list, err = files.ListByTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("list after detach: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("list = %+v, want empty after detach", list)
	}
}

func TestFileBulkAttach(t *testing.T) {
	sqlDB := testutil.NewDB(t)
	ctx := context.Background()
	tasks := db.NewTaskRepo(sqlDB)
	files := db.NewFileRepo(sqlDB)

	task := mustCreate(t, tasks, ctx, "root", nil)

	paths := []string{"src/a.go", "src/b.go", "src/a.go"} // duplicate on purpose
	attached, err := files.BulkAttachToTask(ctx, task.ID, paths)
	if err != nil {
		t.Fatalf("bulk attach: %v", err)
	}
	if len(attached) != 3 {
		t.Fatalf("returned %d files, want 3 (one per input)", len(attached))
	}

	// The task should have two distinct files despite the duplicate path.
	list, err := files.ListByTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("distinct files = %d, want 2", len(list))
	}
}

func TestFileDeletedWhenTaskDeleted(t *testing.T) {
	sqlDB := testutil.NewDB(t)
	ctx := context.Background()
	tasks := db.NewTaskRepo(sqlDB)
	files := db.NewFileRepo(sqlDB)

	task := mustCreate(t, tasks, ctx, "root", nil)
	f, err := files.Upsert(ctx, "src/a.go")
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := files.AttachToTask(ctx, task.ID, f.ID); err != nil {
		t.Fatalf("attach: %v", err)
	}

	// Deleting the task cascades to task_files; the files row itself remains.
	if err := tasks.Delete(ctx, task.ID); err != nil {
		t.Fatalf("delete task: %v", err)
	}
	if _, err := files.GetByID(ctx, f.ID); err != nil {
		t.Errorf("file row should survive task deletion, got err = %v", err)
	}
}
