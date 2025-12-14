package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"alieze-erp/internal/database"
	authhandler "alieze-erp/internal/modules/auth/handler"
	authmiddleware "alieze-erp/internal/modules/auth/middleware"
	authrepository "alieze-erp/internal/modules/auth/repository"
	authservice "alieze-erp/internal/modules/auth/service"
	crmhandler "alieze-erp/internal/modules/crm/handler"
	crmrepository "alieze-erp/internal/modules/crm/repository"
	crmservice "alieze-erp/internal/modules/crm/service"
	productshandler "alieze-erp/internal/modules/products/handler"
	productsrepository "alieze-erp/internal/modules/products/repository"
	productsservice "alieze-erp/internal/modules/products/service"
)

type Server struct {
	port int

	db             database.Service
	authHandler    *authhandler.AuthHandler
	authMiddleware *authmiddleware.AuthMiddleware
	contactHandler *crmhandler.ContactHandler
	productHandler *productshandler.ProductHandler
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	// Initialize database
	dbService := database.New()

	// Run database migrations (disabled for now to avoid database connection issues)
	// if err := dbService.RunMigrations(); err != nil {
	// 	log.Printf("Failed to run database migrations: %v", err)
	// }

	// Initialize auth repository
	authRepo := authrepository.NewAuthRepository(dbService.GetDB())

	// Initialize auth service
	authService := authservice.NewAuthService(authRepo)

	// Initialize auth handler
	authHandler := authhandler.NewAuthHandler(authService)

	// Initialize auth middleware
	authMiddleware := authmiddleware.NewAuthMiddleware()

	// Initialize CRM repositories
	contactRepo := crmrepository.NewContactRepository(dbService.GetDB())

	// Initialize CRM services
	simpleAuthService := NewSimpleAuthServiceAdapter(authService)
	contactService := crmservice.NewContactService(contactRepo, simpleAuthService)

	// Initialize CRM handlers
	contactHandler := crmhandler.NewContactHandler(contactService)

	// Initialize Products repository
	productRepo := productsrepository.NewProductRepository(dbService.GetDB())

	// Initialize Products service
	productService := productsservice.NewProductService(productRepo, simpleAuthService)

	// Initialize Products handler
	productHandler := productshandler.NewProductHandler(productService)

	NewServer := &Server{
		port:           port,
		db:             dbService,
		authHandler:    authHandler,
		authMiddleware: authMiddleware,
		contactHandler: contactHandler,
		productHandler: productHandler,
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
