package lifecycle

import (
	"context"
)

func (s *Supervisor) reconnect(id string) {
	s.tasks.Go(s.ctx, func(ctx context.Context) {
		s.doReconnect(ctx, id)
	})
}

func (s *Supervisor) doReconnect(ctx context.Context, id string) {
	n := s.pool.BeginReconnect(id)
	if n == nil {
		return
	}
	cfg := s.pool.Config()
	var last error
	for attempt := 1; attempt <= cfg.ReconnectAttempts; attempt++ {
		if ctx.Err() != nil || s.pool.Closed() {
			s.pool.FinishClose(n)
			return
		}
		dctx, cancel := context.WithTimeout(ctx, cfg.DialTimeout)
		raw, err := s.pool.Connector().Connect(dctx)
		cancel()
		if err == nil {
			s.pool.CompleteReconnect(n, raw, nil)
			return
		}
		last = err
		s.log.Debug("reconnect attempt failed", "connection_id", id, "attempt", attempt, "error", last)
		delay := Backoff(attempt, cfg.ReconnectBaseDelay, cfg.ReconnectMaxDelay)
		timer := s.pool.Clock().After(delay)
		select {
		case <-ctx.Done():
			s.pool.CompleteReconnect(n, nil, last)
			return
		case <-s.pool.Done():
			s.pool.CompleteReconnect(n, nil, last)
			return
		case <-timer:
		}
	}
	if last != nil {
		s.pool.CompleteReconnect(n, nil, last)
	}
}
