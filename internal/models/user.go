package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Name         string     `json:"name"`
	Gender       string     `json:"gender,omitempty"`
	BirthDate    *time.Time `json:"birth_date,omitempty"`
	Bio          string     `json:"bio,omitempty"`
	AvatarURL    string     `json:"avatar_url,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
