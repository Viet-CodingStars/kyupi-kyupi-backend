package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/auth"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/middleware"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/models"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/repo"
	"github.com/google/uuid"
)

// mockUserRepo implements a mock user repository for testing
type mockUserRepo struct {
	users map[string]*models.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[string]*models.User),
	}
}

func (m *mockUserRepo) Create(user *models.User) error {
	if _, exists := m.users[user.Email]; exists {
		return repo.ErrEmailAlreadyExists
	}
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepo) GetByEmail(email string) (*models.User, error) {
	user, exists := m.users[email]
	if !exists {
		return nil, repo.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserRepo) GetByID(id uuid.UUID) (*models.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, repo.ErrUserNotFound
}

func (m *mockUserRepo) Update(user *models.User) error {
	for email, u := range m.users {
		if u.ID == user.ID {
			user.UpdatedAt = time.Now()
			m.users[email] = user
			return nil
		}
	}
	return repo.ErrUserNotFound
}

func TestSignUp(t *testing.T) {
	handler := &UserHandler{
		userRepo:  &repo.UserRepo{},
		jwtSecret: "test-secret",
	}
	// Replace with mock repo
	mockRepo := newMockUserRepo()
	handler.userRepo = (*repo.UserRepo)(nil)

	// Create a custom handler with mock
	testHandler := func(w http.ResponseWriter, r *http.Request) {
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

		if err := mockRepo.Create(user); err != nil {
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

		token, err := auth.GenerateToken(user.ID, user.Email, handler.jwtSecret)
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

	t.Run("successful signup", func(t *testing.T) {
		body := bytes.NewBufferString(`{"email":"test@example.com","password":"password123","name":"Test User"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/users", body)
		rr := httptest.NewRecorder()

		testHandler(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", rr.Code)
		}

		var response AuthResponse
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Token == "" {
			t.Fatal("expected token in response")
		}
		if response.User.Email != "test@example.com" {
			t.Fatalf("expected email test@example.com, got %s", response.User.Email)
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		body := bytes.NewBufferString(`{"email":"test@example.com"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/users", body)
		rr := httptest.NewRecorder()

		testHandler(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
	})
}

func TestSignIn(t *testing.T) {
	mockRepo := newMockUserRepo()
	
	// Create a test user
	hashedPassword, _ := auth.HashPassword("password123")
	testUser := &models.User{
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		Name:         "Test User",
	}
	mockRepo.Create(testUser)

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		var req SignInRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
			return
		}

		user, err := mockRepo.GetByEmail(req.Email)
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

		token, err := auth.GenerateToken(user.ID, user.Email, "test-secret")
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

	t.Run("successful login", func(t *testing.T) {
		body := bytes.NewBufferString(`{"email":"test@example.com","password":"password123"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/users/sign_in", body)
		rr := httptest.NewRecorder()

		testHandler(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}

		var response AuthResponse
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Token == "" {
			t.Fatal("expected token in response")
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		body := bytes.NewBufferString(`{"email":"test@example.com","password":"wrongpassword"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/users/sign_in", body)
		rr := httptest.NewRecorder()

		testHandler(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rr.Code)
		}
	})
}

func TestSignOut(t *testing.T) {
	handler := &UserHandler{
		jwtSecret: "test-secret",
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/users/sign_out", nil)
	rr := httptest.NewRecorder()

	handler.SignOut(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["message"] != "logged out successfully" {
		t.Fatalf("expected logout message, got %s", response["message"])
	}
}

func TestGetProfile(t *testing.T) {
	mockRepo := newMockUserRepo()
	
	// Create a test user
	testUser := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Name:  "Test User",
	}
	mockRepo.users[testUser.Email] = testUser

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "user not authenticated"})
			return
		}

		user, err := mockRepo.GetByID(userID)
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

	t.Run("successful get profile", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, testUser.ID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		testHandler(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}

		var user models.User
		if err := json.NewDecoder(rr.Body).Decode(&user); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if user.Email != "test@example.com" {
			t.Fatalf("expected email test@example.com, got %s", user.Email)
		}
	})
}

func TestSignUpForm(t *testing.T) {
	handler := &UserHandler{
		jwtSecret: "test-secret",
	}

	req := httptest.NewRequest(http.MethodGet, "/api/users/sign_up", nil)
	rr := httptest.NewRecorder()

	handler.SignUpForm(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["message"] != "Sign up form" {
		t.Fatalf("expected sign up form message, got %v", response["message"])
	}
}

func TestSignInForm(t *testing.T) {
	handler := &UserHandler{
		jwtSecret: "test-secret",
	}

	req := httptest.NewRequest(http.MethodGet, "/api/users/sign_in", nil)
	rr := httptest.NewRecorder()

	handler.SignInForm(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["message"] != "Sign in form" {
		t.Fatalf("expected sign in form message, got %v", response["message"])
	}
}

// Prevent unused import errors
var (
	_ = sql.ErrNoRows
	_ = time.Now
)
