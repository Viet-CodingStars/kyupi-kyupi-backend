package config

import (
  "fmt"
  "os"
  "strings"
)

const (
  EnvDevelopment = "DEVELOPMENT"
  EnvTesting     = "TESTING"
  EnvProduction  = "PRODUCTION"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
  Env         string // APP_ENV: DEVELOPMENT, TESTING, PRODUCTION
  Port        string // PORT, e.g. "8080"
  LogLevel    string // LOG_LEVEL: debug/info/warn/error
  DatabaseURL string // DATABASE_URL, optional
}

// LoadFromEnv reads configuration from environment variables and returns a Config with defaults.
func LoadFromEnv() *Config {
  env := strings.ToUpper(strings.TrimSpace(os.Getenv("APP_ENV")))
  if env == "" {
    env = EnvDevelopment
  }

  port := strings.TrimSpace(os.Getenv("PORT"))
  if port == "" {
    port = "8080"
  }

  logLevel := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
  if logLevel == "" {
    logLevel = "info"
  }

  db := strings.TrimSpace(os.Getenv("DATABASE_URL"))

  return &Config{
    Env:         env,
    Port:        port,
    LogLevel:    logLevel,
    DatabaseURL: db,
  }
}

// Addr returns a host:port address suitable for http.Server
func (c *Config) Addr() string {
  if strings.HasPrefix(c.Port, ":") {
      return c.Port
  }
  return fmt.Sprintf(":%s", c.Port)
}

// IsTesting returns true when APP_ENV=TESTING
func (c *Config) IsTesting() bool { return c.Env == EnvTesting }

// IsProduction returns true when APP_ENV=PRODUCTION
func (c *Config) IsProduction() bool { return c.Env == EnvProduction }
