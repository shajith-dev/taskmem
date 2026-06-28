package main

import (
	"embed"

	"github.com/shajith-dev/taskmem/cmd"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	cmd.SetMigrations(migrationsFS)
	cmd.Execute()
}
