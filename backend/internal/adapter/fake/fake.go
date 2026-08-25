package fake

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"godruid/internal/connx"
)

var (
	ErrInjected = errors.New("injected failure")
	ErrClosed   = errors.New("fake connection closed")
	ErrCanceled = context.Canceled
)

type Options struct {
	DialDelay time.Duration
	PingDelay time.Duration
}

type Connector struct {
	mu        sync.Mutex
	seq       atomic.Uint64
	failPing  bool
	failDial  bool
	dropNext  int
	dialDelay time.Duration
	pingDelay time.Duration
	closes    atomic.Int64
	dials     atomic.Int64
}

func New(opts Options) *Connector {
	return &Connector{dialDelay: opts.DialDelay, pingDelay: opts.PingDelay}
}

func (c *Connector) Kind() string { return "fake" }

func (c *Connector) SetFailPing(v bool) {
	c.mu.Lock()
	c.failPing = v
	c.mu.Unlock()
}

// SetDialDelay updates the connect latency for subsequent Connect calls. It is
// used by tests to inject a slow dial that only context cancellation can end.
func (c *Connector) SetDialDelay(d time.Duration) {
	c.mu.Lock()
	c.dialDelay = d
	c.mu.Unlock()
}

func (c *Connector) SetFailDial(v bool) {
	c.mu.Lock()
	c.failDial = v
	c.mu.Unlock()
}

func (c *Connector) SetDropNext(n int) {
	c.mu.Lock()
	c.dropNext = n
	c.mu.Unlock()
}

func (c *Connector) FaultState() (failPing, failDial bool, dropNext int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.failPing, c.failDial, c.dropNext
}

func (c *Connector) CloseCount() int64 { return c.closes.Load() }
func (c *Connector) DialCount() int64  { return c.dials.Load() }

func (c *Connector) Connect(ctx context.Context) (connx.Connection, error) {
	c.dials.Add(1)
	c.mu.Lock()
	fail := c.failDial
	delay := c.dialDelay
	c.mu.Unlock()
	if delay > 0 {
		if err := wait(ctx, delay); err != nil {
			return nil, err
		}
	}
	if fail {
		return nil, ErrInjected
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	id := c.seq.Add(1)
	return &Conn{parent: c, id: id}, nil
}

type Conn struct {
	parent *Connector
	id     uint64
	closed atomic.Bool
	pings  atomic.Int64
}

func (c *Conn) Ping(ctx context.Context) error {
	if c.closed.Load() {
		return ErrClosed
	}
	c.pings.Add(1)
	c.parent.mu.Lock()
	delay := c.parent.pingDelay
	fail := c.parent.failPing
	if c.parent.dropNext > 0 {
		c.parent.dropNext--
		fail = true
	}
	c.parent.mu.Unlock()
	if delay > 0 {
		if err := wait(ctx, delay); err != nil {
			return err
		}
	}
	if fail {
		return ErrInjected
	}
	return ctx.Err()
}

func (c *Conn) Close() error {
	if c.closed.Swap(true) {
		return nil
	}
	c.parent.closes.Add(1)
	return nil
}

func (c *Conn) Metadata() connx.Metadata {
	return connx.Metadata{Kind: "fake", Remote: "memory", Isolated: true, Note: "in-process simulator"}
}

func (c *Conn) Closed() bool { return c.closed.Load() }

func wait(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
