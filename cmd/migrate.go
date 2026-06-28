package cmd

import (
	"context"
	"embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/shajith-dev/taskmem/internal/config"
	"github.com/shajith-dev/taskmem/internal/db"
)

// migrationsFS holds the embedded migration files, injected from main.
var migrationsFS embed.FS

// SetMigrations injects the embedded migrations filesystem.
func SetMigrations(fsys embed.FS) {
	migrationsFS = fsys
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		if cfg.DatabaseURL == "" {
			fmt.Fprintln(os.Stderr, "DATABASE_URL is not set")
			os.Exit(1)
		}

		return db.Migrate(context.Background(), cfg.DatabaseURL, migrationsFS, "migrations")
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
