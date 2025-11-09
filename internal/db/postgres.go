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
    gender VARCHAR(50),
    birth_date DATE,
    bio TEXT,
    avatar_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
  );
  CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
  
  CREATE TABLE IF NOT EXISTS likes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(10) NOT NULL CHECK (status IN ('like', 'pass')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, target_user_id)
  );
  CREATE INDEX IF NOT EXISTS idx_likes_user_id ON likes(user_id);
  CREATE INDEX IF NOT EXISTS idx_likes_target_user_id ON likes(target_user_id);
  CREATE INDEX IF NOT EXISTS idx_likes_status ON likes(status);
  
  CREATE TABLE IF NOT EXISTS matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user1_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user2_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user1_id, user2_id),
    CHECK (user1_id < user2_id)
  );
  CREATE INDEX IF NOT EXISTS idx_matches_user1_id ON matches(user1_id);
  CREATE INDEX IF NOT EXISTS idx_matches_user2_id ON matches(user2_id);
  `
  _, err := db.Exec(schema)
  return err
}
