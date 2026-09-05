package discovery

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

	"lan_sharing/internal/peer"

	"github.com/grandcat/zeroconf"
)

// MDNSService manages mDNS discovery and broadcasting.
type MDNSService struct {
	nodeID      uint64
	onPeerFound func(peer.Peer)
	server      *zeroconf.Server
}

// NewMDNSService creates a new mDNS service.
func NewMDNSService(nodeID uint64, onPeerFound func(peer.Peer)) *MDNSService {
	return &MDNSService{
		nodeID:      nodeID,
		onPeerFound: onPeerFound,
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

	// Broadcast with NodeID
	txtRecords := []string{"txtv=1", fmt.Sprintf("id=%d", s.nodeID)}
	server, err := zeroconf.Register(instanceName, "_shareapp._tcp", "local.", port, txtRecords, nil)
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

			if len(entry.AddrIPv4) > 0 {
				ip := entry.AddrIPv4[0].String()
				
				// Parse NodeID from TXT records
				var remoteNodeID uint64
				for _, txt := range entry.Text {
					if strings.HasPrefix(txt, "id=") {
						idStr := strings.TrimPrefix(txt, "id=")
						if id, err := strconv.ParseUint(idStr, 10, 64); err == nil {
							remoteNodeID = id
						}
					}
				}

				p := peer.Peer{
					ID:       entry.Instance,
					NodeID:   remoteNodeID,
					Hostname: entry.Instance,
					IP:       ip,
					Port:     entry.Port,
					LastSeen: time.Now(),
				}
				
				log.Printf("Discovered peer: %s (NodeID: %d) at %s:%d\n", p.Hostname, p.NodeID, p.IP, p.Port)
				
				if s.onPeerFound != nil {
					s.onPeerFound(p)
				}
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
