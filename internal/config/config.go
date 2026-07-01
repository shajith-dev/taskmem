package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	DatabaseURL string `mapstructure:"DATABASE_URL"`
	AppEnv      string `mapstructure:"APP_ENV"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	viper.AutomaticEnv()

	cfg := &Config{
		DatabaseURL: viper.GetString("DATABASE_URL"),
		AppEnv:      viper.GetString("APP_ENV"),
	}

	// Zero-config default: store the database under the user's config
	// directory so the CLI works out of the box with no setup.
	if cfg.DatabaseURL == "" {
		path, err := DefaultDatabasePath()
		if err != nil {
			return nil, err
		}
		cfg.DatabaseURL = path
	}

	if cfg.AppEnv == "" {
		cfg.AppEnv = "development"
	}

	return cfg, nil
}

// DefaultDatabasePath returns the default on-disk location of the SQLite
// database, e.g. %AppData%\taskmem\taskmem.db on Windows or
// ~/.config/taskmem/taskmem.db on Linux.
func DefaultDatabasePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(dir, "taskmem", "taskmem.db"), nil
}
