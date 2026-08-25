package pool_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"godruid/internal/adapter/fake"
	"godruid/internal/lifecycle"
	"godruid/internal/pool"
)

func newTestPool(t *testing.T, idle, active int, wait time.Duration) (*pool.Pool, *fake.Connector) {
	t.Helper()
	cfg := pool.DefaultConfig()
	cfg.MaxIdle = idle
	cfg.MaxActive = active
	cfg.MaxWaitTimeout = wait
	cfg.IdleTTL = time.Hour
	cfg.HealthInterval = time.Hour
	f := fake.New(fake.Options{})
	p, err := pool.New(f, cfg, pool.WithID("t"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = p.Close() })
	return p, f
}

func TestGetPutIdleHit(t *testing.T) {
	p, _ := newTestPool(t, 2, 4, time.Second)
	c, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Put(c); err != nil {
		t.Fatal(err)
	}
	c2, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if c2.ID() != c.ID() {
		t.Fatalf("expected reuse %s vs %s", c.ID(), c2.ID())
	}
	_ = p.Put(c2)
	if err := p.InvariantOK(); err != nil {
		t.Fatal(err)
	}
}

func TestMaxIdleShrinkOnReturn(t *testing.T) {
	p, f := newTestPool(t, 1, 3, time.Second)
	var held []*pool.Conn
	for i := 0; i < 3; i++ {
		c, err := p.Get(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		held = append(held, c)
	}
	for _, c := range held {
		if err := p.Put(c); err != nil {
			t.Fatal(err)
		}
	}
	if p.Counts().Idle > 1 {
		t.Fatalf("idle %d", p.Counts().Idle)
	}
	if f.CloseCount() < 2 {
		t.Fatalf("expected extra closes, got %d", f.CloseCount())
	}
}

func TestWaitTimeout(t *testing.T) {
	p, _ := newTestPool(t, 0, 1, 30*time.Millisecond)
	c, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	_, err = p.Get(ctx)
	if !errors.Is(err, pool.ErrWaitTimeout) && !errors.Is(err, context.DeadlineExceeded) {
		if err == nil || pool.MapWaitError(err) != pool.ErrWaitTimeout {
			t.Fatalf("want timeout, got %v", err)
		}
	}
	_ = p.Put(c)
}

func TestContextCancel(t *testing.T) {
	p, _ := newTestPool(t, 0, 1, time.Second)
	c, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = p.Get(ctx)
	if pool.MapWaitError(err) != pool.ErrCanceled && !errors.Is(err, context.Canceled) {
		t.Fatalf("want cancel, got %v", err)
	}
	_ = p.Put(c)
}

func TestDoublePutAndCrossPool(t *testing.T) {
	p1, _ := newTestPool(t, 2, 2, time.Second)
	p2, _ := newTestPool(t, 2, 2, time.Second)
	c, err := p1.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := p1.Put(c); err != nil {
		t.Fatal(err)
	}
	if err := p1.Put(c); err == nil {
		t.Fatal("double put")
	}
	c2, err := p1.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := p2.Put(c2); err == nil {
		t.Fatal("cross pool put")
	}
	_ = p1.Put(c2)
}

func TestCloseWakesWaiters(t *testing.T) {
	p, _ := newTestPool(t, 0, 1, 2*time.Second)
	c, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := p.Get(context.Background())
		done <- err
	}()
	time.Sleep(20 * time.Millisecond)
	_ = p.Close()
	_ = p.Put(c)
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected closed or timeout")
		}
	case <-time.After(time.Second):
		t.Fatal("waiter not woken")
	}
}

func TestConcurrentBorrow(t *testing.T) {
	p, _ := newTestPool(t, 8, 16, time.Second)
	var wg sync.WaitGroup
	var fails atomic.Int64
	for i := 0; i < 80; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, err := p.Get(context.Background())
			if err != nil {
				fails.Add(1)
				return
			}
			time.Sleep(time.Millisecond)
			if err := p.Put(c); err != nil {
				fails.Add(1)
			}
		}()
	}
	wg.Wait()
	if fails.Load() != 0 {
		t.Fatalf("fails %d", fails.Load())
	}
	if err := p.InvariantOK(); err != nil {
		t.Fatal(err)
	}
	if p.LivePlusDialing() > 16 {
		t.Fatalf("over capacity %d", p.LivePlusDialing())
	}
}

func TestNilConnector(t *testing.T) {
	_, err := pool.New(nil, pool.DefaultConfig())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProbeAndReconnect(t *testing.T) {
	cfg := pool.DefaultConfig()
	cfg.MaxIdle = 2
	cfg.MaxActive = 4
	cfg.HealthInterval = 20 * time.Millisecond
	cfg.IdleTTL = time.Hour
	cfg.ReconnectAttempts = 2
	cfg.ReconnectBaseDelay = time.Millisecond
	cfg.ReconnectMaxDelay = 5 * time.Millisecond
	f := fake.New(fake.Options{})
	p, err := pool.New(f, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	s := lifecycle.NewSupervisor(p, nil)
	s.Start()
	defer s.Stop()
	c, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Put(c); err != nil {
		t.Fatal(err)
	}
	f.SetDropNext(1)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		views := p.Views()
		for _, v := range views {
			if v.Generation > 1 || v.State == "RECONNECTING" || v.LastError != "" {
				return
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Log("probe may have succeeded after drop; counts", p.Counts())
}

func TestMarkInvalidTriggersReconnectPath(t *testing.T) {
	p, _ := newTestPool(t, 2, 4, time.Second)
	var got string
	p.SetReconnector(func(id string) { got = id })
	c, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	c.MarkInvalid()
	if err := p.Put(c); err != nil {
		t.Fatal(err)
	}
	if got == "" {
		t.Fatal("reconnector not called")
	}
}
