package lifecycle

import (
	"context"
	"log/slog"
	"sync"

	"godruid/internal/logx"
	"godruid/internal/pool"
)

// Supervisor is the single long-lived goroutine per Pool.
type Supervisor struct {
	pool   *pool.Pool
	log    *slog.Logger
	tasks  *Group
	ctx    context.Context
	stop   context.CancelFunc
	loopWg sync.WaitGroup
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
	s.loopWg.Add(1)
	go s.loop()
}

func (s *Supervisor) loop() {
	defer s.loopWg.Done()
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

// Stop cancels the supervisor context (which exits the loop goroutine and
// interrupts in-flight reconnect/probe dials) and waits for all background
// work to drain. The loop is joined before waiting on tasks so that no new
// probe/reconnect work is admitted concurrently with Wait.
func (s *Supervisor) Stop() {
	s.stop()
	s.loopWg.Wait()
	s.tasks.Close()
	s.tasks.Wait()
}
