package transfer

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"lan_sharing/internal/peer"
	"lan_sharing/util"
)

type ConnectionManager interface {
	SendControlMessage(peerID uint64, msg interface{}) error
	RegisterMessageHandler(handler func(peerID uint64, msgData []byte))
	RegisterHTTPHandler(pattern string, handler http.HandlerFunc)
}

// Manager coordinates the transfer subsystem.
type Manager struct {
	config  *util.Config
	peerMan peer.PeerManager
	connMan ConnectionManager

	sendDir    string
	receiveDir string

	watcher  *Watcher
	queue    *Queue
	receiver *Receiver

	mu          sync.Mutex
	acceptChans map[string]chan TransferAccept
}

// NewManager creates a new TransferManager.
func NewManager(cfg *util.Config, pm peer.PeerManager, cm ConnectionManager) (*Manager, error) {
	sendDir := filepath.Join(cfg.ShareRoot, "Send")
	receiveDir := filepath.Join(cfg.ShareRoot, "Received")

	m := &Manager{
		config:      cfg,
		peerMan:     pm,
		connMan:     cm,
		sendDir:     sendDir,
		receiveDir:  receiveDir,
		acceptChans: make(map[string]chan TransferAccept),
	}

	m.receiver = NewReceiver(receiveDir)

	// Register HTTP and WS handlers
	cm.RegisterHTTPHandler("/transfer", m.receiver.HandleHTTP)
	cm.RegisterMessageHandler(m.handleIncomingMessage)

	return m, nil
}

// Start starts the watcher and queue.
func (m *Manager) Start(ctx context.Context) error {
	// Ensure receive directory exists
	if err := os.MkdirAll(m.receiveDir, 0755); err != nil {
		return err
	}

	w, err := NewWatcher(m.sendDir, m.onFileStable)
	if err != nil {
		return err
	}
	m.watcher = w

	m.queue = NewQueue(ctx, m.processTransfer)

	return m.watcher.Start(ctx)
}

func (m *Manager) onFileStable(absPath, relPath, peerName string) {
	// Find peer by name/hostname
	var targetPeer peer.Peer
	found := false
	for _, p := range m.peerMan.List() {
		if peerFolderName(p) == peerName {
			targetPeer = p
			found = true
			break
		}
	}

	if !found {
		log.Printf("Cannot transfer %s: peer %s not found or disconnected", relPath, peerName)
		return
	}

	m.queue.Enqueue(TransferRequest{
		FilePath: absPath,
		RelPath:  relPath,
		PeerID:   targetPeer.ID,
	})
}

func (m *Manager) handleIncomingMessage(peerID uint64, msgData []byte) {
	var control ControlMessage
	if err := json.Unmarshal(msgData, &control); err != nil {
		return
	}

	payloadBytes, _ := json.Marshal(control.Payload)

	switch control.Type {
	case TypeTransferOffer:
		var offer TransferOffer
		if err := json.Unmarshal(payloadBytes, &offer); err == nil {
			m.handleOffer(peerID, offer)
		}
	case TypeTransferAccept:
		var accept TransferAccept
		if err := json.Unmarshal(payloadBytes, &accept); err == nil {
			m.mu.Lock()
			ch, exists := m.acceptChans[accept.TransferID]
			m.mu.Unlock()
			if exists {
				ch <- accept
			}
		}
	case TypeTransferComplete:
		log.Printf("Peer %d reported transfer complete", peerID)
	case TypeTransferError:
		log.Printf("Peer %d reported transfer error", peerID)
	}
}

func (m *Manager) handleOffer(peerID uint64, offer TransferOffer) {
	// Verify peer
	p, exists := m.peerMan.Get(peerID)
	if !exists {
		return
	}

	tokenBytes := make([]byte, 16)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	log.Printf("Accepting transfer %s from %s for %s", offer.TransferID, p.Hostname, offer.RelativePath)

	m.receiver.ExpectTransfer(
		offer.TransferID,
		token,
		offer.RelativePath,
		offer.Size,
		func(id string) {
			m.connMan.SendControlMessage(peerID, ControlMessage{
				Type:    TypeTransferComplete,
				Payload: TransferComplete{TransferID: id},
			})
		},
		func(id string, err error) {
			log.Printf("Transfer %s failed: %v", id, err)
			m.connMan.SendControlMessage(peerID, ControlMessage{
				Type:    TypeTransferError,
				Payload: TransferError{TransferID: id, Message: err.Error()},
			})
		},
	)

	m.connMan.SendControlMessage(peerID, ControlMessage{
		Type: TypeTransferAccept,
		Payload: TransferAccept{
			TransferID: offer.TransferID,
			Token:      token,
		},
	})
}

