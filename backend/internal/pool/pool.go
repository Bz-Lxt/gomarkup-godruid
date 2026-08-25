package pool

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"godruid/internal/clock"
	"godruid/internal/connx"
	"godruid/internal/logx"
	"godruid/internal/metrics"
)

// Pool is a handwritten connection pool. The doubly-linked list is the
// lifecycle source of truth; idle Channel only carries borrowable tokens.
type Pool struct {
	id   string
	name string
	cfg  Config

	connector   connx.Connector
	clk         clock.Clock
	log         *slog.Logger
	met         *metrics.Registry
	waits       *metrics.WaitWindow
	reconnector Reconnector
	onClosed    func(*Node)

	mu      sync.Mutex
	list    connList
	idle    chan idleToken
	waiters int
	dialing int
	closed  bool
	done    chan struct{}
	closeMu sync.Once

	idSeq atomic.Uint64
	tomb  atomic.Int64
}

func New(connector connx.Connector, cfg Config, opts ...Option) (*Pool, error) {
	if connector == nil {
		return nil, fieldErr("connector", "must not be nil")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	p := &Pool{
		id:        "default",
		name:      connector.Kind(),
		cfg:       cfg,
		connector: connector,
		clk:       clock.Real{},
		log:       logx.L(),
		met:       metrics.NewRegistry(),
		waits:     metrics.NewWaitWindow(time.Minute),
		idle:      make(chan idleToken, cfg.MaxActive*2),
		done:      make(chan struct{}),
	}
	for _, opt := range opts {
		opt(p)
	}
	if p.id == "" {
		p.id = "default"
	}
	return p, nil
}

func (p *Pool) ID() string                 { return p.id }
func (p *Pool) Name() string               { return p.name }
func (p *Pool) Config() Config             { return p.cfg }
func (p *Pool) ConnectorKind() string      { return p.connector.Kind() }
func (p *Pool) Connector() connx.Connector { return p.connector }
func (p *Pool) Clock() clock.Clock         { return p.clk }
func (p *Pool) Metrics() *metrics.Registry { return p.met }
func (p *Pool) Waits() *metrics.WaitWindow { return p.waits }
func (p *Pool) Done() <-chan struct{}      { return p.done }

func (p *Pool) Closed() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.closed
}

func (p *Pool) Get(ctx context.Context) (*Conn, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	start := p.clk.Now()
	if p.met != nil {
		p.met.BorrowReq.Add(1)
	}
	deadline := p.waitDeadline(ctx)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	for {
		if err := p.errIfClosed(); err != nil {
			p.failBorrow(ctx.Err())
			return nil, err
		}
		if n := p.tryPopIdle(); n != nil {
			if p.met != nil {
				p.met.IdleHit.Add(1)
				p.met.BorrowOK.Add(1)
			}
			return &Conn{node: n, pool: p}, nil
		}
		if p.tryReserveDial() {
			n, err := p.dialAndAttach(ctx)
			if err != nil {
				if p.met != nil {
					p.met.CreateFail.Add(1)
				}
				if p.errIfClosed() != nil {
					p.failBorrow(ctx.Err())
					return nil, ErrPoolClosed
				}
				if ctx.Err() != nil {
					p.recordWait(start, false)
					p.failBorrow(ctx.Err())
					return nil, MapWaitError(ctx.Err())
				}
				continue
			}
			if p.met != nil {
				p.met.CreateOK.Add(1)
				p.met.BorrowOK.Add(1)
			}
			return &Conn{node: n, pool: p}, nil
		}

		p.addWaiter()
		n, err := p.waitIdle(ctx)
		p.removeWaiter()
		p.recordWait(start, err == nil && n != nil)
		if err != nil {
			p.failBorrow(err)
			return nil, err
		}
		if n != nil {
			if p.met != nil {
				p.met.BorrowOK.Add(1)
			}
			return &Conn{node: n, pool: p}, nil
		}
	}
}

