package handlers

import (
	"context"
	"net/http"

	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/middleware"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

// MessageRepository defines the interface for message operations
type MessageRepository interface {
	Create(ctx context.Context, message *models.Message) error
	GetByMatchID(ctx context.Context, matchID uuid.UUID) ([]*models.Message, error)
}

type ChatHandler struct {
	messageRepo MessageRepository
	matchRepo   MatchRepository
}

func NewChatHandler(messageRepo MessageRepository, matchRepo MatchRepository) *ChatHandler {
	return &ChatHandler{
		messageRepo: messageRepo,
		matchRepo:   matchRepo,
	}
}

// SendMessageRequest represents the request body for sending a message
type SendMessageRequest struct {
	MatchID    string `json:"match_id" binding:"required"`
	ReceiverID string `json:"receiver_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
}

// SendMessage sends a message to a matched user (POST /api/messages).
// @Summary Send a chat message
// @Description Send a message to a matched user. Requires authentication and active match.
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param message body SendMessageRequest true "Message details"
// @Success 201 {object} models.Message
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/messages [post]
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID, ok := middleware.GetUserID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not authenticated"})
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	// Parse UUIDs
	matchID, err := uuid.Parse(req.MatchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid match_id format"})
		return
	}

	receiverID, err := uuid.Parse(req.ReceiverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid receiver_id format"})
		return
	}

	// Verify match exists and user is part of it
	matchExists, err := h.matchRepo.Exists(userID, receiverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to verify match"})
		return
	}

	if !matchExists {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "no active match found between users"})
		return
	}

	// Create and save message
	message := &models.Message{
		MatchID:    matchID,
		SenderID:   userID,
		ReceiverID: receiverID,
		Content:    req.Content,
	}

	if err := h.messageRepo.Create(c.Request.Context(), message); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to send message"})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// GetMessages retrieves all messages for a specific match (GET /api/matches/:match_id/messages).
// @Summary Get chat messages for a match
// @Description Get all messages for a specific match. Requires authentication and active match.
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param match_id path string true "Match ID"
// @Success 200 {array} models.Message
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/matches/{match_id}/messages [get]
func (h *ChatHandler) GetMessages(c *gin.Context) {
	userID, ok := middleware.GetUserID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not authenticated"})
		return
	}

	matchIDStr := c.Param("match_id")
	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid match_id format"})
		return
	}

	// Get all matches for the user to verify they're part of this match
	matches, err := h.matchRepo.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to verify match"})
		return
	}

	// Check if user is part of the specified match
	isPartOfMatch := false
	for _, match := range matches {
		if match.ID == matchID {
			isPartOfMatch = true
			break
		}
	}

	if !isPartOfMatch {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "not authorized to view messages for this match"})
		return
	}

	// Retrieve messages
	messages, err := h.messageRepo.GetByMatchID(c.Request.Context(), matchID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusOK, []models.Message{})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve messages"})
		return
	}

	if messages == nil {
		messages = []*models.Message{}
	}

	c.JSON(http.StatusOK, messages)
}
