package pool_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"godruid/internal/connx"
	"godruid/internal/pool"
)

type slowCloseConnector struct {
	seq          atomic.Uint64
	closeStarted chan struct{}
	releaseClose chan struct{}
}

func newSlowCloseConnector() *slowCloseConnector {
	return &slowCloseConnector{
		closeStarted: make(chan struct{}),
		releaseClose: make(chan struct{}),
	}
}

func (c *slowCloseConnector) Connect(context.Context) (connx.Connection, error) {
	return &slowCloseConn{parent: c, id: c.seq.Add(1)}, nil
}

func (*slowCloseConnector) Kind() string { return "slow-close" }

type slowCloseConn struct {
	parent *slowCloseConnector
	id     uint64
}

func (*slowCloseConn) Ping(context.Context) error { return nil }

func (c *slowCloseConn) Close() error {
	if c.id == 1 {
		close(c.parent.closeStarted)
		<-c.parent.releaseClose
	}
	return nil
}

func (*slowCloseConn) Metadata() connx.Metadata { return connx.Metadata{Kind: "slow-close"} }

func TestSlowCloseDoesNotBlockIndependentReturn(t *testing.T) {
	cfg := pool.DefaultConfig()
	cfg.MaxActive = 2
	cfg.MaxIdle = 0
	connector := newSlowCloseConnector()
	p, err := pool.New(connector, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = p.Close() })

	first, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	second, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	firstDone := make(chan error, 1)
	go func() { firstDone <- p.Put(first) }()
	select {
	case <-connector.closeStarted:
	case <-time.After(time.Second):
		t.Fatal("first connection did not enter Close")
	}

	secondDone := make(chan error, 1)
	go func() { secondDone <- p.Put(second) }()
	var secondErr error
	blocked := false
	select {
	case secondErr = <-secondDone:
	case <-time.After(100 * time.Millisecond):
		blocked = true
	}

	close(connector.releaseClose)
	if blocked {
		secondErr = <-secondDone
	}
	firstErr := <-firstDone
	if firstErr != nil {
		t.Fatalf("first Put: %v", firstErr)
	}
	if secondErr != nil {
		t.Fatalf("second Put: %v", secondErr)
	}
	if blocked {
		t.Fatal("an unrelated Put waited for another connection's Close")
	}
}
