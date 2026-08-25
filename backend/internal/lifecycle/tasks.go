package lifecycle

import (
	"context"
	"sync"
)

// Group tracks bounded background dial/probe/reconnect work.
//
// Tasks receive a context derived from the Group's own stop context so that
// Stop can interrupt in-flight dials promptly; the parent ctx passed to Go is
// only consulted for admission control. This deliberately does NOT use
// context.WithoutCancel: doing so would detach reconnect/probe work from
// Stop, forcing every shutdown to wait out a full DialTimeout per attempt.
type Group struct {
	sem  chan struct{}
	wg   sync.WaitGroup
	stop context.CancelFunc
	ctx  context.Context
}

func NewGroup(limit int) *Group {
	if limit < 1 {
		limit = 8
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Group{
		sem:  make(chan struct{}, limit),
		ctx:  ctx,
		stop: cancel,
	}
}

// Go admits a task under the admission ctx, then runs fn with the Group's
// cancellable task context. fn's context is cancelled by Close, so in-flight
// dials and backoff waits abort promptly on shutdown.
func (g *Group) Go(ctx context.Context, fn func(context.Context)) bool {
	select {
	case g.sem <- struct{}{}:
	case <-ctx.Done():
		return false
	}
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		defer func() { <-g.sem }()
		fn(g.ctx)
	}()
	return true
}

// Close cancels all in-flight task contexts. Callers must then call Wait to
// drain remaining goroutines.
func (g *Group) Close() {
	g.stop()
}

func (g *Group) Wait() { g.wg.Wait() }
