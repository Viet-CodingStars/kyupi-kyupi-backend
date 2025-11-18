package repo

import (
	"database/sql"
	"errors"

	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/models"
	"github.com/google/uuid"
)

var (
	ErrMatchNotFound      = errors.New("match not found")
	ErrMatchAlreadyExists = errors.New("match already exists")
)

// MatchRepo handles database operations for matches
type MatchRepo struct {
	db *sql.DB
}

// NewMatchRepo creates a new MatchRepo
func NewMatchRepo(db *sql.DB) *MatchRepo {
	return &MatchRepo{db: db}
}

// Create inserts a new match into the database
// Ensures user1_id < user2_id for consistency
func (r *MatchRepo) Create(match *models.Match) error {
	// Ensure user1_id is always less than user2_id for the unique constraint
	user1ID, user2ID := match.User1ID, match.User2ID
	if user1ID.String() > user2ID.String() {
		user1ID, user2ID = user2ID, user1ID
	}

	query := `
		INSERT INTO matches (user1_id, user2_id)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(query, user1ID, user2ID).
		Scan(&match.ID, &match.CreatedAt, &match.UpdatedAt)
	
	match.User1ID = user1ID
	match.User2ID = user2ID
	
	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"matches_user1_id_user2_id_key\"" {
			return ErrMatchAlreadyExists
		}
		return err
	}
	return nil
}

// GetByUserID retrieves all matches for a user
func (r *MatchRepo) GetByUserID(userID uuid.UUID) ([]*models.Match, error) {
	query := `
		SELECT id, user1_id, user2_id, created_at, updated_at
		FROM matches 
		WHERE user1_id = $1 OR user2_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches := []*models.Match{}
	for rows.Next() {
		match := &models.Match{}
		err := rows.Scan(
			&match.ID, &match.User1ID, &match.User2ID,
			&match.CreatedAt, &match.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return matches, nil
}

// Exists checks if a match exists between two users
func (r *MatchRepo) Exists(user1ID, user2ID uuid.UUID) (bool, error) {
	// Normalize the order
	if user1ID.String() > user2ID.String() {
		user1ID, user2ID = user2ID, user1ID
	}

	query := `
		SELECT EXISTS(
			SELECT 1 FROM matches 
			WHERE user1_id = $1 AND user2_id = $2
		)
	`
	var exists bool
	err := r.db.QueryRow(query, user1ID, user2ID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
