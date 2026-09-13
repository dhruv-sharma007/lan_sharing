package transfer

import (
	"os"
	"path/filepath"
	"testing"

	"lan_sharing/internal/peer"
)

func TestPeerFolderName(t *testing.T) {
	p := peer.Peer{
		Hostname: "acer-ACERdhruv",
		ID:       29316217849356388,
	}

	got := peerFolderName(p)
	want := "acer-ACERdhruv-29316217849356388"
	if got != want {
		t.Fatalf("peerFolderName() = %q, want %q", got, want)
	}
}

func TestLegacyPeerFolderName(t *testing.T) {
	p := peer.Peer{
		Hostname: "acer-ACERdhruv",
		ID:       29316217849356388,
	}

	got := legacyPeerFolderName(p)
	want := "acer-ACERdhruv%!(EXTRA string=-%d, uint64=29316217849356388)"
	if got != want {
		t.Fatalf("legacyPeerFolderName() = %q, want %q", got, want)
	}
}

func TestMigrateLegacyPeerDirectory(t *testing.T) {
	p := peer.Peer{Hostname: "acer-ACERdhruv", ID: 29316217849356388}
	sendDir := t.TempDir()
	m := &Manager{sendDir: sendDir}

	legacyDir := filepath.Join(sendDir, legacyPeerFolderName(p))
	if err := os.MkdirAll(legacyDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacyDir, "queued.txt"), []byte("queued"), 0644); err != nil {
		t.Fatal(err)
	}

	peerDir := filepath.Join(sendDir, peerFolderName(p))
	if err := m.migrateLegacyPeerDirectory(p, peerDir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(peerDir, "queued.txt")); err != nil {
		t.Fatalf("queued file was not migrated: %v", err)
	}
	if _, err := os.Stat(legacyDir); !os.IsNotExist(err) {
		t.Fatalf("legacy directory still exists: %v", err)
	}
}
