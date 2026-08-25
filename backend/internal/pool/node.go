package pool

import (
	"sync/atomic"
	"time"

	"godruid/internal/connx"
)

type Node struct {
	id          string
	generation  uint64
	state       State
	raw         connx.Connection
	createdAt   time.Time
	lastBorrow  time.Time
	lastReturn  time.Time
	lastProbe   time.Time
	idleSince   time.Time
	borrowCount uint64
	lastErr     string
	tokenSeq    uint64
	invalid     bool
	closedOnce  atomic.Bool
	pool        *Pool
	prev, next  *Node
}

func (n *Node) ID() string              { return n.id }
func (n *Node) Generation() uint64      { return n.generation }
func (n *Node) State() State            { return n.state }
func (n *Node) Raw() connx.Connection   { return n.raw }

func (n *Node) closeRaw() error {
	if n == nil || n.raw == nil {
		return nil
	}
	if n.closedOnce.Swap(true) {
		return nil
	}
	return n.raw.Close()
}

func (n *Node) replaceRaw(next connx.Connection) connx.Connection {
	old := n.raw
	n.raw = next
	n.closedOnce.Store(false)
	n.generation++
	n.invalid = false
	n.lastErr = ""
	return old
}

// Conn is the caller-facing borrow handle.
type Conn struct {
	node *Node
	pool *Pool
}

func (c *Conn) ID() string {
	if c == nil || c.node == nil {
		return ""
	}
	return c.node.id
}

func (c *Conn) Generation() uint64 {
	if c == nil || c.node == nil {
		return 0
	}
	return c.node.generation
}

func (c *Conn) Raw() connx.Connection {
	if c == nil || c.node == nil {
		return nil
	}
	return c.node.raw
}

func (c *Conn) MarkInvalid() {
	if c == nil || c.node == nil {
		return
	}
	c.pool.markInvalid(c.node)
}
