  package main

  import (
  "context"
  "log"
  "net/http"
  "os"
  "os/signal"
  "time"

  "github.com/Viet-CodingStars/kyupi-kyupi-backend/internal/routes"
  )

  func main() {
  addr := ":8080"
  srv := &http.Server{
    Addr:    addr,
    Handler: routes.NewRouter(),
  }

  // start server
  go func() {
    log.Printf("starting server on %s", addr)
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
      log.Fatalf("server error: %v", err)
    }
  }()

  // graceful shutdown
  quit := make(chan os.Signal, 1)
  signal.Notify(quit, os.Interrupt)
  <-quit
  log.Println("shutting down server...")

  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()
  if err := srv.Shutdown(ctx); err != nil {
    log.Fatalf("server forced to shutdown: %v", err)
  }
  log.Println("server exited")
  }
