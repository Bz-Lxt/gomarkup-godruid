package pool_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"godruid/internal/adapter/fake"
	"godruid/internal/pool"
)

func TestGetAfterClose(t *testing.T) {
	p, _ := newTestPool(t, 1, 2, time.Second)
	_ = p.Close()
	_, err := p.Get(context.Background())
	if err == nil {
		t.Fatal("expected closed")
	}
	if !p.Closed() {
		t.Fatal("closed flag")
	}
}

func TestPutNil(t *testing.T) {
	p, _ := newTestPool(t, 1, 2, time.Second)
	if err := p.Put(nil); err == nil {
		t.Fatal("nil put")
	}
}

func TestCompleteReconnectFail(t *testing.T) {
	p, _ := newTestPool(t, 2, 4, time.Second)
	c, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	id := c.ID()
	c.MarkInvalid()
	_ = p.Put(c)
	n := p.BeginReconnect(id)
	p.CompleteReconnect(n, nil, errors.New("x"))
	if p.Find(id) != nil && p.Find(id).State() != pool.StateClosed && p.Find(id).State() != pool.StateClosing {
		if p.Counts().Live > 2 {
			t.Fatalf("counts %#v", p.Counts())
		}
	}
}

func TestOptionsAndName(t *testing.T) {
	f := fake.New(fake.Options{})
	p, err := pool.New(f, pool.DefaultConfig(), pool.WithName("n"), pool.WithID("p1"), pool.WithOnClose(func(*pool.Node) {}))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	if p.ID() != "p1" || p.Name() != "n" || p.ConnectorKind() != "fake" {
		t.Fatal(p.ID(), p.Name())
	}
	c, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if c.Raw() == nil || c.Generation() == 0 {
		t.Fatal("handle")
	}
	_ = p.Put(c)
}

func TestStateLive(t *testing.T) {
	if pool.StateClosed.Live() {
		t.Fatal("closed live")
	}
	if !pool.StateIdle.Live() {
		t.Fatal("idle live")
	}
}
