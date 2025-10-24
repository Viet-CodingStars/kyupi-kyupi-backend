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
  DatabaseURL string // DATABASE_URL, optional (generic)
  // Database connection URLs (preferred) or their components
  PostgresURL string // POSTGRES_URL, optional full DSN
  MongoURL    string // MONGODB_URL, optional full connection string
  // Postgres individual parts (used when POSTGRES_URL not provided)
  PostgresUser     string
  PostgresPassword string
  PostgresHost     string
  PostgresPort     string
  PostgresDB       string
  // Mongo individual parts (used when MONGODB_URL not provided)
  MongoHost string
  MongoPort string
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

  pgURL := strings.TrimSpace(os.Getenv("POSTGRES_URL"))
  mongoURL := strings.TrimSpace(os.Getenv("MONGODB_URL"))

  // Postgres components (fallbacks)
  pgUser := strings.TrimSpace(os.Getenv("POSTGRES_USER"))
  if pgUser == "" {
    pgUser = "postgres"
  }
  pgPass := strings.TrimSpace(os.Getenv("POSTGRES_PASSWORD"))
  if pgPass == "" {
    pgPass = "postgres"
  }
  pgHost := strings.TrimSpace(os.Getenv("POSTGRES_HOST"))
  if pgHost == "" {
    pgHost = "postgres"
  }
  pgPort := strings.TrimSpace(os.Getenv("POSTGRES_PORT"))
  if pgPort == "" {
    pgPort = "5432"
  }
  pgDB := strings.TrimSpace(os.Getenv("POSTGRES_DB"))
  if pgDB == "" {
    pgDB = "kyupi"
  }

  // Mongo components (fallback host/port)
  mongoHost := strings.TrimSpace(os.Getenv("MONGO_HOST"))
  if mongoHost == "" {
    mongoHost = "mongo"
  }
  mongoPort := strings.TrimSpace(os.Getenv("MONGO_PORT"))
  if mongoPort == "" {
    mongoPort = "27017"
  }

  return &Config{
    Env:              env,
    Port:             port,
    LogLevel:         logLevel,
    DatabaseURL:      db,
    PostgresURL:      pgURL,
    MongoURL:         mongoURL,
    PostgresUser:     pgUser,
    PostgresPassword: pgPass,
    PostgresHost:     pgHost,
    PostgresPort:     pgPort,
    PostgresDB:       pgDB,
    MongoHost:        mongoHost,
    MongoPort:        mongoPort,
  }
}

// PostgresDSN returns the postgres connection string. If POSTGRES_URL is set it is returned;
// otherwise constructs a DSN like postgres://user:pass@host:port/db
func (c *Config) PostgresDSN() string {
  if c.PostgresURL != "" {
    return c.PostgresURL
  }
  return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
    c.PostgresUser, c.PostgresPassword, c.PostgresHost, c.PostgresPort, c.PostgresDB)
}

// MongoConn returns the MongoDB connection string. If MONGODB_URL is set it is returned;
// otherwise constructs a mongodb://host:port URL.
func (c *Config) MongoConn() string {
  if c.MongoURL != "" {
    return c.MongoURL
  }
  return fmt.Sprintf("mongodb://%s:%s", c.MongoHost, c.MongoPort)
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
