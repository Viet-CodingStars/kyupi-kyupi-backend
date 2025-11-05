package middleware

import (
  "context"
  "net/http"
  "strings"

  "github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/auth"
  "github.com/gin-gonic/gin"
  "github.com/google/uuid"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// AuthMiddleware validates JWT tokens and sets user context for Gin handlers.
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
  return func(c *gin.Context) {
    authHeader := c.GetHeader("Authorization")
    if authHeader == "" {
      c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
      return
    }

    parts := strings.Split(authHeader, " ")
    if len(parts) != 2 || parts[0] != "Bearer" {
      c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
      return
    }

    claims, err := auth.ValidateToken(parts[1], jwtSecret)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
      return
    }

    ctx := context.WithValue(c.Request.Context(), UserIDKey, claims.UserID)
    c.Set(string(UserIDKey), claims.UserID)
    c.Request = c.Request.WithContext(ctx)
    c.Next()
  }
}

// GetUserID retrieves the user ID from the request context
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
  userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
  return userID, ok
}
