package repo

import (
	"context"
	"errors"
	"time"

	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/models"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrMessageNotFound = errors.New("message not found")
	ErrInvalidMatchID  = errors.New("invalid match_id")
)

// MessageRepo handles database operations for messages in MongoDB
type MessageRepo struct {
	collection *mongo.Collection
}

// NewMessageRepo creates a new MessageRepo
func NewMessageRepo(db *mongo.Database) *MessageRepo {
	return &MessageRepo{
		collection: db.Collection("messages"),
	}
}

// Create inserts a new message into MongoDB
func (r *MessageRepo) Create(ctx context.Context, message *models.Message) error {
	message.CreatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, message)
	if err != nil {
		return err
	}
	message.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByMatchID retrieves all messages for a specific match, ordered by creation time
func (r *MessageRepo) GetByMatchID(ctx context.Context, matchID uuid.UUID) ([]*models.Message, error) {
	filter := bson.M{"match_id": matchID.String()}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*models.Message
	for cursor.Next(ctx) {
		var msg models.Message
		if err := cursor.Decode(&msg); err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}
