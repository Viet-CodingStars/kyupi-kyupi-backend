package handlers

import (
  "database/sql"
  "fmt"
  "net/http"
  "time"

  "github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/auth"
  "github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/middleware"
  "github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/models"
  "github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/repo"
  "github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/storage"
  "github.com/gin-gonic/gin"
  "github.com/google/uuid"
)

// UserRepository declares the minimal persistence operations required by UserHandler.
type UserRepository interface {
  Create(user *models.User) error
  GetByEmail(email string) (*models.User, error)
  GetByID(id uuid.UUID) (*models.User, error)
  Update(user *models.User) error
  GetUsers(cursor *uuid.UUID, limit int) ([]*models.User, error)
  GetUserDetail(id uuid.UUID) (*models.User, error)
}

type UserHandler struct {
  userRepo      UserRepository
  jwtSecret     string
  avatarStorage storage.AvatarStorage
}

func NewUserHandler(db *sql.DB, jwtSecret string, avatarStorage storage.AvatarStorage) *UserHandler {
  return &UserHandler{
    userRepo:      repo.NewUserRepo(db),
    jwtSecret:     jwtSecret,
    avatarStorage: avatarStorage,
  }
}

// SignUpRequest represents the registration request.
type SignUpRequest struct {
  Email        string  `json:"email"`
  Password     string  `json:"password"`
  Name         string  `json:"name"`
  Gender       int     `json:"gender"`
  BirthDate    string  `json:"birth_date"` // Accepts string "YYYY-MM-DD" for validation
  TargetGender *int    `json:"target_gender,omitempty"`
  Intention    *string `json:"intention,omitempty"`
}

// SignInRequest represents the login request.
type SignInRequest struct {
  Email    string `json:"email"`
  Password string `json:"password"`
}

// AuthResponse represents the authentication response payload.
type AuthResponse struct {
  Token string       `json:"token"`
  User  *models.User `json:"user"`
}

// ErrorResponse represents an error payload.
type ErrorResponse struct {
  Error string `json:"error"`
}

// MessageResponse represents a simple message payload.
type MessageResponse struct {
  Message string `json:"message"`
}

// UpdateUserRequest represents the update user request payload.
type UpdateUserRequest struct {
  Name         *string `json:"name,omitempty"`
  Gender       *int    `json:"gender,omitempty"`
  BirthDate    *string `json:"birth_date,omitempty"` // Remains a *string in "YYYY-MM-DD" format
  Bio          *string `json:"bio,omitempty"`
  AvatarURL    *string `json:"avatar_url,omitempty"`
  TargetGender *int    `json:"target_gender,omitempty"`
  Intention    *string `json:"intention,omitempty"`
}

// UserListItem represents a minimal user item in the list response
type UserListItem struct {
  ID        uuid.UUID `json:"id"`
  Name      string    `json:"name"`
  Gender    int       `json:"gender"`
  AvatarURL string    `json:"avatar_url"`
  Intention string    `json:"intention"`
}

// GetUsersResponse represents the response for the user list endpoint
type GetUsersResponse struct {
  Users      []UserListItem `json:"users"`
  NextCursor *uuid.UUID     `json:"next_cursor"`
}

// UserDetail represents detailed user information
type UserDetail struct {
  ID           uuid.UUID  `json:"id"`
  Name         string     `json:"name"`
  Gender       int        `json:"gender"`
  TargetGender *int       `json:"target_gender,omitempty"`
  Intention    string     `json:"intention"`
  Bio          string     `json:"bio,omitempty"`
  AvatarURL    string     `json:"avatar_url,omitempty"`
  CreatedAt    time.Time  `json:"created_at"`
  UpdatedAt    time.Time  `json:"updated_at"`
}

// validateGender checks if the provided gender matches the supported enum values.
func validateGender(gender int) bool {
  return models.IsValidGender(gender)
}

// validateIntention checks if the provided intention matches the supported enum values.
func validateIntention(intention string) bool {
  return models.IsValidIntention(intention)
}

// validateAge checks if birthdate is at least 18 years ago
func validateAge(birthDate time.Time) bool {
  eighteenYearsAgo := time.Now().AddDate(-18, 0, 0)
  return birthDate.Before(eighteenYearsAgo) || birthDate.Equal(eighteenYearsAgo)
}

// parseBirthDate parses a "YYYY-MM-DD" string
func parseBirthDate(dateString string) (time.Time, error) {
  birthDate, err := time.Parse("2006-01-02", dateString)
  if err != nil {
    return time.Time{}, fmt.Errorf("invalid birth_date format (expected YYYY-MM-DD)")
  }
  return birthDate, nil
}

