package main

import (
  "context"
  "log"
  "net/http"
  "os"
  "os/signal"
  "syscall"
  "time"

  "github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/config"
  "github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/routes"
)

func main() {
  cfg := config.LoadFromEnv()

  addr := cfg.Addr()
  srv := &http.Server{
    Addr:    addr,
    Handler: routes.NewRouter(),
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
