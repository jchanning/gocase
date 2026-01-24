package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"my-app/internal/database"
	"my-app/internal/server"
)

func main() {
	// Load DATABASE_URL from environment variable
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Initialize database service
	db, err := database.NewService(dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Println("Database connection established successfully")

	// Create and configure the server
	srv := server.NewServer(db)

	// Set up graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		addr := ":8080"
		log.Printf("Starting server on %s", addr)
		if err := http.ListenAndServe(addr, srv.Router()); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop
	fmt.Println("\nShutting down gracefully...")
	log.Println("Server stopped")
}