func (p *Pool) Put(c *Conn) error {
	if c == nil || c.node == nil {
		if p.met != nil {
			p.met.ReturnReject.Add(1)
		}
		return ErrNilConn
	}
	n := c.node
	if n.pool != p || c.pool != p {
		if p.met != nil {
			p.met.ReturnReject.Add(1)
		}
		if n.pool != p {
			return ErrCrossPool
		}
		return ErrInvalidPut
	}

	p.mu.Lock()
	if n.state != StateInUse {
		p.mu.Unlock()
		if p.met != nil {
			p.met.ReturnReject.Add(1)
		}
		return ErrDoublePut
	}
	if n.invalid && !p.closed {
		p.transitionLocked(n, StateReconnecting)
		p.mu.Unlock()
		if p.reconnector != nil {
			p.reconnector(n.id)
		}
		if p.met != nil {
			p.met.ReturnOK.Add(1)
		}
		return nil
	}
	if p.closed || p.countLocked(StateIdle) >= p.cfg.MaxIdle {
		p.transitionLocked(n, StateClosing)
		p.mu.Unlock()
		p.finishClose(n)
		if p.met != nil {
			p.met.ReturnOK.Add(1)
		}
		return nil
	}
	p.parkIdleLocked(n)
	p.mu.Unlock()
	if p.met != nil {
		p.met.ReturnOK.Add(1)
	}
	return nil
}

func (p *Pool) Close() error {
	var nodes []*Node
	p.closeMu.Do(func() {
		p.mu.Lock()
		p.closed = true
		close(p.done)
		nodes = p.list.Snapshot()
		for _, n := range nodes {
			if n.state != StateClosed && n.state != StateClosing {
				p.transitionLocked(n, StateClosing)
			}
		}
		p.mu.Unlock()
		p.drainIdle()
	})
	for _, n := range nodes {
		p.finishClose(n)
	}
	return nil
}

func (p *Pool) markInvalid(n *Node) {
	p.mu.Lock()
	n.invalid = true
	p.mu.Unlock()
}

func (p *Pool) tryPopIdle() *Node {
	select {
	case tok := <-p.idle:
		n, _ := p.claimToken(tok)
		return n
	default:
		return nil
	}
}

func (p *Pool) claimToken(tok idleToken) (*Node, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, ErrPoolClosed
	}
	if !tok.valid() {
		return nil, nil
	}
	p.transitionLocked(tok.node, StateInUse)
	tok.node.lastBorrow = p.clk.Now()
	tok.node.borrowCount++
	return tok.node, nil
}

func (p *Pool) tryReserveDial() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return false
	}
	if p.list.Len()+p.dialing >= p.cfg.MaxActive {
		return false
	}
	p.dialing++
	return true
}

func (p *Pool) releaseDial() {
	p.mu.Lock()
	if p.dialing > 0 {
		p.dialing--
	}
	p.mu.Unlock()
}

func (p *Pool) dialAndAttach(ctx context.Context) (*Node, error) {
	dctx, cancel := context.WithTimeout(ctx, p.cfg.DialTimeout)
	defer cancel()
	raw, err := p.connector.Connect(dctx)
	if err != nil {
		p.releaseDial()
		return nil, err
	}
	p.mu.Lock()
	p.dialing--
	if p.closed {
		p.mu.Unlock()
		_ = raw.Close()
		return nil, ErrPoolClosed
	}
	n := p.newNodeLocked(raw, StateInUse)
	n.lastBorrow = p.clk.Now()
	n.borrowCount = 1
	p.mu.Unlock()
	return n, nil
}

func (p *Pool) newNodeLocked(raw connx.Connection, state State) *Node {
	seq := p.idSeq.Add(1)
	now := p.clk.Now()
	n := &Node{
		id:        fmt.Sprintf("c-%04d", seq),
		generation: 1,
		state:     StateConnecting,
		raw:       raw,
		createdAt: now,
		pool:      p,
	}
	p.list.PushBack(n)
	p.transitionLocked(n, state)
	return n
}

func (p *Pool) parkIdleLocked(n *Node) {
	now := p.clk.Now()
	p.transitionLocked(n, StateIdle)
	n.lastReturn = now
	n.idleSince = now
	n.tokenSeq++
	tok := idleToken{node: n, seq: n.tokenSeq}
	select {
	case p.idle <- tok:
	default:
		p.purgeStaleIdleLocked()
		select {
		case p.idle <- tok:
		default:
			p.transitionLocked(n, StateClosing)
			go p.finishClose(n)
		}
	}
}

func (p *Pool) isolateLocked(n *Node, next State) bool {
	if n == nil || n.state != StateIdle {
		return false
	}
	n.tokenSeq++
	p.transitionLocked(n, next)
	return true
}

