package pool

import (
	"log/slog"

	"godruid/internal/clock"
	"godruid/internal/metrics"
)

type Option func(*Pool)

func WithClock(c clock.Clock) Option {
	return func(p *Pool) {
		if c != nil {
			p.clk = c
		}
	}
}

func WithLogger(l *slog.Logger) Option {
	return func(p *Pool) {
		if l != nil {
			p.log = l
		}
	}
}

func WithMetrics(r *metrics.Registry) Option {
	return func(p *Pool) {
		if r != nil {
			p.met = r
		}
	}
}

func WithID(id string) Option {
	return func(p *Pool) { p.id = id }
}

func WithName(name string) Option {
	return func(p *Pool) { p.name = name }
}

func WithOnClose(fn func(*Node)) Option {
	return func(p *Pool) { p.onClosed = fn }
}

func WithReconnector(fn Reconnector) Option {
	return func(p *Pool) { p.reconnector = fn }
}

// Reconnector starts an asynchronous reconnect for an isolated node.
type Reconnector func(nodeID string)

func (p *Pool) SetReconnector(fn Reconnector) {
	if p != nil {
		p.reconnector = fn
	}
}
