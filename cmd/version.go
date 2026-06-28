package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// These are set at build time via -ldflags by GoReleaser (or the Makefile).
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			return printJSON(map[string]string{
				"version": version,
				"commit":  commit,
				"date":    date,
			})
		}
		fmt.Printf("taskmem %s (commit %s, built %s)\n", version, commit, date)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
