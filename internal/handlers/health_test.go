package handlers

import (
  "encoding/json"
  "net/http"
  "net/http/httptest"
  "testing"
)

func TestHealthHandler(t *testing.T) {
  req := httptest.NewRequest(http.MethodGet, "/health", nil)
  rr := httptest.NewRecorder()
  HealthHandler(rr, req)

  if rr.Code != http.StatusOK {
    t.Fatalf("expected status 200, got %d", rr.Code)
  }

  var body map[string]string
  if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
    t.Fatalf("invalid json response: %v", err)
  }
  if body["status"] != "ok" {
    t.Fatalf("expected status 'ok', got %q", body["status"])
  }
}
