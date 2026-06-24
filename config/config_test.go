package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("SERVICE_NAME", "")
	t.Setenv("HTTP_ADDR", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Name != "servkit" {
		t.Errorf("Name = %q, want servkit", cfg.Name)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Errorf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("ShutdownTimeout = %v", cfg.ShutdownTimeout)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("SERVICE_NAME", "payments-api")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("CLIENT_RETRIES", "5")
	t.Setenv("SHUTDOWN_TIMEOUT", "15s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Name != "payments-api" {
		t.Errorf("Name = %q", cfg.Name)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Errorf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if cfg.ClientRetries != 5 {
		t.Errorf("ClientRetries = %d", cfg.ClientRetries)
	}
	if cfg.ShutdownTimeout != 15*time.Second {
		t.Errorf("ShutdownTimeout = %v", cfg.ShutdownTimeout)
	}
}

func TestLoadInvalidRetriesIgnored(t *testing.T) {
	t.Setenv("CLIENT_RETRIES", "not-a-number")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ClientRetries != 3 {
		t.Errorf("ClientRetries = %d, want default 3", cfg.ClientRetries)
	}
}

func TestMain(m *testing.M) {
	os.Clearenv()
	os.Exit(m.Run())
}
