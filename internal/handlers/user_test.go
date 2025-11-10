package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/auth"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/middleware"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/models"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/repo"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// mockUserRepo implements a mock user repository for testing
type mockUserRepo struct {
	users map[string]*models.User
}

type mockAvatarStorage struct {
	saved map[uuid.UUID]string
}

var _ storage.AvatarStorage = (*mockAvatarStorage)(nil)

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[string]*models.User),
	}
}

func (m *mockAvatarStorage) Save(userID uuid.UUID, data io.Reader, originalFilename string) (string, error) {
	if m.saved == nil {
		m.saved = make(map[uuid.UUID]string)
	}
	io.Copy(io.Discard, data)
	path := "/avatars/" + userID.String() + "/mock-file.png"
	m.saved[userID] = path
	return path, nil
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
	handler := &UserHandler{userRepo: mockRepo, jwtSecret: "test-secret", avatarStorage: &mockAvatarStorage{}}

	t.Run("successful signup", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := bytes.NewBufferString(`{"email":"test@example.com","password":"password123","name":"Test User", "gender": 1, "birth_date": "2000-01-01", "target_gender": 2}`)
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
		if response.User.TargetGender == nil || *response.User.TargetGender != models.GenderFemale {
			if response.User.TargetGender == nil {
				t.Fatal("expected target gender female, got <nil>")
			}
			t.Fatalf("expected target gender female, got %d", *response.User.TargetGender)
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

	t.Run("invalid target gender", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := bytes.NewBufferString(`{"email":"test2@example.com","password":"password123","name":"Test User", "gender": 1, "birth_date": "2000-01-01", "target_gender": 9}`)
		req := httptest.NewRequest(http.MethodPost, "/api/users", body)
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		handler.SignUp(c)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("user under 18", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := bytes.NewBufferString(`{"email":"test3@example.com","password":"password123","name":"Test User", "gender": 1, "birth_date": "2010-01-01"}`)
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
		Gender:       models.GenderMale,
	})

	handler := &UserHandler{userRepo: mockRepo, jwtSecret: "test-secret", avatarStorage: &mockAvatarStorage{}}

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
		ID:     uuid.New(),
		Email:  "test@example.com",
		Name:   "Test User",
		Gender: models.GenderMale,
	}
	mockRepo.users[testUser.Email] = testUser

	handler := &UserHandler{userRepo: mockRepo, jwtSecret: "test-secret", avatarStorage: &mockAvatarStorage{}}

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

func TestUploadAvatar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := newMockUserRepo()
	mockStorage := &mockAvatarStorage{}
	testUser := &models.User{
		ID:     uuid.New(),
		Email:  "test@example.com",
		Name:   "Test User",
		Gender: models.GenderMale,
	}
	mockRepo.users[testUser.Email] = testUser

	handler := &UserHandler{userRepo: mockRepo, jwtSecret: "test-secret", avatarStorage: mockStorage}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("avatar", "avatar.png")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := fileWriter.Write([]byte("fakeimagecontent")); err != nil {
		t.Fatalf("failed to write mock file: %v", err)
	}
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/users/profile/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, testUser.ID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.UploadAvatar(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var user models.User
	if err := json.Unmarshal(w.Body.Bytes(), &user); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if user.AvatarURL == "" {
		t.Fatal("expected avatar URL in response")
	}
	if _, exists := mockStorage.saved[testUser.ID]; !exists {
		t.Fatal("expected avatar to be saved in storage")
	}
}