func (m *Manager) processTransfer(ctx context.Context, req TransferRequest) {
	p, exists := m.peerMan.Get(req.PeerID)
	if !exists {
		log.Printf("Transfer failed: peer %d offline", req.PeerID)
		return
	}

	info, err := os.Stat(req.FilePath)
	if err != nil {
		log.Printf("Transfer failed: cannot stat file %s", req.FilePath)
		return
	}

	transferIDBytes := make([]byte, 16)
	rand.Read(transferIDBytes)
	transferID := hex.EncodeToString(transferIDBytes)

	offer := TransferOffer{
		ProtocolVersion: 1,
		TransferID:      transferID,
		SenderNodeID:    m.config.NodeID,
		Filename:        filepath.Base(req.FilePath),
		RelativePath:    req.RelPath,
		Size:            info.Size(),
		ModTime:         info.ModTime().Unix(),
	}

	acceptCh := make(chan TransferAccept, 1)
	m.mu.Lock()
	m.acceptChans[transferID] = acceptCh
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.acceptChans, transferID)
		m.mu.Unlock()
	}()

	err = m.connMan.SendControlMessage(req.PeerID, ControlMessage{
		Type:    TypeTransferOffer,
		Payload: offer,
	})
	if err != nil {
		log.Printf("Transfer failed: %v", err)
		return
	}

	// Wait for accept
	var accept TransferAccept
	select {
	case <-ctx.Done():
		return
	case <-time.After(30 * time.Second):
		log.Printf("Transfer timed out waiting for accept")
		return
	case accept = <-acceptCh:
	}

	// Send file over HTTP POST
	file, err := os.Open(req.FilePath)
	if err != nil {
		log.Printf("Transfer failed: cannot open file %s", req.FilePath)
		return
	}
	defer file.Close()

	url := fmt.Sprintf("http://%s:%d/transfer?id=%s&token=%s", p.IP, p.Port, transferID, accept.Token)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, file)
	if err != nil {
		log.Printf("Transfer failed: %v", err)
		return
	}

	// We must set ContentLength so the client can stream accurately
	httpReq.ContentLength = info.Size()

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("Transfer HTTP POST failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		log.Printf("Transfer rejected by receiver: %s - %s", resp.Status, buf.String())
		return
	}

	log.Printf("Transfer %s to %s completed successfully", req.RelPath, p.Hostname)
}

// HandlePeerConnected creates a directory for the newly connected peer.
func (m *Manager) HandlePeerConnected(p peer.Peer) {
	peerName := peerFolderName(p)
	peerDir := filepath.Join(m.sendDir, peerName)
	if err := m.migrateLegacyPeerDirectory(p, peerDir); err != nil {
		log.Printf("Failed to migrate legacy peer directory for %s: %v", peerName, err)
	}

	if err := os.MkdirAll(peerDir, 0755); err != nil {
		log.Printf("Failed to create peer directory %s: %v", peerDir, err)
	} else {
		log.Printf("Created peer directory: %s", peerDir)
		if m.watcher != nil {
			if err := m.watcher.AddDirectory(peerDir); err != nil {
				log.Printf("Failed to add watch for %s: %v", peerDir, err)
			}
		}
	}
}

// peerFolderName returns the stable, human-readable Send directory for a peer.
// The node ID keeps names unique when multiple devices share a hostname.
func peerFolderName(p peer.Peer) string {
	return fmt.Sprintf("%s-%d", p.Hostname, p.ID)
}

// legacyPeerFolderName is the malformed name created before peerFolderName
// was introduced. It is retained only to migrate existing Send folders.
func legacyPeerFolderName(p peer.Peer) string {
	return fmt.Sprintf("%s%%!(EXTRA string=-%%d, uint64=%d)", p.Hostname, p.ID)
}

func (m *Manager) migrateLegacyPeerDirectory(p peer.Peer, peerDir string) error {
	legacyDir := filepath.Join(m.sendDir, legacyPeerFolderName(p))

	if _, err := os.Stat(peerDir); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	if _, err := os.Stat(legacyDir); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	if err := os.Rename(legacyDir, peerDir); err != nil {
		return err
	}

	log.Printf("Migrated peer directory from %s to %s", legacyDir, peerDir)
	return nil
}
