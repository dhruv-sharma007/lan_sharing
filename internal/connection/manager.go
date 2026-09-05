package connection

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"lan_sharing/internal/peer"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all for LAN
	},
}

// ConnectionManager handles WebSocket connections between peers.
type ConnectionManager struct {
	myNodeID    uint64
	peerManager peer.PeerManager
	
	mu    sync.Mutex
	conns map[string]*websocket.Conn // key: Peer ID
}

// NewConnectionManager creates a new connection manager.
func NewConnectionManager(nodeID uint64, pm peer.PeerManager) *ConnectionManager {
	return &ConnectionManager{
		myNodeID:    nodeID,
		peerManager: pm,
		conns:       make(map[string]*websocket.Conn),
	}
}

// StartServer starts the HTTP server for accepting incoming WebSocket connections.
func (cm *ConnectionManager) StartServer(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", cm.handleIncomingWS)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("WebSocket server listening on %s", addr)
	
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Printf("WebSocket server error: %v", err)
	}
}

func (cm *ConnectionManager) handleIncomingWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade failed: %v", err)
		return
	}

	// Wait for peer to send its identity
	var p peer.Peer
	if err := conn.ReadJSON(&p); err != nil {
		log.Printf("Failed to read peer identity: %v", err)
		conn.Close()
		return
	}

	log.Printf("Accepted incoming connection from %s", p.Hostname)
	cm.registerConnection(p, conn)
}

// ConnectToPeer initiates a WebSocket connection to a discovered peer.
func (cm *ConnectionManager) ConnectToPeer(p peer.Peer) {
	url := fmt.Sprintf("ws://%s:%d/ws", p.IP, p.Port)
	log.Printf("Attempting to connect to peer %s at %s...", p.Hostname, url)

	// Add a short timeout
	dialer := websocket.Dialer{HandshakeTimeout: 5 * time.Second}
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		log.Printf("Failed to dial peer %s: %v", p.Hostname, err)
		return
	}

	// Send our identity to the peer
	// For this, we just need a dummy Peer struct representing ourselves, 
	// but with our ID and Hostname. The other side receives it to know who connected.
	myIdentity := peer.Peer{
		ID:       fmt.Sprintf("node-%d", cm.myNodeID), // Or our actual instance ID if we had it
		NodeID:   cm.myNodeID,
		Hostname: "me", // We could pass actual hostname here
	}
	
	if err := conn.WriteJSON(myIdentity); err != nil {
		log.Printf("Failed to send identity to peer %s: %v", p.Hostname, err)
		conn.Close()
		return
	}

	log.Printf("Successfully connected to %s", p.Hostname)
	cm.registerConnection(p, conn)
}

func (cm *ConnectionManager) registerConnection(p peer.Peer, conn *websocket.Conn) {
	cm.mu.Lock()
	if existing, exists := cm.conns[p.ID]; exists {
		log.Printf("Closing duplicate connection for %s", p.ID)
		existing.Close()
	}
	cm.conns[p.ID] = conn
	cm.mu.Unlock()

	// Update PeerManager now that connection is established
	cm.peerManager.Add(p)

	// Keep connection alive / read loop
	go cm.readLoop(p, conn)
}

func (cm *ConnectionManager) readLoop(p peer.Peer, conn *websocket.Conn) {
	defer func() {
		conn.Close()
		cm.mu.Lock()
		if cm.conns[p.ID] == conn {
			delete(cm.conns, p.ID)
		}
		cm.mu.Unlock()
		
		log.Printf("Connection lost with %s", p.Hostname)
		cm.peerManager.Remove(p.ID)
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		// Handle incoming messages later
	}
}
