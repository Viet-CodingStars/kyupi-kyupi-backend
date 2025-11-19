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
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/repo"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestCreatePass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         uuid.UUID
		requestBody    CreatePassRequest
		mockCreateErr  error
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful pass creation",
			userID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			requestBody: CreatePassRequest{
				TargetUserID: "22222222-2222-2222-2222-222222222222",
			},
			mockCreateErr:  nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "cannot pass yourself",
			userID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			requestBody: CreatePassRequest{
				TargetUserID: "11111111-1111-1111-1111-111111111111",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "cannot pass yourself",
		},
		{
			name:   "invalid target_user_id",
			userID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			requestBody: CreatePassRequest{
				TargetUserID: "invalid-uuid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid target_user_id",
		},
		{
			name:   "duplicate pass returns conflict",
			userID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			requestBody: CreatePassRequest{
				TargetUserID: "22222222-2222-2222-2222-222222222222",
			},
			mockCreateErr:  repo.ErrLikeAlreadyExists,
			expectedStatus: http.StatusConflict,
			expectedError:  "you have already liked or passed this user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock repository using the existing mockLikeRepo from like_test.go
			mockRepo := newMockLikeRepo()
			
			// If we need to simulate an error, pre-populate the mock
			if tt.mockCreateErr == repo.ErrLikeAlreadyExists {
				mockRepo.Create(&models.Like{
					UserID:       tt.userID,
					TargetUserID: uuid.MustParse(tt.requestBody.TargetUserID),
					Status:       "pass",
				})
			}

			handler := &PassHandler{
				likeRepo: mockRepo,
			}

			// Prepare request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/passes", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Add user ID to context
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			// Record response
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Call handler
			handler.CreatePass(c)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check error message if expected
			if tt.expectedError != "" {
				var errResp ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if errResp.Error != tt.expectedError {
					t.Errorf("expected error '%s', got '%s'", tt.expectedError, errResp.Error)
				}
			}

			// Check successful response
			if tt.expectedStatus == http.StatusCreated {
				var resp PassResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp.Pass == nil {
					t.Error("expected pass in response")
				}
				if resp.Pass.Status != "pass" {
					t.Errorf("expected status 'pass', got '%s'", resp.Pass.Status)
				}
			}
		})
	}
}

func TestCreatePassUnauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &PassHandler{
		likeRepo: newMockLikeRepo(),
	}

	// Prepare request without user ID in context
	body, _ := json.Marshal(CreatePassRequest{
		TargetUserID: "22222222-2222-2222-2222-222222222222",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/passes", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Call handler
	handler.CreatePass(c)

	// Check status code
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
