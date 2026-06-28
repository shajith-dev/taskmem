package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/shajith-dev/taskmem/internal/app"
)

var (
	cfgFile    string
	jsonOutput bool
	currentApp *app.App
)

var rootCmd = &cobra.Command{
	Use:          "taskmem",
	Short:        "taskmem CLI",
	SilenceUsage: true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// These commands don't touch the database, so don't require a connection.
		// (migrate manages its own connection.)
		switch cmd.Name() {
		case "migrate", "version", "help", "completion":
			return nil
		}
		a, err := app.New(context.Background())
		if err != nil {
			return fmt.Errorf("init app: %w", err)
		}
		currentApp = a
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if jsonOutput {
			b, merr := json.Marshal(map[string]string{"error": err.Error()})
			if merr != nil {
				fmt.Fprintln(os.Stderr, err)
			} else {
				fmt.Fprintf(os.Stderr, "%s\n", b)
			}
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		if currentApp != nil {
			currentApp.Close()
		}
		os.Exit(1)
	}
	if currentApp != nil {
		currentApp.Close()
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: .env)")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output as JSON instead of human-readable tables")
}
