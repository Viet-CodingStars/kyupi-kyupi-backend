package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/middleware"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/models"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/repo"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// mockLikeRepo implements a mock like repository for testing
type mockLikeRepo struct {
	likes       map[string]*models.Like
	mutualLikes map[string]bool
}

// mockMatchRepo implements a mock match repository for testing
type mockMatchRepo struct {
	matches      []*models.Match
	existingPair map[string]bool
}

func newMockLikeRepo() *mockLikeRepo {
	return &mockLikeRepo{
		likes:       make(map[string]*models.Like),
		mutualLikes: make(map[string]bool),
	}
}

func newMockMatchRepo() *mockMatchRepo {
	return &mockMatchRepo{
		matches:      []*models.Match{},
		existingPair: make(map[string]bool),
	}
}

func (m *mockLikeRepo) Create(like *models.Like) error {
	key := like.UserID.String() + "-" + like.TargetUserID.String()
	if _, exists := m.likes[key]; exists {
		return repo.ErrLikeAlreadyExists
	}
	like.ID = uuid.New()
	like.CreatedAt = time.Now()
	like.UpdatedAt = time.Now()
	m.likes[key] = like
	return nil
}

func (m *mockLikeRepo) GetByUserAndTarget(userID, targetUserID uuid.UUID) (*models.Like, error) {
	key := userID.String() + "-" + targetUserID.String()
	like, exists := m.likes[key]
	if !exists {
		return nil, repo.ErrLikeNotFound
	}
	return like, nil
}

func (m *mockLikeRepo) CheckMutualLike(userID, targetUserID uuid.UUID) (bool, error) {
	key := targetUserID.String() + "-" + userID.String()
	if mutual, exists := m.mutualLikes[key]; exists {
		return mutual, nil
	}
	// Check if a like exists from target to user with "like" status
	if like, exists := m.likes[key]; exists && like.Status == "like" {
		return true, nil
	}
	return false, nil
}

func (m *mockMatchRepo) Create(match *models.Match) error {
	user1ID, user2ID := match.User1ID, match.User2ID
	if user1ID.String() > user2ID.String() {
		user1ID, user2ID = user2ID, user1ID
	}
	
	key := user1ID.String() + "-" + user2ID.String()
	if m.existingPair[key] {
		return repo.ErrMatchAlreadyExists
	}
	
	match.ID = uuid.New()
	match.CreatedAt = time.Now()
	match.UpdatedAt = time.Now()
	match.User1ID = user1ID
	match.User2ID = user2ID
	
	m.matches = append(m.matches, match)
	m.existingPair[key] = true
	return nil
}

func (m *mockMatchRepo) GetByUserID(userID uuid.UUID) ([]*models.Match, error) {
	result := []*models.Match{}
	for _, match := range m.matches {
		if match.User1ID == userID || match.User2ID == userID {
			result = append(result, match)
		}
	}
	return result, nil
}

func (m *mockMatchRepo) Exists(user1ID, user2ID uuid.UUID) (bool, error) {
	if user1ID.String() > user2ID.String() {
		user1ID, user2ID = user2ID, user1ID
	}
	key := user1ID.String() + "-" + user2ID.String()
	return m.existingPair[key], nil
}

func TestCreateLike(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful like creation", func(t *testing.T) {
		mockLikeRepo := newMockLikeRepo()
		mockMatchRepo := newMockMatchRepo()
		handler := &LikeHandler{
			likeRepo:  mockLikeRepo,
			matchRepo: mockMatchRepo,
		}

		userID := uuid.New()
		targetUserID := uuid.New()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := bytes.NewBufferString(`{"target_user_id":"` + targetUserID.String() + `","status":"like"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/likes", body)
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		c.Request = req

		handler.CreateLike(c)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", w.Code)
		}

		var response LikeResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Like == nil {
			t.Fatal("expected like in response")
		}
		if response.Like.UserID != userID {
			t.Fatalf("expected user_id %s, got %s", userID, response.Like.UserID)
		}
		if response.Like.TargetUserID != targetUserID {
			t.Fatalf("expected target_user_id %s, got %s", targetUserID, response.Like.TargetUserID)
		}
		if response.Matched {
			t.Fatal("expected matched to be false")
		}
	})

	t.Run("successful like creation with mutual match", func(t *testing.T) {
		mockLikeRepo := newMockLikeRepo()
		mockMatchRepo := newMockMatchRepo()
		handler := &LikeHandler{
			likeRepo:  mockLikeRepo,
			matchRepo: mockMatchRepo,
		}

		userID := uuid.New()
		targetUserID := uuid.New()

		// Create a like from target to user first
		mockLikeRepo.Create(&models.Like{
			UserID:       targetUserID,
			TargetUserID: userID,
			Status:       "like",
		})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := bytes.NewBufferString(`{"target_user_id":"` + targetUserID.String() + `","status":"like"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/likes", body)
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		c.Request = req

		handler.CreateLike(c)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", w.Code)
		}

		var response LikeResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Matched {
			t.Fatal("expected matched to be true")
		}
		if response.Match == nil {
			t.Fatal("expected match in response")
		}
	})

	t.Run("pass does not create match", func(t *testing.T) {
		mockLikeRepo := newMockLikeRepo()
		mockMatchRepo := newMockMatchRepo()
		handler := &LikeHandler{
			likeRepo:  mockLikeRepo,
			matchRepo: mockMatchRepo,
		}

		userID := uuid.New()
		targetUserID := uuid.New()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := bytes.NewBufferString(`{"target_user_id":"` + targetUserID.String() + `","status":"pass"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/likes", body)
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		c.Request = req

		handler.CreateLike(c)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", w.Code)
		}

		var response LikeResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Matched {
			t.Fatal("expected matched to be false for pass")
		}
		if response.Match != nil {
			t.Fatal("expected no match for pass")
		}
	})

	t.Run("cannot like yourself", func(t *testing.T) {
		mockLikeRepo := newMockLikeRepo()
		mockMatchRepo := newMockMatchRepo()
		handler := &LikeHandler{
			likeRepo:  mockLikeRepo,
			matchRepo: mockMatchRepo,
		}

		userID := uuid.New()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := bytes.NewBufferString(`{"target_user_id":"` + userID.String() + `","status":"like"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/likes", body)
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		c.Request = req

		handler.CreateLike(c)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid target_user_id", func(t *testing.T) {
		mockLikeRepo := newMockLikeRepo()
		mockMatchRepo := newMockMatchRepo()
		handler := &LikeHandler{
			likeRepo:  mockLikeRepo,
			matchRepo: mockMatchRepo,
		}

		userID := uuid.New()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := bytes.NewBufferString(`{"target_user_id":"invalid-uuid","status":"like"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/likes", body)
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		c.Request = req

		handler.CreateLike(c)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("duplicate like returns conflict", func(t *testing.T) {
		mockLikeRepo := newMockLikeRepo()
		mockMatchRepo := newMockMatchRepo()
		handler := &LikeHandler{
			likeRepo:  mockLikeRepo,
			matchRepo: mockMatchRepo,
		}

		userID := uuid.New()
		targetUserID := uuid.New()

		// Create first like
		mockLikeRepo.Create(&models.Like{
			UserID:       userID,
			TargetUserID: targetUserID,
			Status:       "like",
		})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := bytes.NewBufferString(`{"target_user_id":"` + targetUserID.String() + `","status":"like"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/likes", body)
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		c.Request = req

		handler.CreateLike(c)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d", w.Code)
		}
	})
}
