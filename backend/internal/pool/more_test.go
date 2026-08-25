package pool_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"godruid/internal/adapter/fake"
	"godruid/internal/pool"
)

func TestCompleteReconnectAndViews(t *testing.T) {
	p, f := newTestPool(t, 2, 4, time.Second)
	c, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	id := c.ID()
	c.MarkInvalid()
	if err := p.Put(c); err != nil {
		t.Fatal(err)
	}
	n := p.BeginReconnect(id)
	if n == nil {
		t.Fatal("missing node")
	}
	raw, err := f.Connect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	p.CompleteReconnect(n, raw, nil)
	if n.Generation() < 2 {
		t.Fatalf("generation %d", n.Generation())
	}
	if p.Waiting() != 0 {
		t.Fatal("waiters")
	}
	views := p.Views()
	if len(views) == 0 {
		t.Fatal("views")
	}
	_ = p.Close()
	_ = p.Close()
}

func TestTTLIsolateAndProbe(t *testing.T) {
	p, _ := newTestPool(t, 2, 4, time.Second)
	c, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Put(c); err != nil {
		t.Fatal(err)
	}
	nodes := p.TakeIdleForTTL(time.Now().Add(2 * time.Hour))
	if len(nodes) == 0 {
		t.Fatal("ttl isolate")
	}
	p.FinishClose(nodes[0])
	c2, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Put(c2); err != nil {
		t.Fatal(err)
	}
	probed := p.TakeIdleForProbe(time.Now())
	if len(probed) == 0 {
		t.Fatal("probe isolate")
	}
	p.CompleteProbe(probed[0], nil)
	p.PurgeStaleIdle()
	if err := p.InvariantOK(); err != nil {
		t.Fatal(err)
	}
}

func TestSanitizeAndCounts(t *testing.T) {
	p, _ := newTestPool(t, 1, 2, time.Second)
	c, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	counts := p.Counts()
	if counts.InUse != 1 {
		t.Fatalf("inuse %v", counts)
	}
	_ = p.Put(c)
	if p.Tombstones() < 0 {
		t.Fatal("tomb")
	}
	if !strings.HasPrefix(c.ID(), "c-") {
		t.Fatal(c.ID())
	}
}

func TestDialFailure(t *testing.T) {
	cfg := pool.DefaultConfig()
	cfg.MaxIdle = 0
	cfg.MaxActive = 1
	cfg.MaxWaitTimeout = 50 * time.Millisecond
	cfg.DialTimeout = 20 * time.Millisecond
	f := fake.New(fake.Options{})
	f.SetFailDial(true)
	p, err := pool.New(f, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	_, err = p.Get(context.Background())
	if err == nil {
		t.Fatal("expected dial/wait error")
	}
}
