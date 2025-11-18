package models

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Message represents a chat message between matched users in MongoDB
type Message struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MatchID    uuid.UUID          `bson:"match_id" json:"match_id"`
	SenderID   uuid.UUID          `bson:"sender_id" json:"sender_id"`
	ReceiverID uuid.UUID          `bson:"receiver_id" json:"receiver_id"`
	Content    string             `bson:"content" json:"content"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}
