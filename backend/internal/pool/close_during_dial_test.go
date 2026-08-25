package pool_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"godruid/internal/connx"
	"godruid/internal/pool"
)

type stagedConnector struct {
	started chan struct{}
	release chan struct{}
	closed  atomic.Bool
}

func (c *stagedConnector) Kind() string { return "staged" }

func (c *stagedConnector) Connect(ctx context.Context) (connx.Connection, error) {
	close(c.started)
	select {
	case <-c.release:
		return &stagedConnection{closed: &c.closed}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type stagedConnection struct {
	closed *atomic.Bool
}

func (c *stagedConnection) Ping(context.Context) error { return nil }

func (c *stagedConnection) Close() error {
	c.closed.Store(true)
	return nil
}

func (c *stagedConnection) Metadata() connx.Metadata {
	return connx.Metadata{Kind: "staged", Isolated: true}
}

func TestCloseDuringDialLeavesNoLiveConnection(t *testing.T) {
	connector := &stagedConnector{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	cfg := pool.DefaultConfig()
	cfg.MaxActive = 1
	cfg.MaxIdle = 1
	p, err := pool.New(connector, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = p.Close() })

	getDone := make(chan error, 1)
	go func() {
		_, err := p.Get(context.Background())
		getDone <- err
	}()

	select {
	case <-connector.started:
	case <-time.After(time.Second):
		t.Fatal("connection attempt did not start")
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
	close(connector.release)

	select {
	case err := <-getDone:
		if !errors.Is(err, pool.ErrPoolClosed) {
			t.Fatalf("Get error = %v, want %v", err, pool.ErrPoolClosed)
		}
	case <-time.After(time.Second):
		t.Fatal("Get did not return after the connection attempt completed")
	}

	if !connector.closed.Load() {
		t.Fatal("connection created during shutdown was not closed")
	}
	counts := p.Counts()
	if counts.Live != 0 {
		t.Fatalf("closed pool reports %d live connection(s): %#v", counts.Live, counts)
	}
}
