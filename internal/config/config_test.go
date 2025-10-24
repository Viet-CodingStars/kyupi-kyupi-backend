package config

import (
  "os"
  "testing"
)

func TestLoadFromEnv_Defaults(t *testing.T) {
  // clear env vars that may be set in CI or local
  os.Unsetenv("APP_ENV")
  os.Unsetenv("PORT")
  os.Unsetenv("LOG_LEVEL")
  os.Unsetenv("DATABASE_URL")

  cfg := LoadFromEnv()
  if cfg.Env != EnvDevelopment {
      t.Fatalf("expected default env %s, got %s", EnvDevelopment, cfg.Env)
  }
  if cfg.Port != "8080" {
      t.Fatalf("expected default port 8080, got %s", cfg.Port)
  }
  if cfg.LogLevel != "info" {
      t.Fatalf("expected default log level 'info', got %s", cfg.LogLevel)
  }
  if cfg.DatabaseURL != "" {
      t.Fatalf("expected empty DatabaseURL by default, got %s", cfg.DatabaseURL)
  }
}

func TestLoadFromEnv_Custom(t *testing.T) {
  os.Setenv("APP_ENV", "testing")
  os.Setenv("PORT", "9000")
  os.Setenv("LOG_LEVEL", "DEBUG")
  os.Setenv("DATABASE_URL", "postgres://user:pass@localhost/db")
  defer func() {
    os.Unsetenv("APP_ENV")
    os.Unsetenv("PORT")
    os.Unsetenv("LOG_LEVEL")
    os.Unsetenv("DATABASE_URL")
  }()

  cfg := LoadFromEnv()
  if cfg.Env != EnvTesting {
    t.Fatalf("expected env %s, got %s", EnvTesting, cfg.Env)
  }
  if cfg.Port != "9000" {
    t.Fatalf("expected port 9000, got %s", cfg.Port)
  }
  if cfg.LogLevel != "debug" {
    t.Fatalf("expected log level debug, got %s", cfg.LogLevel)
  }
  if cfg.DatabaseURL == "" {
    t.Fatalf("expected DatabaseURL to be set")
  }
}
