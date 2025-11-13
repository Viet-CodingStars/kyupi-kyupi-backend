package repo

import (
  "database/sql"
  "errors"
  "strings"
  "time"

  "github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/models"
  "github.com/google/uuid"
)

var (
  ErrUserNotFound       = errors.New("user not found")
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
  if user.Intention == "" {
    user.Intention = models.DefaultIntention()
  }

  query := `
    INSERT INTO users (email, password_hash, name, gender, birth_date, target_gender, intention, created_at, updated_at)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    RETURNING id
  `
  now := time.Now()
  user.CreatedAt = now
  user.UpdatedAt = now

  var targetGender interface{}
  if user.TargetGender != nil && models.IsValidGender(*user.TargetGender) {
    targetGender = *user.TargetGender
  } else {
    targetGender = nil
  }

  err := r.db.QueryRow(query,
    user.Email, user.PasswordHash, user.Name, user.Gender, user.BirthDate,
    targetGender, user.Intention, user.CreatedAt, user.UpdatedAt,
  ).Scan(&user.ID)

  if err != nil {
    if strings.Contains(err.Error(), "users_email_key") {
      return ErrEmailAlreadyExists
    }
    return err
  }
  return nil
}

// GetByEmail retrieves a user by email
func (r *UserRepo) GetByEmail(email string) (*models.User, error) {
  query := `
    SELECT id, email, password_hash, name, gender, birth_date, target_gender, intention, bio, avatar_url, created_at, updated_at
    FROM users WHERE email = $1
  `
  user := &models.User{}

  var targetGender sql.NullInt64
  var intention sql.NullString
  var bio sql.NullString
  var avatarURL sql.NullString

  err := r.db.QueryRow(query, email).Scan(
    &user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Gender,
    &user.BirthDate, &targetGender, &intention, &bio, &avatarURL,
    &user.CreatedAt, &user.UpdatedAt,
  )
  if err == sql.ErrNoRows {
    return nil, ErrUserNotFound
  }
  if err != nil {
    return nil, err
  }

  if targetGender.Valid {
    val := int(targetGender.Int64)
    user.TargetGender = &val
  }
  if intention.Valid {
    user.Intention = intention.String
  } else {
    user.Intention = models.DefaultIntention()
  }
  if bio.Valid {
    user.Bio = bio.String
  }
  if avatarURL.Valid {
    user.AvatarURL = avatarURL.String
  }

  return user, nil
}

// GetByID retrieves a user by ID
func (r *UserRepo) GetByID(id uuid.UUID) (*models.User, error) {
  query := `
    SELECT id, email, password_hash, name, gender, birth_date, target_gender, intention, bio, avatar_url, created_at, updated_at
    FROM users WHERE id = $1
  `
  user := &models.User{}

  var targetGender sql.NullInt64
  var intention sql.NullString
  var bio sql.NullString
  var avatarURL sql.NullString

  err := r.db.QueryRow(query, id).Scan(
    &user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Gender,
    &user.BirthDate, &targetGender, &intention, &bio, &avatarURL,
    &user.CreatedAt, &user.UpdatedAt,
  )
  if err == sql.ErrNoRows {
    return nil, ErrUserNotFound
  }
  if err != nil {
    return nil, err
  }

  if targetGender.Valid {
    val := int(targetGender.Int64)
    user.TargetGender = &val
  }
  if intention.Valid {
    user.Intention = intention.String
  } else {
    user.Intention = models.DefaultIntention()
  }
  if bio.Valid {
    user.Bio = bio.String
  }
  if avatarURL.Valid {
    user.AvatarURL = avatarURL.String
  }

  return user, nil
}

// Update updates a user's information
func (r *UserRepo) Update(user *models.User) error {
  if user.Intention == "" {
    user.Intention = models.DefaultIntention()
  }

  query := `
    UPDATE users
    SET name = $1, gender = $2, birth_date = $3, bio = $4, avatar_url = $5, target_gender = $6, intention = $7, updated_at = NOW()
    WHERE id = $8
    RETURNING updated_at
  `
  var targetGender interface{}
  if user.TargetGender != nil && models.IsValidGender(*user.TargetGender) {
    targetGender = *user.TargetGender
  } else {
    targetGender = nil
  }

  err := r.db.QueryRow(query,
    user.Name, user.Gender, user.BirthDate, user.Bio, user.AvatarURL,
    targetGender, user.Intention, user.ID,
  ).Scan(&user.UpdatedAt)

  if err == sql.ErrNoRows {
    return ErrUserNotFound
  }
  return err
}
