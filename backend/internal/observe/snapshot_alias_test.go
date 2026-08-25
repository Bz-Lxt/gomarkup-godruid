package observe_test

import (
	"context"
	"testing"

	"godruid/internal/adapter/fake"
	"godruid/internal/metrics"
	"godruid/internal/observe"
	"godruid/internal/pool"
)

func TestPublishedSnapshotsKeepConnectionState(t *testing.T) {
	cfg := pool.DefaultConfig()
	cfg.MaxIdle = 1
	cfg.MaxActive = 1
	p, err := pool.New(fake.New(fake.Options{}), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = p.Close() })

	hub := observe.NewHub()
	updates, cancel := hub.Subscribe()
	t.Cleanup(cancel)

	conn, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	first := metrics.Snapshot{
		Seq:         1,
		Counts:      p.Counts(),
		Connections: p.Views(),
	}
	hub.Publish(first)

	if err := p.Put(conn); err != nil {
		t.Fatal(err)
	}
	second := metrics.Snapshot{
		Seq:         2,
		Counts:      p.Counts(),
		Connections: p.Views(),
	}
	hub.Publish(second)

	receivedFirst := <-updates
	if receivedFirst.Counts.InUse != 1 || len(receivedFirst.Connections) != 1 {
		t.Fatalf("unexpected first snapshot: %+v", receivedFirst)
	}
	if got := receivedFirst.Connections[0].State; got != string(pool.StateInUse) {
		t.Fatalf("seq=1 reports in_use=1 but connection state = %q, want %q", got, pool.StateInUse)
	}

	receivedSecond := <-updates
	if receivedSecond.Counts.Idle != 1 || len(receivedSecond.Connections) != 1 {
		t.Fatalf("unexpected second snapshot: %+v", receivedSecond)
	}
	if got := receivedSecond.Connections[0].State; got != string(pool.StateIdle) {
		t.Fatalf("second snapshot connection state = %q, want %q", got, pool.StateIdle)
	}
}