func (p *Pool) TakeIdleForProbe(now time.Time) []*Node {
	p.mu.Lock()
	defer p.mu.Unlock()
	var out []*Node
	p.list.ForEach(func(n *Node) bool {
		if n.state == StateIdle && (n.lastProbe.IsZero() || now.Sub(n.lastProbe) >= p.cfg.HealthInterval) {
			if p.isolateLocked(n, StateProbing) {
				out = append(out, n)
			}
		}
		return true
	})
	return out
}

func (p *Pool) TakeIdleForTTL(now time.Time) []*Node {
	p.mu.Lock()
	defer p.mu.Unlock()
	var out []*Node
	p.list.ForEach(func(n *Node) bool {
		if n.state == StateIdle && !n.idleSince.IsZero() && now.Sub(n.idleSince) >= p.cfg.IdleTTL {
			if p.isolateLocked(n, StateClosing) {
				out = append(out, n)
			}
		}
		return true
	})
	return out
}

func (p *Pool) CompleteProbe(n *Node, err error) {
	p.mu.Lock()
	if n.state != StateProbing {
		p.mu.Unlock()
		return
	}
	n.lastProbe = p.clk.Now()
	if err == nil {
		n.lastErr = ""
		p.parkIdleLocked(n)
		p.mu.Unlock()
		if p.met != nil {
			p.met.ProbeOK.Add(1)
		}
		return
	}
	n.lastErr = sanitizeErr(err)
	p.transitionLocked(n, StateReconnecting)
	p.mu.Unlock()
	if p.met != nil {
		p.met.ProbeFail.Add(1)
	}
	if p.reconnector != nil {
		p.reconnector(n.id)
	}
}

func (p *Pool) BeginReconnect(id string) *Node {
	p.mu.Lock()
	defer p.mu.Unlock()
	n := p.findLocked(id)
	if n == nil {
		return nil
	}
	if n.state != StateReconnecting && n.state != StateClosing && n.state != StateClosed {
		if n.state == StateIdle {
			p.isolateLocked(n, StateReconnecting)
		} else if n.state == StateInUse || n.state == StateProbing {
			n.tokenSeq++
			p.transitionLocked(n, StateReconnecting)
		}
	}
	return n
}

func (p *Pool) CompleteReconnect(n *Node, raw connx.Connection, err error) {
	if n == nil {
		if raw != nil {
			_ = raw.Close()
		}
		return
	}
	p.mu.Lock()
	if p.closed || n.state == StateClosing || n.state == StateClosed {
		p.mu.Unlock()
		if raw != nil {
			_ = raw.Close()
		}
		return
	}
	if err != nil || raw == nil {
		n.lastErr = sanitizeErr(err)
		p.transitionLocked(n, StateClosing)
		p.mu.Unlock()
		p.finishClose(n)
		if p.met != nil {
			p.met.ReconnectFail.Add(1)
		}
		return
	}
	n.replaceRaw(raw)
	n.lastProbe = p.clk.Now()
	p.parkIdleLocked(n)
	p.mu.Unlock()
	if p.met != nil {
		p.met.ReconnectOK.Add(1)
	}
}

func (p *Pool) FinishClose(n *Node) {
	p.mu.Lock()
	if n != nil && n.state != StateClosing && n.state != StateClosed {
		if n.state == StateIdle {
			n.tokenSeq++
		}
		p.transitionLocked(n, StateClosing)
	}
	p.mu.Unlock()
	p.finishClose(n)
}

func (p *Pool) Find(id string) *Node {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.findLocked(id)
}

func (p *Pool) findLocked(id string) *Node {
	var found *Node
	p.list.ForEach(func(n *Node) bool {
		if n.id == id {
			found = n
			return false
		}
		return true
	})
	return found
}

func (p *Pool) Counts() metrics.Counts {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.countsLocked()
}

// CountsAndViews returns counts and connection views captured atomically
// under a single lock. This prevents torn frames where the aggregate counts
// and the per-connection states are observed at different points in time.
func (p *Pool) CountsAndViews() (metrics.Counts, []metrics.ConnView) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.countsLocked(), p.viewsLocked()
}

