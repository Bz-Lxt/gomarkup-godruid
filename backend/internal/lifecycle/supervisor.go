package lifecycle

import (
	"context"
	"log/slog"

	"godruid/internal/logx"
	"godruid/internal/pool"
)

// Supervisor is the single long-lived goroutine per Pool.
type Supervisor struct {
	pool  *pool.Pool
	log   *slog.Logger
	tasks *Group
	ctx   context.Context
	stop  context.CancelFunc
}

func NewSupervisor(p *pool.Pool, log *slog.Logger) *Supervisor {
	if log == nil {
		log = logx.L()
	}
	ctx, cancel := context.WithCancel(context.Background())
	s := &Supervisor{
		pool:  p,
		log:   log,
		tasks: NewGroup(16),
		ctx:   ctx,
		stop:  cancel,
	}
	p.SetReconnector(s.reconnect)
	return s
}

func (s *Supervisor) Start() {
	go s.loop()
}

func (s *Supervisor) loop() {
	tick := s.pool.Clock().NewTicker(s.pool.Config().HealthInterval)
	defer tick.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.pool.Done():
			s.stop()
			return
		case <-tick.C():
			s.pool.PurgeStaleIdle()
			s.shrinkTTL()
			s.probeIdle()
		}
	}
}

func (s *Supervisor) Stop() {
	s.stop()
	s.tasks.Wait()
}
