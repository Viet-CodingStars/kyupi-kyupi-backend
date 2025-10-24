package routes

import (
  "net/http"

  "github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/handlers"
)

// NewRouter creates the http.Handler for the application
func NewRouter() http.Handler {
  mux := http.NewServeMux()
  mux.HandleFunc("/", handlers.RootHandler)
  mux.HandleFunc("/health", handlers.HealthHandler)
  return mux
}
