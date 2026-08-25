package pool

import (
	"context"
	"time"
)

func (p *Pool) waitDeadline(ctx context.Context) time.Time {
	now := p.clk.Now()
	limit := now.Add(p.cfg.MaxWaitTimeout)
	if dl, ok := ctx.Deadline(); ok && dl.Before(limit) {
		return dl
	}
	return limit
}

func (p *Pool) addWaiter() {
	p.mu.Lock()
	p.waiters++
	p.mu.Unlock()
}

func (p *Pool) removeWaiter() {
	p.mu.Lock()
	if p.waiters > 0 {
		p.waiters--
	}
	p.mu.Unlock()
}

func (p *Pool) Waiting() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.waiters
}

func (p *Pool) waitIdle(ctx context.Context) (*Node, error) {
	select {
	case tok := <-p.idle:
		return p.claimToken(tok)
	case <-ctx.Done():
		return nil, MapWaitError(ctx.Err())
	case <-p.done:
		return nil, ErrPoolClosed
	}
}
