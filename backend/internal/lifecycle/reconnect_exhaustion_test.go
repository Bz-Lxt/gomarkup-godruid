package lifecycle_test

import (
	"context"
	"testing"
	"time"

	"godruid/internal/adapter/fake"
	"godruid/internal/lifecycle"
	"godruid/internal/pool"
)

func TestReconnectExhaustionReleasesCapacity(t *testing.T) {
	cfg := pool.DefaultConfig()
	cfg.MaxActive = 1
	cfg.MaxIdle = 1
	cfg.MaxWaitTimeout = 50 * time.Millisecond
	cfg.HealthInterval = time.Hour
	cfg.IdleTTL = time.Hour
	cfg.DialTimeout = 20 * time.Millisecond
	cfg.ReconnectAttempts = 2
	cfg.ReconnectBaseDelay = time.Millisecond
	cfg.ReconnectMaxDelay = time.Millisecond

	connector := fake.New(fake.Options{})
	p, err := pool.New(connector, cfg)
	if err != nil {
		t.Fatal(err)
	}
	s := lifecycle.NewSupervisor(p, nil)
	t.Cleanup(func() {
		s.Stop()
		_ = p.Close()
	})

	conn, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	failedID := conn.ID()
	connector.SetFailDial(true)
	conn.MarkInvalid()
	if err := p.Put(conn); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) && connector.DialCount() < 3 {
		time.Sleep(time.Millisecond)
	}
	if got := connector.DialCount(); got < 3 {
		t.Fatalf("reconnect attempts did not run: dials=%d counts=%+v", got, p.Counts())
	}
	time.Sleep(10 * time.Millisecond)
	connector.SetFailDial(false)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	recovered, err := p.Get(ctx)
	if err != nil {
		t.Fatalf("capacity was not released after reconnect errors: err=%v counts=%+v reconnect_fail=%d dials=%d", err, p.Counts(), p.Metrics().ReconnectFail.Load(), connector.DialCount())
	}
	if recovered.ID() == failedID {
		t.Fatalf("failed logical connection was borrowed again: id=%s generation=%d", recovered.ID(), recovered.Generation())
	}
	if got := p.Metrics().ReconnectFail.Load(); got != 1 {
		t.Fatalf("reconnect failure count=%d, want 1", got)
	}
}
