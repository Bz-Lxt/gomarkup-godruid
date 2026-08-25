package tcp

import (
	"context"
	"net"
	"sync/atomic"
	"time"

	"godruid/internal/connx"
)

type Connector struct {
	Addr    string
	Timeout time.Duration
}

func New(addr string) *Connector {
	return &Connector{Addr: addr, Timeout: 3 * time.Second}
}

func (c *Connector) Kind() string { return "tcp" }

func (c *Connector) Connect(ctx context.Context) (connx.Connection, error) {
	d := c.Timeout
	if d <= 0 {
		d = 3 * time.Second
	}
	dialer := net.Dialer{Timeout: d}
	conn, err := dialer.DialContext(ctx, "tcp", c.Addr)
	if err != nil {
		return nil, err
	}
	return &Conn{c: conn}, nil
}

type Conn struct {
	c      net.Conn
	closed atomic.Bool
}

func (c *Conn) Ping(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if c.c == nil {
		return net.ErrClosed
	}
	one := []byte{}
	_ = c.c.SetReadDeadline(time.Now().Add(time.Millisecond))
	_, err := c.c.Read(one)
	_ = c.c.SetReadDeadline(time.Time{})
	if err == nil {
		return nil
	}
	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		return nil
	}
	return err
}

func (c *Conn) Close() error {
	if c.c == nil || c.closed.Swap(true) {
		return nil
	}
	return c.c.Close()
}

func (c *Conn) Metadata() connx.Metadata {
	remote := ""
	if c.c != nil {
		remote = c.c.RemoteAddr().String()
	}
	return connx.Metadata{Kind: "tcp", Remote: connx.SanitizeRemote(remote), Isolated: true}
}
