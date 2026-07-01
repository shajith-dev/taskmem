package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/shajith-dev/taskmem/internal/db"
	"github.com/shajith-dev/taskmem/internal/testutil"
)

func TestTaskGraph(t *testing.T) {
	sqlDB := testutil.NewDB(t)
	ctx := context.Background()
	tasks := db.NewTaskRepo(sqlDB)
	graph := db.NewTaskGraphRepo(sqlDB)

	a := mustCreate(t, tasks, ctx, "a", nil)
	b := mustCreate(t, tasks, ctx, "b", nil)

	// a depends on b.
	if err := graph.AddDependency(ctx, a.ID, b.ID); err != nil {
		t.Fatalf("add dependency: %v", err)
	}

	// Adding the same edge again is a no-op (ON CONFLICT DO NOTHING).
	if err := graph.AddDependency(ctx, a.ID, b.ID); err != nil {
		t.Fatalf("re-add dependency: %v", err)
	}

	deps, err := graph.GetDependencies(ctx, a.ID)
	if err != nil {
		t.Fatalf("get dependencies: %v", err)
	}
	if len(deps) != 1 || deps[0].TaskID != a.ID || deps[0].DependsOn != b.ID {
		t.Errorf("dependencies = %+v, want a->b once", deps)
	}

	dependents, err := graph.GetDependents(ctx, b.ID)
	if err != nil {
		t.Fatalf("get dependents: %v", err)
	}
	if len(dependents) != 1 || dependents[0].TaskID != a.ID {
		t.Errorf("dependents = %+v, want [a]", dependents)
	}

	if err := graph.RemoveDependency(ctx, a.ID, b.ID); err != nil {
		t.Fatalf("remove dependency: %v", err)
	}
	if err := graph.RemoveDependency(ctx, a.ID, b.ID); !errors.Is(err, db.ErrNotFound) {
		t.Errorf("remove missing dependency err = %v, want ErrNotFound", err)
	}
}

func TestTaskGraphCascade(t *testing.T) {
	sqlDB := testutil.NewDB(t)
	ctx := context.Background()
	tasks := db.NewTaskRepo(sqlDB)
	graph := db.NewTaskGraphRepo(sqlDB)

	a := mustCreate(t, tasks, ctx, "a", nil)
	b := mustCreate(t, tasks, ctx, "b", nil)
	if err := graph.AddDependency(ctx, a.ID, b.ID); err != nil {
		t.Fatalf("add dependency: %v", err)
	}

	// Deleting a task must remove its graph edges via ON DELETE CASCADE.
	if err := tasks.Delete(ctx, b.ID); err != nil {
		t.Fatalf("delete b: %v", err)
	}
	deps, err := graph.GetDependencies(ctx, a.ID)
	if err != nil {
		t.Fatalf("get dependencies: %v", err)
	}
	if len(deps) != 0 {
		t.Errorf("edges = %+v, want none after cascade", deps)
	}
}

// TestTaskGraphSelfDependencyRejected covers the CHECK (task_id <> depends_on)
// constraint in the schema.
func TestTaskGraphSelfDependencyRejected(t *testing.T) {
	sqlDB := testutil.NewDB(t)
	ctx := context.Background()
	tasks := db.NewTaskRepo(sqlDB)
	graph := db.NewTaskGraphRepo(sqlDB)

	a := mustCreate(t, tasks, ctx, "a", nil)
	if err := graph.AddDependency(ctx, a.ID, a.ID); err == nil {
		t.Fatal("expected self-dependency to be rejected by CHECK constraint")
	}
}
