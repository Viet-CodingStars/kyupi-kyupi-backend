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

type MatchHandler struct {
	matchRepo MatchRepository
	userRepo  UserRepository
}

func NewMatchHandler(db *sql.DB) *MatchHandler {
	return &MatchHandler{
		matchRepo: repo.NewMatchRepo(db),
		userRepo:  repo.NewUserRepo(db),
	}
}

// MatchWithUser represents a match with the matched user's details
type MatchWithUser struct {
	ID          uuid.UUID    `json:"id"`
	MatchedUser *models.User `json:"matched_user"`
	CreatedAt   string       `json:"created_at"`
}

// GetMatches retrieves all matches for the current user (GET /api/v1/matches).
// @Summary Get all matches with user details
// @Description Returns all matches for the authenticated user with matched user information.
// @Tags Matches
// @Produce json
// @Security BearerAuth
// @Success 200 {array} MatchWithUser
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

	// Build response with matched user details
	matchesWithUsers := []MatchWithUser{}
	for _, match := range matches {
		// Determine which user is the matched user (not the current user)
		var matchedUserID uuid.UUID
		if match.User1ID == userID {
			matchedUserID = match.User2ID
		} else {
			matchedUserID = match.User1ID
		}

		// Fetch the matched user's details
		matchedUser, err := h.userRepo.GetByID(matchedUserID)
		if err != nil {
			// Skip this match if we can't get user details
			continue
		}

		matchesWithUsers = append(matchesWithUsers, MatchWithUser{
			ID:          match.ID,
			MatchedUser: matchedUser,
			CreatedAt:   match.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	c.JSON(http.StatusOK, matchesWithUsers)
}