func (p *Pool) countsLocked() metrics.Counts {
	var c metrics.Counts
	p.list.ForEach(func(n *Node) bool {
		switch n.state {
		case StateIdle:
			c.Idle++
		case StateInUse:
			c.InUse++
		case StateProbing:
			c.Probing++
		case StateReconnecting:
			c.Reconnecting++
		case StateConnecting:
			c.Connecting++
		case StateClosing:
			c.Closing++
		}
		if n.state.Live() {
			c.Live++
		}
		return true
	})
	c.Dialing = p.dialing
	c.Waiting = p.waiters
	return c
}

func (p *Pool) Views() []metrics.ConnView {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.viewsLocked()
}

func (p *Pool) viewsLocked() []metrics.ConnView {
	nodes := p.list.Snapshot()
	out := make([]metrics.ConnView, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, metrics.ConnView{
			ConnectionID: n.id,
			Generation:   n.generation,
			State:        string(n.state),
			CreatedAt:    n.createdAt,
			LastBorrowAt: n.lastBorrow,
			LastReturnAt: n.lastReturn,
			LastProbeAt:  n.lastProbe,
			BorrowCount:  n.borrowCount,
			LastError:    n.lastErr,
		})
	}
	return out
}

func (p *Pool) LivePlusDialing() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.list.Len() + p.dialing
}

func (p *Pool) InvariantOK() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.list.Len()+p.dialing > p.cfg.MaxActive {
		return fmt.Errorf("live+dialing %d > MaxActive %d", p.list.Len()+p.dialing, p.cfg.MaxActive)
	}
	if p.countLocked(StateIdle) > p.cfg.MaxIdle {
		return fmt.Errorf("idle %d > MaxIdle %d", p.countLocked(StateIdle), p.cfg.MaxIdle)
	}
	if p.waiters < 0 {
		return fmt.Errorf("negative waiters")
	}
	return nil
}

func (p *Pool) Tombstones() int64 { return p.tomb.Load() }

func (p *Pool) transitionLocked(n *Node, to State) {
	if n.state == to {
		return
	}
	if n.state != StateConnecting && !AllowedTransition(n.state, to) && !(n.state == StateClosing && to == StateClosed) {
		p.log.Warn("illegal state transition", "id", n.id, "from", n.state, "to", to)
	}
	n.state = to
}

func (p *Pool) countLocked(s State) int {
	n := 0
	p.list.ForEach(func(node *Node) bool {
		if node.state == s {
			n++
		}
		return true
	})
	return n
}

func (p *Pool) finishClose(n *Node) {
	if n == nil {
		return
	}
	_ = n.closeRaw()
	p.mu.Lock()
	if n.state != StateClosed {
		n.state = StateClosed
		p.list.Remove(n)
		p.tomb.Add(1)
	}
	p.mu.Unlock()
	if p.met != nil {
		p.met.CloseOK.Add(1)
	}
	if p.onClosed != nil {
		p.onClosed(n)
	}
}

func (p *Pool) drainIdle() {
	for {
		select {
		case <-p.idle:
		default:
			return
		}
	}
}

func (p *Pool) purgeStaleIdleLocked() {
	var keep []idleToken
	for {
		select {
		case tok := <-p.idle:
			if tok.valid() {
				keep = append(keep, tok)
			}
		default:
			for _, tok := range keep {
				select {
				case p.idle <- tok:
				default:
				}
			}
			return
		}
	}
}

func (p *Pool) PurgeStaleIdle() {
	p.mu.Lock()
	p.purgeStaleIdleLocked()
	p.mu.Unlock()
}

func (p *Pool) errIfClosed() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return ErrPoolClosed
	}
	return nil
}

func (p *Pool) failBorrow(err error) {
	if p.met == nil {
		return
	}
	if err == nil {
		return
	}
	if err == ErrWaitTimeout || MapWaitError(err) == ErrWaitTimeout {
		p.met.BorrowTimeout.Add(1)
		return
	}
	if err == ErrCanceled || MapWaitError(err) == ErrCanceled {
		p.met.BorrowCancel.Add(1)
	}
}

func (p *Pool) recordWait(start time.Time, _ bool) {
	if p.waits == nil {
		return
	}
	p.waits.Add(p.clk.Now(), p.clk.Now().Sub(start))
}

func sanitizeErr(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	if len(s) > 160 {
		s = s[:160]
	}
	return s
}
