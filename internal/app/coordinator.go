package app

import (
	"context"
	"fmt"
	"log"

	"lan_sharing/internal/discovery"
	"lan_sharing/internal/peer"
)

// App is the main application coordinator.
type App struct {
	peerManager peer.PeerManager
	discovery   *discovery.MDNSService
}

// New creates a new application instance.
func New() *App {
	pm := peer.NewManager()
	mdns := discovery.NewMDNSService(pm)
	
	return &App{
		peerManager: pm,	
		discovery:   mdns,
	}
}

// Run starts the application and blocks until context is cancelled.
func (a *App) Run(ctx context.Context, port int) error {
	log.Printf("Starting ShareApp on port %d...", port)

	if err := a.discovery.Start(ctx, port); err != nil {
		return fmt.Errorf("failed to start discovery: %w", err)
	}
	defer a.discovery.Stop()

	log.Println("Discovery started. Waiting for peers... (Press Ctrl+C to stop)")

	// Block until context is done
	<-ctx.Done()
	
	log.Println("Shutting down ShareApp...")
	return nil
}

// GetPeers returns the list of currently known peers.
func (a *App) GetPeers() []peer.Peer {
	return a.peerManager.List()
}
