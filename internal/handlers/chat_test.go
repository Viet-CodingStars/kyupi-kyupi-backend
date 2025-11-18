package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/middleware"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// mockMessageRepo implements a mock message repository for testing
type mockMessageRepo struct {
	messages []*models.Message
}

func newMockMessageRepo() *mockMessageRepo {
	return &mockMessageRepo{
		messages: []*models.Message{},
	}
}

func (m *mockMessageRepo) Create(ctx context.Context, message *models.Message) error {
	m.messages = append(m.messages, message)
	return nil
}

func (m *mockMessageRepo) GetByMatchID(ctx context.Context, matchID uuid.UUID) ([]*models.Message, error) {
	result := []*models.Message{}
	for _, msg := range m.messages {
		if msg.MatchID == matchID {
			result = append(result, msg)
		}
	}
	return result, nil
}

func TestSendMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful message send", func(t *testing.T) {
		mockMessageRepo := newMockMessageRepo()
		mockMatchRepo := newMockMatchRepo()
		handler := NewChatHandler(mockMessageRepo, mockMatchRepo)

		userID := uuid.New()
		receiverID := uuid.New()
		matchID := uuid.New()

		// Create a match between users
		match := &models.Match{
			User1ID: userID,
			User2ID: receiverID,
		}
		mockMatchRepo.Create(match)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := SendMessageRequest{
			MatchID:    matchID.String(),
			ReceiverID: receiverID.String(),
			Content:    "Hello, this is a test message!",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		c.Request = req

		handler.SendMessage(c)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
		}

		var response models.Message
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Content != "Hello, this is a test message!" {
			t.Fatalf("expected content 'Hello, this is a test message!', got %s", response.Content)
		}
		if response.SenderID != userID {
			t.Fatalf("expected sender_id %s, got %s", userID, response.SenderID)
		}
		if response.ReceiverID != receiverID {
			t.Fatalf("expected receiver_id %s, got %s", receiverID, response.ReceiverID)
		}
	})

	t.Run("cannot send message without match", func(t *testing.T) {
		mockMessageRepo := newMockMessageRepo()
		mockMatchRepo := newMockMatchRepo()
		handler := NewChatHandler(mockMessageRepo, mockMatchRepo)

		userID := uuid.New()
		receiverID := uuid.New()
		matchID := uuid.New()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := SendMessageRequest{
			MatchID:    matchID.String(),
			ReceiverID: receiverID.String(),
			Content:    "Hello!",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		c.Request = req

		handler.SendMessage(c)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", w.Code)
		}
	})

	t.Run("invalid match_id format", func(t *testing.T) {
		mockMessageRepo := newMockMessageRepo()
		mockMatchRepo := newMockMatchRepo()
		handler := NewChatHandler(mockMessageRepo, mockMatchRepo)

		userID := uuid.New()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := bytes.NewBufferString(`{"match_id":"invalid-uuid","receiver_id":"` + uuid.New().String() + `","content":"Hello"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", body)
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		c.Request = req

		handler.SendMessage(c)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("unauthenticated request", func(t *testing.T) {
		mockMessageRepo := newMockMessageRepo()
		mockMatchRepo := newMockMatchRepo()
		handler := NewChatHandler(mockMessageRepo, mockMatchRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := bytes.NewBufferString(`{"match_id":"` + uuid.New().String() + `","receiver_id":"` + uuid.New().String() + `","content":"Hello"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", body)
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		handler.SendMessage(c)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", w.Code)
		}
	})
}

func TestGetMessages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful get messages", func(t *testing.T) {
		mockMessageRepo := newMockMessageRepo()
		mockMatchRepo := newMockMatchRepo()
		handler := NewChatHandler(mockMessageRepo, mockMatchRepo)

		userID := uuid.New()
		receiverID := uuid.New()
		matchID := uuid.New()

		// Create a match between users
		match := &models.Match{
			ID:      matchID,
			User1ID: userID,
			User2ID: receiverID,
		}
		mockMatchRepo.matches = append(mockMatchRepo.matches, match)

		// Add some messages
		msg1 := &models.Message{
			MatchID:    matchID,
			SenderID:   userID,
			ReceiverID: receiverID,
			Content:    "Hello",
		}
		msg2 := &models.Message{
			MatchID:    matchID,
			SenderID:   receiverID,
			ReceiverID: userID,
			Content:    "Hi there!",
		}
		mockMessageRepo.messages = append(mockMessageRepo.messages, msg1, msg2)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/matches/"+matchID.String()+"/messages", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		c.Request = req
		c.Params = gin.Params{{Key: "match_id", Value: matchID.String()}}

		handler.GetMessages(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var messages []*models.Message
		if err := json.Unmarshal(w.Body.Bytes(), &messages); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(messages) != 2 {
			t.Fatalf("expected 2 messages, got %d", len(messages))
		}
	})

	t.Run("cannot get messages for unauthorized match", func(t *testing.T) {
		mockMessageRepo := newMockMessageRepo()
		mockMatchRepo := newMockMatchRepo()
		handler := NewChatHandler(mockMessageRepo, mockMatchRepo)

		userID := uuid.New()
		matchID := uuid.New()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/matches/"+matchID.String()+"/messages", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		c.Request = req
		c.Params = gin.Params{{Key: "match_id", Value: matchID.String()}}

		handler.GetMessages(c)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", w.Code)
		}
	})

	t.Run("empty messages returns empty array", func(t *testing.T) {
		mockMessageRepo := newMockMessageRepo()
		mockMatchRepo := newMockMatchRepo()
		handler := NewChatHandler(mockMessageRepo, mockMatchRepo)

		userID := uuid.New()
		receiverID := uuid.New()
		matchID := uuid.New()

		// Create a match between users
		match := &models.Match{
			ID:      matchID,
			User1ID: userID,
			User2ID: receiverID,
		}
		mockMatchRepo.matches = append(mockMatchRepo.matches, match)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/matches/"+matchID.String()+"/messages", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		c.Request = req
		c.Params = gin.Params{{Key: "match_id", Value: matchID.String()}}

		handler.GetMessages(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var messages []*models.Message
		if err := json.Unmarshal(w.Body.Bytes(), &messages); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(messages) != 0 {
			t.Fatalf("expected 0 messages, got %d", len(messages))
		}
	})

	t.Run("unauthenticated request", func(t *testing.T) {
		mockMessageRepo := newMockMessageRepo()
		mockMatchRepo := newMockMatchRepo()
		handler := NewChatHandler(mockMessageRepo, mockMatchRepo)

		matchID := uuid.New()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/matches/"+matchID.String()+"/messages", nil)
		c.Request = req
		c.Params = gin.Params{{Key: "match_id", Value: matchID.String()}}

		handler.GetMessages(c)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", w.Code)
		}
	})
}
