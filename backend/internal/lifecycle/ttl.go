package lifecycle

func (s *Supervisor) shrinkTTL() {
	nodes := s.pool.TakeIdleForTTL(s.pool.Clock().Now())
	for _, n := range nodes {
		s.pool.FinishClose(n)
	}
}
