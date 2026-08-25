package lifecycle

import (
	"context"

	"godruid/internal/pool"
)

func (s *Supervisor) probeIdle() {
	nodes := s.pool.TakeIdleForProbe(s.pool.Clock().Now())
	for _, n := range nodes {
		node := n
		s.tasks.Go(s.ctx, func(ctx context.Context) {
			s.pingNode(ctx, node)
		})
	}
}

func (s *Supervisor) pingNode(ctx context.Context, n *pool.Node) {
	cfg := s.pool.Config()
	pctx, cancel := context.WithTimeout(ctx, cfg.HealthTimeout)
	defer cancel()
	var err error
	if n.Raw() != nil {
		err = n.Raw().Ping(pctx)
	}
	s.pool.CompleteProbe(n, err)
}
