package connection

import (
	"net"
	"testing"
	"time"

	"lan_sharing/internal/peer"
)

func TestConnectToPeerRecordsInboundEndpoint(t *testing.T) {
	firstPort := reserveTCPPort(t)
	secondPort := reserveTCPPort(t)

	firstPeers := peer.NewManager()
	first := NewConnectionManager(1, firstPeers)
	if err := first.StartServer(firstPort); err != nil {
		t.Fatalf("start first server: %v", err)
	}

	secondPeers := peer.NewManager()
	second := NewConnectionManager(2, secondPeers)
	if err := second.StartServer(secondPort); err != nil {
		t.Fatalf("start second server: %v", err)
	}

	first.ConnectToPeer(peer.Peer{
		ID:       2,
		Hostname: "second",
		IP:       "127.0.0.1",
		Port:     secondPort,
	})

	stored := waitForPeer(t, secondPeers, 1)
	if stored.IP != "127.0.0.1" {
		t.Errorf("stored IP = %q, want 127.0.0.1", stored.IP)
	}
	if stored.Port != firstPort {
		t.Errorf("stored port = %d, want %d", stored.Port, firstPort)
	}
}

func TestStartServerReturnsPortConflict(t *testing.T) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	manager := NewConnectionManager(1, peer.NewManager())
	if err := manager.StartServer(port); err == nil {
		t.Fatal("StartServer succeeded while the port was already in use")
	}
}

func reserveTCPPort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}

func waitForPeer(t *testing.T, peers peer.PeerManager, id uint64) peer.Peer {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if stored, found := peers.Get(id); found {
			return stored
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("peer %d was not registered", id)
	return peer.Peer{}
}
