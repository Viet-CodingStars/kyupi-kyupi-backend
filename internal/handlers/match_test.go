package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/middleware"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestGetMatches(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful get matches", func(t *testing.T) {
		mockMatchRepo := newMockMatchRepo()
		handler := &MatchHandler{
			matchRepo: mockMatchRepo,
		}

		userID := uuid.New()
		otherUserID := uuid.New()

		// Create some matches
		match1 := &models.Match{
			ID:        uuid.New(),
			User1ID:   userID,
			User2ID:   otherUserID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		mockMatchRepo.matches = append(mockMatchRepo.matches, match1)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/matches", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		c.Request = req

		handler.GetMatches(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var matches []*models.Match
		if err := json.Unmarshal(w.Body.Bytes(), &matches); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(matches) != 1 {
			t.Fatalf("expected 1 match, got %d", len(matches))
		}
		if matches[0].User1ID != userID {
			t.Fatalf("expected user1_id %s, got %s", userID, matches[0].User1ID)
		}
	})

	t.Run("no matches returns empty array", func(t *testing.T) {
		mockMatchRepo := newMockMatchRepo()
		handler := &MatchHandler{
			matchRepo: mockMatchRepo,
		}

		userID := uuid.New()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/matches", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		c.Request = req

		handler.GetMatches(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var matches []*models.Match
		if err := json.Unmarshal(w.Body.Bytes(), &matches); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(matches) != 0 {
			t.Fatalf("expected 0 matches, got %d", len(matches))
		}
	})

	t.Run("unauthenticated request", func(t *testing.T) {
		mockMatchRepo := newMockMatchRepo()
		handler := &MatchHandler{
			matchRepo: mockMatchRepo,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/matches", nil)
		c.Request = req

		handler.GetMatches(c)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", w.Code)
		}
	})
}
