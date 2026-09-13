package transfer

import (
	"context"
	"log"
	"sync"
)

// TransferRequest represents a file to be sent to a specific peer.
type TransferRequest struct {
	FilePath string // Absolute path to the local file
	RelPath  string // Relative path to preserve structure
	PeerID   uint64 // Target peer ID
}

// Queue manages pending transfers, ensuring only 1 active transfer per peer.
type Queue struct {
	mu      sync.Mutex
	workers map[uint64]chan TransferRequest
	ctx     context.Context
	process func(ctx context.Context, req TransferRequest)
}

// NewQueue creates a new transfer queue.
func NewQueue(ctx context.Context, processFunc func(ctx context.Context, req TransferRequest)) *Queue {
	return &Queue{
		workers: make(map[uint64]chan TransferRequest),
		ctx:     ctx,
		process: processFunc,
	}
}

// Enqueue adds a transfer request to the queue for the specified peer.
// If a worker for the peer doesn't exist, it starts one.
func (q *Queue) Enqueue(req TransferRequest) {
	q.mu.Lock()
	defer q.mu.Unlock()

	ch, exists := q.workers[req.PeerID]
	if !exists {
		ch = make(chan TransferRequest, 100) // Buffer to hold pending requests
		q.workers[req.PeerID] = ch
		go q.workerLoop(req.PeerID, ch)
	}

	select {
	case ch <- req:
		log.Printf("Transfer queued for peer %d: %s", req.PeerID, req.RelPath)
	default:
		log.Printf("Queue full for peer %d, dropping transfer request for %s", req.PeerID, req.RelPath)
	}
}

func (q *Queue) workerLoop(peerID uint64, ch <-chan TransferRequest) {
	log.Printf("Started transfer worker for peer %d", peerID)
	for {
		select {
		case <-q.ctx.Done():
			log.Printf("Stopping transfer worker for peer %d", peerID)
			return
		case req := <-ch:
			log.Printf("Starting transfer for %s to %d", req.RelPath, req.PeerID)
			q.process(q.ctx, req)
			log.Printf("Finished transfer attempt for %s to %d", req.RelPath, req.PeerID)
		}
	}
}
