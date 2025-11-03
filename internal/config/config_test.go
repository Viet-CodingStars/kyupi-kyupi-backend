package config

import (
  "os"
  "testing"
  "time"
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
  os.Unsetenv("POSTGRES_SSL_MODE")
  os.Unsetenv("POSTGRES_MAX_OPEN_CONNS")
  os.Unsetenv("POSTGRES_MAX_IDLE_CONNS")
  os.Unsetenv("POSTGRES_CONN_MAX_LIFETIME")
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
  if cfg.PostgresSSLMode != "disable" {
    t.Fatalf("expected default sslmode disable, got %s", cfg.PostgresSSLMode)
  }
  if cfg.PostgresMaxOpenConns != 25 {
    t.Fatalf("expected default max open conns 25, got %d", cfg.PostgresMaxOpenConns)
  }
  if cfg.PostgresConnMaxLifetime != 30*time.Minute {
    t.Fatalf("expected default conn max lifetime 30m, got %s", cfg.PostgresConnMaxLifetime)
  }
  if cfg.MongoDatabase != "kyupi" {
    t.Fatalf("expected default mongo database kyupi, got %s", cfg.MongoDatabase)
  }
  if got := cfg.MongoConn(); got != "mongodb://mongo:27017/kyupi" {
    t.Fatalf("expected default mongo URI, got %s", got)
  }
}

func TestLoadFromEnv_Custom(t *testing.T) {
  os.Setenv("APP_ENV", "testing")
  os.Setenv("PORT", "9000")
  os.Setenv("LOG_LEVEL", "DEBUG")
  os.Setenv("DATABASE_URL", "postgres://user:pass@localhost/db")
  os.Setenv("POSTGRES_URL", "")
  os.Setenv("POSTGRES_USER", "pguser")
  os.Setenv("POSTGRES_PASSWORD", "pgpass")
  os.Setenv("POSTGRES_HOST", "pg-host")
  os.Setenv("POSTGRES_PORT", "5433")
  os.Setenv("POSTGRES_DB", "pgdb")
  os.Setenv("POSTGRES_SSL_MODE", "require")
  os.Setenv("POSTGRES_MAX_OPEN_CONNS", "50")
  os.Setenv("POSTGRES_MAX_IDLE_CONNS", "20")
  os.Setenv("POSTGRES_CONN_MAX_LIFETIME", "1h")
  os.Setenv("MONGODB_URL", "")
  os.Setenv("MONGO_HOST", "mongo-host")
  os.Setenv("MONGO_PORT", "27018")
  os.Setenv("MONGO_USER", "mongouser")
  os.Setenv("MONGO_PASSWORD", "mongopass")
  os.Setenv("MONGO_DATABASE", "kyupiapp")
  os.Setenv("MONGO_AUTH_SOURCE", "admin")
  os.Setenv("MONGO_REPLICA_SET", "rs0")
  defer func() {
    os.Unsetenv("APP_ENV")
    os.Unsetenv("PORT")
    os.Unsetenv("LOG_LEVEL")
    os.Unsetenv("DATABASE_URL")
    os.Unsetenv("POSTGRES_URL")
    os.Unsetenv("POSTGRES_USER")
    os.Unsetenv("POSTGRES_PASSWORD")
    os.Unsetenv("POSTGRES_HOST")
    os.Unsetenv("POSTGRES_PORT")
    os.Unsetenv("POSTGRES_DB")
    os.Unsetenv("POSTGRES_SSL_MODE")
    os.Unsetenv("POSTGRES_MAX_OPEN_CONNS")
    os.Unsetenv("POSTGRES_MAX_IDLE_CONNS")
    os.Unsetenv("POSTGRES_CONN_MAX_LIFETIME")
    os.Unsetenv("MONGODB_URL")
    os.Unsetenv("MONGO_HOST")
    os.Unsetenv("MONGO_PORT")
    os.Unsetenv("MONGO_USER")
    os.Unsetenv("MONGO_PASSWORD")
    os.Unsetenv("MONGO_DATABASE")
    os.Unsetenv("MONGO_AUTH_SOURCE")
    os.Unsetenv("MONGO_REPLICA_SET")
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
  expectedPostgres := "postgres://pguser:pgpass@pg-host:5433/pgdb?sslmode=require"
  if got := cfg.PostgresDSN(); got != expectedPostgres {
    t.Fatalf("unexpected PostgresDSN: got %s, want %s", got, expectedPostgres)
  }
  if cfg.PostgresMaxOpenConns != 50 {
    t.Fatalf("expected max open conns 50, got %d", cfg.PostgresMaxOpenConns)
  }
  if cfg.PostgresConnMaxLifetime != time.Hour {
    t.Fatalf("expected conn lifetime 1h, got %s", cfg.PostgresConnMaxLifetime)
  }
  expectedMongo := "mongodb://mongouser:mongopass@mongo-host:27018/kyupiapp?authSource=admin&replicaSet=rs0"
  if got := cfg.MongoConn(); got != expectedMongo {
    t.Fatalf("unexpected MongoConn: got %s, want %s", got, expectedMongo)
  }
}
