package handlers

import (
	"database/sql"
	"net/http"

	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/middleware"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/models"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/repo"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LikeRepository declares the minimal persistence operations required by LikeHandler.
type LikeRepository interface {
	Create(like *models.Like) error
	GetByUserAndTarget(userID, targetUserID uuid.UUID) (*models.Like, error)
	CheckMutualLike(userID, targetUserID uuid.UUID) (bool, error)
}

// MatchRepository declares the minimal persistence operations required by LikeHandler.
type MatchRepository interface {
	Create(match *models.Match) error
	GetByUserID(userID uuid.UUID) ([]*models.Match, error)
	Exists(user1ID, user2ID uuid.UUID) (bool, error)
}

type LikeHandler struct {
	likeRepo  LikeRepository
	matchRepo MatchRepository
}

func NewLikeHandler(db *sql.DB) *LikeHandler {
	return &LikeHandler{
		likeRepo:  repo.NewLikeRepo(db),
		matchRepo: repo.NewMatchRepo(db),
	}
}

// CreateLikeRequest represents the like/pass request.
type CreateLikeRequest struct {
	TargetUserID string `json:"target_user_id" binding:"required"`
	Status       string `json:"status" binding:"required,oneof=like pass"`
}

// LikeResponse represents the response after creating a like
type LikeResponse struct {
	Like    *models.Like  `json:"like"`
	Match   *models.Match `json:"match,omitempty"`
	Matched bool          `json:"matched"`
}

// CreateLike handles creating a like or pass (POST /api/v1/likes).
// @Summary Create a like or pass
// @Description Creates a like or pass for a target user. If both users like each other, a match is automatically created.
// @Tags Likes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body CreateLikeRequest true "Like/Pass payload"
// @Success 201 {object} LikeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/likes [post]
func (h *LikeHandler) CreateLike(c *gin.Context) {
	userID, ok := middleware.GetUserID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not authenticated"})
		return
	}

	var req CreateLikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	targetUserID, err := uuid.Parse(req.TargetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid target_user_id"})
		return
	}

	// Prevent users from liking themselves
	if userID == targetUserID {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "cannot like yourself"})
		return
	}

	// Create the like
	like := &models.Like{
		UserID:       userID,
		TargetUserID: targetUserID,
		Status:       req.Status,
	}

	if err := h.likeRepo.Create(like); err != nil {
		if err == repo.ErrLikeAlreadyExists {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "you have already liked or passed this user"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create like"})
		return
	}

	response := LikeResponse{
		Like:    like,
		Matched: false,
	}

	// If status is "like", check for mutual like and create match
	if req.Status == "like" {
		mutualLike, err := h.likeRepo.CheckMutualLike(userID, targetUserID)
		if err != nil {
			// Log error but don't fail the request
			c.JSON(http.StatusCreated, response)
			return
		}

		if mutualLike {
			// Check if match already exists
			matchExists, err := h.matchRepo.Exists(userID, targetUserID)
			if err != nil {
				c.JSON(http.StatusCreated, response)
				return
			}

			if !matchExists {
				// Create match
				match := &models.Match{
					User1ID: userID,
					User2ID: targetUserID,
				}

				if err := h.matchRepo.Create(match); err != nil {
					// Log error but don't fail the request
					c.JSON(http.StatusCreated, response)
					return
				}

				response.Match = match
				response.Matched = true
			}
		}
	}

	c.JSON(http.StatusCreated, response)
}
