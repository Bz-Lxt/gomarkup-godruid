package lifecycle_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"godruid/internal/connx"
	"godruid/internal/lifecycle"
	"godruid/internal/pool"
)

var errTransientDial = errors.New("transient dial failure")

type retryConnector struct {
	calls atomic.Int64
}

func (c *retryConnector) Kind() string { return "retry-test" }

func (c *retryConnector) Connect(ctx context.Context) (connx.Connection, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if c.calls.Add(1) == 2 {
		return nil, errTransientDial
	}
	return &retryConnection{}, nil
}

type retryConnection struct {
	closed atomic.Bool
}

func (c *retryConnection) Ping(ctx context.Context) error {
	if c.closed.Load() {
		return errors.New("connection closed")
	}
	return ctx.Err()
}

func (c *retryConnection) Close() error {
	c.closed.Store(true)
	return nil
}

func (c *retryConnection) Metadata() connx.Metadata {
	return connx.Metadata{Kind: "retry-test", Remote: "memory"}
}

func TestReconnectRecoversAfterTransientDialError(t *testing.T) {
	cfg := pool.DefaultConfig()
	cfg.MaxIdle = 1
	cfg.MaxActive = 1
	cfg.MaxWaitTimeout = 200 * time.Millisecond
	cfg.DialTimeout = 100 * time.Millisecond
	cfg.IdleTTL = time.Hour
	cfg.HealthInterval = time.Hour
	cfg.ReconnectAttempts = 2
	cfg.ReconnectBaseDelay = time.Millisecond
	cfg.ReconnectMaxDelay = time.Millisecond

	connector := &retryConnector{}
	p, err := pool.New(connector, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	supervisor := lifecycle.NewSupervisor(p, nil)
	supervisor.Start()
	defer supervisor.Stop()

	conn, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	id, generation := conn.ID(), conn.Generation()
	conn.MarkInvalid()
	if err := p.Put(conn); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(time.Second)
	for p.Metrics().ReconnectOK.Load()+p.Metrics().ReconnectFail.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if p.Metrics().ReconnectOK.Load()+p.Metrics().ReconnectFail.Load() == 0 {
		t.Fatal("reconnect did not complete")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	reconnected, err := p.Get(ctx)
	if err != nil {
		t.Fatalf("borrow after transient reconnect error: %v", err)
	}
	defer p.Put(reconnected)
	if reconnected.ID() != id {
		t.Fatalf("transient reconnect replaced logical connection: got id %q, want %q", reconnected.ID(), id)
	}
	if reconnected.Generation() != generation+1 {
		t.Fatalf("generation after successful retry = %d, want %d", reconnected.Generation(), generation+1)
	}
}
