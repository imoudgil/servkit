// Package config loads service configuration from environment variables with
// sensible defaults suitable for local development and container deployment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Service holds runtime configuration shared across HTTP servers and clients.
type Service struct {
	Name            string
	Environment     string
	HTTPAddr        string
	ShutdownTimeout time.Duration
	LogLevel        string
	ClientTimeout   time.Duration
	ClientRetries   int
}

// Load reads configuration from the environment. Missing values fall back to
// defaults documented on each field.
func Load() (Service, error) {
	cfg := Service{
		Name:            envOr("SERVICE_NAME", "servkit"),
		Environment:     envOr("SERVICE_ENV", "development"),
		HTTPAddr:        envOr("HTTP_ADDR", ":8080"),
		ShutdownTimeout: envDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		LogLevel:        envOr("LOG_LEVEL", "info"),
		ClientTimeout:   envDuration("CLIENT_TIMEOUT", 5*time.Second),
		ClientRetries:   envInt("CLIENT_RETRIES", 3),
	}
	if cfg.ClientRetries < 0 {
		return Service{}, fmt.Errorf("CLIENT_RETRIES must be >= 0")
	}
	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return n
}

func envDuration(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return d
}
