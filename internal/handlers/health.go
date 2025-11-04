package handlers

import (
  "net/http"

  "github.com/gin-gonic/gin"
)

// StatusResponse represents a simple status payload used in health endpoints.
type StatusResponse struct {
  Status string `json:"status"`
}

// RootHandler responds with a small info JSON.
// @Summary Root message
// @Description Returns a basic message confirming the API is reachable.
// @Tags Health
// @Produce json
// @Success 200 {object} MessageResponse
// @Router / [get]
func RootHandler(c *gin.Context) {
  c.JSON(http.StatusOK, MessageResponse{Message: "kyupi-kyupi-backend"})
}

// HealthHandler returns a simple status payload used for health checks.
// @Summary Health check
// @Description Returns OK when the service is healthy.
// @Tags Health
// @Produce json
// @Success 200 {object} StatusResponse
// @Router /health [get]
func HealthHandler(c *gin.Context) {
  c.JSON(http.StatusOK, StatusResponse{Status: "ok"})
}
