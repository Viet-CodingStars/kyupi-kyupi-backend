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

type PassHandler struct {
	likeRepo LikeRepository
}

func NewPassHandler(db *sql.DB) *PassHandler {
	return &PassHandler{
		likeRepo: repo.NewLikeRepo(db),
	}
}

// CreatePassRequest represents the pass request.
type CreatePassRequest struct {
	TargetUserID string `json:"target_user_id" binding:"required"`
}

// PassResponse represents the response after creating a pass
type PassResponse struct {
	Pass *models.Like `json:"pass"`
}

// CreatePass handles creating a pass (POST /api/passes).
// @Summary Create a pass
// @Description Creates a pass for a target user (indicating no interest).
// @Tags Passes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body CreatePassRequest true "Pass payload"
// @Success 201 {object} PassResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/passes [post]
func (h *PassHandler) CreatePass(c *gin.Context) {
	userID, ok := middleware.GetUserID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not authenticated"})
		return
	}

	var req CreatePassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	targetUserID, err := uuid.Parse(req.TargetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid target_user_id"})
		return
	}

	// Prevent users from passing themselves
	if userID == targetUserID {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "cannot pass yourself"})
		return
	}

	// Create the pass
	pass := &models.Like{
		UserID:       userID,
		TargetUserID: targetUserID,
		Status:       "pass",
	}

	if err := h.likeRepo.Create(pass); err != nil {
		if err == repo.ErrLikeAlreadyExists {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "you have already liked or passed this user"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create pass"})
		return
	}

	response := PassResponse{
		Pass: pass,
	}

	c.JSON(http.StatusCreated, response)
}
