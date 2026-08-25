package lifecycle

import (
	"context"
	"sync"
)

// Group tracks bounded background dial/probe/reconnect work.
type Group struct {
	sem chan struct{}
	wg  sync.WaitGroup
	mu  sync.Mutex
}

func NewGroup(limit int) *Group {
	if limit < 1 {
		limit = 8
	}
	return &Group{sem: make(chan struct{}, limit)}
}

func (g *Group) Go(ctx context.Context, fn func(context.Context)) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	select {
	case g.sem <- struct{}{}:
	case <-ctx.Done():
		return false
	}
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		defer func() { <-g.sem }()
		fn(ctx)
	}()
	return true
}

func (g *Group) Wait() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.wg.Wait()
}
