package metrics

import (
	"sort"
	"sync"
	"time"
)

type waitSample struct {
	at      time.Time
	latency time.Duration
}

// WaitWindow keeps completed wait samples for the last 60 seconds.
type WaitWindow struct {
	mu      sync.Mutex
	samples []waitSample
	retain  time.Duration
}

func NewWaitWindow(retain time.Duration) *WaitWindow {
	if retain <= 0 {
		retain = time.Minute
	}
	return &WaitWindow{retain: retain}
}

func (w *WaitWindow) Add(now time.Time, d time.Duration) {
	if w == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.samples = append(w.samples, waitSample{at: now, latency: d})
	w.gcLocked(now)
}

func (w *WaitWindow) Stats(now time.Time) (avg, p50, p95, p99 time.Duration, n int) {
	if w == nil {
		return 0, 0, 0, 0, 0
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.gcLocked(now)
	n = len(w.samples)
	if n == 0 {
		return 0, 0, 0, 0, 0
	}
	vals := make([]time.Duration, n)
	var sum time.Duration
	for i, s := range w.samples {
		vals[i] = s.latency
		sum += s.latency
	}
	sort.Slice(vals, func(i, j int) bool { return vals[i] < vals[j] })
	avg = sum / time.Duration(n)
	p50 = percentile(vals, 0.50)
	p95 = percentile(vals, 0.95)
	p99 = percentile(vals, 0.99)
	return avg, p50, p95, p99, n
}

func (w *WaitWindow) gcLocked(now time.Time) {
	cut := now.Add(-w.retain)
	i := 0
	for i < len(w.samples) && w.samples[i].at.Before(cut) {
		i++
	}
	if i > 0 {
		w.samples = append([]waitSample(nil), w.samples[i:]...)
	}
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
