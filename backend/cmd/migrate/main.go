package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/meta-boy/mech-alligator/internal/config"
)

func main() {
	var direction = flag.String("direction", "up", "Migration direction: up or down")
	var steps = flag.Int("steps", 0, "Number of migration steps (0 for all)")
	var forceVersion = flag.Int("force", -1, "Force migration version and clear dirty state (-1 to disable)")
	flag.Parse()

	cfg := config.LoadDatabaseConfig()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid database config: %v", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create migration driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create migration driver: %v", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/database/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	// Run migrations or force version
	if *forceVersion != -1 {
		err = m.Force(*forceVersion)
		if err != nil {
			log.Fatalf("Failed to force migration version: %v", err)
		}
		fmt.Printf("Successfully forced migration version to %d and cleared dirty state\n", *forceVersion)
	} else {
		switch *direction {
		case "up":
			if *steps == 0 {
				err = m.Up()
			} else {
				err = m.Steps(*steps)
			}
		case "down":
			if *steps == 0 {
				err = m.Down()
			} else {
				err = m.Steps(-*steps)
			}
		default:
			log.Fatalf("Invalid direction: %s. Use 'up' or 'down'. Or use -force flag.", *direction)
		}

		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration failed: %v", err)
		}

		if err == migrate.ErrNoChange {
			fmt.Println("No migrations to run")
		} else {
			fmt.Printf("Migrations completed successfully (%s)\n", *direction)
		}
	}

	// Print current version
	version, dirty, err := m.Version()
	if err != nil {
		log.Printf("Failed to get migration version: %v", err)
	} else {
		fmt.Printf("Current migration version: %d (dirty: %t)\n", version, dirty)
	}
}
