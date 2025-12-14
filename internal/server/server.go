package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"alieze-erp/internal/database"
	"alieze-erp/internal/modules/auth/handler"
	"alieze-erp/internal/modules/auth/middleware"
	"alieze-erp/internal/modules/auth/repository"
	"alieze-erp/internal/modules/auth/service"
)

type Server struct {
	port int

	db             database.Service
	authHandler    *handler.AuthHandler
	authMiddleware *middleware.AuthMiddleware
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	// Initialize database
	dbService := database.New()

	// Run database migrations
	if err := dbService.RunMigrations(); err != nil {
		log.Printf("Failed to run database migrations: %v", err)
		// Continue with server startup even if migrations fail
	}

	// Initialize auth repository
	authRepo := repository.NewAuthRepository(dbService.GetDB())

	// Initialize auth service
	authService := service.NewAuthService(authRepo)

	// Initialize auth handler
	authHandler := handler.NewAuthHandler(authService)

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware()

	NewServer := &Server{
		port:           port,
		db:             dbService,
		authHandler:    authHandler,
		authMiddleware: authMiddleware,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
