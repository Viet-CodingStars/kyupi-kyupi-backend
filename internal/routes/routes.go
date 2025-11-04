package routes

import (
	"database/sql"
	"net/http"

	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/config"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/handlers"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/middleware"
)

// NewRouter creates the http.Handler for the application
func NewRouter(db *sql.DB, cfg *config.Config) http.Handler {
	mux := http.NewServeMux()
	
	// Health endpoints
	mux.HandleFunc("/", handlers.RootHandler)
	mux.HandleFunc("/health", handlers.HealthHandler)

	// User authentication handlers
	userHandler := handlers.NewUserHandler(db, cfg.JWTSecret)
	
	// Public endpoints
	mux.HandleFunc("POST /api/users", userHandler.SignUp)
	mux.HandleFunc("POST /api/users/sign_in", userHandler.SignIn)
	
	// Protected endpoints (require authentication)
	authMw := middleware.AuthMiddleware(cfg.JWTSecret)
	
	mux.Handle("/api/users/sign_out", authMw(http.HandlerFunc(userHandler.SignOut)))
	mux.Handle("/api/users/profile", authMw(http.HandlerFunc(userHandler.GetProfile)))
	
	// Handle PATCH/PUT for /api/users/profile
	mux.Handle("PATCH /api/users/profile", authMw(http.HandlerFunc(userHandler.UpdateProfile)))
	mux.Handle("PUT /api/users/profile", authMw(http.HandlerFunc(userHandler.UpdateProfile)))

	return mux
}
