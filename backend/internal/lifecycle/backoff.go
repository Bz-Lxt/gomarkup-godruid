package lifecycle

import (
	"math/rand"
	"time"
)

func Backoff(attempt int, base, max time.Duration) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	d := base
	for i := 1; i < attempt; i++ {
		if d > max/2 {
			d = max
			break
		}
		d *= 2
	}
	if d > max {
		d = max
	}
	if d <= 0 {
		return base
	}
	jitter := time.Duration(rand.Int63n(int64(d/2) + 1))
	out := d/2 + jitter
	if out > max {
		return max
	}
	return out
}
