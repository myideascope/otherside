package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/myideascope/otherside/internal/config"
	"github.com/myideascope/otherside/internal/repository"
	"github.com/myideascope/otherside/internal/service"
)

func main() {
	// Parse command line flags
	var (
		migrate = flag.Bool("migrate", false, "Run database migrations and exit")
		cleanup = flag.Bool("cleanup", false, "Run cleanup operations and exit")
		status  = flag.Bool("status", false, "Show migration status and exit")
	)
	flag.Parse()

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := repository.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Handle specific commands
	if *migrate {
		if err := db.Migrator.Up(context.Background()); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("Migrations completed successfully")
		return
	}

	if *status {
		if err := db.MigrationStatus(context.Background()); err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}
		return
	}

	if *cleanup {
		runCleanup(db, cfg)
		return
	}

	// Initialize application components
	app, err := initializeApp(db, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start application
	go func() {
		log.Println("Starting OtherSide application...")
		// TODO: Start HTTP server, WebSocket handlers, etc.
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received, gracefully shutting down...")

	// Shutdown application
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Println("Application shutdown complete")
}

// Application holds all application components
type Application struct {
	sessionManager *service.SessionStateManager
	fileManager    *repository.FileManager
	cleanupManager *repository.CleanupManager
	db             *repository.DB
}

// initializeApp sets up all application components
func initializeApp(db *repository.DB, cfg *config.Config) (*Application, error) {
	app := &Application{
		db: db,
	}

	// Initialize file manager
	app.fileManager = repository.NewFileManager(db.DB, cfg.Storage.DataPath)

	// Initialize session manager
	app.sessionManager = service.NewSessionStateManager(db.DB)
	if err := app.sessionManager.Initialize(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize session manager: %w", err)
	}

	// Initialize cleanup manager
	app.cleanupManager = repository.NewCleanupManager(db.DB, app.fileManager)

	return app, nil
}

// Shutdown gracefully shuts down the application
func (app *Application) Shutdown(ctx context.Context) error {
	log.Println("Shutting down application components...")

	// Save session states
	if err := app.sessionManager.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down session manager: %v", err)
	}

	// Close database
	if err := app.db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	return nil
}

// runCleanup performs cleanup operations
func runCleanup(db *repository.DB, cfg *config.Config) {
	ctx := context.Background()
	fileManager := repository.NewFileManager(db.DB, cfg.Storage.DataPath)
	cleanupManager := repository.NewCleanupManager(db.DB, fileManager)

	// Default cleanup configuration
	cleanupConfig := repository.CleanupConfig{
		MaxSessionAge:         time.Duration(cfg.Storage.RetentionDays) * 24 * time.Hour,
		MaxFileSizeBytes:      int64(cfg.Storage.MaxSizeGB) * 1024 * 1024 * 1024 / 10, // 10% of max storage
		MaxStorageSizeBytes:   int64(cfg.Storage.MaxSizeGB) * 1024 * 1024 * 1024,
		EnableFileCleanup:     true,
		EnableDatabaseCleanup: true,
		EnableOrphanCleanup:   true,
	}

	// Show cleanup plan first
	fmt.Println("Cleanup Plan:")
	plan, err := cleanupManager.GetCleanupPlan(ctx, cleanupConfig)
	if err != nil {
		log.Fatalf("Failed to generate cleanup plan: %v", err)
	}

	fmt.Printf("Sessions to clean: %d\n", plan.SessionsCleaned)
	fmt.Printf("Files to clean: %d\n", plan.FilesCleaned)
	fmt.Printf("Bytes to free: %d\n", plan.BytesFreed)

	if len(plan.Errors) > 0 {
		fmt.Println("Plan errors:")
		for _, err := range plan.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	// Confirm cleanup
	fmt.Print("Proceed with cleanup? (y/N): ")
	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		fmt.Println("Cleanup cancelled")
		return
	}

	// Run cleanup
	fmt.Println("Running cleanup...")
	stats, err := cleanupManager.RunCleanup(ctx, cleanupConfig)
	if err != nil {
		log.Fatalf("Cleanup failed: %v", err)
	}

	fmt.Printf("Cleanup completed in %v\n", stats.Duration)
	fmt.Printf("Sessions cleaned: %d\n", stats.SessionsCleaned)
	fmt.Printf("Files cleaned: %d\n", stats.FilesCleaned)
	fmt.Printf("Bytes freed: %d\n", stats.BytesFreed)

	if len(stats.Errors) > 0 {
		fmt.Println("Errors during cleanup:")
		for _, err := range stats.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}
}
