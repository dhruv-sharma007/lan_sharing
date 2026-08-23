package discovery

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"os/user"
	"time"

	"lan_sharing/internal/peer"

	"github.com/grandcat/zeroconf"
)

// MDNSService manages mDNS discovery and broadcasting.
type MDNSService struct {
	manager peer.PeerManager
	server  *zeroconf.Server
}

// NewMDNSService creates a new mDNS service.
func NewMDNSService(manager peer.PeerManager) *MDNSService {
	return &MDNSService{
		manager: manager,
	}
}

// Start begins broadcasting this device and discovering others.
func (s *MDNSService) Start(ctx context.Context, port int) error {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "Unknown"
	}

	myUser, err := user.Current()
	var username string
	if err == nil {
		username = myUser.Username
	} else {
		username = "user_" + fmt.Sprint(rand.IntN(1000-100+1))
	}

	instanceName := fmt.Sprintf("%s-%s", hostname, username)

	// Broadcast
	server, err := zeroconf.Register(instanceName, "_shareapp._tcp", "local.", port, []string{"txtv=1"}, nil)
	if err != nil {
		return fmt.Errorf("failed to register mDNS service: %w", err)
	}
	s.server = server

	// Discover
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		server.Shutdown()
		return fmt.Errorf("failed to create mDNS resolver: %w", err)
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			if entry.Instance == instanceName {
				continue // Skip self
			}

			// Add to peer manager
			if len(entry.AddrIPv4) > 0 {
				ip := entry.AddrIPv4[0].String()
				p := peer.Peer{
					ID:       entry.Instance,
					Hostname: entry.Instance,
					IP:       ip,
					Port:     entry.Port,
					LastSeen: time.Now(),
				}
				s.manager.Add(p)
				log.Printf("Discovered peer: %s at %s:%d\n", p.Hostname, p.IP, p.Port)
			}
		}
	}(entries)

	err = resolver.Browse(ctx, "_shareapp._tcp", "local.", entries)
	if err != nil {
		server.Shutdown()
		return fmt.Errorf("failed to browse mDNS: %w", err)
	}

	return nil
}

// Stop shuts down the mDNS service.
func (s *MDNSService) Stop() {
	if s.server != nil {
		s.server.Shutdown()
	}
}
