package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/shajith-dev/taskmem/internal/models"
	"github.com/shajith-dev/taskmem/internal/service"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
}

// ── create ────────────────────────────────────────────────────────────────────

var (
	createParent      int64
	createModel       string
	createUseSubagent bool
)

var taskCreateCmd = &cobra.Command{
	Use:   "create <description>",
	Short: "Create a new task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var parentID *int64
		if cmd.Flags().Changed("parent") {
			parentID = &createParent
		}

		t, err := currentApp.Tasks.Create(context.Background(), args[0], createModel, parentID, createUseSubagent)
		if err != nil {
			return err
		}

		if jsonOutput {
			return printJSON(t)
		}
		fmt.Printf("Created task #%d\n", t.ID)
		return nil
	},
}

// ── create-bulk ───────────────────────────────────────────────────────────────

var (
	bulkParent      int64
	bulkModel       string
	bulkUseSubagent bool
	bulkFile        string
)

var taskCreateBulkCmd = &cobra.Command{
	Use:   "create-bulk [description...]",
	Short: "Create multiple tasks at once (args or --file tasks.json)",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var inputs []service.BulkTaskInput

		if cmd.Flags().Changed("file") {
			data, err := os.ReadFile(bulkFile)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}
			// Strip UTF-8 BOM if present.
			data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
			if err := json.Unmarshal(data, &inputs); err != nil {
				return fmt.Errorf("parse JSON: %w", err)
			}
		} else {
			if len(args) == 0 {
				return fmt.Errorf("provide at least one description or use --file")
			}
			var parentID *int64
			if cmd.Flags().Changed("parent") {
				parentID = &bulkParent
			}
			for _, desc := range args {
				inputs = append(inputs, service.BulkTaskInput{
					Description: desc,
					Model:       bulkModel,
					ParentID:    parentID,
					UseSubagent: bulkUseSubagent,
				})
			}
		}

		created, err := currentApp.Tasks.BulkCreate(context.Background(), inputs)
		if err != nil {
			return err
		}

		if jsonOutput {
			if created == nil {
				created = []*models.Task{}
			}
			return printJSON(created)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tSTATUS\tDESCRIPTION")
		for _, t := range created {
			fmt.Fprintf(w, "%d\t%s\t%s\n", t.ID, t.Status, t.Description)
		}
		w.Flush()
		return nil
	},
}

// ── get ───────────────────────────────────────────────────────────────────────

var taskGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a task by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid id: %w", err)
		}

		t, err := currentApp.Tasks.Get(context.Background(), id)
		if err != nil {
			return err
		}

		if jsonOutput {
			return printJSON(t)
		}
		printTask(t)
		return nil
	},
}

// ── list ──────────────────────────────────────────────────────────────────────

var listParent int64

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks (root tasks by default)",
	RunE: func(cmd *cobra.Command, args []string) error {
		var parentID *int64
		if cmd.Flags().Changed("parent") {
			parentID = &listParent
		}

		tasks, err := currentApp.Tasks.ListChildren(context.Background(), parentID)
		if err != nil {
			return err
		}

		if jsonOutput {
			if tasks == nil {
				tasks = []*models.Task{}
			}
			return printJSON(tasks)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tSTATUS\tSUBAGENT\tDESCRIPTION")
		for _, t := range tasks {
			fmt.Fprintf(w, "%d\t%s\t%v\t%s\n", t.ID, t.Status, t.UseSubagent, t.Description)
		}
		w.Flush()
		return nil
	},
}

// ── status ────────────────────────────────────────────────────────────────────

var taskStatusCmd = &cobra.Command{
	Use:   "status <id> <status>",
	Short: "Update task status (PENDING|IN_PROGRESS|COMPLETED|PARTIALLY_COMPLETED)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid id: %w", err)
		}

		status := models.TaskStatus(args[1])
		if err := currentApp.Tasks.UpdateStatus(context.Background(), id, status); err != nil {
			return err
		}
		if jsonOutput {
			return printJSON(map[string]any{"id": id, "status": string(status)})
		}
		return nil
	},
}

// ── update ────────────────────────────────────────────────────────────────────

var (
	updateParent      int64
	updateNoParent    bool
	updateStatus      string
	updateDescription string
	updateModel       string
	updateSubagent    bool
)

var taskUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update fields of a task (only the flags you pass are changed)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid id: %w", err)
		}

		if cmd.Flags().Changed("parent") && cmd.Flags().Changed("no-parent") {
			return fmt.Errorf("--parent and --no-parent are mutually exclusive")
		}

		// Load the current task, then override only the flags the user set.
		t, err := currentApp.Tasks.Get(context.Background(), id)
		if err != nil {
			return err
		}

		if cmd.Flags().Changed("description") {
			t.Description = updateDescription
		}
		if cmd.Flags().Changed("model") {
			t.Model = updateModel
		}
		if cmd.Flags().Changed("subagent") {
			t.UseSubagent = updateSubagent
		}
		if cmd.Flags().Changed("status") {
			status := models.TaskStatus(updateStatus)
			if !status.Valid() {
				return fmt.Errorf("invalid status %q", updateStatus)
			}
			t.Status = status
		}
		if cmd.Flags().Changed("parent") {
			p := updateParent
			t.Parent = &p
		}
		if cmd.Flags().Changed("no-parent") {
			t.Parent = nil
		}

		updated, err := currentApp.Tasks.Update(context.Background(), t)
		if err != nil {
			return err
		}

		if jsonOutput {
			return printJSON(updated)
		}
		printTask(updated)
		return nil
	},
}

// ── delete ────────────────────────────────────────────────────────────────────

var taskDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid id: %w", err)
		}

		if err := currentApp.Tasks.Delete(context.Background(), id); err != nil {
			return err
		}
		if jsonOutput {
			return printJSON(map[string]any{"deleted": id})
		}
		return nil
	},
}

// ── dep ───────────────────────────────────────────────────────────────────────

var depCmd = &cobra.Command{
	Use:   "dep",
	Short: "Manage task dependencies",
}

var depAddCmd = &cobra.Command{
	Use:   "add <task-id> <depends-on-id>",
	Short: "Add a dependency between tasks",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid task-id: %w", err)
		}
		dependsOn, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid depends-on-id: %w", err)
		}

		if err := currentApp.Tasks.AddDependency(context.Background(), taskID, dependsOn); err != nil {
			return err
		}
		if jsonOutput {
			return printJSON(map[string]any{"task_id": taskID, "depends_on": dependsOn})
		}
		return nil
	},
}

var depRemoveCmd = &cobra.Command{
	Use:   "remove <task-id> <depends-on-id>",
	Short: "Remove a dependency between tasks",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid task-id: %w", err)
		}
		dependsOn, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid depends-on-id: %w", err)
		}

		if err := currentApp.Tasks.RemoveDependency(context.Background(), taskID, dependsOn); err != nil {
			return err
		}
		if jsonOutput {
			return printJSON(map[string]any{"removed": true, "task_id": taskID, "depends_on": dependsOn})
		}
		return nil
	},
}

var depListCmd = &cobra.Command{
	Use:   "list <task-id>",
	Short: "List dependencies of a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid task-id: %w", err)
		}

		deps, err := currentApp.Tasks.GetDependencies(context.Background(), taskID)
		if err != nil {
			return err
		}

		if jsonOutput {
			if deps == nil {
				deps = []*models.TaskGraph{}
			}
			return printJSON(deps)
		}
		if len(deps) == 0 {
			fmt.Println("No dependencies.")
			return nil
		}

		for _, d := range deps {
			fmt.Printf("task %d depends on task %d\n", d.TaskID, d.DependsOn)
		}
		return nil
	},
}

var depDependentsCmd = &cobra.Command{
	Use:   "dependents <task-id>",
	Short: "List tasks that depend on this task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid task-id: %w", err)
		}

		deps, err := currentApp.Tasks.GetDependents(context.Background(), taskID)
		if err != nil {
			return err
		}

		if jsonOutput {
			if deps == nil {
				deps = []*models.TaskGraph{}
			}
			return printJSON(deps)
		}
		if len(deps) == 0 {
			fmt.Println("No dependents.")
			return nil
		}

		for _, d := range deps {
			fmt.Printf("task %d depends on task %d\n", d.TaskID, d.DependsOn)
		}
		return nil
	},
}

// ── scratchpad ────────────────────────────────────────────────────────────────

var scratchpadCmd = &cobra.Command{
	Use:   "scratchpad",
	Short: "Read or write a task's scratchpad (agent working memory)",
}

var scratchpadGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Print the task's scratchpad",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid id: %w", err)
		}

		t, err := currentApp.Tasks.Get(context.Background(), id)
		if err != nil {
			return err
		}

		if jsonOutput {
			return printJSON(map[string]any{"id": t.ID, "scratchpad": t.Scratchpad})
		}
		if t.Scratchpad != nil && *t.Scratchpad != "" {
			fmt.Println(*t.Scratchpad)
		}
		return nil
	},
}

var scratchpadSetCmd = &cobra.Command{
	Use:   "set <id> <text>",
	Short: "Replace the task's scratchpad with <text>",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid id: %w", err)
		}

		t, err := currentApp.Tasks.SetScratchpad(context.Background(), id, args[1])
		if err != nil {
			return err
		}

		if jsonOutput {
			return printJSON(map[string]any{"id": t.ID, "scratchpad": t.Scratchpad})
		}
		return nil
	},
}

var scratchpadAppendCmd = &cobra.Command{
	Use:   "append <id> <text>",
	Short: "Append <text> to the task's scratchpad",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid id: %w", err)
		}

		t, err := currentApp.Tasks.AppendScratchpad(context.Background(), id, args[1])
		if err != nil {
			return err
		}

		if jsonOutput {
			return printJSON(map[string]any{"id": t.ID, "scratchpad": t.Scratchpad})
		}
		return nil
	},
}

// ── helpers ───────────────────────────────────────────────────────────────────

func printTask(t *models.Task) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ID:\t%d\n", t.ID)
	if t.Parent != nil {
		fmt.Fprintf(w, "Parent:\t%d\n", *t.Parent)
	}
	fmt.Fprintf(w, "Status:\t%s\n", t.Status)
	fmt.Fprintf(w, "Model:\t%s\n", t.Model)
	fmt.Fprintf(w, "Subagent:\t%v\n", t.UseSubagent)
	fmt.Fprintf(w, "Description:\t%s\n", t.Description)
	if t.Scratchpad != nil {
		fmt.Fprintf(w, "Scratchpad:\t%s\n", *t.Scratchpad)
	}
	fmt.Fprintf(w, "Created:\t%s\n", t.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "Updated:\t%s\n", t.UpdatedAt.Format("2006-01-02 15:04:05"))
	w.Flush()
}

func init() {
	// task create flags
	taskCreateCmd.Flags().Int64Var(&createParent, "parent", 0, "Parent task ID")
	taskCreateCmd.Flags().StringVar(&createModel, "model", "inherit", "Model to use")
	taskCreateCmd.Flags().BoolVar(&createUseSubagent, "subagent", false, "Use subagent")

	// task create-bulk flags
	taskCreateBulkCmd.Flags().Int64Var(&bulkParent, "parent", 0, "Parent task ID (applied to all)")
	taskCreateBulkCmd.Flags().StringVar(&bulkModel, "model", "inherit", "Model (applied to all)")
	taskCreateBulkCmd.Flags().BoolVar(&bulkUseSubagent, "subagent", false, "Use subagent (applied to all)")
	taskCreateBulkCmd.Flags().StringVar(&bulkFile, "file", "", "JSON file with task definitions")

	// task list flags
	taskListCmd.Flags().Int64Var(&listParent, "parent", 0, "Filter by parent task ID")

	// task update flags
	taskUpdateCmd.Flags().StringVar(&updateDescription, "description", "", "New description")
	taskUpdateCmd.Flags().StringVar(&updateStatus, "status", "", "New status (PENDING|IN_PROGRESS|COMPLETED|PARTIALLY_COMPLETED)")
	taskUpdateCmd.Flags().StringVar(&updateModel, "model", "", "New model")
	taskUpdateCmd.Flags().Int64Var(&updateParent, "parent", 0, "New parent task ID")
	taskUpdateCmd.Flags().BoolVar(&updateNoParent, "no-parent", false, "Clear the parent (make it a root task)")
	taskUpdateCmd.Flags().BoolVar(&updateSubagent, "subagent", false, "Set the use-subagent flag")

	// wire subcommands
	depCmd.AddCommand(depAddCmd, depRemoveCmd, depListCmd, depDependentsCmd)
	scratchpadCmd.AddCommand(scratchpadGetCmd, scratchpadSetCmd, scratchpadAppendCmd)
	taskCmd.AddCommand(taskCreateCmd, taskCreateBulkCmd, taskGetCmd, taskListCmd, taskStatusCmd, taskUpdateCmd, taskDeleteCmd, depCmd, scratchpadCmd)
	rootCmd.AddCommand(taskCmd)
}
