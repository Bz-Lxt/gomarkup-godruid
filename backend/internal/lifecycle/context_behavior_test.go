package lifecycle_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"godruid/internal/connx"
	"godruid/internal/lifecycle"
	"godruid/internal/pool"
)

type blockingReconnectConnector struct {
	dials       atomic.Int32
	started     chan struct{}
	startedOnce sync.Once
	release     chan struct{}
	releaseOnce sync.Once
}

func newBlockingReconnectConnector() *blockingReconnectConnector {
	return &blockingReconnectConnector{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
}

func (c *blockingReconnectConnector) Kind() string { return "blocking-reconnect" }

func (c *blockingReconnectConnector) Connect(ctx context.Context) (connx.Connection, error) {
	if c.dials.Add(1) == 1 {
		return blockingReconnectConn{}, nil
	}
	c.startedOnce.Do(func() { close(c.started) })
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.release:
		return nil, context.Canceled
	}
}

func (c *blockingReconnectConnector) unblock() {
	c.releaseOnce.Do(func() { close(c.release) })
}

type blockingReconnectConn struct{}

func (blockingReconnectConn) Ping(context.Context) error { return nil }
func (blockingReconnectConn) Close() error               { return nil }
func (blockingReconnectConn) Metadata() connx.Metadata {
	return connx.Metadata{Kind: "blocking-reconnect", Isolated: true}
}

func TestSupervisorStopCancelsInFlightReconnect(t *testing.T) {
	connector := newBlockingReconnectConnector()
	t.Cleanup(connector.unblock)

	cfg := pool.DefaultConfig()
	cfg.MaxIdle = 1
	cfg.MaxActive = 1
	cfg.IdleTTL = time.Hour
	cfg.HealthInterval = time.Hour
	cfg.DialTimeout = 2 * time.Second
	cfg.ReconnectAttempts = 1
	cfg.ReconnectBaseDelay = time.Millisecond
	cfg.ReconnectMaxDelay = time.Millisecond
	p, err := pool.New(connector, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = p.Close() })

	supervisor := lifecycle.NewSupervisor(p, nil)
	supervisor.Start()
	c, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	c.MarkInvalid()
	if err := p.Put(c); err != nil {
		t.Fatal(err)
	}

	select {
	case <-connector.started:
	case <-time.After(time.Second):
		t.Fatal("reconnect did not start")
	}

	stopped := make(chan struct{})
	started := time.Now()
	go func() {
		supervisor.Stop()
		close(stopped)
	}()

	const stopLimit = 150 * time.Millisecond
	select {
	case <-stopped:
	case <-time.After(stopLimit):
		connector.unblock()
		<-stopped
		t.Fatalf("Stop remained blocked for %v after reconnect started; want cancellation within %v", time.Since(started), stopLimit)
	}
}
