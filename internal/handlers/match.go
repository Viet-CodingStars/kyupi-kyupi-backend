package handlers

import (
	"database/sql"
	"net/http"

	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/middleware"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/models"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/repo"
	"github.com/gin-gonic/gin"
)

type MatchHandler struct {
	matchRepo MatchRepository
}

func NewMatchHandler(db *sql.DB) *MatchHandler {
	return &MatchHandler{
		matchRepo: repo.NewMatchRepo(db),
	}
}

// GetMatches retrieves all matches for the current user (GET /api/v1/matches).
// @Summary Get all matches
// @Description Returns all matches for the authenticated user.
// @Tags Matches
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Match
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/matches [get]
func (h *MatchHandler) GetMatches(c *gin.Context) {
	userID, ok := middleware.GetUserID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not authenticated"})
		return
	}

	matches, err := h.matchRepo.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve matches"})
		return
	}

	// Return empty array instead of null if no matches
	if matches == nil {
		matches = []*models.Match{}
	}

	c.JSON(http.StatusOK, matches)
}
