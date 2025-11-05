package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	docs "github.com/Viet-CodingStars/kyupi-kyupi-backend/docs"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/config"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/db"
	"github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/routes"
)

func main() {
	cfg := config.LoadFromEnv()

	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Title = "Kyupi Kyupi Backend API"
	docs.SwaggerInfo.Description = "API documentation for the Kyupi Kyupi backend service."
	docs.SwaggerInfo.Version = "1.0"

	pg, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pg.Close()

	mongoClient, err := db.ConnectMongo(cfg)
	if err != nil {
		log.Fatalf("failed to connect to mongo: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("mongo disconnect error: %v", err)
		}
	}()

	addr := cfg.Addr()
	srv := &http.Server{
		Addr:    addr,
		Handler: routes.NewRouter(pg, cfg),
	}

	log.Printf("starting server (env=%s) on %s", cfg.Env, addr)

	// start server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// graceful shutdown (handle SIGINT/SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server exited")
}
