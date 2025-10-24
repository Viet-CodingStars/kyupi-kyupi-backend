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
  os.Unsetenv("POSTGRES_URL")
  os.Unsetenv("MONGODB_URL")
  os.Unsetenv("POSTGRES_USER")
  os.Unsetenv("POSTGRES_PASSWORD")
  os.Unsetenv("POSTGRES_HOST")
  os.Unsetenv("POSTGRES_PORT")
  os.Unsetenv("POSTGRES_DB")
  os.Unsetenv("MONGO_HOST")
  os.Unsetenv("MONGO_PORT")

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
  if cfg.PostgresURL != "" {
    t.Fatalf("expected empty PostgresURL by default, got %s", cfg.PostgresURL)
  }
  if cfg.MongoURL != "" {
    t.Fatalf("expected empty MongoURL by default, got %s", cfg.MongoURL)
  }
  if cfg.PostgresUser == "" || cfg.PostgresHost == "" {
    t.Fatalf("expected Postgres defaults to be set")
  }
}

func TestLoadFromEnv_Custom(t *testing.T) {
  os.Setenv("APP_ENV", "testing")
  os.Setenv("PORT", "9000")
  os.Setenv("LOG_LEVEL", "DEBUG")
  os.Setenv("DATABASE_URL", "postgres://user:pass@localhost/db")
  os.Setenv("POSTGRES_URL", "postgres://u:p@pg:5432/mydb")
  os.Setenv("MONGODB_URL", "mongodb://mongo:27017/mydb")
  defer func() {
    os.Unsetenv("APP_ENV")
    os.Unsetenv("PORT")
    os.Unsetenv("LOG_LEVEL")
    os.Unsetenv("DATABASE_URL")
    os.Unsetenv("POSTGRES_URL")
    os.Unsetenv("MONGODB_URL")
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
  if cfg.PostgresDSN() == "" {
    t.Fatalf("expected PostgresDSN to be set")
  }
  if cfg.MongoConn() == "" {
    t.Fatalf("expected MongoConn to be set")
  }
}
