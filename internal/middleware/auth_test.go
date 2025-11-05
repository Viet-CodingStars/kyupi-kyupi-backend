package middleware

import (
  "net/http"
  "net/http/httptest"
  "testing"

  "github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/auth"
  "github.com/gin-gonic/gin"
  "github.com/google/uuid"
)

func TestAuthMiddleware(t *testing.T) {
  gin.SetMode(gin.TestMode)
  secret := "test-secret"
  userID := uuid.New()
  email := "test@example.com"

  token, err := auth.GenerateToken(userID, email, secret)
  if err != nil {
    t.Fatalf("failed to generate token: %v", err)
  }

  t.Run("valid token", func(t *testing.T) {
    handlerCalled := false
    router := gin.New()
    router.Use(AuthMiddleware(secret))
    router.GET("/test", func(c *gin.Context) {
      handlerCalled = true
      id, ok := GetUserID(c.Request.Context())
      if !ok {
        t.Fatal("expected user ID in context")
      }
      if id != userID {
        t.Fatalf("expected user ID %v, got %v", userID, id)
      }
      ginUserID, exists := c.Get(string(UserIDKey))
      if !exists {
        t.Fatal("expected user ID in gin context")
      }
      if idFromGin, ok := ginUserID.(uuid.UUID); !ok || idFromGin != userID {
        t.Fatalf("expected user ID %v in gin context, got %v", userID, ginUserID)
      }
      c.Status(http.StatusOK)
    })

    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    rr := httptest.NewRecorder()
    router.ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
      t.Fatalf("expected status 200, got %d", rr.Code)
    }
    if !handlerCalled {
      t.Fatal("expected handler to execute")
    }
  })

  t.Run("missing authorization header", func(t *testing.T) {
    handlerCalled := false
    router := gin.New()
    router.Use(AuthMiddleware(secret))
    router.GET("/test", func(c *gin.Context) {
      handlerCalled = true
      c.Status(http.StatusOK)
    })

    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    rr := httptest.NewRecorder()
    router.ServeHTTP(rr, req)

    if rr.Code != http.StatusUnauthorized {
      t.Fatalf("expected status 401, got %d", rr.Code)
    }
    if handlerCalled {
      t.Fatal("expected handler not to execute")
    }
  })

  t.Run("invalid authorization header format", func(t *testing.T) {
    handlerCalled := false
    router := gin.New()
    router.Use(AuthMiddleware(secret))
    router.GET("/test", func(c *gin.Context) {
      handlerCalled = true
      c.Status(http.StatusOK)
    })

    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    req.Header.Set("Authorization", "InvalidFormat")
    rr := httptest.NewRecorder()
    router.ServeHTTP(rr, req)

    if rr.Code != http.StatusUnauthorized {
      t.Fatalf("expected status 401, got %d", rr.Code)
    }
    if handlerCalled {
      t.Fatal("expected handler not to execute")
    }
  })

  t.Run("invalid token", func(t *testing.T) {
    handlerCalled := false
    router := gin.New()
    router.Use(AuthMiddleware(secret))
    router.GET("/test", func(c *gin.Context) {
      handlerCalled = true
      c.Status(http.StatusOK)
    })

    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    req.Header.Set("Authorization", "Bearer invalid-token")
    rr := httptest.NewRecorder()
    router.ServeHTTP(rr, req)

    if rr.Code != http.StatusUnauthorized {
      t.Fatalf("expected status 401, got %d", rr.Code)
    }
    if handlerCalled {
      t.Fatal("expected handler not to execute")
    }
  })
}
