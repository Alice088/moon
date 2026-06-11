package handler

import (
	"encoding/json"
	"go-template/internal/dto"
	"net/http"
	"time"
)

// HealthHandler serves the /health endpoint.
type HealthHandler struct {
	version string
}

// NewHealthHandler creates a HealthHandler with the given app version.
func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{version: version}
}

// Health godoc
// @Summary      Health check
// @Description  Returns the service health status
// @Tags         system
// @Produce      json
// @Success      200  {object}  dto.HealthResponse
// @Router       /health [get]
func (h *HealthHandler) Health(w http.ResponseWriter, _ *http.Request) {
	resp := dto.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   h.version,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
