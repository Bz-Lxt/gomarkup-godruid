package redis

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"godruid/internal/connx"
)

// Connector opens a single TCP connection and speaks RESP PING.
// It does not use a Redis client pool.
type Connector struct {
	Addr    string
	Timeout time.Duration
}

func New(addr string) *Connector {
	return &Connector{Addr: addr, Timeout: 3 * time.Second}
}

func (c *Connector) Kind() string { return "redis" }

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
	if dl, ok := ctx.Deadline(); ok {
		_ = c.c.SetDeadline(dl)
		defer func() { _ = c.c.SetDeadline(time.Time{}) }()
	}
	if _, err := c.c.Write([]byte("*1\r\n$4\r\nPING\r\n")); err != nil {
		return err
	}
	br := bufio.NewReader(c.c)
	line, err := br.ReadString('\n')
	if err != nil {
		return err
	}
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "+PONG") || strings.HasPrefix(line, "+") {
		return nil
	}
	if strings.HasPrefix(line, "-") {
		return fmt.Errorf("redis error: %s", strings.TrimPrefix(line, "-"))
	}
	return fmt.Errorf("unexpected redis reply")
}

func (c *Conn) Close() error {
	if c.closed.Swap(true) {
		return nil
	}
	return c.c.Close()
}

func (c *Conn) Metadata() connx.Metadata {
	return connx.Metadata{
		Kind:     "redis",
		Remote:   connx.SanitizeRemote(c.c.RemoteAddr().String()),
		Isolated: true,
		Note:     "single RESP connection; no client-side pool",
	}
}
