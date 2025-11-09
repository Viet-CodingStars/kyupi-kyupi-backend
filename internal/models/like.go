package models

import (
	"time"

	"github.com/google/uuid"
)

// Like represents a user liking or passing another user
type Like struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	TargetUserID uuid.UUID `json:"target_user_id"`
	Status       string    `json:"status"` // "like" or "pass"
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
