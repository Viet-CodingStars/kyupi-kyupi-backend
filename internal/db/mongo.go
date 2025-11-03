package db

import (
  "context"
  "time"

  "github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/config"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
)

// ConnectMongo opens a MongoDB client using configuration values and validates connectivity.
func ConnectMongo(cfg *config.Config) (*mongo.Client, error) {
  clientOpts := options.Client().ApplyURI(cfg.MongoConn())

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  client, err := mongo.Connect(ctx, clientOpts)
  if err != nil {
    return nil, err
  }

  if err := client.Ping(ctx, nil); err != nil {
    _ = client.Disconnect(context.Background())
    return nil, err
  }

  return client, nil
}
