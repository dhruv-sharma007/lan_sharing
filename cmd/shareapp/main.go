package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"lan_sharing/internal/app"
)

func main() {
	// Setup context that cancels on interrupt signals (Ctrl+C)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received termination signal, shutting down...")
		cancel()
	}()

	application := app.New()
	
	// Default port for the application (could be made configurable)
	port := 3498
	
	if err := application.Run(ctx, port); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
