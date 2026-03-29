package testhelpers

import (
	"time"

	"backend-sport-team-report-go/internal/config"
)

func DefaultTestConfig() config.Config {
	return config.Config{
		App: config.AppConfig{
			Name: "soccer-team-report",
			Env:  config.EnvTest,
		},
		Database: config.DatabaseConfig{},
		Auth: config.AuthConfig{
			JWTSecret:      "integration-test-secret",
			AccessTokenTTL: 15 * time.Minute,
		},
		Security: config.SecurityConfig{
			MaxJSONBodyBytes: 1 << 20,
			RateLimit: config.RateLimitConfig{
				Login: config.RateLimitRule{
					Window:      time.Minute,
					MaxRequests: 5,
				},
				AuthenticatedWrite: config.RateLimitRule{
					Window:      time.Minute,
					MaxRequests: 30,
				},
			},
		},
	}
}
