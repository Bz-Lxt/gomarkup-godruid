package demo

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"godruid/internal/pool"
)

type WorkloadState struct {
	Running     bool `json:"running"`
	Concurrency int  `json:"concurrency"`
	HoldMS      int  `json:"hold_ms"`
	ThinkMS     int  `json:"think_ms"`
}

type Engine struct {
	pool *pool.Pool

	mu     sync.Mutex
	cancel context.CancelFunc
	state  WorkloadState
	active atomic.Int64
}

func NewEngine(p *pool.Pool) *Engine {
	return &Engine{
		pool: p,
		state: WorkloadState{
			Running:     true,
			Concurrency: 24,
			HoldMS:      40,
			ThinkMS:     15,
		},
	}
}

func (e *Engine) State() WorkloadState {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.state
}

func (e *Engine) Apply(running bool, concurrency, holdMS, thinkMS int) WorkloadState {
	e.mu.Lock()
	if e.cancel != nil {
		e.cancel()
		e.cancel = nil
	}
	e.state = WorkloadState{Running: running, Concurrency: concurrency, HoldMS: holdMS, ThinkMS: thinkMS}
	st := e.state
	e.mu.Unlock()
	if running {
		e.Start()
	}
	return st
}

func (e *Engine) Start() {
	e.mu.Lock()
	if e.cancel != nil {
		e.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel
	n := e.state.Concurrency
	hold := time.Duration(e.state.HoldMS) * time.Millisecond
	think := time.Duration(e.state.ThinkMS) * time.Millisecond
	e.state.Running = true
	e.mu.Unlock()
	for i := 0; i < n; i++ {
		go e.worker(ctx, hold, think)
	}
}

func (e *Engine) Stop() {
	e.mu.Lock()
	if e.cancel != nil {
		e.cancel()
		e.cancel = nil
	}
	e.state.Running = false
	e.mu.Unlock()
}

func (e *Engine) worker(ctx context.Context, hold, think time.Duration) {
	e.active.Add(1)
	defer e.active.Add(-1)
	for {
		if ctx.Err() != nil {
			return
		}
		c, err := e.pool.Get(ctx)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(20 * time.Millisecond):
			}
			continue
		}
		if hold > 0 {
			t := time.NewTimer(hold)
			select {
			case <-ctx.Done():
				t.Stop()
				_ = e.pool.Put(c)
				return
			case <-t.C:
			}
		}
		_ = e.pool.Put(c)
		if think > 0 {
			t := time.NewTimer(think)
			select {
			case <-ctx.Done():
				t.Stop()
				return
			case <-t.C:
			}
		}
	}
}
