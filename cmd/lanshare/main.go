package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"lan_sharing/internal/app"
)

var Version = "dev"

func main() {
	versionFlag := flag.Bool("version", false, "Print the version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("LanShare %s\n", Version)
		os.Exit(0)
	}

	if err := app.InitLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Setup context that cancels on interrupt signals (Ctrl+C)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start auto-updater background routine
	app.StartUpdater(ctx, Version)

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
