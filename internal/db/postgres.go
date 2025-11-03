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
  return db, nil
}
