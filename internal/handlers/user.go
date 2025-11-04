package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/auth"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/middleware"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/models"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/repo"
)

type UserHandler struct {
	userRepo  *repo.UserRepo
	jwtSecret string
}

func NewUserHandler(db *sql.DB, jwtSecret string) *UserHandler {
	return &UserHandler{
		userRepo:  repo.NewUserRepo(db),
		jwtSecret: jwtSecret,
	}
}

// SignUpRequest represents the registration request
type SignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// SignInRequest represents the login request
type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token string        `json:"token"`
	User  *models.User  `json:"user"`
}

// UpdateUserRequest represents the update user request
type UpdateUserRequest struct {
	Name      *string `json:"name,omitempty"`
	Gender    *string `json:"gender,omitempty"`
	BirthDate *string `json:"birth_date,omitempty"`
	Bio       *string `json:"bio,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// SignUp handles user registration (POST /api/users)
func (h *UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "email, password, and name are required"})
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to hash password"})
		return
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Name:         req.Name,
	}

	if err := h.userRepo.Create(user); err != nil {
		if err == repo.ErrEmailAlreadyExists {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"error": "email already exists"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create user"})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, h.jwtSecret)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to generate token"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{Token: token, User: user})
}

// SignIn handles user login (POST /api/users/sign_in)
func (h *UserHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid email or password"})
		return
	}

	if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid email or password"})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, h.jwtSecret)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to generate token"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{Token: token, User: user})
}

// SignOut handles user logout (DELETE /api/users/sign_out)
func (h *UserHandler) SignOut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Since JWT is stateless, logout is handled on the client side by discarding the token
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "logged out successfully"})
}

// GetProfile returns the current user's profile (GET /api/users)
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "user not authenticated"})
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// UpdateProfile updates the current user's profile (PATCH/PUT /api/users)
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "user not authenticated"})
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Gender != nil {
		user.Gender = *req.Gender
	}
	if req.BirthDate != nil {
		// Parse the date string (expected format: YYYY-MM-DD)
		birthDate, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid birth_date format (expected YYYY-MM-DD)"})
			return
		}
		user.BirthDate = &birthDate
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
	}
	if req.AvatarURL != nil {
		user.AvatarURL = *req.AvatarURL
	}

	if err := h.userRepo.Update(user); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to update user"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
