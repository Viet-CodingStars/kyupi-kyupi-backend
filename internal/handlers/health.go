package handlers

import (
  "encoding/json"
  "net/http"
)

// RootHandler responds with a small info JSON
func RootHandler(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json; charset=utf-8")
  _ = json.NewEncoder(w).Encode(map[string]string{"message": "kyupi-kyupi-backend"})
}

// HealthHandler returns a simple status payload used for health checks.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json; charset=utf-8")
  w.WriteHeader(http.StatusOK)
  _ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
