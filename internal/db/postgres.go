package db

import (
  "context"
  "database/sql"
  "time"

  _ "github.com/lib/pq"
  "github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/config"
)

// Connect opens a sql.DB to Postgres using the provided config and validates connectivity.
func Connect(cfg *config.Config) (*sql.DB, error) {
  dsn := cfg.PostgresDSN()
  db, err := sql.Open("postgres", dsn)
  if err != nil {
    return nil, err
  }

  if cfg.PostgresConnMaxLifetime > 0 {
    db.SetConnMaxLifetime(cfg.PostgresConnMaxLifetime)
  }
  if cfg.PostgresMaxOpenConns > 0 {
    db.SetMaxOpenConns(cfg.PostgresMaxOpenConns)
  }
  if cfg.PostgresMaxIdleConns > 0 {
    db.SetMaxIdleConns(cfg.PostgresMaxIdleConns)
  }

  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()
  if err := db.PingContext(ctx); err != nil {
    db.Close()
    return nil, err
  }

  // Initialize schema
  if err := initSchema(db); err != nil {
    db.Close()
    return nil, err
  }

  return db, nil
}

// initSchema creates the users table if it doesn't exist
func initSchema(db *sql.DB) error {
  schema := `
  CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    gender VARCHAR(50) NOT NULL,
    birth_date DATE NOT NULL,
    target_gender VARCHAR(50),
    bio TEXT,
    avatar_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
  );
  CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
  `
  _, err := db.Exec(schema)
  return err
}
