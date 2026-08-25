package metrics_test

import (
	"reflect"
	"testing"
	"time"

	"godruid/internal/metrics"
)

func TestRingWindowResultRemainsStableAcrossQueries(t *testing.T) {
	base := time.Date(2026, time.August, 26, 10, 0, 0, 0, time.UTC)
	r := metrics.NewRing(10*time.Minute, time.Minute)
	for _, sample := range []struct {
		at   time.Time
		live int
	}{
		{at: base, live: 10},
		{at: base.Add(4 * time.Minute), live: 20},
		{at: base.Add(5 * time.Minute), live: 30},
	} {
		r.Append(metrics.Snapshot{
			ServerTime: sample.at,
			Counts:     metrics.Counts{Live: sample.live},
		})
	}

	fiveMinute := r.Window(base.Add(5*time.Minute), 5*time.Minute)
	before := liveValues(fiveMinute)
	oneMinute := r.Window(base.Add(5*time.Minute), time.Minute)

	if got, want := liveValues(oneMinute), []int{20, 30}; !reflect.DeepEqual(got, want) {
		t.Fatalf("one-minute result live values = %v, want %v", got, want)
	}
	if got, want := liveValues(fiveMinute), before; !reflect.DeepEqual(got, want) {
		t.Fatalf("five-minute result changed after one-minute query: before=%v after=%v", want, got)
	}
}

func liveValues(points []metrics.SeriesPoint) []int {
	values := make([]int, len(points))
	for i, point := range points {
		values[i] = point.Live
	}
	return values
}
