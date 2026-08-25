package observe

import (
	"sync"
	"sync/atomic"

	"godruid/internal/metrics"
)

const subBuf = 8

type sub struct {
	ch     chan metrics.Snapshot
	closed atomic.Bool
}

// Hub fans snapshots to SSE clients with bounded buffers.
type Hub struct {
	mu   sync.Mutex
	subs map[*sub]struct{}
}

func NewHub() *Hub {
	return &Hub{subs: map[*sub]struct{}{}}
}

func (h *Hub) Subscribe() (<-chan metrics.Snapshot, func()) {
	s := &sub{ch: make(chan metrics.Snapshot, subBuf)}
	h.mu.Lock()
	h.subs[s] = struct{}{}
	h.mu.Unlock()
	cancel := func() {
		if s.closed.Swap(true) {
			return
		}
		h.mu.Lock()
		delete(h.subs, s)
		h.mu.Unlock()
		close(s.ch)
	}
	return s.ch, cancel
}

func (h *Hub) Publish(s metrics.Snapshot) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for sub := range h.subs {
		select {
		case sub.ch <- s:
		default:
			// drop frame; client detects seq gap and resyncs
		}
	}
}

func (h *Hub) Len() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.subs)
}
