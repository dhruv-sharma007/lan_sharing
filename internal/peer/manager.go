package peer

import "sync"

type manager struct {
	mu    sync.RWMutex
	peers map[string]Peer
}

// NewManager creates a new thread-safe PeerManager.
func NewManager() PeerManager {
	return &manager{
		peers: make(map[string]Peer),
	}
}

func (m *manager) Add(peer Peer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.peers[peer.ID] = peer
}

func (m *manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.peers, id)
}

func (m *manager) Get(id string) (Peer, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.peers[id]
	return p, ok
}

func (m *manager) List() []Peer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]Peer, 0, len(m.peers))
	for _, p := range m.peers {
		list = append(list, p)
	}
	return list
}
