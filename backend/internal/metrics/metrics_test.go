package metrics

import (
	"testing"
	"time"
)

func TestWaitWindowPercentiles(t *testing.T) {
	w := NewWaitWindow(time.Minute)
	now := time.Now()
	for i := 1; i <= 100; i++ {
		w.Add(now, time.Duration(i)*time.Millisecond)
	}
	avg, p50, p95, p99, n := w.Stats(now)
	if n != 100 {
		t.Fatalf("n=%d", n)
	}
	if avg <= 0 || p50 <= 0 || p95 < p50 || p99 < p95 {
		t.Fatalf("stats %v %v %v %v", avg, p50, p95, p99)
	}
}

func TestRingWindow(t *testing.T) {
	r := NewRing(15*time.Minute, time.Second)
	now := time.Now()
	for i := 0; i < 10; i++ {
		r.Append(Snapshot{
			Seq:        r.NextSeq(),
			ServerTime: now.Add(time.Duration(i) * time.Second),
			Rates:      Rates{BorrowRPS: float64(i)},
			Counts:     Counts{Live: i},
		})
	}
	pts := r.Window(now.Add(10*time.Second), time.Minute)
	if len(pts) != 10 {
		t.Fatalf("points %d", len(pts))
	}
	if _, ok := ParseWindow("5m"); !ok {
		t.Fatal("5m")
	}
	if _, ok := ParseWindow("nope"); ok {
		t.Fatal("bad window")
	}
}

func TestRegistryRates(t *testing.T) {
	r := NewRegistry()
	r.BorrowOK.Add(10)
	r.IdleHit.Add(8)
	r.ReturnOK.Add(9)
	b, ret, hit, sampled := r.SnapshotRates(time.Second)
	if !sampled || b != 10 || ret != 9 || hit != 0.8 {
		t.Fatalf("%v %v %v %v", b, ret, hit, sampled)
	}
	b, _, _, sampled = r.SnapshotRates(time.Second)
	if sampled || b != 0 {
		t.Fatalf("second window should be zero sampled=%v b=%v", sampled, b)
	}
}
