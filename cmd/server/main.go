package main

import (
	"fmt"
	"log"
	"net/http"

	"saml-poc/internal/config"
	"saml-poc/internal/database"
	"saml-poc/internal/handlers"
	"saml-poc/internal/middleware"
	"saml-poc/internal/saml"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := database.New(cfg.DatabaseConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := database.NewUserRepository(db)

	// Initialize SAML provider
	samlProvider, err := saml.NewProvider(cfg)
	if err != nil {
		log.Fatalf("Failed to create SAML provider: %v", err)
	}

	// Initialize JIT service
	jitService := saml.NewJITService(userRepo, &cfg.JIT)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jitService)

	// Initialize handlers
	homeHandler := handlers.NewHomeHandler()
	debugHandler := handlers.NewDebugHandler(cfg)

	// Setup routes
	setupRoutes(samlProvider, authMiddleware, homeHandler, debugHandler)

	// Print startup information
	printStartupInfo(cfg)

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Starting server on %s", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}

// setupRoutes configures all HTTP routes
func setupRoutes(
	samlProvider *saml.Provider,
	authMiddleware *middleware.AuthMiddleware,
	homeHandler *handlers.HomeHandler,
	debugHandler *handlers.DebugHandler,
) {
	// SAML endpoints - register with prefix pattern
	http.Handle("/saml/", samlProvider.GetMiddleware())

	// Debug endpoint (unprotected)
	http.Handle("/debug", debugHandler)

	// Protected home endpoint with database validation middleware
	http.Handle("/home", samlProvider.GetMiddleware().RequireAccount(
		authMiddleware.DatabaseValidation(homeHandler),
	))

	// Root redirect to protected home - this will trigger SAML auth if not authenticated
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/home", http.StatusFound)
	})
}

// printStartupInfo prints server startup information
func printStartupInfo(cfg *config.Config) {
	fmt.Printf("Server started at http://%s\n", cfg.ServerAddress())
	fmt.Println("Database connection established")
	fmt.Printf("JIT (Just-In-Time) user creation: %s\n",
		map[bool]string{true: "ENABLED", false: "DISABLED"}[cfg.JIT.Enabled])

	if cfg.JIT.Enabled {
		fmt.Printf("  - Default user status: %s\n",
			map[bool]string{true: "Active", false: "Inactive"}[cfg.JIT.DefaultUserActive])
		fmt.Printf("  - Required attributes: %s\n",
			map[bool]string{true: "Enforced", false: "Optional"}[cfg.JIT.RequiredAttributesMode])
	}

	fmt.Println("SAML endpoints:")
	fmt.Printf("  - SSO: http://%s/saml/sso\n", cfg.ServerAddress())
	fmt.Printf("  - ACS: http://%s/saml/acs\n", cfg.ServerAddress())
	fmt.Printf("  - Metadata: http://%s/saml/metadata\n", cfg.ServerAddress())
}
