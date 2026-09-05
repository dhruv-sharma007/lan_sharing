package app

import (
	"context"
	"fmt"
	"log"

	"lan_sharing/internal/connection"
	"lan_sharing/internal/discovery"
	"lan_sharing/internal/peer"
	"lan_sharing/util"
)

// App is the main application coordinator.
type App struct {
	config            *util.Config
	peerManager       peer.PeerManager
	connectionManager *connection.ConnectionManager
	discovery         *discovery.MDNSService
}

// New creates a new application instance.
func New() *App {
	cfg, err := util.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	pm := peer.NewManager()
	cm := connection.NewConnectionManager(cfg.NodeID, pm)

	mdns := discovery.NewMDNSService(cfg.NodeID, func(p peer.Peer) {
		// Connection Logic Rule: Lower ID initiates connection
		if cfg.NodeID < p.NodeID {
			go cm.ConnectToPeer(p)
		} else {
			log.Printf("Peer %s has lower ID (%d < %d), waiting for them to connect.", p.Hostname, p.NodeID, cfg.NodeID)
		}
	})
	
	return &App{
		config:            cfg,
		peerManager:       pm,
		connectionManager: cm,
		discovery:         mdns,
	}
}

// Run starts the application and blocks until context is cancelled.
func (a *App) Run(ctx context.Context, port int) error {
	log.Printf("Starting ShareApp on port %d with NodeID %d...", port, a.config.NodeID)

	go a.connectionManager.StartServer(port)

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
