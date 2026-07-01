package config_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/shajith-dev/taskmem/internal/config"
)

func TestDefaultDatabasePath(t *testing.T) {
	path, err := config.DefaultDatabasePath()
	if err != nil {
		t.Fatalf("default path: %v", err)
	}
	if filepath.Base(path) != "taskmem.db" {
		t.Errorf("path = %q, want it to end in taskmem.db", path)
	}
	if !strings.Contains(filepath.ToSlash(path), "taskmem/taskmem.db") {
		t.Errorf("path = %q, want it under a taskmem directory", path)
	}
}

func TestLoadDefaultsDatabaseURL(t *testing.T) {
	// With no DATABASE_URL set, Load must fall back to the default file path
	// so the CLI works with zero configuration.
	t.Setenv("DATABASE_URL", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.DatabaseURL == "" {
		t.Fatal("DatabaseURL is empty; expected a default path")
	}
	if filepath.Base(cfg.DatabaseURL) != "taskmem.db" {
		t.Errorf("DatabaseURL = %q, want default taskmem.db path", cfg.DatabaseURL)
	}
}

func TestLoadHonoursDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "./custom.db")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.DatabaseURL != "./custom.db" {
		t.Errorf("DatabaseURL = %q, want ./custom.db", cfg.DatabaseURL)
	}
}
