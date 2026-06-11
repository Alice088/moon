// Package main Go Template API
//
// @title          Go Template API
// @version        0.1.0
// @description    A Go API project template
// @termsOfService https://github.com
//
// @contact.name   Developer
// @contact.email  dev@example.com
//
// @license.name MIT
// @license.url  https://opensource.org/licenses/MIT
//
// @host      localhost:8080
// @BasePath  /v1
//
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description               Type "Bearer " followed by a JWT token
package main

import (
	"context"
	"fmt"
	"go-template/internal/config"
	"go-template/internal/database"
	"go-template/internal/handler"
	"go-template/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"

	v1 "go-template/internal/router/v1"
)

const version = "0.1.0"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Warn().Err(err).Msg("no .env file found, using defaults")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	logger.Setup(cfg.Log)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	db, err := database.NewPostgres(ctx, cfg.Database.DSN())
	cancel()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	healthHandler := handler.NewHealthHandler(version)

	router := v1.NewRouter(healthHandler)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Str("addr", addr).Msg("starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	<-quit
	log.Info().Msg("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server forced shutdown")
	}

	log.Info().Msg("server stopped gracefully")
}
