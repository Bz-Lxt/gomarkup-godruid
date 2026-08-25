package metrics

import (
	"sync"
	"time"
)

// Ring stores immutable snapshots for the retention window.
type Ring struct {
	mu     sync.RWMutex
	items  []Snapshot
	max    int
	seq    uint64
}

func NewRing(retention, interval time.Duration) *Ring {
	if interval <= 0 {
		interval = time.Second
	}
	n := int(retention / interval)
	if n < 60 {
		n = 60
	}
	return &Ring{max: n, items: make([]Snapshot, 0, n)}
}

func (r *Ring) NextSeq() uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	return r.seq
}

func (r *Ring) Append(s Snapshot) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.items) == r.max {
		r.items = r.items[1:]
	}
	r.items = append(r.items, s)
}

func (r *Ring) Latest() (Snapshot, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.items) == 0 {
		return Snapshot{}, false
	}
	return r.items[len(r.items)-1], true
}

func (r *Ring) Window(now time.Time, d time.Duration) []SeriesPoint {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cut := now.Add(-d)
	out := make([]SeriesPoint, 0, len(r.items))
	for _, s := range r.items {
		if s.ServerTime.Before(cut) {
			continue
		}
		out = append(out, SeriesPoint{
			T:         s.ServerTime,
			BorrowRPS: s.Rates.BorrowRPS,
			ReturnRPS: s.Rates.ReturnRPS,
			HitRate:   s.Rates.HitRate,
			Live:      s.Counts.Live,
			Waiting:   s.Counts.Waiting,
		})
	}
	return out
}

func ParseWindow(s string) (time.Duration, bool) {
	switch s {
	case "", "1m":
		return time.Minute, true
	case "5m":
		return 5 * time.Minute, true
	case "15m":
		return 15 * time.Minute, true
	default:
		return 0, false
	}
}
