package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/shajith-dev/taskmem/internal/models"
)

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Manage files attached to tasks",
}

var fileAttachCmd = &cobra.Command{
	Use:   "attach <task-id> <file-path>",
	Short: "Attach a file path to a task",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid task-id: %w", err)
		}

		f, err := currentApp.Files.AttachPath(context.Background(), taskID, args[1])
		if err != nil {
			return err
		}

		if jsonOutput {
			return printJSON(f)
		}
		fmt.Printf("Attached file #%d (%s) to task #%d\n", f.ID, f.FilePath, taskID)
		return nil
	},
}

var fileAttachBulkCmd = &cobra.Command{
	Use:   "attach-bulk <task-id> <file-path>...",
	Short: "Attach multiple file paths to a task at once",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid task-id: %w", err)
		}

		files, err := currentApp.Files.BulkAttachPaths(context.Background(), taskID, args[1:])
		if err != nil {
			return err
		}

		if jsonOutput {
			if files == nil {
				files = []*models.File{}
			}
			return printJSON(files)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tPATH")
		for _, f := range files {
			fmt.Fprintf(w, "%d\t%s\n", f.ID, f.FilePath)
		}
		w.Flush()
		return nil
	},
}

var fileDetachCmd = &cobra.Command{
	Use:   "detach <task-id> <file-path>",
	Short: "Detach a file path from a task",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid task-id: %w", err)
		}

		if err := currentApp.Files.DetachPath(context.Background(), taskID, args[1]); err != nil {
			return err
		}
		if jsonOutput {
			return printJSON(map[string]any{"detached": true})
		}
		return nil
	},
}

var fileListCmd = &cobra.Command{
	Use:   "list <task-id>",
	Short: "List files attached to a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid task-id: %w", err)
		}

		files, err := currentApp.Files.ListByTask(context.Background(), taskID)
		if err != nil {
			return err
		}

		if jsonOutput {
			if files == nil {
				files = []*models.File{}
			}
			return printJSON(files)
		}
		if len(files) == 0 {
			fmt.Println("No files attached.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tPATH")
		for _, f := range files {
			fmt.Fprintf(w, "%d\t%s\n", f.ID, f.FilePath)
		}
		w.Flush()
		return nil
	},
}

func init() {
	fileCmd.AddCommand(fileAttachCmd, fileAttachBulkCmd, fileDetachCmd, fileListCmd)
	rootCmd.AddCommand(fileCmd)
}