// UploadAvatar handles avatar uploads for the current user.
// @Summary Upload avatar image
// @Description Uploads an avatar image for the current user and updates the profile.
// @Tags Users
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "Avatar image"
// @Success 200 {object} models.User
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/users/profile/avatar [post]
func (h *UserHandler) UploadAvatar(c *gin.Context) {
  userID, ok := middleware.GetUserID(c.Request.Context())
  if !ok {
    c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not authenticated"})
    return
  }

  if h.avatarStorage == nil {
    c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "avatar storage not configured"})
    return
  }

  fileHeader, err := c.FormFile("avatar")
  if err != nil {
    c.JSON(http.StatusBadRequest, ErrorResponse{Error: "avatar file is required"})
    return
  }

  const maxAvatarSize = 5 << 20
  if fileHeader.Size > maxAvatarSize {
    c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("avatar file too large (max %d bytes)", maxAvatarSize)})
    return
  }

  file, err := fileHeader.Open()
  if err != nil {
    c.JSON(http.StatusBadRequest, ErrorResponse{Error: "unable to read avatar file"})
    return
  }
  defer file.Close()

  avatarURL, err := h.avatarStorage.Save(userID, file, fileHeader.Filename)
  if err != nil {
    c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to store avatar"})
    return
  }

  user, err := h.userRepo.GetByID(userID)
  if err != nil {
    c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
    return
  }

  user.AvatarURL = avatarURL
  if err := h.userRepo.Update(user); err != nil {
    c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update user"})
    return
  }

  c.JSON(http.StatusOK, user)
}

// SignUp handles user registration (POST /api/users).
// @Summary Register a new user
// @Description Creates a new account and returns a JWT token for the user.
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body SignUpRequest true "User sign up payload"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/users [post]
func (h *UserHandler) SignUp(c *gin.Context) {
  var req SignUpRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
    return
  }

  if req.Email == "" || req.Password == "" || req.Name == "" || req.BirthDate == "" || req.Gender == 0 {
    c.JSON(http.StatusBadRequest, ErrorResponse{Error: "email, password, name, gender, and birth_date are required"})
    return
  }

  if !validateGender(req.Gender) {
    c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid gender (1=male, 2=female, 3=others)"})
    return
  }

  birthDate, err := parseBirthDate(req.BirthDate)
  if err != nil {
    c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
    return
  }

  if !validateAge(birthDate) {
    c.JSON(http.StatusBadRequest, ErrorResponse{Error: "user must be at least 18 years old"})
    return
  }

  hashedPassword, err := auth.HashPassword(req.Password)
  if err != nil {
    c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to hash password"})
    return
  }

  var targetGender *int
  if req.TargetGender != nil {
    if !validateGender(*req.TargetGender) {
      c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid target_gender (1=male, 2=female, 3=others)"})
      return
    }
    val := *req.TargetGender
    targetGender = &val
  }

  intention := models.DefaultIntention()
  if req.Intention != nil {
    if !validateIntention(*req.Intention) {
      c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid intention value"})
      return
    }
    intention = *req.Intention
  }

  user := &models.User{
    Email:        req.Email,
    PasswordHash: hashedPassword,
    Name:         req.Name,
    Gender:       req.Gender,
    BirthDate:    birthDate,
    TargetGender: targetGender,
    Intention:    intention,
  }

  if err := h.userRepo.Create(user); err != nil {
    if err == repo.ErrEmailAlreadyExists {
      c.JSON(http.StatusConflict, ErrorResponse{Error: "email already exists"})
      return
    }
    c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create user"})
    return
  }

  token, err := auth.GenerateToken(user.ID, user.Email, h.jwtSecret)
  if err != nil {
    c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate token"})
    return
  }

  c.JSON(http.StatusCreated, AuthResponse{Token: token, User: user})
}

// SignIn handles user login (POST /api/users/sign_in).
// @Summary Authenticate a user
// @Description Verifies credentials and returns a JWT token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body SignInRequest true "User sign in payload"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/users/sign_in [post]
func (h *UserHandler) SignIn(c *gin.Context) {
  var req SignInRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
    return
  }

  user, err := h.userRepo.GetByEmail(req.Email)
  if err != nil {
    c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid email or password"})
    return
  }

  if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
    c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid email or password"})
    return
  }

  token, err := auth.GenerateToken(user.ID, user.Email, h.jwtSecret)
  if err != nil {
    c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate token"})
    return
  }

  c.JSON(http.StatusOK, AuthResponse{Token: token, User: user})
}

// SignOut handles user logout (DELETE /api/users/sign_out).
// @Summary Sign out the current user
// @Description Stateless logout helper that simply acknowledges the request.
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} MessageResponse
// @Router /api/users/sign_out [delete]
func (h *UserHandler) SignOut(c *gin.Context) {
  c.JSON(http.StatusOK, MessageResponse{Message: "logged out successfully"})
}

// GetProfile returns the current user's profile (GET /api/users/profile).
// @Summary Get current user profile
// @Description Returns the authenticated user's profile details.
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
  userID, ok := middleware.GetUserID(c.Request.Context())
  if !ok {
    c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not authenticated"})
    return
  }

  user, err := h.userRepo.GetByID(userID)
  if err != nil {
    c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
    return
  }

  c.JSON(http.StatusOK, user)
}

