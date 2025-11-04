package handlers

import (
  "encoding/json"
  "net/http"
  "net/http/httptest"
  "testing"

  "github.com/gin-gonic/gin"
)

func TestHealthHandler(t *testing.T) {
  gin.SetMode(gin.TestMode)
  responseRecorder := httptest.NewRecorder()
  ctx, _ := gin.CreateTestContext(responseRecorder)
  ctx.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

  HealthHandler(ctx)

  if responseRecorder.Code != http.StatusOK {
    t.Fatalf("expected status 200, got %d", responseRecorder.Code)
  }

  var body StatusResponse
  if err := json.NewDecoder(responseRecorder.Body).Decode(&body); err != nil {
    t.Fatalf("invalid json response: %v", err)
  }
  if body.Status != "ok" {
    t.Fatalf("expected status 'ok', got %q", body.Status)
  }
}
