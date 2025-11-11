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
  Intention    string    `json:"intention"`
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

const (
  IntentionLongTermPartner     = "long_term_partner"
  IntentionLongTermOpenToShort = "long_term_open_to_short"
  IntentionShortTermOpenToLong = "short_term_open_to_long"
  IntentionShortTermFun        = "short_term_fun"
  IntentionNewFriends          = "new_friends"
  IntentionStillFiguringOut    = "still_figuring_out"
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

// IsValidIntention returns true when the provided intention matches a supported enum value.
func IsValidIntention(val string) bool {
  switch val {
  case IntentionLongTermPartner,
    IntentionLongTermOpenToShort,
    IntentionShortTermOpenToLong,
    IntentionShortTermFun,
    IntentionNewFriends,
    IntentionStillFiguringOut:
    return true
  default:
    return false
  }
}

// DefaultIntention returns the default intention value for new users.
func DefaultIntention() string {
  return IntentionStillFiguringOut
}
