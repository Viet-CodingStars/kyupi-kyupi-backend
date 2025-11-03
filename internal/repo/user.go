package repo

import (
	"database/sql"
	"errors"

	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/models"
	"github.com/google/uuid"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// UserRepo handles database operations for users
type UserRepo struct {
	db *sql.DB
}

// NewUserRepo creates a new UserRepo
func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create inserts a new user into the database
func (r *UserRepo) Create(user *models.User) error {
	query := `
		INSERT INTO users (email, password_hash, name, gender, birth_date, bio, avatar_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(query, user.Email, user.PasswordHash, user.Name, user.Gender, user.BirthDate, user.Bio, user.AvatarURL).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
			return ErrEmailAlreadyExists
		}
		return err
	}
	return nil
}

// GetByEmail retrieves a user by email
func (r *UserRepo) GetByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, name, gender, birth_date, bio, avatar_url, created_at, updated_at
		FROM users WHERE email = $1
	`
	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Gender,
		&user.BirthDate, &user.Bio, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByID retrieves a user by ID
func (r *UserRepo) GetByID(id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, name, gender, birth_date, bio, avatar_url, created_at, updated_at
		FROM users WHERE id = $1
	`
	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Gender,
		&user.BirthDate, &user.Bio, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Update updates a user's information
func (r *UserRepo) Update(user *models.User) error {
	query := `
		UPDATE users
		SET name = $1, gender = $2, birth_date = $3, bio = $4, avatar_url = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at
	`
	err := r.db.QueryRow(query, user.Name, user.Gender, user.BirthDate, user.Bio, user.AvatarURL, user.ID).
		Scan(&user.UpdatedAt)
	if err == sql.ErrNoRows {
		return ErrUserNotFound
	}
	return err
}
