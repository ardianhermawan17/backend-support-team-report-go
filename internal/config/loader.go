package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

func Load() (Config, error) {
	cfg := defaultConfig()

	if err := mergeYAML(filepath.Join("configs", "app.yaml"), &cfg); err != nil {
		return Config{}, err
	}

	env := getenv("APP_ENV", cfg.App.Env)
	if env != "" {
		cfg.App.Env = env
	}

	envPath := filepath.Join("configs", fmt.Sprintf("app.%s.yaml", cfg.App.Env))
	if err := mergeYAML(envPath, &cfg); err != nil {
		return Config{}, err
	}

	applyEnvOverrides(&cfg)
	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		App: AppConfig{
			Name:            "soccer-team-report",
			Env:             EnvLocal,
			Host:            "0.0.0.0",
			Port:            "8080",
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 10 * time.Second,
		},
		Database: DatabaseConfig{
			DSN:                 "postgres://postgres:postgres@localhost:5432/soccer_team_report?sslmode=disable",
			MaxOpenConns:        10,
			MaxIdleConns:        5,
			ConnMaxLifetime:     30 * time.Minute,
			HealthCheckTimeout:  2 * time.Second,
			StartupPingTimeout:  30 * time.Second,
			StartupPingInterval: 2 * time.Second,
		},
		Auth: AuthConfig{
			JWTSecret:      "local-dev-only-change-me",
			AccessTokenTTL: 15 * time.Minute,
		},
		Security: SecurityConfig{
			MaxJSONBodyBytes: 1 << 20,
			RateLimit: RateLimitConfig{
				Login: RateLimitRule{
					Window:      time.Minute,
					MaxRequests: 5,
				},
				AuthenticatedWrite: RateLimitRule{
					Window:      time.Minute,
					MaxRequests: 30,
				},
			},
		},
	}
}

func mergeYAML(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read config %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("unmarshal config %s: %w", path, err)
	}

	return nil
}

func applyEnvOverrides(cfg *Config) {
	cfg.App.Name = getenv("APP_NAME", cfg.App.Name)
	cfg.App.Env = getenv("APP_ENV", cfg.App.Env)
	cfg.App.Host = getenv("APP_HOST", cfg.App.Host)
	cfg.App.Port = getenv("APP_PORT", cfg.App.Port)
	cfg.Database.DSN = getenv("DATABASE_DSN", cfg.Database.DSN)
	cfg.Auth.JWTSecret = getenv("AUTH_JWT_SECRET", cfg.Auth.JWTSecret)
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
