package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/auth"
	"github.com/google/uuid"
)

func TestAuthMiddleware(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New()
	email := "test@example.com"

	// Generate a valid token
	token, err := auth.GenerateToken(userID, email, secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := GetUserID(r.Context())
		if !ok {
			t.Fatal("expected user ID in context")
		}
		if id != userID {
			t.Fatalf("expected user ID %v, got %v", userID, id)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with auth middleware
	authHandler := AuthMiddleware(secret)(handler)

	t.Run("valid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		authHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}
	})

	t.Run("missing authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		authHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rr.Code)
		}
	})

	t.Run("invalid authorization header format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		rr := httptest.NewRecorder()

		authHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rr.Code)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rr := httptest.NewRecorder()

		authHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rr.Code)
		}
	})
}
