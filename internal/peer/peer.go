package peer

import "time"

// Peer represents a discovered device on the network.
type Peer struct {
	ID       string
	NodeID   uint64
	Hostname string
	IP       string
	Port     int
	LastSeen time.Time
}

// PeerManager defines the interface for managing discovered peers.
type PeerManager interface {
	Add(peer Peer)
	Remove(id string)
	Get(id string) (Peer, bool)
	List() []Peer
}
