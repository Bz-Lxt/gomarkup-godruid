package clock

import (
	"sync"
	"time"
)

// Clock abstracts time so pool tests can advance without sleeping.
type Clock interface {
	Now() time.Time
	Since(t time.Time) time.Duration
	NewTicker(d time.Duration) Ticker
	After(d time.Duration) <-chan time.Time
	Sleep(d time.Duration)
}

type Ticker interface {
	C() <-chan time.Time
	Stop()
}

type Real struct{}

func (Real) Now() time.Time                         { return time.Now() }
func (Real) Since(t time.Time) time.Duration        { return time.Since(t) }
func (Real) NewTicker(d time.Duration) Ticker       { return realTicker{time.NewTicker(d)} }
func (Real) After(d time.Duration) <-chan time.Time { return time.After(d) }
func (Real) Sleep(d time.Duration)                  { time.Sleep(d) }

type realTicker struct{ *time.Ticker }

func (t realTicker) C() <-chan time.Time { return t.Ticker.C }

// Fake is a manually advanced clock. Tickers fire when Now reaches their next tick.
type Fake struct {
	mu      sync.Mutex
	now     time.Time
	tickers []*fakeTicker
}

func NewFake(start time.Time) *Fake {
	return &Fake{now: start}
}

func (f *Fake) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

func (f *Fake) Since(t time.Time) time.Duration {
	return f.Now().Sub(t)
}

func (f *Fake) Sleep(d time.Duration) {
	f.Advance(d)
}

func (f *Fake) After(d time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	t := f.NewTicker(d).(*fakeTicker)
	go func() {
		ts, ok := <-t.ch
		if !ok {
			return
		}
		ch <- ts
		t.Stop()
	}()
	return ch
}

func (f *Fake) NewTicker(d time.Duration) Ticker {
	if d <= 0 {
		d = time.Nanosecond
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	t := &fakeTicker{
		period: d,
		next:   f.now.Add(d),
		ch:     make(chan time.Time, 8),
	}
	f.tickers = append(f.tickers, t)
	return t
}

func (f *Fake) Advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = f.now.Add(d)
	now := f.now
	for _, t := range f.tickers {
		t.fire(now)
	}
}

type fakeTicker struct {
	mu     sync.Mutex
	period time.Duration
	next   time.Time
	ch     chan time.Time
	stopped bool
}

func (t *fakeTicker) C() <-chan time.Time { return t.ch }

func (t *fakeTicker) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.stopped {
		return
	}
	t.stopped = true
}

func (t *fakeTicker) fire(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.stopped {
		return
	}
	for !t.next.After(now) {
		select {
		case t.ch <- t.next:
		default:
		}
		t.next = t.next.Add(t.period)
	}
}
