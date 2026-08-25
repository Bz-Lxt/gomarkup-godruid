package lifecycle

import (
	"context"
	"testing"
)

func TestBackoffBounded(t *testing.T) {
	d := Backoff(8, 100, 400)
	if d <= 0 || d > 400 {
		t.Fatalf("backoff %v", d)
	}
	if Backoff(1, 100, 400) > 400 {
		t.Fatal("first attempt")
	}
}

func TestGroupRuns(t *testing.T) {
	g := NewGroup(2)
	done := make(chan struct{}, 3)
	ctx := t.Context()
	for i := 0; i < 3; i++ {
		if !g.Go(ctx, func(context.Context) { done <- struct{}{} }) {
			t.Fatal("go rejected")
		}
	}
	g.Wait()
	if len(done) != 3 {
		t.Fatalf("done %d", len(done))
	}
}