// UpdateProfile updates the current user's profile (PATCH/PUT /api/users/profile).
// @Summary Update current user profile
// @Description Applies partial or full updates to the authenticated user's profile.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body UpdateUserRequest true "Profile update payload"
// @Success 200 {object} models.User
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/users/profile [patch]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
  userID, ok := middleware.GetUserID(c.Request.Context())
  if !ok {
    c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not authenticated"})
    return
  }

  user, err := h.userRepo.GetByID(userID)
  if err != nil {
    c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
    return
  }

  var req UpdateUserRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
    return
  }

  if req.Name != nil {
    if *req.Name == "" {
      c.JSON(http.StatusBadRequest, ErrorResponse{Error: "name cannot be empty"})
      return
    }
    user.Name = *req.Name
  }
  if req.Gender != nil {
    if !validateGender(*req.Gender) {
      c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid gender (1=male, 2=female, 3=others)"})
      return
    }
    user.Gender = *req.Gender
  }
  if req.BirthDate != nil {
    birthDate, err := parseBirthDate(*req.BirthDate)
    if err != nil {
      c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
      return
    }
    if !validateAge(birthDate) {
      c.JSON(http.StatusBadRequest, ErrorResponse{Error: "user must be at least 18 years old"})
      return
    }
    user.BirthDate = birthDate
  }
  if req.Bio != nil {
    user.Bio = *req.Bio
  }
  if req.AvatarURL != nil {
    user.AvatarURL = *req.AvatarURL
  }
  if req.TargetGender != nil {
    if !validateGender(*req.TargetGender) {
      c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid target_gender (1=male, 2=female, 3=others)"})
      return
    }
    val := *req.TargetGender
    user.TargetGender = &val
  }
  if req.Intention != nil {
    if !validateIntention(*req.Intention) {
      c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid intention value"})
      return
    }
    user.Intention = *req.Intention
  }

  if err := h.userRepo.Update(user); err != nil {
    c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update user"})
    return
  }

  c.JSON(http.StatusOK, user)
}

// GetUsers returns a paginated list of users (GET /api/users).
// @Summary Get list of users
// @Description Returns a paginated list of users with minimal information for swipe screen.
// @Tags Users
// @Produce json
// @Param cursor query string false "Cursor for pagination (UUID)"
// @Param limit query int false "Number of users to return (default: 20, max: 100)"
// @Success 200 {object} GetUsersResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
  // Parse cursor parameter
  var cursor *uuid.UUID
  if cursorStr := c.Query("cursor"); cursorStr != "" {
    cursorUUID, err := uuid.Parse(cursorStr)
    if err != nil {
      c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid cursor format"})
      return
    }
    cursor = &cursorUUID
  }

  // Parse limit parameter with default of 20
  limit := 20
  if limitStr := c.Query("limit"); limitStr != "" {
    fmt.Sscanf(limitStr, "%d", &limit)
    if limit < 1 || limit > 100 {
      c.JSON(http.StatusBadRequest, ErrorResponse{Error: "limit must be between 1 and 100"})
      return
    }
  }

  // Get users from repository
  users, err := h.userRepo.GetUsers(cursor, limit)
  if err != nil {
    c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve users"})
    return
  }

  // Convert to response format
  userItems := make([]UserListItem, len(users))
  for i, user := range users {
    userItems[i] = UserListItem{
      ID:        user.ID,
      Name:      user.Name,
      Gender:    user.Gender,
      AvatarURL: user.AvatarURL,
      Intention: user.Intention,
    }
  }

  // Determine next cursor
  var nextCursor *uuid.UUID
  if len(users) == limit {
    lastID := users[len(users)-1].ID
    nextCursor = &lastID
  }

  c.JSON(http.StatusOK, GetUsersResponse{
    Users:      userItems,
    NextCursor: nextCursor,
  })
}

// GetUserDetail returns detailed information about a specific user (GET /api/users/:user_id/detail).
// @Summary Get user detail
// @Description Returns detailed information about a specific user.
// @Tags Users
// @Produce json
// @Param user_id path string true "User ID (UUID)"
// @Success 200 {object} UserDetail
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/users/{user_id}/detail [get]
func (h *UserHandler) GetUserDetail(c *gin.Context) {
  userIDStr := c.Param("user_id")
  userID, err := uuid.Parse(userIDStr)
  if err != nil {
    c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user ID format"})
    return
  }

  user, err := h.userRepo.GetUserDetail(userID)
  if err != nil {
    if err == repo.ErrUserNotFound {
      c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
      return
    }
    c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve user detail"})
    return
  }

  detail := UserDetail{
    ID:           user.ID,
    Name:         user.Name,
    Gender:       user.Gender,
    TargetGender: user.TargetGender,
    Intention:    user.Intention,
    Bio:          user.Bio,
    AvatarURL:    user.AvatarURL,
    CreatedAt:    user.CreatedAt,
    UpdatedAt:    user.UpdatedAt,
  }

  c.JSON(http.StatusOK, detail)
}
