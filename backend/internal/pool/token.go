package pool

// idleToken is the only value stored in the idle Channel.
// Isolation (probe/TTL/close) increments node.tokenSeq so leftover tokens
// become stale and are discarded on receive.
type idleToken struct {
	node *Node
	seq  uint64
}

func (t idleToken) valid() bool {
	return t.node != nil && t.node.state == StateIdle && t.seq == t.node.tokenSeq
}
