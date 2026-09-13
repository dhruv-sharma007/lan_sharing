package connection

import (
	"fmt"
	"log"
	"net/http"
	"os"
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

// peerConn wraps a websocket connection with a mutex for thread-safe writes.
type peerConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

// ConnectionManager handles WebSocket connections between peers.
type ConnectionManager struct {
	myNodeID    uint64
	peerManager peer.PeerManager

	mu    sync.Mutex
	conns map[uint64]*peerConn // key: Peer ID

	msgHandler           func(peerID uint64, msgData []byte)
	peerConnectedHandler func(p peer.Peer)
	mux                  *http.ServeMux
}

// NewConnectionManager creates a new connection manager.
func NewConnectionManager(nodeID uint64, pm peer.PeerManager) *ConnectionManager {
	return &ConnectionManager{
		myNodeID:    nodeID,
		peerManager: pm,
		conns:       make(map[uint64]*peerConn),
		mux:         http.NewServeMux(),
	}
}

// RegisterMessageHandler registers a callback for incoming WebSocket messages.
func (cm *ConnectionManager) RegisterMessageHandler(handler func(peerID uint64, msgData []byte)) {
	cm.msgHandler = handler
}

// RegisterPeerConnectedHandler allows registering a callback for when a peer connects.
func (cm *ConnectionManager) RegisterPeerConnectedHandler(handler func(p peer.Peer)) {
	cm.peerConnectedHandler = handler
}

// RegisterHTTPHandler allows other components to register HTTP endpoints on the same server.
func (cm *ConnectionManager) RegisterHTTPHandler(pattern string, handler http.HandlerFunc) {
	cm.mux.HandleFunc(pattern, handler)
}

// StartServer starts the HTTP server for accepting incoming WebSocket connections.
func (cm *ConnectionManager) StartServer(port int) {
	cm.mux.HandleFunc("/ws", cm.handleIncomingWS)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("WebSocket server listening on %s", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: cm.mux,
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

	// Attempt to get hostname, fallback to "peer" if it fails
	myHostname, err := os.Hostname()
	if err != nil {
		myHostname = "peer"
	}

	// Send our identity to the peer
	myIdentity := peer.Peer{
		ID:       cm.myNodeID,
		Hostname: myHostname,
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
	// An incoming WebSocket identity only contains the peer ID and hostname.
	// Preserve the address details previously obtained through mDNS so later
	// HTTP file transfers still have a valid destination.
	if discoveredPeer, exists := cm.peerManager.Get(p.ID); exists {
		if p.IP == "" {
			p.IP = discoveredPeer.IP
		}
		if p.Port == 0 {
			p.Port = discoveredPeer.Port
		}
		if p.LastSeen.IsZero() {
			p.LastSeen = discoveredPeer.LastSeen
		}
	}

	cm.mu.Lock()
	if existing, exists := cm.conns[p.ID]; exists {
		log.Printf("Closing duplicate connection for %s", p.ID)
		existing.mu.Lock()
		existing.conn.Close()
		existing.mu.Unlock()
	}
	pc := &peerConn{conn: conn}
	cm.conns[p.ID] = pc
	cm.mu.Unlock()

	// Update PeerManager now that connection is established
	cm.peerManager.Add(p)

	if cm.peerConnectedHandler != nil {
		cm.peerConnectedHandler(p)
	}

	// Keep connection alive / read loop
	go cm.readLoop(p, pc)
}

func (cm *ConnectionManager) readLoop(p peer.Peer, pc *peerConn) {
	defer func() {
		pc.mu.Lock()
		pc.conn.Close()
		pc.mu.Unlock()

		cm.mu.Lock()
		if cm.conns[p.ID] == pc {
			delete(cm.conns, p.ID)
		}
		cm.mu.Unlock()

		log.Printf("Connection lost with %s", p.Hostname)
		cm.peerManager.Remove(p.ID)
	}()

	for {
		_, msg, err := pc.conn.ReadMessage()
		if err != nil {
			break
		}
		if cm.msgHandler != nil {
			cm.msgHandler(p.ID, msg)
		}
	}
}

// SendControlMessage safely sends a JSON message to a peer over the WebSocket connection.
func (cm *ConnectionManager) SendControlMessage(peerID uint64, msg interface{}) error {
	cm.mu.Lock()
	pc, exists := cm.conns[peerID]
	cm.mu.Unlock()

	if !exists {
		return fmt.Errorf("no active connection for peer %s", peerID)
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()
	return pc.conn.WriteJSON(msg)
}
