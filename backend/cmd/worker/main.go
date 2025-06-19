package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/meta-boy/mech-alligator/internal/config"
	"github.com/meta-boy/mech-alligator/internal/database"
	"github.com/meta-boy/mech-alligator/internal/queue"
	"github.com/meta-boy/mech-alligator/internal/queue/jobs"
	"github.com/meta-boy/mech-alligator/internal/repository/postgres"
	"github.com/meta-boy/mech-alligator/internal/scraper"
	"github.com/meta-boy/mech-alligator/internal/scraper/plugins/shopify"
)

func main() {
	log.Println("Starting worker...")

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

	// Create queue
	jobQueue := queue.NewDatabaseQueue(jobRepo)

	// Create scraper manager and register plugins
	scraperManager := scraper.NewManager()

	// Register Shopify plugin
	shopifyPlugin := shopify.NewShopifyPlugin()
	if err := scraperManager.RegisterPlugin(shopifyPlugin); err != nil {
		log.Fatalf("Failed to register Shopify plugin: %v", err)
	}

	log.Printf("Registered scraper plugins: Shopify")

	// Create job handlers
	scrapeHandler := jobs.NewScrapeJobHandler(db, scraperManager, jobQueue)
	tagHandler, err := jobs.NewTagJobHandler(db.DB)
	if err != nil {
		log.Fatalf("Failed to create tag job handler: %v", err)
	}

	// Create scheduler
	workers := getWorkerCount()
	scheduler := queue.NewJobScheduler(jobQueue, workers)

	// Register handlers
	scheduler.RegisterHandler(scrapeHandler)
	scheduler.RegisterHandler(tagHandler)

	log.Printf("Job scheduler configured with %d workers", workers)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduler
	if err := scheduler.Start(ctx); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}

	log.Println("Worker started successfully")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	log.Println("Received shutdown signal...")

	// Graceful shutdown
	cancel()
	if err := scheduler.Stop(); err != nil {
		log.Printf("Error stopping scheduler: %v", err)
	}

	log.Println("Worker stopped")
}

func getWorkerCount() int {
	workers := 3 // Default worker count

	if workerEnv := os.Getenv("WORKER_COUNT"); workerEnv != "" {
		// Parse worker count from environment if available
		// For simplicity, using default here
		log.Printf("Using default worker count: %d", workers)
	}

	return workers
}
