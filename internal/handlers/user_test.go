package handlers

import (
  "bytes"
  "context"
  "encoding/json"
  "net/http"
  "net/http/httptest"
  "testing"
  "time"

  "github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/auth"
  "github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/middleware"
  "github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/models"
  "github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/repo"
  "github.com/gin-gonic/gin"
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
  gin.SetMode(gin.TestMode)
  mockRepo := newMockUserRepo()
  handler := &UserHandler{userRepo: mockRepo, jwtSecret: "test-secret"}

  t.Run("successful signup", func(t *testing.T) {
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    body := bytes.NewBufferString(`{"email":"test@example.com","password":"password123","name":"Test User"}`)
    req := httptest.NewRequest(http.MethodPost, "/api/users", body)
    req.Header.Set("Content-Type", "application/json")
    c.Request = req

    handler.SignUp(c)

    if w.Code != http.StatusCreated {
      t.Fatalf("expected status 201, got %d", w.Code)
    }

    var response AuthResponse
    if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
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
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    body := bytes.NewBufferString(`{"email":"test@example.com"}`)
    req := httptest.NewRequest(http.MethodPost, "/api/users", body)
    req.Header.Set("Content-Type", "application/json")
    c.Request = req

    handler.SignUp(c)

    if w.Code != http.StatusBadRequest {
      t.Fatalf("expected status 400, got %d", w.Code)
    }
  })
}

func TestSignIn(t *testing.T) {
  gin.SetMode(gin.TestMode)
  mockRepo := newMockUserRepo()
  hashedPassword, _ := auth.HashPassword("password123")
  mockRepo.Create(&models.User{
    Email:        "test@example.com",
    PasswordHash: hashedPassword,
    Name:         "Test User",
  })

  handler := &UserHandler{userRepo: mockRepo, jwtSecret: "test-secret"}

  t.Run("successful login", func(t *testing.T) {
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    body := bytes.NewBufferString(`{"email":"test@example.com","password":"password123"}`)
    req := httptest.NewRequest(http.MethodPost, "/api/users/sign_in", body)
    req.Header.Set("Content-Type", "application/json")
    c.Request = req

    handler.SignIn(c)

    if w.Code != http.StatusOK {
      t.Fatalf("expected status 200, got %d", w.Code)
    }

    var response AuthResponse
    if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
      t.Fatalf("failed to decode response: %v", err)
    }

    if response.Token == "" {
      t.Fatal("expected token in response")
    }
  })

  t.Run("invalid credentials", func(t *testing.T) {
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    body := bytes.NewBufferString(`{"email":"test@example.com","password":"wrongpassword"}`)
    req := httptest.NewRequest(http.MethodPost, "/api/users/sign_in", body)
    req.Header.Set("Content-Type", "application/json")
    c.Request = req

    handler.SignIn(c)

    if w.Code != http.StatusUnauthorized {
      t.Fatalf("expected status 401, got %d", w.Code)
    }
  })
}

func TestSignOut(t *testing.T) {
  gin.SetMode(gin.TestMode)
  handler := &UserHandler{jwtSecret: "test-secret"}

  w := httptest.NewRecorder()
  c, _ := gin.CreateTestContext(w)
  c.Request = httptest.NewRequest(http.MethodDelete, "/api/users/sign_out", nil)

  handler.SignOut(c)

  if w.Code != http.StatusOK {
    t.Fatalf("expected status 200, got %d", w.Code)
  }

  var response map[string]string
  if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
    t.Fatalf("failed to decode response: %v", err)
  }

  if response["message"] != "logged out successfully" {
    t.Fatalf("expected logout message, got %s", response["message"])
  }
}

func TestGetProfile(t *testing.T) {
  gin.SetMode(gin.TestMode)
  mockRepo := newMockUserRepo()
  testUser := &models.User{
    ID:    uuid.New(),
    Email: "test@example.com",
    Name:  "Test User",
  }
  mockRepo.users[testUser.Email] = testUser

  handler := &UserHandler{userRepo: mockRepo, jwtSecret: "test-secret"}

  t.Run("successful get profile", func(t *testing.T) {
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    req := httptest.NewRequest(http.MethodGet, "/api/users/profile", nil)
    ctx := context.WithValue(req.Context(), middleware.UserIDKey, testUser.ID)
    req = req.WithContext(ctx)
    c.Request = req

    handler.GetProfile(c)

    if w.Code != http.StatusOK {
      t.Fatalf("expected status 200, got %d", w.Code)
    }

    var user models.User
    if err := json.Unmarshal(w.Body.Bytes(), &user); err != nil {
      t.Fatalf("failed to decode response: %v", err)
    }

    if user.Email != "test@example.com" {
      t.Fatalf("expected email test@example.com, got %s", user.Email)
    }
  })
}
