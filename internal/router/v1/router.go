// Package v1 provides the API v1 router with middleware.
package v1

import (
	"go-template/internal/handler"
	"go-template/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// NewRouter builds the v1 chi router with all routes and middleware.
func NewRouter(healthHandler *handler.HealthHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(middleware.Logging)
	r.Use(middleware.Recovery)

	r.Handle("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Get("/health", healthHandler.Health)

	return r
}
