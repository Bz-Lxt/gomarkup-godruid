package lifecycle_test

import (
	"context"
	"testing"
	"time"

	"godruid/internal/adapter/fake"
	"godruid/internal/lifecycle"
	"godruid/internal/pool"
)

func TestTTLShrink(t *testing.T) {
	cfg := pool.DefaultConfig()
	cfg.MaxIdle = 2
	cfg.MaxActive = 4
	cfg.IdleTTL = 40 * time.Millisecond
	cfg.HealthInterval = 20 * time.Millisecond
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
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if p.Counts().Live == 0 && f.CloseCount() >= 1 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("ttl did not shrink live=%d closes=%d", p.Counts().Live, f.CloseCount())
}

func TestReconnectAfterPingFail(t *testing.T) {
	cfg := pool.DefaultConfig()
	cfg.MaxIdle = 2
	cfg.MaxActive = 4
	cfg.IdleTTL = time.Hour
	cfg.HealthInterval = 15 * time.Millisecond
	cfg.HealthTimeout = 50 * time.Millisecond
	cfg.ReconnectAttempts = 3
	cfg.ReconnectBaseDelay = time.Millisecond
	cfg.ReconnectMaxDelay = 8 * time.Millisecond
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
	id := c.ID()
	if err := p.Put(c); err != nil {
		t.Fatal(err)
	}
	f.SetFailPing(true)
	time.Sleep(80 * time.Millisecond)
	f.SetFailPing(false)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		for _, v := range p.Views() {
			if v.ConnectionID == id && v.Generation >= 2 {
				return
			}
		}
		if p.Metrics().ReconnectOK.Load() > 0 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("no reconnect gen views=%v reconn=%d", p.Views(), p.Metrics().ReconnectOK.Load())
}
