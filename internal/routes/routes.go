package routes

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/config"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/handlers"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/middleware"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewRouter creates the Gin engine for the application.
func NewRouter(db *sql.DB, cfg *config.Config, avatarStorage storage.AvatarStorage) http.Handler {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5174"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	if cfg.AvatarStorageDir != "" && cfg.AvatarURLPrefix != "" {
		router.Static(cfg.AvatarURLPrefix, cfg.AvatarStorageDir)
	}

	router.GET("/", handlers.RootHandler)
	router.GET("/health", handlers.HealthHandler)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	userHandler := handlers.NewUserHandler(db, cfg.JWTSecret, avatarStorage)
	authMw := middleware.AuthMiddleware(cfg.JWTSecret)

	api := router.Group("/api")
	{
		api.POST("/users", userHandler.SignUp)
		api.POST("/users/sign_in", userHandler.SignIn)

		users := api.Group("/users")
		users.Use(authMw)
		{
			users.DELETE("/sign_out", userHandler.SignOut)
			users.GET("/profile", userHandler.GetProfile)
			users.PATCH("/profile", userHandler.UpdateProfile)
			users.PUT("/profile", userHandler.UpdateProfile)
			users.POST("/profile/avatar", userHandler.UploadAvatar)
		}
	}

	return router
}
