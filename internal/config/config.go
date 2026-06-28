package config

import (
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

	if cfg.AppEnv == "" {
		cfg.AppEnv = "development"
	}

	return cfg, nil
}
