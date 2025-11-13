package repo

import (
	"database/sql"
	"errors"

	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/models"
	"github.com/google/uuid"
)

var (
	ErrLikeNotFound      = errors.New("like not found")
	ErrLikeAlreadyExists = errors.New("like already exists for this pair")
)

// LikeRepo handles database operations for likes
type LikeRepo struct {
	db *sql.DB
}

// NewLikeRepo creates a new LikeRepo
func NewLikeRepo(db *sql.DB) *LikeRepo {
	return &LikeRepo{db: db}
}

// Create inserts a new like into the database
func (r *LikeRepo) Create(like *models.Like) error {
	query := `
		INSERT INTO likes (user_id, target_user_id, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(query, like.UserID, like.TargetUserID, like.Status).
		Scan(&like.ID, &like.CreatedAt, &like.UpdatedAt)
	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"likes_user_id_target_user_id_key\"" {
			return ErrLikeAlreadyExists
		}
		return err
	}
	return nil
}

// GetByUserAndTarget retrieves a like by user_id and target_user_id
func (r *LikeRepo) GetByUserAndTarget(userID, targetUserID uuid.UUID) (*models.Like, error) {
	query := `
		SELECT id, user_id, target_user_id, status, created_at, updated_at
		FROM likes WHERE user_id = $1 AND target_user_id = $2
	`
	like := &models.Like{}
	err := r.db.QueryRow(query, userID, targetUserID).Scan(
		&like.ID, &like.UserID, &like.TargetUserID, &like.Status,
		&like.CreatedAt, &like.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrLikeNotFound
	}
	if err != nil {
		return nil, err
	}
	return like, nil
}

// CheckMutualLike checks if there is a mutual like between two users
// Returns true if targetUser has liked the originalUser
func (r *LikeRepo) CheckMutualLike(userID, targetUserID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM likes 
			WHERE user_id = $1 AND target_user_id = $2 AND status = 'like'
		)
	`
	var exists bool
	err := r.db.QueryRow(query, targetUserID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
