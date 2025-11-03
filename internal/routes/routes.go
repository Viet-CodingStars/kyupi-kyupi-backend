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
	
	// Handle both GET and PATCH/PUT for /api/users
	mux.Handle("/api/users", authMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			userHandler.GetProfile(w, r)
		} else if r.Method == http.MethodPatch || r.Method == http.MethodPut {
			userHandler.UpdateProfile(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	return mux
}
