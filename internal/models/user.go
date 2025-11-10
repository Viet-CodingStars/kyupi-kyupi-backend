package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Gender       int       `json:"gender"`
	BirthDate    time.Time `json:"birth_date"`
	TargetGender *int      `json:"target_gender,omitempty"`
	Bio          string    `json:"bio,omitempty"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

const (
	GenderMale   = 1
	GenderFemale = 2
	GenderOthers = 3
)

// IsValidGender returns true when the provided gender matches a supported enum value.
func IsValidGender(g int) bool {
	switch g {
	case GenderMale, GenderFemale, GenderOthers:
		return true
	default:
		return false
	}
}
