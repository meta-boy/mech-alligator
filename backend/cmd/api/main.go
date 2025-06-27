package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/meta-boy/mech-alligator/internal/api/handlers"
	"github.com/meta-boy/mech-alligator/internal/api/routes"
	"github.com/meta-boy/mech-alligator/internal/config"
	"github.com/meta-boy/mech-alligator/internal/database"
	"github.com/meta-boy/mech-alligator/internal/queue"
	"github.com/meta-boy/mech-alligator/internal/repository/postgres"
	"github.com/meta-boy/mech-alligator/internal/service"
)

func main() {
	log.Println("Starting API server...")

	// Load database configuration
	cfg := config.LoadDatabaseConfig()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid database config: %v", err)
	}

	// Connect to database
	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Health(); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}

	log.Println("Database connection established")

	// Create repositories
	jobRepo := postgres.NewJobRepository(db)
	productRepo := postgres.NewProductRepository(db)

	// Create queue (same as worker)
	jobQueue := queue.NewDatabaseQueue(jobRepo)

	// Create services
	jobService := service.NewJobService(db, jobQueue, nil) // scheduler not needed for API
	productService := service.NewProductService(productRepo)

	// Create handlers
	jobHandler := handlers.NewJobHandler(jobService)
	productHandler := handlers.NewProductHandler(productService)

	// Setup routes
	mux := http.NewServeMux()

	// Add basic health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	// Setup job routes
	routes.SetupJobRoutes(mux, jobHandler)

	// Setup product routes
	routes.SetupProductRoutes(mux, productHandler)

	// Add basic logging middleware
	loggedMux := loggingMiddleware(mux)

	// Setup server
	port := getPort()
	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", port),
		Handler: loggedMux,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("API server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Println("API server started successfully")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	log.Println("Received shutdown signal...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Println("API server stopped")
}

func getPort() string {
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call the next handler
		next.ServeHTTP(w, r)

		// Log the request
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}
